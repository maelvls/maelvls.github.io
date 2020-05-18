---
title: "The Client-go Transitive Hell"
description: ""
date: 2020-04-15
images: [title/cover.png]
url: /client-go-transitive-hell
draft: true
---

> ⚠️ I'm not sure this is a transitive issue. It might just not be due to
> transitivity at all!

Looks like the Kubernetes people chose to break the client-go API without
bumping the major version (with a new import path e.g
`k8s.io/client-go/v13`) when adding support for `context.Context` that
comes with Kubernetes v1.18. Darren Shepherd reported the
[issue](https://issues.k8s.io/88472) in February 2020. We [were warned](https://github.com/kubernetes/client-go#compatibility-your-code---client-go):

> The v0.x.y tags indicate that go APIs may change in incompatible ways in
> different versions.

At first, I thought that it would not affect us at Ori since we don’t use
any extra dependency that rely on client-go — so no transitive dependencies
on client-go, we are the sole user of it. We then began working on some
tooling depending on the API of our main project. And the transitive hell
began.

Here is what the error looks like:

```plain
# k8s.io/client-go/rest
../../../go/pkg/mod/k8s.io/client-go@v11.0.0+incompatible/rest/request.go:598:31: not enough arguments in call to watch.NewStreamWatcher
        have (*versioned.Decoder)
        want (watch.Decoder, watch.Reporter)
```

On top of that, you might also have a version mismatch between `k8s.io/api`
and `k8s.io/apimachinery` with the following error:

```plain
../../../go/pkg/mod/k8s.io/client-go@v11.0.0+incompatible/tools/clientcmd/api/v1/conversion.go:52:12: s.DefaultConvert undefined (type conversion.Scope has no field or method DefaultConvert)
```

The workaround is to set client-go to use the latest pre-v1.18 version:

```diff
 require (
-    k8s.io/client-go v11.0.0+incompatible
-    k8s.io/apimachinery v0.18.1
-    k8s.io/api v0.18.1 //indirect
+    k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
+    k8s.io/apimachinery v0.17.4
+    k8s.io/api v0.17.4 //indirect
 )
```

The `v0.17.4` version is the last version of apimachinery and api that
stays compatible with client-go pre-v1.18.

**Long-term**: you want to move away from client-go `v11.0.0` (i.e. tag
`kubernetes-1.14.0`) to an earlier version (e.g. `v12.0.0`).

What’s funny is that we can’t use `v12.0.0` because the client-go project
doesn't follow the "[semantic import
versioning](https://research.swtch.com/vgo-import)" rule so you can’t just
do

```sh
% go get k8s.io/client-go@v12.0.0
go get k8s.io/client-go@v12.0.0: k8s.io/client-go@v12.0.0: invalid version: module contains a go.mod file, so major version must be compatible: should be v0 or v1, not v12
```

Client-go also maintains another set of tags that begin with `v0.*` (e.g.,
`v0.18.0`) for that specific reason. Just a clever way of escaping the
semantic import versioning, but all these different versions make it very
confusing...

And when you use the `kubernetes-1.17.4` tag, it redirects to the `v0.17.4`
tag (my guess is that it is infered by `go get`):

```sh
% go get k8s.io/client-go@kubernetes-1.17.4
go: k8s.io/client-go kubernetes-1.17.4 => v0.17.4
go: downloading k8s.io/client-go v0.17.4
```

And to make even more confusing, these tags (kubernetes-v1.17.4, v0.17.4
and v12.0.0) are not set on the master branch; instead, they all live on
headless branches.

This issue is a reminder that we (Kubernetes hackers who write controllers
for a living) heavily rely on the “good will” of the Kubernetes team. These
decisions as they might affect anyone relying on Kubernetes "as a
platform"... 🤔



<!--
https://ori-edge.slack.com/archives/C96DU1WDC/p1583851517147600?thread_ts=1582966579.027200&cid=C96DU1WDC
https://ori-edge.slack.com/archives/C96DU1WDC/p1586525068128100
-->