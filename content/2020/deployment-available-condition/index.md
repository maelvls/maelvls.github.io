---
title: "Understanding the Available condition of a Kubernetes deployment"
description: "Although the Kubernetes documentation is excellent, the API reference does not document the conditions that can be found in a deployment's status. The Available condition has always eluded me!"
date: 2020-07-07T16:33:55+02:00
url: /deployment-available-condition
images: [deployment-available-condition/cover-deployment-available-condition.png]
tags: [kubernetes, conditions]
author: Maël Valais
---

The various conditions that may appear in a deployment status are not
documented in the [API reference][api-ref]. For example, we can see what are the
fields for [DeploymentConditions][deploy-cond-doc], but it lacks the
description of what conditions can appear in this field.

In this post, I dig into what the `Available` condition type is about and
how it is computed.

Since the [API reference][api-ref] does not contain any information about
conditions, the only way to learn more is to dig into the Kubernetes
codebase; quite deep into the code, [we can read some
comments][deploy-cond-code] about the three possible conditions for a
deployment:

```go
// Available means the deployment is available, ie. at least the minimum available
// replicas required are up and running for at least minReadySeconds.
DeploymentAvailable DeploymentConditionType = "Available"

// Progressing means the deployment is progressing. Progress for a deployment is
// considered when a new replica set is created or adopted, and when new pods scale
// up or old pods scale down. Progress is not estimated for paused deployments or
// when progressDeadlineSeconds is not specified.
DeploymentProgressing DeploymentConditionType = "Progressing"

// ReplicaFailure is added in a deployment when one of its pods fails to be created
// or deleted.
DeploymentReplicaFailure DeploymentConditionType = "ReplicaFailure"
```

The description given to the `Available` condition type is quite
mysterious:

> At least the minimum available replicas required are up.

What does "minimum available replicas" mean? Is this minimum 1? I cannot
see any `minAvailable` field in the deployment spec, so my initial guess
was that it would be 1.

Before going further, let's the description attached to the reason
[`MinimumReplicasAvailable`][reason-min-avail]. Apparently, this reason is
the only reason for the `Available` condition type.

```go
// MinimumReplicasAvailable is added in a deployment when it has its minimum
// replicas required available.
MinimumReplicasAvailable = "MinimumReplicasAvailable"
```

The description doesn't help either. Let's see what the [deployment
sync][deploy-sync] function does:

```go
if availableReplicas + deploy.MaxUnavailable(deployment) >= deployment.Spec.Replicas {
    minAvailability := deploy.NewCondition("Available", "True", "MinimumReplicasAvailable", "Deployment has minimum availability.")
    deploy.SetCondition(&status, *minAvailability)
}
```

> Note: in the real code, the max unavailable is on the right side of the
> inequality. I find it easier to reason about this inequality when the
> single value to the right is the desired replica number.

Ahhh, the actual logic being `Available`! If `maxUnavailable` is 0, then it
becomes obvious: the "minimum availability" means that number of available
replicas is greater or equal to the number of replicas in the spec; the
deployment has minimum availability if and only if the following inequality
holds:

|                  available |   +   |            acceptable unavailable            |   ≥   |     desired     |
| -------------------------: | :---: | :------------------------------------------: | :---: | :-------------: |
| `status.availableReplicas` |       | `spec.strategy.rollingUpdate.maxUnavailable` |       | `spec.replicas` |

Let's take an example:

```yaml
kind: Deployment
spec:
  replicas: 10                # Desired
  strategy:
    rollingUpdate:
      maxUnavailable: 2       # Acceptable unavailable
status:
  availableReplicas: 8        # Available
  conditions:
  - type: "Available"
    status: "True"
```

In this example, the inequality holds which means this deployment has
"minimum availability" (= `Available = True`):

| `status.availableReplicas` |   +   | `spec.strategy.rollingUpdate.maxUnavailable` |   ≥   | `spec.replicas` |
| -------------------------: | :---: | :------------------------------------------: | :---: | :-------------: |
|                          8 |       |                      2                       |       |       10        |

## Default value for `maxUnavailable` is 25%

Now, what happens when `maxUnavailable` is not set? The official
documentation [maxUnavailable][max-unavailable] says:

> `maxUnavailable` is an optional field that specifies the maximum number
> of Pods that can be unavailable during the update process. The value can
> be an absolute number (for example, 5) or a percentage of desired Pods
> (for example, 10%). The absolute number is calculated from percentage by
> rounding down. The value cannot be 0 if `maxSurge` is 0. **The default
> value is 25%**.
>
> For example, when this value is set to 30%, the old ReplicaSet can be
> scaled down to 70% of desired Pods immediately when the rolling update
> starts. Once new Pods are ready, old ReplicaSet can be scaled down
> further, followed by scaling up the new ReplicaSet, ensuring that the
> total number of Pods available at all times during the update is at least
> 70% of the desired Pods.

Let's take an example with a deployment that has no `maxUnavailable` field
set, and imagine that 4 pods are unavailable due to a resource quota that
only allows for 5 pods to start:

```yaml
kind: Deployment
spec:
  replicas: 10                
status:
  availableReplicas: 5        
  unavailableReplicas: 5       
  conditions:
  - type: "Available"
    status: "False"
```

This time, the inequality does not hold:

| `status.availableReplicas` |   +   | `spec.strategy.rollingUpdate.maxUnavailable` |   ≱   | `spec.replicas` |
| -------------------------: | :---: | :------------------------------------------: | :---: | :-------------: |
|                          5 |       |             lower(25% * 10) = 2              |       |       10        |

Let's dig a bit more and see how [`MaxAvailable`][deploy-max-unavail] is
defined:

```go
// MaxUnavailable returns the maximum unavailable pods a rolling deployment
// can take.
func MaxUnavailable(deployment apps.Deployment) int {
    if !IsRollingUpdate(&deployment) || deployment.Spec.Replicas == 0 {
        return 0
    }
    _, maxUnavailable, _ := ResolveFenceposts(deployment.Spec.Strategy.RollingUpdate.MaxSurge, deployment.Spec.Strategy.RollingUpdate.MaxUnavailable, *(deployment.Spec.Replicas))
    if maxUnavailable > *deployment.Spec.Replicas {
        return *deployment.Spec.Replicas
    }
    return maxUnavailable
}
```

The core of the logic behind maxUnavailable is in
[`ResolveFenceposts`][deploy-fenceposts] (note: I simplified the code a bit
to make it more readable):

```go
// ResolveFenceposts resolves both maxSurge and maxUnavailable. This needs to happen in one
// step. For example:
//
// 2 desired, max unavailable 1%, surge 0% - should scale old(-1), then new(+1), then old(-1), then new(+1)
// 1 desired, max unavailable 1%, surge 0% - should scale old(-1), then new(+1)
// 2 desired, max unavailable 25%, surge 1% - should scale new(+1), then old(-1), then new(+1), then old(-1)
// 1 desired, max unavailable 25%, surge 1% - should scale new(+1), then old(-1)
// 2 desired, max unavailable 0%, surge 1% - should scale new(+1), then old(-1), then new(+1), then old(-1)
// 1 desired, max unavailable 0%, surge 1% - should scale new(+1), then old(-1)
func ResolveFenceposts(maxSurge, maxUnavailable *instr.IntOrString, desired int) (int, int, error) {
    surge, _       := instr.GetValueFromIntOrPercent(instr.ValueOrDefault(maxSurge, instr.FromInt(0)), desired, true)
    unavailable, _ := instr.GetValueFromIntOrPercent(instr.ValueOrDefault(maxUnavailable, instr.FromInt(0)), desired, false)
    return surge, unavailable, nil
}
```

The `false` boolean turns the integer rounding "up", which means `0.5` will
be rounded to `0` instead of `1`.

The `maxUnavailable` and `maxSurge` (they call them "fenceposts" values)
are simply read from the deployment's spec. In the following example, the
deployment will become `Available = True` only if there are at least 5
replicas available:

```yaml
kind: Deployment
spec:
  replicas: 10
  strategy:
    rollingUpdate:
      maxUnavailable: 0%    # 🔰 10 * 0.0 = 0 replicas
```

[GetValueFromIntOrPercent]: https://godoc.org/k8s.io/apimachinery/pkg/util/intstr#GetValueFromIntOrPercent
[api-ref]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/
[deploy-cond-doc]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#deploymentcondition-v1-apps
[deploy-cond-code]: https://github.com/kubernetes/kubernetes/blob/3615291/pkg/apis/apps/types.go#L461-L473
[reason-min-avail]: https://github.com/kubernetes/kubernetes/blob/3615291/pkg/controller/deployment/util/deployment_util.go#L96-L97
[deploy-sync]: https://github.com/kubernetes/kubernetes/blob/3615291/pkg/controller/deployment/sync.go#L513-L516
[deploy-max-unavail]: https://github.com/kubernetes/kubernetes/blob/3615291/pkg/controller/deployment/util/deployment_util.go#L434-L445
[deploy-fenceposts]: https://github.com/kubernetes/kubernetes/blob/3615291/pkg/controller/deployment/util/deployment_util.go#L874-L902
[max-unavailable]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#max-unavailable

## Hands-on example with the `Available` condition

Imagine that we have a namespace named `restricted` that only allows for
200 MiB, and our pod requires 50 MiB. The first 4 pods will be successfully
created, but the fifth one will fail.

Let us first apply the following manifest:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: restricted
---
apiVersion: v1
kind: ResourceQuota
metadata:
  namespace: restricted
  name: mem-cpu-demo
spec:
  hard:
    requests.memory: 200Mi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: restricted
  name: test
spec:
  replicas: 5
  strategy:
    rollingUpdate:
      maxUnavailable: 0          # 🔰
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
        - name: test
          image: nginx:alpine
          resources:
            requests:
              memory: "50Mi"     # the 5th pod will fail (on purpose)
          ports:
            - containerPort: 80
```

After a few seconds, the `Available` condition stabilizes to `False`:

```yaml
# kubectl -n restricted get deploy test -oyaml
kind: Deployment
spec:
  replicas: 5
  strategy:
    rollingUpdate:
      maxUnavailable: 0          # 🔰
status:
  conditions:
  - lastTransitionTime: "2020-07-07T14:04:27Z"
    lastUpdateTime: "2020-07-07T14:04:27Z"
    message: Deployment does not have minimum availability.
    reason: MinimumReplicasUnavailable
    status: "False"
    type: Available
  - lastTransitionTime: "2020-07-07T14:04:27Z"
    lastUpdateTime: "2020-07-07T14:04:27Z"
    message: 'pods "test-7df57bd99d-5qw47" is forbidden: exceeded quota: mem-cpu-demo,
      requested: requests.memory=50Mi, used: requests.memory=200Mi, limited: requests.memory=200Mi'
    reason: FailedCreate
    status: "True"
    type: ReplicaFailure
  - lastTransitionTime: "2020-07-07T14:04:27Z"
    lastUpdateTime: "2020-07-07T14:04:38Z"
    message: ReplicaSet "test-7df57bd99d" is progressing.
    reason: ReplicaSetUpdated
    status: "True"
    type: Progressing
  replicas: 4
  availableReplicas: 4
  readyReplicas: 4
  unavailableReplicas: 1
  updatedReplicas: 4
```

We are asking for at most 0 unavailable replicas and there is 1 unavailable
replica (due to the resource quota). Thus, the "minimum availability"
inequality does not hold which means the deployment has the condition
`Available = False`:

| `status.availableReplicas` |   +   | `spec.strategy.rollingUpdate.maxUnavailable` |   ≱   | `spec.replicas` |
| -------------------------: | :---: | :------------------------------------------: | :---: | :-------------: |
|                          4 |       |                      0                       |       |        5        |

**Update 9 July 2020:** added a paragraph on the default value for
[maxUnavailable][max-unavailable], and fixed the yaml example where I had
mixed `unavailableReplicas` with `maxUnavailable`.
