---
title: "Why Kubernetes Compares With Hash"
description: ""
date: 2020-07-08T11:55:27+02:00
url: /why-kubernetes-compares-with-hash
images: [why-kubernetes-compares-with-hash/cover-why-kubernetes-compares-with-hash.png]
draft: true
tags: []
author: Maël Valais
---

In Kubernetes, the ReplicaSet controller uses a hashing mechanism to
compare objects. In this post, I want to present the advantages and
limitations of using a hash function in this context.

- **advantage**: works around the fact that child object might get mutated
  or defaulted, which means the `reflect.DeepEqual` can't work (**what about performance???**)
- **disadvantage**: if the child gets updated, the parent cannot know that
  it has been changed since the hash only works in one way. (**talk about
  ReplicaSet mapping to multiple pods**)
- why is it a label and not an annotation?

```yaml
kind: Pod
metadata:
  labels:
    pod-template-hash: 775855699b
```

Looks like the hash is created [in the replicaset sync func][rs-sync] and
it uses the [`ComputeHash`][ComputeHash] func which uses
[fnv.New32a](https://golang.org/pkg/hash/fnv/#New32) from the std lib:

```go
// ComputeHash returns a hash value calculated from pod template and
// a collisionCount to avoid hash collision. The hash will be safe encoded to
// avoid bad words.
func ComputeHash(template *v1.PodTemplateSpec, collisionCount *int32) string {
    podTemplateSpecHasher := fnv.New32a()
    hashutil.DeepHashObject(podTemplateSpecHasher, *template)

    // Add collisionCount in the hash if it exists.
    if collisionCount != nil {
        collisionCountBytes := make([]byte, 8)
        binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
        podTemplateSpecHasher.Write(collisionCountBytes)
    }

    return rand.SafeEncodeString(fmt.Sprint(podTemplateSpecHasher.Sum32()))
}
```

and the [`DeepHashObject`][DeepHashObject] looks like this:

```go
// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
    hasher.Reset()
    printer := spew.ConfigState{
        Indent:         " ",
        SortKeys:       true,
        DisableMethods: true,
        SpewKeys:       true,
    }
    printer.Fprintf(hasher, "%#v", objectToWrite)
}
```

So the hashing mechanism uses
[davecgh/go-spew](https://github.com/davecgh/go-spew) to turn the object
into a string, and then uses the `fnv` std library to hash that string.

[rs-sync]: https://github.com/kubernetes/kubernetes/blob/7e75a5ef/pkg/controller/deployment/sync.go#L189
[ComputeHash]: https://github.com/kubernetes/kubernetes/blob/7e75a5ef/pkg/controller/controller_utils.go#L1130-L1145
[DeepHashObject]: https://github.com/kubernetes/kubernetes/blob/7e75a5ef/pkg/util/hash/hash.go#L25-L37


## Benchmarking two hashing functions

```go
package main

import (
    "hash"
    "hash/fnv"
    "testing"

    "github.com/davecgh/go-spew/spew"
    "github.com/mitchellh/hashstructure"
)

type ComplexStruct struct {
    Name     string
    Age      uint
    Metadata map[string]interface{}
}

var v = ComplexStruct{
    Name: "mitchellh",
    Age:  64,
    Metadata: map[string]interface{}{
        "car":      true,
        "location": "California",
        "siblings": []string{"Bob", "John"},
    },
}

func BenchmarkMitchellhHashstructure(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, _ = hashstructure.Hash(v, nil)
    }
}

func BenchmarkKubernetesComputeHash(b *testing.B) {
    for i := 0; i < b.N; i++ {
        hasher := fnv.New32a()
        DeepHashObject(hasher, v)
        _ = hasher.Sum32()
    }
}

// From https://github.com/kubernetes/kubernetes/blob/7e75a5ef/pkg/util/hash/hash.go#L25-L37
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
    hasher.Reset()
    printer := spew.ConfigState{
        Indent:         " ",
        SortKeys:       true,
        DisableMethods: true,
        SpewKeys:       true,
    }
    printer.Fprintf(hasher, "%#v", objectToWrite)
}
```

```sh
% go test -bench .
goos: darwin
goarch: amd64
pkg: test
BenchmarkMitchellhHashstructure-8         368101              3.127 µs/op
BenchmarkKubernetesComputeHash-8          456028              2.704 µs/op
```
