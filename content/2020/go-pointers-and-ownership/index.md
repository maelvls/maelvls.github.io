---
title: "Go, pointers and data ownership"
description: ""
date: 2020-04-21T08:17:37+02:00
url: /go-pointers-and-ownership
images: [go-pointers-and-ownership/cover-go-pointers-and-ownership.png]
draft: true
---

After reading the venerable "[Using Pointers In Go
(2014)](https://www.ardanlabs.com/blog/2014/12/using-pointers-in-go.html)".
In this post, William Kennedy mentions the "tribal knowledge" around the
use of pointers:

> My use of pointers is based on discoveries I have made
> looking at code from the standard library. There are always exceptions to
> these rules, but what I will show you is common practice. It starts with
> classifying the type of value that needs to be shared. These type
> classifications are built-in, struct and reference types. Let’s look at
> each one individually. — William Kennedy

As far as I remember, Go was created with a minimal set of features. Why
did pointers make it into Go when we all know how confusing they are? I
really wich the Go team had gone with a pointer-less language but with
"move" semantics... That would have avoided the need for `nil` entirely!

At least, Go doesn't have arithmetic on pointers! With this, we avoid a
whole class of bugs that can be typically found in C and C++ and is the
source of endless security patches.

Anyway; now that we are stuck with pointers, I want to talk about how I
deal with them and how to "model" some of the decisions (using pointer vs
passing by value) based on the "ownership" of the data they point to.

When developing Kubernetes controllers, you end up with a lot of pointer
manipulation. A good practice in Go is to announce that an argument will be
mutated by passing it by pointer (I mean, that's what ............)

<!-- ![Example of a PR comment mentioning that a pointer automatically means "watch out, this data is modified by the function"](pointer-and-ownership.png) -->

<!--
https://github.com/ori-edge/edge-platform-controllers/pull/24
-->

Imagine you have:

```go
type Condition struct {
    Type string
}

// Update mutates the given slice.
func Update (current []Condition, new Condition) []Condition {
    for i, c := range current {
        if c.Type == new.Type {
            current[i] = new
            return current
        }
    }
    // Condition type doesn't currently exist, add it.
    return append(current, new)
}
```

We are mutating `current` but we still return it, which is confusing: is the
function mutating `current`?

One better way would be to write:

```go
func Update (current []Condition, new Condition) []Condition {
    for i, c := range current {
        if c.Type == new.Type {
            return append(current[0:i-1]..., new, current[i+1...])
        }
    }
    // Condition type doesn't currently exist, add it.
    return append(current, new)
}
```

That's much better.

## Pointer or not pointer?

```go
type Dictionary struct {
    words []string
}

func AddWord(d Dictionary, entry string) Dictionary
```

With that signature, I immediately see the data flow: a new Dictionary is
returned. But what about `words`? If we don't copy the whole slice, both `words`
will point to the same slice.

The `words` slice is a shared resource.

```go
func AddWord(d *Dictionary, entry string)
```

We kind of avoid that kind of signature in go -- at least when it's possible to
avoid it. We can't see the flow of data. Even worse: imagine that this function
works with an interface, which means the `*` won't even help us guess that this
data is shared.

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
