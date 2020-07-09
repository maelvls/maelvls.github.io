---
title: "Deepequal Is Slow"
description: ""
date: 2020-07-09T09:02:48+02:00
url: /deepequal-is-slow
images: [deepequal-is-slow/cover-deepequal-is-slow.png]
draft: true
tags: []
author: Maël Valais
---

<!--
From https://github.com/ori-edge/deploy-controllers/pull/14#discussion_r432398876
-->

```go
package main_test

import (
    "reflect"
    "testing"
)

type Status struct {
    Available bool
    Addresses []string
}

var (
    s1 = Status{
        Available: true,
        Addresses: []string{"87.35.9.18", "87.34.9.11", "some.tld.org"},
    }
    s2 = Status{
        Available: true,
        Addresses: []string{"87.35.9.18", "87.34.9.11", "some1.tld.org"},
    }
)

func BenchmarkDeepEqual(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = reflect.DeepEqual(s1, s2)
    }
}

func BenchmarkFastEqual(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = FastEqual(s1, s2)
    }
}

func FastEqual(s1 Status, s2 Status) bool {
    if len(s1.Addresses) != len(s2.Addresses) || s1.Available != s2.Available {
        return false
    }

    for i, s1addr := range s1.Addresses {
        if s1addr != s2.Addresses[i] {
            return false
        }
    }

    return true
}
```

```sh
% go test -bench .
BenchmarkDeepEqual-8   12256587        96.6 ns/op
BenchmarkFastEqual-8    135698862      8.75 ns/op
```
