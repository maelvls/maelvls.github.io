---
title: "Kubernetes Kind vs Resource Type"
description: ""
date: 2021-04-09T12:04:10+02:00
url: /kubernetes-kind-vs-resource-type
images:
  [kubernetes-kind-vs-resource-type/cover-kubernetes-kind-vs-resource-type.png]
draft: true
tags: []
author: Maël Valais
devtoId: 0
devtoPublished: false
---

This is a summary of the official
[api-concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)
that I wrote to remember the difference between the various terms used to
designate "objects" in Kubernetes.

In Kubernetes, the term "resource" is a term coming from the RESTful
terminology. Let us have a YAML manifest example:

```yaml
# (1)
apiVersion: acme.cert-manager.io/v1alpha2
kind: Challenge
metadata:
  name: example-1
  namespace: foo
```

The resource corresponding would be:

```sh
GET /apis/acme.cert-manager.io/v1alpha2/namespaces/foo/challenges/example-1
#           <---------------> <-(2)-->                <--(1)---> <--(5)-->
#           <-------------------------(2) resource-------------------------->
```

In the following table, I focus on the "object" class of resource types. For
example, the resource type `pods` belongs to the "object" class of resource
types like most resource types in Kubernetes. The other possible resource type
category in Kubernetes is the "virtual" class of resource types, for example
`subjectaccessreviews`.

| Term                     |                                                | Example                                                                   |
| ------------------------ | ---------------------------------------------- | ------------------------------------------------------------------------- |
| (1) object resource type | Lower-case plural noun that appears in the URL | `challenges`                                                              |
| (2) object resource      | URL of one instance of the resource type       | `/apis/acme.cert-manager.io/v1alpha2/namespaces/foo/challenges/example-1` |
| (3) object resource name | Name of one instance of the resource type      | `example-1`                                                               |
| (4) object               | one concrete representation of a resource      | The above YAML manifest                                                   |
| kind                     | represents the object schema of the resource   | `Challenge`                                                               |
| apiVersion               |                                                | `acme.cert-manager.io/v1alpha2`                                           |
| version                  |                                                | `acme.cert-manager.io/v1alpha2`                                           |
| collection               | all instance of the resource type              | `/apis/acme.cert-manager.io/v1alpha2/challenges`                          |
