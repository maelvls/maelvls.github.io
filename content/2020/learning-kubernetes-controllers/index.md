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
draft: true
---

Kubernetes' extensibility is probably its biggest strength. Controllers and
CRDs are all over the place. But finding the right information to begin
writing a controller isn't easy due to the sheer amount of tribal knowledge
scattered everywhere. Here are some links to help you start. And let us
begin with some notes on terminology:

- when I say "controller" (singular noun), I mean one single loop that
  watches some objects. I often call this loop "controller loop" or "sync
  loop" or even "reconcile loop".
- When I say "controllers" (plural), I mean one binary that runs multiple
  sync loops.
- A CRD (custom resource definition) is a simple YAML manifest that
  describes a custom object. After applying it to a Kubernetes cluster, it
  will start accepting manifests of this custom kind.
- "CRD" is just a schema and doesn't carry any logic (except for the basic
  validation the apiserver does). The actual logic happens in controllers.
- Many people also use the term "operator" to mean "controllers". I do not
  make a distinction between an operator (e.g., elatic operator) or
  controllers.

And here are the links that I would give to anyone interested in writing
their own controller:

- [Kubebuilder book](https://book.kubebuilder.io/quick-start.html) is a
  nice starting point. Kubebuilder uses code generation a lot and that's
  what most controllers use nowadays (at Rancher, they use a somehow forked
  version of controller-runtime and controller-tools that generates code).
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
- The Kubernetes codebase itself is also a very nice read. For
  example, [syncReplicaSet](https://github.com/kubernetes/kubernetes/blob/5bac42bf/pkg/controller/replicaset/replica_set.go#L653-L721)
  shows how the Kubernetes team structures their sync functions.

And a final note: CRDs are not necessary for writing a controller! You can
write a tiny controller that watches the "standard" Kubernetes objects.
That's exactly what ingress controllers do: they watch for Service objects.

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
