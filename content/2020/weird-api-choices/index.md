---
title: "Weird Api Choices"
description: ""
date: 2020-05-29T08:35:36+02:00
url: /weird-api-choices
images: [weird-api-choices/cover-weird-api-choices.png]
draft: true
tags: []
---

# Map vs. discriminative arrays (map key -> discriminative name)

From `docker inspect`:

```json
{
    Networks: {
        "bridge": {
            IPAddress: "172.17.0.2"
        },
        "kind": {
            IPAddress: "172.18.0.2"
        }
    }
}
```
versus:
```json
{
    Networks: [
        {
            Name: "bridge"
            IPAddress: "172.17.0.2"
        },
        {
            Name: "kind",
            IPAddress: "172.18.0.2"
        }
    ]
}
```

# Polymorphic vs. discriminative array (polymorphic object -> discriminative type)

From the [Service](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#loadbalanceringress-v1-core) type in Kubernetes:

```yaml
status:
  loadBalancer:
    ingress:
    - ip: 35.67.89.10
    - ip: 35.67.89.11
    - hostname: nginx-test.gcp-1.helix.engineering
```

It is confusing: can we have an object with both "ip" and "hostname" set? Using a discriminative type helps:

```yaml
status:
  loadBalancer:
    ingress:
    - type: ip
      address: 35.67.89.10
    - type: ip
      address: 35.67.89.11
    - type: hostname
      address: nginx-test.gcp-1.helix.engineering
```
