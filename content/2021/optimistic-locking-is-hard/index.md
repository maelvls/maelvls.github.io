---
title: "Optimistic Locking Is Hard"
description: ""
date: 2021-03-31T14:51:20+02:00
url: /optimistic-locking-is-hard
images: [optimistic-locking-is-hard/cover-optimistic-locking-is-hard.png]
draft: true
tags: []
author: Maël Valais
devtoId: 0
devtoPublished: false
---

> Each one of these controllers actually only manages a very specific subset of the Certificate status

That's something I have always been wary about: although it is recommended that each part of the status is only managed by a single controller, we currently have multiple controllers operating on the same `Issuing` condition.

This idea will seem a bit far-fetched, but why wouldn't we use the notion of "one state-level = one condition"? (as in "[Kubernetes is level-based](https://stackoverflow.com/questions/31041766/what-does-edge-based-and-level-based-mean)")

If we want to keep the controller loops separated, they should all clearly rely on waiting for different levels instead of having the level muddled like we have right now:

```
if trigger controller sees         Issuing=False         and         lastFailure set              →           set Issuing=True
if issuing controller sees        Issuing=False         and         lastFailure set              →           set Issuing=True
```