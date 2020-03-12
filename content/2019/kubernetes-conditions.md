---
title: Use of conditions in Kubernetes controllers
tags: [kubernetes, go, controller]
date: 2019-11-12
description: |
    Although progress is being made, Kubernetes controllers and operators
    still require prior knowledge about Kubernetes internals. Information on
    how to set the status is scattered across comments, issues, PRs and the
    Kubernetes code itself. Conditions may be a good solution for your
    controller, but for what?
url: /kubernetes-conditions
---

While building a Kubernetes controller using CRDs, I stumbled across
'conditions' in the status field. What are conditions and how should I
implement them in my controller?

In this post, I will explain what 'status conditions' are in Kubernetes and
show how they can be used in your own controllers.

---

Table of contents:

1. [Pod example](#pod-example)
2. [What other projects do](#what-other-projects-do)
3. [Conditions vs. State machine](#conditions-vs-state-machine)
4. [Conditions vs. Events](#conditions-vs-events)
5. [Orthogonality vs. Extensibility](#orthogonality-vs-extensibility)
6. [Are Conditions still used?](#are-conditions-still-used)
7. [Conditions vs. Reasons](#conditions-vs-reasons)
8. [How many conditions?](#how-many-conditions)

---

In the following, a 'component' is considered to be one sync loop. A sync
loop (also called reconciliation loop) is what must be done in order to
synchronize the 'desired state' with the 'observed state'.

Kubernetes itself is made of multiple binaries (kubelet on each node, one
apiserver, one kube-controller-manager and one kube-scheduler). And each of
these binaries have multiple components (i.e., sync loops):

| binary                  | sync loop = component              | reads | creates    | updates    |
| ----------------------- | ---------------------------------- | ----- | ---------- | ---------- |
| kube-controller-manager | [`syncDeployment`][syncdeployment] | Pod   | ReplicaSet | Deployment |
| kube-controller-manager | [`syncReplicaSet`][syncreplicaset] |       | Pod        |            |
| kubelet                 | [`syncPod`][syncpod]               |       |            | Pod        |
| kube-scheduler          | [`scheduleOne`][scheduleone]       |       |            | Pod        |
| kubelet                 | [`syncNodeStatus`][syncnodestatus] |       |            | Node       |

[scheduleone]: https://github.com/kubernetes/kubernetes/blob/5bac42bf/pkg/scheduler/scheduler.go#L589-L762
[syncdeployment]: https://github.com/kubernetes/kubernetes/blob/5bac42bf/pkg/controller/deployment/deployment_controller.go#L560-L649
[syncreplicaset]: https://github.com/kubernetes/kubernetes/blob/5bac42bf/pkg/controller/replicaset/replica_set.go#L653-L721
[syncpod]: https://github.com/kubernetes/kubernetes/blob/5bac42bf/pkg/kubelet/status/status_manager.go#L514-L567
[syncnodestatus]: https://github.com/kubernetes/kubernetes/blob/5bac42bff9bfb9dfe0f2ea40f1c80cac47fc12b2/pkg/kubelet/kubelet_node_status.go#L374-L391

We can see that one single object (Pod) can be read, edited and updated by
different components. When I say 'edited', I mean the sync loop edits the
status (which contains the conditions), not the rest. The status is a way
of communicating between components/sync loops.

As Brian Grant put it,

{{< twitter 1111473660479430656 >}}

## Pod example

Let's read a Pod manifest:

```yaml
# kubectl -n cert-manager get pods -oyaml
kind: Pod
status:
  conditions:
    - lastTransitionTime: "2019-10-22T16:29:24Z"
      status: "True"
      type: PodScheduled
    - lastTransitionTime: "2019-10-22T16:29:24Z"
      status: "True"
      type: Initialized
    - lastTransitionTime: "2019-10-22T16:29:31Z"
      status: "True"
      type: ContainersReady
    - lastTransitionTime: "2019-10-22T16:29:31Z"
      status: "True"
      type: Ready
  containerStatuses:
    - image: quay.io/jetstack/cert-manager-controller:v0.6.1
      ready: true
      state: { running: { startedAt: "2019-10-22T16:29:31Z" } }
  phase: Running
```

A condition contains a Status and a Type. It may also contain a Reason and
a Message. For example:

```yaml
- lastTransitionTime: "2019-10-22T16:29:31Z"
  status: "True"
  type: ContainersReady
```

The [pod-lifecycle][] documentation explains the difference between the
'phase' and 'conditions:

1. The top-level `phase` is an aggregated state that answers some
   user-facing questions such as _is my pod in a terminal state?_ but has
   gaps since the actual state is contained in the conditions.
2. The `conditions` array is a set of types (Ready, PodScheduled...) with a
   status (True, False or Unknown) that make up the 'computed state' of a
   Pod at any time. As we will see later, the state is almost always
   'partial' (_open-ended conditions_).

Now, let's see what I mean by 'components'. The status of a Pod is not
updated by a single Sync loop: it is updated by multiple components: the
kubelet, and the kube-scheduler. Here is a list of the condition types per
component:

| Possible condition types for a Pod | Component that updates this condition type    |
| ---------------------------------- | --------------------------------------------- |
| [PodScheduled][podscheduled]       | [`scheduleOne`][scheduleone] (kube-scheduler) |
| [Unschedulable][unschedulable]     | [`scheduleOne`][scheduleone] (kube-scheduler) |
| [Initialized][initialized]         | [`syncPod`][syncpod] (kubelet)                |
| [ContainersReady][containersready] | [`syncPod`][syncpod] (kubelet)                |
| [Ready][ready]                     | [`syncPod`][syncpod] (kubelet)                |

[pod-lifecycle]: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions
[podscheduled]: https://github.com/kubernetes/kubernetes/blob/beaf3a2f0/pkg/scheduler/scheduler.go#L435
[unschedulable]: https://github.com/kubernetes/kubernetes/blob/beaf3a2f0/pkg/scheduler/scheduler.go#L615
[initialized]: https://github.com/kubernetes/kubernetes/blob/beaf3a2f0/pkg/kubelet/kubelet_pods.go#L1368
[ready]: https://github.com/kubernetes/kubernetes/blob/beaf3a2f0/pkg/kubelet/kubelet_pods.go#L1369
[containersready]: https://github.com/kubernetes/kubernetes/blob/beaf3a2f0/pkg/kubelet/kubelet_pods.go#L1370
[kube-controller-manager]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/

Although conditions are a good way to convey information to the user, they
also serve as a way of communicating between components (e.g., between
kube-scheduler and apiserver) but also to external components (e.g. a
custom controller that wants to trigger something as soon as a pod becomes
'Unschedulable', and maybe order more VMs to the cloud provider and add it
as a node.

As you will see below, the 'conditions' array is considered to be
containing all the 'ground truth'. The 'phase' just an abstraction of these
conditions.

## What other projects do

- [cluster-api](https://github.com/kubernetes-sigs/cluster-api) and its
  providers (e.g., cluster-api-gcp-provider) do not use Conditions at all:

  > ([Feng Min, Nov
  > 2018](https://github.com/kubernetes-sigs/cluster-api/blob/112951ee/docs/proposals/20181121-machine-api.md))
  > Brian Grant (@bgrant0607) and Eric Tune (@erictune) have indicated that
  > the API pattern of having "Conditions" lists in object statuses is soon
  > to be deprecated. These have generally been used as a timeline of state
  > transitions for the object's reconciliation, and difficult to consume
  > for clients that just want a meaningful representation of the object's
  > current state. There are no existing examples of the new pattern to
  > follow instead, just the guidance that we should use top-level fields
  > in the status to represent meaningful information. We can revisit the
  > specifics when new patterns start to emerge in core.

  Instead of using a list of conditions, cluster-api projects go like that:

  ```yaml
  status:
    ready: false
    phase: "Failed"
    failureReason: "GcpMachineCrashed"
    failureMessage: "VM crashed: ..."
  ```

- [cert-manager](https://github.com/jetstack/cert-manager) uses Conditions,
  but the only type is 'Ready'. The rest of the information is given in
  Ready's Reason field. That might be unfortunate for users or other
  components that are willing to use cert-manager's API, similarly to what
  happens with the PodSchedulable condition and its reason,
  PodReasonUnschedulable. For example, the [Issuer's
  Sync](https://github.com/jetstack/cert-manager/blob/b91b7d8d3/pkg/controller/issuers/sync.go#L61-L69):

  ```yaml
  kind: Issuer
  status:
    conditions:
      - type: "Ready"
        status: "False"
        reason: "ErrorConfig"
        message: "Resource validation failed: ..."
  ```

- Openshift uses conditions a lot but does not use the standard 'Ready'
  type. For example, cluster-version-operators (CVO) has a [`kind:
  ClusterOperator`](https://github.com/openshift/api/blob/b1bcdbc3/config/v1/types_cluster_operator.go#L123-L134)
  with condition types 'Available', 'Progressing', 'Degraded' and
  'Upgradeable'.

- Hyperconverged Cluster Operator (HCO) uses conditions but in a slightly
  different way that cert-manager or openshift does. Each condition is set
  by a different component

  > It's important to point out that the Condition types on the HCO don't
  > represent the condition the HCO is in, rather the condition of the
  > component CRs.

- Gardener also uses conditions, and like Openshift, it doesn't rely on the
  standard 'Ready' type. For example, the [`kind:
  ControllerInstallation`](https://github.com/gardener/gardener/blob/161695ada/pkg/apis/core/v1alpha1/types_controllerinstallation.go)
  has conditions types 'Valid' and 'Installed'.

We see a lot of variance across projects. In late 2017, we saw some
uncertainty about Conditions disappearing. In the remaining of this post, I
will try to give answers as to why you should use conditions and how to do
so.

## Conditions vs. State machine

[api-conventions]: https://github.com/kubernetes/community/blob/a2cdce5/contributors/devel/sig-architecture/api-conventions.md
[principles]: https://github.com/kubernetes/community/blob/06207d2/contributors/design-proposals/architecture/principles

> ([api-conventions][], July 2017) Conditions are observations and not, themselves, state machines.

So, what is the difference between a state in a state machine and an
observation? A state machine has a known a fixed state. In comparison,
conditions offers an 'open-world' perspective with the Unknown value. For
example, the status of a Pod is partly constructed by the kube-scheduler,
partly by the kubelet.

But when an object gets created, no conditions are present. What does that
mean?

> ([Tim Hockin, May
> 2015](https://issues.k8s.io/7856#issuecomment-99657942)) The absence of a
> condition should be interpreted the same as Unknown.

In the end, this set of conditions are just a way of communicating changes
of state between components, and this state can always be reconstructed by
observing the system:

> ([principles][], 2015) Object status must be 100% reconstructable by
> observation. Any history kept (E.g., through conditions or other fields in
> 'status') must be just an optimization and not required for correct
> operation.
>
> Do not define comprehensive state machines for objects with behaviors
> associated with state transitions and/or "assumed" states that cannot be
> ascertained by observation.

One last comment that explains why the Kubernetes docs don't have any
state-machine diagram:

> ([Tim Hockin, Apr 2016](https://issues.k8s.io/24130)) This
> [state-machine-alike diagram representing the Pod states] is not strictly
> correct though, because a pod can cease to be "ready" but not die. We've
> intentionally NOT drawn this as a state-machine because conditions are
> really orthogonal concepts. We report a singular status as part of
> kubectl, so maybe we should document THAT, but it's more complicated than
> this diagram captures.

## Conditions vs. Events

Events are meant to save history (for non-machine consumption, i.e., for humans). The set of conditions describes the 'current' state:

> ([Brian Grant, Aug
> 2017](https://issues.k8s.io/7856#issuecomment-323196033) Status,
> represented as conditions or otherwise, should be complementary to events
> pertaining to the resource. Events provide additional information for
> debugging and tracing, akin to log messages, but aren't intended to be
> used for event-driven automation.

It was reminded in 2019:

> ([Kenneth Owens, Oct
> 2019](https://issues.k8s.io/51594#issuecomment-334646068)) Events are
> used to express point in time occurrences, and Conditions are used to
> express the current state of an ongoing process. For instance, Pod
> creation is an event, Deployment progressing is a condition.

## Orthogonality vs. Extensibility

In 2015, the Kubernetes team seemed to be attached to the idea of
'orthogonality' between conditions.

> ([David Oppenheimer, Apr
> 2015](https://issues.k8s.io/6951#issuecomment-94203495)) To avoid
> confusion, Conditions should be orthogonal. The ones you suggested --
> Instantiated, Initialized, and Terminated -- don't seem to be orthogonal
> (in fact I would expect a pod to transition from FFF to TFF to TTF to
> TTT). What you've described sounds more like a state machine than a set
> of orthogonal conditions, and I think Conditions should only be used for
> the latter.

In 2017, things start to get confusing: _each condition represents one
non-orthogonal "state" of a resource_?! Conditions are orthogonal to each
other, but the states that they represent may be non-orthogonal...

> ([Brian Grant, Aug
> 2017](https://issues.k8s.io/7856#issuecomment-323196033)) Conditions were
> created as a way to wean contributors off of the idea of state machines
> with mutually exclusive states (called "phase").
>
> Each condition should represent a non-orthogonal "state" of a resource --
> is the resource in that state or not (or unknown).
>
> Arbitrarily typed/structured status properties specific to a particular
> resource should just be status fields.

I did not find anywhere a definition of 'orthogonal condition'; my guess is
that two conditions are orthogonal when they explain two uncorrelated parts
of the system.

But in 2019, the narrative seems to have changed: Brian Grant talks about
'non-orthogonal & extensible conditions'.

> ([Brian Grant, Mar
> 2019](https://twitter.com/bgrant0607/status/1111473660479430656)) Rather
> than rigid fine-grained state enumerations that couldn't be evolved, we
> initially adopted simple basic states that could report open-ended
> reasons for being in each state (<http://issues.k8s.io/1146>), and later
> non-orthogonal, extensible conditions (<http://issues.k8s.io/7856>).

Orthogonality was hard to implement anyway:

> ([Steven E. Harris, Aug
> 2019](https://issues.k8s.io/7856#issuecomment-325978016)) The burden is
> on the status type author to come up with dimensions that are orthogonal;
> that required careful thought when writing my own condition types.

Here is what I think: you should definitely use conditions, but don't
bother with orthogonality. Just use conditions that represent important
changes for the object, beginning with the 'Ready' condition type. 'Ready'
is the strongest condition of all and indicates something 100% operational.

## Are Conditions still used?

There was an extensive dicussion about removing or keeping these
conditions. They were described by [Brian Grant in Aug
2017](https://issues.k8s.io/7856#issuecomment-323196033) as cumbersome (an
array is harder to deal with than top-level fields) and confusing because
of the open-ended statuses (True, False, Unknown). In 2019, the Kubernetes
stated that conditions are still what controller authors should use:

> ([Daniel Smith, May 2019](https://issues.k8s.io/7856#issuecomment-492812566))
> Conditions are not going to be removed.
>
> The ground truth is set as conditions by the components that are nearby,
> e.g. kubelet sets "DiskPressure = True".
>
> The set of conditions is summarized into phases (or secondary conditions)
> for consumption by general controllers. E.g., if there is disk pressure,
> that can be aggregated into `conditionSummary.schedulable: false`. The
> process doing this needs to know all possible gound truth statements; the
> process of summarizing them isolates the rest of the cluster from needing
> to know this.

## Conditions vs. Reasons

Do I really need to bother with this `.status.conditions` array, can I just
use `.status.phase` with a simple enum? Or just a `reason` as part of an
existing 'Ready' condition?

At first, Kubernetes would rely on many Reasons:

> ([api-conventions][], July 2017) In condition types, and everywhere else
> they appear in the API, `Reason` is intended to be a one-word, CamelCase
> representation of the category of cause of the current status, and
> `Message` is intended to be a human-readable phrase or sentence, which
> may contain specific details of the individual occurrence.
>
> `Reason` is intended to be used in concise output, such as one-line
> kubectl get output, and in summarizing occurrences of causes, whereas
> `Message` is intended to be presented to users in detailed status
> explanations, such as kubectl describe output.

Later, they mode to non-orthogonal (read: may-be-corrolated conditions):

> ([Brian Grant, Mar 2019](https://twitter.com/bgrant0607/status/1111473660479430656)) Rather than rigid fine-grained state enumerations that couldn't be evolved, we initially adopted simple basic states that could report [open-ended reasons for being in each state](http://issues.k8s.io/1146) (2014), and later [non-orthogonal, extensible conditions](http://issues.k8s.io/7856) (2017).

Since the Pod conditions do not have a 'Reason' field (in the previous Pod example), let's take a look at the Node conditions instead:

```yaml
kind: Node
status:
  conditions:
    - type: DiskPressure
      status: "False"
      reason: KubeletHasNoDiskPressure
      message: kubelet has no disk pressure
      lastHeartbeatTime: "2019-11-17T14:18:26Z"
      lastTransitionTime: "2019-10-22T16:27:53Z"
    - type: Ready
      status: "True"
      reason: KubeletReady
      message: kubelet is posting ready status. AppArmor enabled
      lastHeartbeatTime: "2019-11-17T14:18:26Z"
      lastTransitionTime: "2019-10-22T16:27:53Z"
```

Here, the 'Ready' condition type has an additional 'Reason' field that
gives the user more information about the transition. The Reason field is
sometimes used by other components: for example, `PodReasonUnschedulable`
is a reason of the `PodSchedulable` condition type. This reason is set by
the kube-scheduler and is 'consumed' by the kubelet (or the reverse, I
can't really tell).

So, Reason of another condition or Condition? If you feel that this state
might be interesting to the rest of the system, then use a proper
Condition.

Regarding the 'phase' top-level field, I would recommend to offer it to
your users when the conditions alone cannot help them answer questions such
as _has my pod terminated?_.

## How many conditions?

If you think that the 'Errored' state can be useful for other components or
the user (e.g. `kubectl wait --for=condition=errored`), then add it. If you
think it is important for other components or for the user to know that
something is in progress, go ahead and add a 'InProgress' condition.

Regarding the naming of condition types, here is some advice:

> ([api-conventions][], July 2017) Condition types should indicate state in
> the "abnormal-true" polarity. For example, if the condition indicates
> when a policy is invalid, the "is valid" case is probably the norm, so
> the condition should be called "Invalid".

Also, remember that these types are part of your API and you should keep in
mind that they require maintaining backwards- and forwards-compatibility.
Adding a new condition is not free: you must maintain them over time.

> ([api-conventions][], July 2017) The meaning of a Condition can not be
> changed arbitrarily - it becomes part of the API, and has the same
> backwards- and forwards-compatibility concerns of any other part of the
> API.

Conditions are also a clean way of letting third-party components (such as
the cluster-api controller) to add their own 'named' conditions to an
existing object, e.g. to a Pod (see [pod-readiness-gate][]):

```yaml
Kind: Pod
spec:
  readinessGates:
    - conditionType: "www.example.com/feature-1"
status:
  conditions:
    - type: Ready # this is a builtin PodCondition
      status: "False"
    - type: "www.example.com/feature-1" # an extra PodCondition
      status: "False"
```

[pod-readiness-gate]: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-readiness-gate

---

Discussion is happening here:

{{< twitter 1195058503695642627 >}}
