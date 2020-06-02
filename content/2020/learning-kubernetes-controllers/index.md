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
author: Maël Valais
---

Kubernetes' extensibility is probably its biggest strength. Controllers and
CRDs are all over the place. But finding the right information to begin
writing a controller isn't easy due to the sheer amount of tribal knowledge
scattered everywhere. This post intends to help you start with controllers.

<!--

> A "Kubernetes controller" is a binary that runs reconciliation loops. A
> reconciliation loop watches the objects stored in Kubernetes. When it
> notices a discrepency between what the object specifies (e.g. 4 replicas)
> and the observed reality (e.g., the reconcialiation loop asks kubelet,
> and it answers there are only 2 replicas), the reconcialiation loop will
> take actions in order to satisfy what is specified in the object. The
> "controller" binary is run as a simple Kubernetes Deployment. Sometimes,
> when the Kubernetes API is not enough, it may also come with some
> CustomResourceDefinitions YAML files.

I use interchangeably the term "sync loop" and "controller". The word
"controller" is quite overloaded: we use to qualify the binary that runs in
a pod and watches for objects ("Kubernetes controller"), but we also use it
to mean "one sync loop" that is running inside this binary.

Anyone writing Kubernetes controllers might want to take a look at the
following resources.
One controller

**[Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)**

> When a new version of an object is POSTed or PUT, the "spec" is updated
> and available immediately. Over time the system will work to bring the
> "status" into line with the "spec". The system will drive toward the most
> recent "spec" regardless of previous versions of that stanza. In other
> words, if a value is changed from 2 to 5 in one PUT and then back down to
> 3 in another PUT the system is not required to 'touch base' at 5 before
> changing the "status" to 3. In other words, the system's behavior is
> level-based rather than edge-based. This enables robust behavior in the
> presence of missed intermediate state changes.

-->

---

Let us begin with some terminology:

- **controller**: a single loop that watches some objects. We often refer
  to this loop as "controller loop" or "sync loop" or "reconcile loop".
- **controller binary** is a binary that runs one or multiple sync loops.
  We often refer to it as "controllers".
- **CRD** (Custom Resource Definition) is a simple YAML manifest that
  describes a custom object, for example [this
  CRD](https://github.com/jetstack/cert-manager/blob/a04d2f0935/deploy/crds/crd-orders.yaml#L2)
  defines the acme.cert-manager.io/v1alpha3 Order resource. After applying
  this CRD to a Kubernetes cluster, you can apply manifests that have the
  kind "Order"

  > Note: CRDs and controllers are decoupled. You can apply a CRD manifest
  > without having any controller binary running. It works in both ways:
  > you can have a controller binary running that doesn't require any
  > custom objects. Traefik is a controller binary which relies on built-in
  > Service objects.

  > Note: the "CRD" manifest is just a schema. It doesn't carry any logic
  > (except for the basic validation the apiserver does). The actual logic
  > happens in the controller binary.

- **operator**: the term "operator" is often used to mean a controller
  binary with its CRDs, for example the [elastic
  operator](https://github.com/elastic/cloud-on-k8s).

---

Here are the links that I would give to anyone interested in writing their
own controller:

- [sig-api-machinery/controllers.md](https://github.com/kubernetes/community/blob/712590c108bd4533b80e8f2753cadaa617d9bdf2/contributors/devel/sig-api-machinery/controllers.md)
  gives a good intuition as to what a "controller" is:

  > A Kubernetes controller is an active reconciliation process. That is,
  > it watches some object for the world's desired state, and it watches
  > the world's actual state, too. Then, it sends instructions to try and
  > make the world's current state be more like the desired state.

  Note: the client-go's informers and listers and workqueue are not
  mandatory for writing a controller: you can just rely on client-go's
  `Watch` primitive to reconcile state. The informers and workqueue add
  important scalability and reliability features but these also come with
  the cost of heavy abstractions. Use client-go's `Watch` first to have a
  sense of what it can offer, and then try out informers and workqueue.
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

- **Update 23 April 2020**: I added a quote from Kris Nova! 😁
- **Update 2 May 2020**: Rephrased the "terminology" bullet points to make
  them clearer and added a note on CRD vs. controller binary.
