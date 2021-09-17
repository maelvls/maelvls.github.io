---
title: "Go Flaky Test"
description: ""
date: 2021-09-02T15:07:38+02:00
url: /go-flaky-test
images: [go-flaky-test/cover-go-flaky-test.png]
draft: true
tags: []
author: Maël Valais
devtoId: 0
devtoPublished: false
---

## Debugging a flaky test

- `go test` does not allow you to run a single test many times in parallel. If
  your test takes 10 seconds (in our case, it had to spin up an apiserver for
  every test case), and that you want to run this test 100 times, it will take
  16 minutes. The flags `-count`, `-cpu`, `-parallel` and the function
  `t.Parallel` only work for tests in separate packages. Multiple instances of a
  single test never run in parallel with each other.
- ```sh
  parallel "go test ./test/integration/ -count=1" ::: {1..100}
  ```
