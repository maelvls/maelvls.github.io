---
title: "Quick Kubernetes Controller in Bash"
description: ""
date: 2023-06-13T10:16:06+02:00
url: /quick-kubernetes-controller-in-bash
images:
  [
    quick-kubernetes-controller-in-bash/cover-quick-kubernetes-controller-in-bash.png,
  ]
draft: true
tags: []
author: Maël Valais
devtoId: 0
devtoPublished: false
---

<!--

https://github.com/maelvls/hack-your-controller-in-bash

-->

## Step 1: Controlling a built-in resource: Example With Secret



```bash
kubectl get externalsecret --watch -ojson \
  | jq 'select(.status.conditions[]?.reason == "SecretSyncedError")' --unbuffered \
  | jq '.spec.data[0].remoteRef | "\(.key) \(.property)"' -r --unbuffered \
  | while read key property; do
    vault kv put $key $property=somerandomvalue
  done
```

## Step 2: Controlling an external resource with a CRD: Example With

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must be in the form: <plural>.<group>
  name: posts.maelvls.dev
spec:
  group: maelvls.dev
  scope: Namespaced
  names: { kind: Post, singular: post, plural: posts }
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              x-kubernetes-preserve-unknown-fields: true
```
