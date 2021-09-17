---
title: "Cert Manager Webhook Debug Guide"
description: ""
date: 2021-09-14T11:17:19+02:00
url: /cert-manager-webhook-debug-guide
images:
  [cert-manager-webhook-debug-guide/cover-cert-manager-webhook-debug-guide.png]
draft: true
tags: []
author: Maël Valais
devtoId: 0
devtoPublished: false
---

```
failed calling webhook "webhook.cert-manager.io": Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp 10.43.183.232:443: connect: connection refused
```

This happens when the `SYN` packet is answered with `RST`:

```
192.168.1.43  ->  34.142.22.43   TCP   59466 → 442 [SYN]
34.142.22.43  ->  34.142.22.43   TCP   442 → 59466 [RST, ACK]
```

This error message is quite rare. In most cases, the `SYN` packet is dropped by
some firewall; this error message indicates that the `SYN` packet was able to
reach the webhook pod. The webhook deployment is likely misconfigured (e.g.,
listening on the wrong port).

Issues:

- https://github.com/jetstack/cert-manager/issues/3445
- https://github.com/jetstack/cert-manager/issues/3133
- https://serverfault.com/questions/1076563/creating-issuer-for-kubernetes-cert-manager-is-causing-404-and-500-error
- https://github.com/jetstack/cert-manager/issues/3195
- https://github.com/jetstack/cert-manager/issues/2736

> My answer to https://github.com/jetstack/cert-manager/issues/3133:
>
> Hi!
>
> The error message `connect: connection refused` suggests that the apiserver is able to hit the webhook pod's TCP stack since the `SYN` packet was answered with `RST, ACK`, which means that nothing is listening on port 443 (if it was a firewall issue, the `SYN` would have been dropped).
>
> The Helm chart creates a Service that redirects traffic from `cluster-ip:443` to `pod-ip:10250`, and we can confirm by reading the logs that the webhook is listening on port 10250:
>
> ```
> I0925 14:18:30.749655       1 server.go:159] cert-manager/webhook "msg"="listening for secure connections"  "address"=":10250"
> ```
>
> At this point, the only thing that I could think of is that there is no pod IP associated with this service IP, as described by @r0bnet in https://github.com/jetstack/cert-manager/issues/3445#issue-738472291.

---

```
failed calling webhook "webhook.cert-manager.io": Post "https://cert-manager-webhook.cert-manager.svc:443/mutate?timeout=10s": dial tcp 10.43.183.232:443: i/o timeout
```

This time, the `SYN` packet is never answered, and the connection times out:

```
192.168.1.43  ->  34.142.22.43   TCP   44772 → 442 [SYN]
192.168.1.43  ->  34.142.22.43   TCP   [TCP Retransmission] 44772 → 442 [SYN]
192.168.1.43  ->  34.142.22.43   TCP   [TCP Retransmission] 44772 → 442 [SYN]
192.168.1.43  ->  34.142.22.43   TCP   [TCP Retransmission] 44772 → 442 [SYN]
...
```

This issue is caused by the `SYN` packet being dropped by the firewall. The
packet drop may have happened in many possible ways. One reason may be that the
cluster is configured with a network policy controller such as Calico, but there
is no policy allowing traffic from the apiserver to the webhook.
