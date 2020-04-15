---
title: Resources for learning how to write Kubernetes controllers
description: ""
date: 2020-04-19
url: /resources-for-writing-controllers
images: []
tags: [kubernetes, go]
draft: true
---

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

> When a new version of an object is POSTed or PUT, the "spec" is updated
> and available immediately. Over time the system will work to bring the
> "status" into line with the "spec". The system will drive toward the most
> recent "spec" regardless of previous versions of that stanza. In other
> words, if a value is changed from 2 to 5 in one PUT and then back down to
> 3 in another PUT the system is not required to 'touch base' at 5 before
> changing the "status" to 3. In other words, the system's behavior is
> level-based rather than edge-based. This enables robust behavior in the
> presence of missed intermediate state changes.


<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
