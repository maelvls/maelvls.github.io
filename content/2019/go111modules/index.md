---
title: "Why is GO111MODULE everywhere, and everything about Go Modules (updated with Go 1.20)"
tags: [go, go-modules]
date: 2019-11-13
description: "GO111MODULE is all over the place. It appears in README install instructions, in Dockerfiles, in makefiles. On top of that, the behavior of GO111MODULE has changed from Go 1.11 to 1.12, changed again with Go 1.13 and Go 1.15 and changed a last time in Go 1.16, and is stable since then."
url: /go111module-everywhere
images: [go111module-everywhere/cover-go-modules-sad.jpg]
devtoId: 204994
devtoPublished: true
devtoUrl: https://dev.to/maelvls/why-is-go111module-everywhere-and-everything-about-go-modules-24k
---

You might have noticed that `GO111MODULE=on` is flourishing everywhere. Many readmes have that:

```sh
GO111MODULE=on go get golang.org/x/tools/gopls@latest
```

Note that `go get` has been deprecated for installing binaries since Go 1.17 and will be become impossible in Go 1.18. If you are using Go 1.16 or above, you should use instead:

```
go install golang.org/x/tools/gopls@latest
```

In this short post, I will explain why `GO111MODULE` was introduced in Go 1.11, its caveats and interesting bits that you need to know when dealing with Go Modules.

---

**Table of content:**

1. [From `GOPATH` to `GO111MODULE`](#from-gopath-to-go111module)
2. [The `GO111MODULE` environment variable](#the-go111module-environment-variable)
   1. [`GO111MODULE` with Go 1.11 and 1.12](#go111module-with-go-111-and-112)
   2. [`GO111MODULE` with Go 1.13](#go111module-with-go-113)
   3. [`GO111MODULE` with Go 1.14](#go111module-with-go-114)
   4. [`GO111MODULE` with Go 1.15](#go111module-with-go-115)
   5. [`GO111MODULE` with Go 1.16](#go111module-with-go-116)
   6. [`GO111MODULE` with Go 1.17](#go111module-with-go-117)
      1. [Faster downloading of dependencies if you are using Git to fetch modules](#faster-downloading-of-dependencies-if-you-are-using-git-to-fetch-modules)
      2. [Installing binaries with `GO111MODULE=on go get` is deprecated](#installing-binaries-with-go111moduleon-go-get-is-deprecated)
      3. [`go run` knows about `@version` (finally!)](#go-run-knows-about-version-finally)
   7. [`GO111MODULE` with Go 1.18](#go111module-with-go-118)
   8. [`GO111MODULE` with Go 1.19](#go111module-with-go-119)
   9. [`GO111MODULE` with Go 1.20](#go111module-with-go-120)
   10. [Why was `GO111MODULE` everywhere? (Go 1.15 and below)](#why-was-go111module-everywhere-go-115-and-below)
   11. [The pitfall of `go.mod` being silently updated (Go 1.15 and below)](#the-pitfall-of-gomod-being-silently-updated-go-115-and-below)
   12. [The `-u` and `@version` pitfall](#the--u-and-version-pitfall)
3. [Caveats when using Go Modules](#caveats-when-using-go-modules)
   1. [Remember that `go get` also updates your `go.mod`](#remember-that-go-get-also-updates-your-gomod)
   2. [Where are the sources of the dependencies with Go Modules](#where-are-the-sources-of-the-dependencies-with-go-modules)
   3. [Set `GO111MODULE` on a per-folder basis with `direnv`](#set-go111module-on-a-per-folder-basis-with-direnv)
   4. [Private Go Modules and Dockerfile](#private-go-modules-and-dockerfile)
      1. [Solution 1: vendoring](#solution-1-vendoring)
      2. [Solution 2: no vendoring](#solution-2-no-vendoring)

---

## From `GOPATH` to `GO111MODULE`

First off, let's talk about GOPATH. When Go was first introduced in 2009, it was not shipped with a package manager. Instead, `go get` would fetch all the sources by using their import paths and store them in `$GOPATH/src`. There was no versioning and the 'master' branch would represent a stable version of the package.

Go Modules (previously called vgo -- versioned Go) were introduced with Go 1.11. Instead of using the GOPATH for storing a single git checkout of every package, Go Modules stores tagged versions with `go.mod` keeping track of each package's version.

Since then, the interaction between the 'GOPATH behavior' and the 'Go Modules behavior' has become one of the biggest gotchas of Go. One environment variable is responsible for 95% of this pain: `GO111MODULE`.

## The `GO111MODULE` environment variable

`GO111MODULE` is an environment variable that can be set when using `go` for changing how Go imports packages. One of the first pain-points is that depending on the Go version, its semantics change.

`GO111MODULE` can be set in two different ways:

- The environment variable in your shell, e.g.: `export GO111MODULE=on`.
- The "hidden" global configuration only known by `go env` using `go env -w GO111MODULE=on` (only available since Go 1.12).

If Go seems to behave the way you think it should, it is recommended to check whether, in the past, you have set the global configuration using `go env -w GO111MODULE=on` (or `off`):

```bash
go env GO111MODULE
```

Note that the environment variable takes precedence over the value that was stored using `go env -w GO111MODULE`.

To unset the global configuration, you can do:

```bash
go env -u GO111MODULE
```

### `GO111MODULE` with Go 1.11 and 1.12

- `GO111MODULE=on` will force using Go modules even if the project is in your GOPATH. Requires `go.mod` to work.
- `GO111MODULE=off` forces Go to behave the GOPATH way, even outside of GOPATH.
- `GO111MODULE=auto` is the default mode. In this mode, Go will behave

  - similarly to `GO111MODULE=on` when you are outside of `GOPATH`,
  - similarly to `GO111MODULE=off` when you are inside the `GOPATH` even if a `go.mod` is present.

Whenever you are in your GOPATH and you want to do an operation that requires Go modules (e.g., `go get` a specific version of a binary), you need to do:

```sh
GO111MODULE=on go get github.com/golang/mock/tree/master/mockgen@v1.3.1
```

### `GO111MODULE` with Go 1.13

Using Go 1.13, `GO111MODULE`'s default (`auto`) changes:

- behaves like `GO111MODULE=on` anywhere there is a `go.mod` OR anywhere outside the GOPATH even if there is no `go.mod`. So you can keep all your repositories in your GOPATH with Go 1.13.
- behaves like `GO111MODULE=off` in the GOPATH with no `go.mod`.

### `GO111MODULE` with Go 1.14

The `GO111MODULE` variable has the same behavior as with Go 1.13.

Note that some slight changes in behaviors unrelated to `GO111MODULE` happened:

- The `vendor/` is picked up automatically. That has the tendency of breaking Gomock ([issue](https://github.com/golang/mock/issues/415)) which were unknowingly not using `vendor/` before 1.14.
- You still need to use `cd && GO111MODULE=on go get` when you don't want to mess up your current project’s `go.mod` (that's so annoying).

### `GO111MODULE` with Go 1.15

Nothing with regards to `GO111MODULE` has changed with Go 1.15.

### `GO111MODULE` with Go 1.16

As of Go 1.16, the default behavior is `GO111MODULE=on`, meaning that if you want to keep using the old `GOPATH` way, you will have to force Go not to use the Go Modules feature:

```sh
export GO111MODULE=off
```

The best news in Go 1.16 is that we finally get a dedicated command for installing Go tools instead of relying on the does-it-all `go get` that keeps [updating](#the-pitfall-of-gomod-being-silently-updated) your `go.mod`. Instead of:

```sh
# Old way
(cd && go install golang.org/x/tools/gopls@latest)
```

you can now run:

```sh
go install golang.org/x/tools/gopls@latest
```

One caveat is that the semantics of `go install` is slightly different from `go get`. As detailed on the [Go
blog](https://blog.golang.org/go116-module-changes):

> In order to eliminate ambiguity about which versions are used, there are several restrictions on what directives may be present in the program's `go.mod` file when using this install syntax. In particular, **replace and exclude directives are not allowed**, at least for now. In the long term, once the new `go install program@version` is working well for enough use cases, we plan to make `go get` stop installing command binaries. See [issue 43684](https://github.com/golang/go/issues/43684) for details.

As an example, you cannot (as of April 2021) install the master version of [gopls](https://github.com/golang/tools/tree/master/gopls):

```sh
% go install golang.org/x/tools/gopls@master
go: downloading golang.org/x/tools v0.1.1-0.20210316190639-9e9211a98eaa
go: downloading golang.org/x/tools/gopls v0.0.0-20210316190639-9e9211a98eaa
go install golang.org/x/tools/gopls@master: golang.org/x/tools/gopls@v0.0.0-20210316190639-9e9211a98eaa
        The go.mod file for the module providing named packages contains one or
        more replace directives. It must not contain directives that would cause
        it to be interpreted differently than if it were the main module.
```

The replace directive in the [go.mod](https://github.com/golang/tools/blob/master/gopls/go.mod) file looks like this:

```go
replace golang.org/x/tools => ../
```

Fortunately, the gopls project makes sure to remove the `replace` directive before each release, which means you can use the `latest` tag:

```sh
go install golang.org/x/tools/gopls@latest
#                                   ^^^^^^
```

### `GO111MODULE` with Go 1.17

Go 1.17 was released on August 16, 2021. As for 1.16, `GO111MODULE=on` is the default behavior, and `GO111MODULE=auto` is equivalent to `GO111MODULE=on`. If you still want to use the `GOPATH` way, you will have to force Go not to use the Go Modules feature using `GO111MODULE=off` (see [the section about `direnv`](#set-go111module-on-a-per-folder-basis-with-direnv)).

The three important changes that may affect how you use `GO111MODULE` are the following:

#### Faster downloading of dependencies if you are using Git to fetch modules

If you decide to update the `go` line in your `go.mod` to take advantage of the new [module graph pruning](https://golang.org/ref/mod#graph-pruning), your `go.mod` will want to get updated. The first time you will want to use `go build`, you will see the following error:

```
go build ./...
go: updates to go.mod needed; to update it:
        go mod tidy
```

After running `go mod tidy`, you will see that your `go.mod` changed quite a lot with a new `require` block:

```diff
diff --git a/go.mod b/go.mod
index 3af719e..26ed354 100644
--- a/go.mod
+++ b/go.mod
@@ -1,12 +1,12 @@
   module github.com/maelvls/users-grpc

-go 1.12
+go 1.17

   require (
   github.com/MakeNowJust/heredoc/v2 v2.0.1
   github.com/fsnotify/fsnotify v1.4.9 // indirect
   github.com/golang/mock v1.4.4
-	github.com/golang/protobuf v1.4.3
+	github.com/golang/protobuf v1.5.2
   github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
   github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
   github.com/hashicorp/go-memdb v1.3.0
@@ -37,8 +37,25 @@ require (
   golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
   google.golang.org/genproto v0.0.0-20201116205149-79184cff4dfe // indirect
   google.golang.org/grpc v1.33.2
-	google.golang.org/protobuf v1.25.0
+	google.golang.org/protobuf v1.27.1
   gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
   gopkg.in/ini.v1 v1.62.0 // indirect
   gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
   )
+
+require (
+	github.com/beorn7/perks v1.0.0 // indirect
+	github.com/davecgh/go-spew v1.1.1 // indirect
+	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
+	github.com/hashicorp/golang-lru v0.5.4 // indirect
+	github.com/hashicorp/hcl v1.0.0 // indirect
+	github.com/inconshreveable/mousetrap v1.0.0 // indirect
+	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
+	github.com/pmezard/go-difflib v1.0.0 // indirect
+	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
+	github.com/prometheus/common v0.4.0 // indirect
+	github.com/prometheus/procfs v0.0.0-20190507164030-5867b95ac084 // indirect
+	github.com/spf13/pflag v1.0.5 // indirect
+	github.com/subosito/gotenv v1.2.0 // indirect
+	gopkg.in/yaml.v2 v2.3.0 // indirect
+)
```

The new `require` block adds even more `// indirect` lines. Thanks to these new lines, Go commands such as `go get` have less downloading to do (the mechanism is called [lazy modules loading](https://go.googlesource.com/proposal/+/master/design/36460-lazy-module-loading.md)). Previously, `go get` had to download every single project in order to access their `go.mod` files, even if some of these projects were not actually used by your code. That was a big problem for git-based projects, and less of a problem for people using `GOPROXY` (i.e., everyone else) since the `go.mod` can be fetched without fetching the whole repository when using the `GOPROXY` mechanism.

For example, using the `GOPROXY` mechanism, the number of HTTP GET calls is reduced from 252 to 169 (49% less HTTP requests) with the cert-manager project for the tag v1.4.0. The number of HTTP calls was calculated using mitmproxy using `GOCACHE=off`. Although the number of HTTP calls has been reduced, I did not notice any significant time reduction.

The biggest difference will happen for people who are relying on Git for downloading repositories. Each time Go needs to read a `go.mod`, it needs to download the whole Git repository. We can force Go to clone the Git repositories by using `GOPRIVATE=*`:

```sh
# With Go 1.16:
$ time GOPATH=$(mktemp -d) GOPRIVATE='*' go get ./...
161.65s user 34.07s system 38% cpu 8:23.58 total


# With Go 1.17:
$ time GOPATH=$(mktemp -d) GOPRIVATE='*' go get ./...
158.09s user 33.39s system 39% cpu 7:59.63 total
```

That's 5% faster (24 seconds less with a 30 Mbit/s DSL connection). I would have expected a better result, but that number will depend on the `go.mod` you have!

#### Installing binaries with `GO111MODULE=on go get` is deprecated

In Go 1.17, the use of `go get` to install binaries is now showing the following warning:

```sh
$ GO111MODULE=on go get golang.org/x/tools/gopls@latest
go get: installing executables with 'go get' in module mode is deprecated.
     Use 'go install pkg@version' instead.
     For more information, see https://golang.org/doc/go-get-install-deprecation
     or run 'go help get' or 'go help install'.
```

To work around this warning, you can use `go install` that was taught how to deal with `@version` in Go 1.16:

```sh
go install golang.org/x/tools/gopls@latest
```

Notice that `@version` implies the Go module mode, meaning that you don't have to add `GO111MODULE=on` at the beginning of the command when you are not in a Go project that has a `go.mod`.

#### `go run` knows about `@version` (finally!)

The `go run` command was taught how to deal with `@version`! Like `go install`, it implies `GO111MODULE=on`. Up to now, if you wanted to run a binary once, you had two ways:

- If you were running the binary from a `go.mod` enabled project, you could specify the version of the binary you want to run in your `go.mod` and then run `go run -mod=mod`.

```
# Before Go 1.17:
go get github.com/golang/mock/mockgen@v1.3.1
go run -mod=mod github.com/golang/mock/mockgen

# With Go 1.17:
go run github.com/golang/mock/mockgen@v1.3.1
```

This is great if you use `//go:generate` to generate mocks. For example, in the project [users-grpc](https://github.com/maelvls/users-grpc/blob/ master/schema/placeholder.go), I was able to replace my `//go:generate` directive:

```go
// Old way, before 1.17:
//go:generate go run -mod=mod github.com/golang/mock/mockgen -build_flags=-mod=mod -package mocks -destination ./mock_service.go -source=../  user.go
// New way, Go 1.17:
//go:generate go run github.com/golang/mock/mockgen@v1.4.4 -package mocks -destination ./mock_service.go -source=../user.go
```

### `GO111MODULE` with Go 1.18

If you still need to use `go get` to install a binary, you will need to set `GO111MODULE=off`. The recommended way is to switch to using `go install` instead. Without `GO111MODULE=off`, `go get` will only update your `go.mod`. Otherwise, it won't do anything. All the READMEs on the Internet have to be updated; if they don't, people will get confused by not seeing the binary being installed.

### `GO111MODULE` with Go 1.19

Nothing with regards to `GO111MODULE` has changed with Go 1.19.

### `GO111MODULE` with Go 1.20

Nothing with regards to `GO111MODULE` has changed with Go 1.20.

### Why was `GO111MODULE` everywhere? (Go 1.15 and below)

Now that we know that `GO111MODULE` was very useful with Go 1.15 and below for enabling the Go Modules behavior, here is the answer: that's because `GO111MODULE=on` allows you to select a version. Without Go Modules, `go get` fetches the latest commit from master. With Go Modules, you can select a specific version based on git tags.

Before Go 1.16 got released, I used to use `GO111MODULE=on` very often when I wanted to switch between the latest version and the HEAD version of `gopls` (the Go Language Server):

```bash
# With Go 1.15.
GO111MODULE=on go get golang.org/x/tools/gopls@latest
GO111MODULE=on go get golang.org/x/tools/gopls@master
GO111MODULE=on go get golang.org/x/tools/gopls@v0.1
GO111MODULE=on go get golang.org/x/tools/gopls@v0.1.8
GO111MODULE="on" go get sigs.k8s.io/kind@v0.7.0
```

This became much easier with Go 1.16 though. I could do the same thing with:

```bash
# With Go 1.16.
go install golang.org/x/tools/gopls@latest
```

### The pitfall of `go.mod` being silently updated (Go 1.15 and below)

With Go 1.15 and below, our only option to install binaries was to use `go get`. You may have encountered this weird one-liner in READMEs:

```sh
(cd && GO111MODULE=on go get golang.org/x/tools/gopls@latest)
```

> **Note:** the `@latest` suffix will use the latest git tag of gopls. Note that `-u` (which means 'update') is not needed for `@v0.1.8` since this is a 'fixed' version, and updating a fixed version does not really make sense. It is also interesting to note that with `@v0.1`, `go get` will fetch the latest patch version for that tag.

So, why did we need to use this contrived command that calls a subshell and moves to your HOME? That's yet another Go ideocracy that was fixed with Go 1.16: in Go 1.15 and below, by default (and you can't turn that off), if you are in a folder that has a `go.mod`, `go get` will update that `go.mod` with what you just installed. And in the case of development binaries like [gopls](https://github.com/golang/tools/tree/master/gopls) or [kind](https://github.com/kubernetes-sigs/kind), you definitely don't want to have these appearing in the `go.mod` file!

So the workaround for anyone using Go 1.15 and below was to give a one-liner that makes sure that you won't be in a `go.mod`-enabled folder: `(cd && go get)` does exactly that.

Fortunately, with Go 1.16, there is now a clear separation of concerns between `go get` that is adding a dependency to your `go.mod` (like npm install) and `go install` that is meant to install a binary without messing up your `go.mod`. With Go 1.16, the above `go get` becomes:

```sh
go install golang.org/x/tools/gopls@latest
```

### The `-u` and `@version` pitfall

I have been bitten multiple times by this: when using `go get @latest` (for a binary, at least), you should avoid using `-u` so that it uses the dependencies as defined in `go.sum`. Otherwise, it will update all the dependencies to their latest minor revision... And since a ton of projects choose to have breaking changes between minor versions (e.g. v0.2.0 to v0.3.0), using `-u` has a large chance of breaking things.

So if you see this:

```sh
# Both -u and @latest!
GO111MODULE=on go get -u golang.org/x/tools/gopls@latest
```

then you will immediately realize that (1) it uses the old preGo .1.16 way of installing a Go binary, and (2) you want to be using the recorded versions given in `go.sum` when go-getting a binary!

Rebecca Stambler [reminds us](https://github.com/golang/go/issues/35868#issuecomment-564151454) that we should not use `-u` in conjunction with a version:

> `-u` should not be used in conjunction with the `@latest` tag, as it will give you incorrect versions of the dependencies.

But it's kind of hidden in this issue... I guess it is written somewhere in the Go help (btw, what a hideous help compared to `git help`) but that kind of caveat should be more visible: maybe print a warning when installing a binary with both `@version` and `-u`?

## Caveats when using Go Modules

Now, let's go through some caveats I encountered when working with Go Modules.

### Remember that `go get` also updates your `go.mod`

Before Go 1.16 came out, that was one of the weird things with `go get`: sometimes, it served the purpose of installing binaries or downloading packages. And with Go modules, if you were in a repo with a `go.mod`, it would silently add the binary that you were `go get`-ing to your `go.mod`!

Fortunately, with Go 1.16, `go install` has [learnt](https://blog.golang.org/go116-module-changes) about the `@version` suffix. With `go install foo@version`, your local `go.mod` won't be affected! And in Go 1.18, `go get` won't install binaries anymore, and will only ever be used for adding dependencies to your `go.mod`.

### Where are the sources of the dependencies with Go Modules

When using Go Modules, the packages that are used during `go build` are stored in `$GOPATH/pkg/mod`. When trying to inspect an 'import' in vim or VSCode, you might end up in the GOPATH version of the package instead of the pkg/mod one used during compilation.

A second issue that arises is when you want to hack one of your dependencies, for example for testing purposes.

**Solution 1**: use `go mod vendor` + `go build -mod=vendor`. That will force `go` to use the vendor/ files instead of using the `$GOPATH/pkg/mod` ones. This option also solves the problem of vim and VSCode not opening the right version of a package’s file.

**Solution 2**: add a 'replace' line at the end of your `go.mod`:

```plain
replace github.com/maelvls/beers => ../beers
```

where `../beers` is a local copy I made of the dependency I want to inspect and hack.

### Set `GO111MODULE` on a per-folder basis with `direnv`

On older versions of Go (1.15 and below), while migrating from GOPATH-based projects (mainly using Dep) to Go Modules, I found myself struggling with two different places: inside and outside GOPATH. All Go Modules had to be kept outside of GOPATH, which meant my projects were in different folders. To remediate that, I used `GO111MODULE` extensively. I would keep all my projects in the GOPATH, and for the Go Modules-enabled projects, I would set `export GO111MODULE=on`.

> **Note:** since the default behavior in Go 1.16 is now `GO111MODULE=on`, this trick isn't necessary anymore.

This is where [`direnv`](https://direnv.net/) came in handy. Direnv is a lightweight command written in Go that will load a file, `.envrc`, whenever you enter a directory and `.envrc` is present. For every Go Module-enabled project, I would have this `.envrc`:

```sh
# .envrc
export GO111MODULE=on
export GOPRIVATE=github.com/mycompany/\*
export GOFLAGS=-mod=vendor
```

The `GOPRIVATE` environment variable disables the Go Proxy (introduced in Go 1.13) for certain import paths. I also found useful to set `-mod=vendor` so that every command uses the `vendor` folder (`go mod vendor`).

### Private Go Modules and Dockerfile

Many companies choose to use private repositories as import paths. As explained above, we can use `GOPRIVATE` in order to tell Go (as of Go 1.13) to skip the package proxy and fetch our private packages directly from Github.

But what about building Docker images? How can `go get` fetch our private repositories from a docker build?

#### Solution 1: vendoring

With `go mod vendor`, no need to pass Github credentials to the docker build context. We can just put everything in `vendor/` and the problem is solved. In the Dockerfile, `-mod=vendor` will be required, but developers don't even have to bother with `-mod=vendor` since they have access to the private Github repositories anyway using their local Git config

- Pros: faster build on CI (~10 to 30 seconds less)
- Cons: PRs are bloated with `vendor/` changes and the repo's size might be big

#### Solution 2: no vendoring

If `vendor/` is just too big (e.g., for Kubernetes controllers, `vendor/` is about 30MB), we can very well do it without vendoring. That would require to pass some form of GITHUB_TOKEN as argument of `docker build`, and in the Dockerfile, set something like:

```sh
git config --global url."https://foo:${GITHUB_TOKEN}@github.com/company".insteadOf "https://github.com/company"
export GOPRIVATE=github.com/company/\*
```

_Illustration by Bailey Beougher, from The Illustrated Children's Guide to Kubernetes._

- **Update 22 June 2020:** it said `use replace` instead of just `replace`.
- **Update 8 April 2021:** update with Go 1.16.
- **Update 20 Sept 2021:** update with Go 1.17.
- **Update 18 Feb 2023:** mention the `go env GO111MODULE` gotcha reported by [Josh Soref](https://github.com/jsoref). I also mentioned Go 1.18, Go 1.19, and Go 1.20. Finally, I reformulated "Why is GO111MODULE everywhere" to "Why was GO111MODULE everywhere in Go 1.15 and below" thanks to [Ed Randall](https://github.com/edrandall)'s remark.
