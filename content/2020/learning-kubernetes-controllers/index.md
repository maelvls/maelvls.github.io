---
title: "Learning Kubernetes Controllers"
description: |
  Kubernetes' extensibility is probably its biggest strength. Controllers and
  CRDs are all over the place. But finding the right information to begin
  writing a controller isn't easy due to the sheer amount of tribal knowledge
  scattered everywhere. Here are some links to help you start.
date: 2020-04-22T11:58:26+02:00
url: /learning-kubernetes-controllers
images: [learning-kubernetes-controllers/cover-learning-kubernetes-controllers.png]
tags: [kubernetes, controllers]
---

Kubernetes' extensibility is probably its biggest strength. Controllers and
CRDs are all over the place. But finding the right information to begin
writing a controller isn't easy due to the sheer amount of tribal knowledge
scattered everywhere. This post intends to help you start with controllers.

---

Let us begin with some terminology:

- By "controller" (singular noun), I mean one single loop that watches some
  objects. I often call this loop "controller loop" or "sync loop" or even
  "reconcile loop".
- By "controllers" (plural), I mean one binary that runs multiple sync
  loops.
- A CRD (custom resource definition) is a simple YAML manifest that
  describes a custom object. After applying it to a Kubernetes cluster, it
  will start accepting manifests of this custom kind.
- "CRD" is just a schema and doesn't carry any logic (except for the basic
  validation the apiserver does). The actual logic happens in controllers.
- Many people also use the term "operator" to mean "controllers". I do not
  make a distinction between an operator (e.g., the [elastic
  operator](https://github.com/elastic/cloud-on-k8s)) or controllers.

---

Here are the links that I would give to anyone interested in writing their
own controller:

- [Kubebuilder book](https://book.kubebuilder.io/quick-start.html) is a
  nice starting point. Kubebuilder uses code generation a lot and that's
  what most controllers use nowadays (Rancher uses a somehow forked version
  of controller-runtime and controller-tools,
  [Wrangler](https://github.com/rancher/wrangler), that also generates code
  but with their own "style" – for example, simple flat interfaces instead
  of [client-go](https://github.com/kubernetes/client-go)'s deeply nested
  interfaces that don't feel like Go).
- The [Kubernetes API
  conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
  is an amazing document. It summarizes a lot of the "tribal knowledge"
  around naming and how sync loops are conceived and what they mean by
  "level-based behaviour".
- Github search "[language:yaml language:go kubernetes controllers](https://github.com/search?q=language%3Ayaml+language%3Ago+kubernetes+controllers)", tons of nice examples of controllers
- [cert-manager](https://github.com/jetstack/cert-manager)'s codebase is a nice controller to take a look at
- ClusterAPI [proposals](https://github.com/kubernetes-sigs/cluster-api/blob/master/docs/proposals/20190610-machine-states-preboot-bootstrapping.md) and codebase ([capi](https://github.com/kubernetes-sigs/cluster-api), [capa](https://github.com/kubernetes-sigs/cluster-api-provider-aws)) and  (we took a lot of inspiration from what they do)
- The ClusterAPI [Meeting
  notes](https://docs.google.com/document/d/1fQNlqsDkvEggWFi51GVxOglL2P1Bvo2JhZlMhm2d-Co/edit#) contains
  a ton of useful information on Machine, MachinePool... (crazy how much I
  learned from that).
- The Kubernetes `status` field is tricky. You can take a look at
  "[conditions vs. phases vs.
  reasons](https://maelvls.dev/kubernetes-conditions/)".
- The [Kubernetes codebase](https://github.com/kubernetes/kubernetes)
  itself is also a very nice read. It might feel overwhelming at first; I
  invite you to take a look at a few of the following sync loops contained
  in the `kube-controller-manager`, `kube-scheduler` and `kubelet`. Since
  each sync loop reads or updates different objects, I also detail which
  objects are updated or created by each sync loop:

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
- The podcast episode "[Gotime #105 – Kubernetes and Cloud
  Native](https://changelog.com/gotime/105)" (Oct. 2019) with Joe Beda
  (initiator of Kubernetes) and Kris Nova is very interesting and tells us
  more about the genesis of the project, which things like why is
  Kubernetes written in Go and why client-go feels like Java. For example:
  > **Kris Nova:** I think there’s a fourth role. I think there’s what we
  > called in the book an infrastructure engineer. These are effectively
  > the folks like Joe and myself. These are the folks who are writing
  > software to manage and mutate infrastructure behind the scenes. Folks
  > who are contributing to Kubernetes, folks who are writing the software
  > for the operators, folks who are writing admission controller
  > implementations and so forth… I think it’s this very new engineer role,
  > that we haven’t seen until we’ve started having – effectively, as Joe
  > likes to put it, a platform-platform.
- The [operator-sdk](https://github.com/operator-framework/operator-sdk)
  (RedHat) is a package that aims at helping dealing with the whole
  scafollding when writing a sync loop. It relies on
  [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime).
  I don't use either of them but taking a look at these projects helps
  getting more understanding about the challenges (read: boilerplate) that
  comes when writing controllers. I personally write all the
  controller-related boilerplate myself (creating the queue, setting event
  handlers, running the loop itself...).

And a final note: CRDs are not necessary for writing a controller! You can
write a tiny controller that watches the "standard" Kubernetes objects.
That's exactly what ingress controllers do: they watch for Service objects.

**Update 23 April 2020**: I added a quote from Kris Nova! 😁

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
