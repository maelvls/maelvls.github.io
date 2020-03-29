---
title: "GO111MODULE is everywhere: history and tips"
tags: [go]
date: 2019-11-13
description: |
    GO111MODULE is all over the place. It appears in README install
    instructions, in Dockerfiles, in makefiles. On top of that, the behavior of
    GO111MODULE has changed from Go 1.11 to 1.12, changed again with 1.13 and
    will changed a last time in 1.14.
url: /go111module-everywhere
images: [go111module-everywhere/cover-go-modules-sad.jpg]
---

You might have noticed that `GO111MODULE=on` is flourishing everywhere.
Many readmes have that:

```sh
GO111MODULE=on go get -u golang.org/x/tools/gopls@latest
```

In this short post, I will explain why `GO111MODULE` exists, its caveats
and interesting bits that you need to know when dealing with Go Modules.

---

Table of contents:

1. [From `GOPATH` to `GO111MODULE`](#from-gopath-to-go111module)
2. [The `GO111MODULE` environment variable](#the-go111module-environment-variable)
   1. [`GO111MODULE` with Go 1.11 and 1.12](#go111module-with-go-111-and-112)
   2. [`GO111MODULE` with Go 1.13](#go111module-with-go-113)
   3. [`GO111MODULE` with Go 1.14](#go111module-with-go-114)
   4. [So, why is `GO111MODULE` everywhere?!](#so-why-is-go111module-everywhere)
   5. [The pitfall of `go.mod` being silently updated](#the-pitfall-of-gomod-being-silently-updated)
   6. [The `-u` and `@version` pitfall](#the--u-and-version-pitfall)
3. [Caveats when using Go Modules](#caveats-when-using-go-modules)
   1. [Remember that `go get` also updates your `go.mod`](#remember-that-go-get-also-updates-your-gomod)
   2. [Where are the sources of the dependencies with Go Modules](#where-are-the-sources-of-the-dependencies-with-go-modules)
   3. [Set `GO111MODULE` on a per-folder basis with `direnv`](#set-go111module-on-a-per-folder-basis-with-direnv)
   4. [Private Go Modules and Dockerfile](#private-go-modules-and-dockerfile)
      1. [Solution 1: vendoring](#solution-1-vendoring)
      2. [Solution 2: no vendoring](#solution-2-no-vendoring)

---

## From `GOPATH` to `GO111MODULE`

First off, let's talk about GOPATH. When Go was first introduced in 2009,
it was not shipped with a package manager. Instead, `go get` would fetch
all the sources by using their import paths and store them in
`$GOPATH/src`. There was no versioning and the 'master' branch would
represent a stable version of the package.

Go Modules (previously called vgo -- versioned Go) were introduced with Go
1.11. Instead of using the GOPATH for storing a single git checkout of
every package, Go Modules stores tagged versions with `go.mod` keeping
track of each package's version.

Since then, the interaction between the 'GOPATH behavior' and the 'Go
Modules behavior' has become one of the biggest gotchas of Go. One
environment variable is responsible for 95% of this pain: `GO111MODULE`.

## The `GO111MODULE` environment variable

`GO111MODULE` is an environment variable that can be set when using `go`
for changing how Go imports packages. One of the first pain-points is that
depending on the Go version, its semantics change.

### `GO111MODULE` with Go 1.11 and 1.12

- `GO111MODULE=on` will force using Go modules even if the project is in
  your GOPATH. Requires `go.mod` to work.
- `GO111MODULE=off` forces Go to behave the GOPATH way, even outside of
  GOPATH.
- `GO111MODULE=auto` is the default mode. In this mode, Go will behave

  - similarly to `GO111MODULE=on` when you are outside of `GOPATH`,
  - similarly to `GO111MODULE=off` when you are inside the `GOPATH` even if
    a `go.mod` is present.

Whenever you are in your GOPATH and you want to do an operation that
requires Go modules (e.g., `go get` a specific version of a binary), you
need to do:

```sh
GO111MODULE=on go get github.com/golang/mock/tree/master/mockgen@v1.3.1
```

### `GO111MODULE` with Go 1.13

Using Go 1.13, `GO111MODULE`'s default (`auto`) changes:

- behaves like `GO111MODULE=on` anywhere there is a `go.mod` OR anywhere
  outside the GOPATH even if there is no `go.mod`. So you can keep all your
  repositories in your GOPATH with Go 1.13.
- behaves like `GO111MODULE=off` in the GOPATH with no `go.mod`.

### `GO111MODULE` with Go 1.14

The `GO111MODULE` variable has the same behavior as with Go 1.13.

Note that some slight changes in behaviors unrelated to `GO111MODULE`
happened:

- The `vendor/` is picked up automatically. That has the tendency of
  breaking Gomock ([issue](https://github.com/golang/mock/issues/415))
  which were unknowingly not using `vendor/` before 1.14.
- You still need to use `cd && GO111MODULE=on go get` when you don't want
  to mess up your current project’s `go.mod` (that's so annoying).

### So, why is `GO111MODULE` everywhere?!

Now that we know that `GO111MODULE` can be very useful for enabling the Go
Modules behavior, here is the answer: that's because `GO111MODULE=on`
allows you to select a version. Without Go Modules, `go get` fetches the
latest commit from master. With Go Modules, you can select a specific
version based on git tags.

I use `GO111MODULE=on` very often when I want to switch between the latest
version and the HEAD version of `gopls` (the Go Language Server):

```sh
GO111MODULE=on go get -u golang.org/x/tools/gopls@latest
GO111MODULE=on go get -u golang.org/x/tools/gopls@master
GO111MODULE=on go get -u golang.org/x/tools/gopls@v0.1
GO111MODULE=on go get golang.org/x/tools/gopls@v0.1.8
GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0
```

### The pitfall of `go.mod` being silently updated

And to make things even worse, some projects have an even more complicated
one-liners:

```sh
(cd && GO111MODULE=on go get golang.org/x/tools/gopls@latest)
```

> Note: the `@latest` suffix will use the latest git tag of gopls. Note
> that `-u` (which means 'update') is not needed for `@v0.1.8` since this
> is a 'fixed' version, and updating a fixed version does not really make
> sense. It is also interesting to note that with `@v0.1`, `go get` will
> fetch the latest patch version for that tag.

That's yet another Go ideocracy: by default (and you can't turn that off),
if you are in a folder that has a `go.mod`, `go get` will update that
`go.mod` with what you just installed. And in the case of development
binaries like [gopls](https://github.com/golang/tools/tree/master/gopls) or
[kind](https://github.com/kubernetes-sigs/kind), you definitely don't want
to have these appearing in the `go.mod` file!

So the workaround is to give a one-liner that makes sure that you won't be
in a `go.mod`-enabled folder: `(cd && go get)` does exactly that.

I hope that (sooner or later) we will have a clear separation of concerns
between `go get` that is adding a dependency to your `go.mod` (like npm
install) and `go install` that is meant to install a binary without messing
up your `go.mod`. But then, two issues:

- We all use `go get` to install dev dependencies, so moving to `go
  install` would kind of not work (habits...)
- `go install` doesn’t allow you to give a version (e.g., `@latest` or
  `@v1.4.5`), and `go run` either by the way. 😞

### The `-u` and `@version` pitfall

I have been bitten multiple times by this: when using `go get @latest` (for
a binary, at least), you should avoid using `-u` so that it uses the
dependencies as defined in `go.sum`. Otherwise, it will update all the
dependencies to their latest minor revision... And since a ton of projects
choose to have breaking changes between minor versions (e.g. v0.2.0 to
v0.3.0), using `-u` has a large chance of breaking things.

Rebecca Stambler [reminds
us](https://github.com/golang/go/issues/35868#issuecomment-564151454) that
we should not use `-u` in conjunction with a version:

> `-u` should not be used in conjunction with the `@latest` tag, as it will
> give you incorrect versions of the dependencies.

But it's kind of hidden in this issue... I guess it is written somewhere in
the Go help (btw, what a hiddeous help compared to `git help`) but that
kind of caveat should be more visible: maybe print a warning when
installing a binary with both `@version` and `-u`?

## Caveats when using Go Modules

Now, let's go through some caveats I encountered when working with Go
Modules.

### Remember that `go get` also updates your `go.mod`

That’s one of the weird things with `go get`: sometimes, it serves the
purpose of installing binaries or downloading packages. But with Go
modules, if you are in a repo with a `go.mod`, it will silently add the
package you go get to your go.mod.

That’s one of the catches of Go modules! 😁

### Where are the sources of the dependencies with Go Modules

When using Go Modules, the packages that are used during `go build` are
stored in `$GOPATH/pkg/mod`. When trying to inspect an 'import' in vim or
VSCode, you might end up in the GOPATH version of the package instead of
the pkg/mod one used during compilation.

A second issue that arises is when you want to hack one of your dependencies, for example for testing purposes.

**Solution 1**: use `go mod vendor` + `go build -mod=vendor`. That will
force `go` to use the vendor/ files instead of using the `$GOPATH/pkg/mod`
ones. This option also solves the problem of vim and VSCode not opening the
right version of a package’s file.

**Solution 2**: add a 'replace' line at the end of your `go.mod`:

```plain
use replace github.com/maelvls/beers => ../beers
```

where `../beers` is a local copy I made of the dependency I want to inspect
and hack.

### Set `GO111MODULE` on a per-folder basis with `direnv`

During the migration from GOPATH-based projects (mainly using Dep) to Go
Modules, I found myself struggling with two different places: inside and
outside GOPATH. All Go Modules had to be kept outside of GOPATH, which
meant my projects were in different folders.

To remediate that, I used `GO111MODULE` extensively. I would keep all my
projects in the GOPATH, and for the Go Modules-enabled projects, I would
set `export GO111MODULE=on`.

This is where [`direnv`](https://direnv.net/) comes in handy. Direnv is a
lightweight command written in Go that will load a file, `.envrc`, whenever
you enter a directory and `.envrc` is present. For every Go Module-enabled
project, I would have this `.envrc`:

```sh
# .envrc
export GO111MODULE=on
export GOPRIVATE=github.com/mycompany/\*
export GOFLAGS=-mod=vendor
```

The GOPRIVATE disables the Go Proxy (Go 1.13) for certain import paths. I
also found useful to set `-mod=vendor` so that every command uses the
`vendor` folder (`go mod vendor`).

### Private Go Modules and Dockerfile

At my company, we use a lot of private repositories. As explained above, we
can use `GOPRIVATE` in order to tell Go 1.13 to skip the package proxy and
fetch our private packages directly from Github.

But what about building Docker images? How can `go get` fetch our private
repositories from a docker build?

#### Solution 1: vendoring

With `go mod vendor`, no need to pass Github credentials to the docker
build context. We can just put everything in `vendor/` and the problem is
solved. In the Dockerfile, `-mod=vendor` will be required, but developers
don't even have to bother with `-mod=vendor` since they have access to the
private Github repositories anyway using their local Git config

- Pros: faster build on CI (~10 to 30 seconds less)
- Cons: PRs are bloated with `vendor/` changes and the repo's size might be
  big

#### Solution 2: no vendoring

If `vendor/` is just too big (e.g., for Kubernetes controllers, `vendor/`
is about 30MB), we can very well do it without vendoring. That would
require to pass some form of GITHUB_TOKEN as argument of `docker build`,
and in the Dockerfile, set something like:

```sh
git config --global url."https://foo:${GITHUB_TOKEN}@github.com/company".insteadOf "https://github.com/company"
export GOPRIVATE=github.com/company/\*
```

_Illustration by Bailey Beougher, from The Illustrated Children's Guide to Kubernetes._

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
