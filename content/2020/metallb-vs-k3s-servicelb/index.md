---
title: "MetalLB vs. K3s servicelb"
description: ""
date: 2020-04-25T12:25:02+02:00
url: /metallb-vs-k3s-servicelb
images: [metallb-vs-k3s-servicelb/cover-metallb-vs-k3s-servicelb.png]
draft: true
tags: []
---

Great post by Duffie Cooley in Apr 2019 "[Using MetalLb with
Kind](https://mauilion.dev/posts/kind-metallb)":

```sh
% kind create cluster --name metallb --config - <<EOF
heredoc> kind: Cluster
apiVersion: kind.sigs.k8s.io/v1alpha3
nodes:
- role: control-plane
- role: worker
heredoc> EOF
```

```sh
% k get nodes
NAME                    STATUS     ROLES    AGE   VERSION
metallb-control-plane   Ready      master   42s   v1.17.2
metallb-worker          NotReady   <none>   9s    v1.17.2
```

```sh
% docker network inspect bridge | jq -r '.[].IPAM.Config[].Subnet'
172.17.0.0/16
```

```sh
% arp -a -n
? (192.168.1.1) at a4:3e:51:cf:e7:96 on en9 ifscope [ethernet]
? (192.168.1.1) at a4:3e:51:cf:e7:96 on en0 ifscope [ethernet]
? (192.168.1.12) at 60:f4:45:70:a9:82 on en9 ifscope [ethernet]
? (192.168.1.21) at 80:49:71:10:b4:ba on en0 ifscope [ethernet]
? (192.168.1.255) at ff:ff:ff:ff:ff:ff on en9 ifscope [ethernet]
? (224.0.0.251) at 1:0:5e:0:0:fb on en9 ifscope permanent [ethernet]
? (224.0.0.251) at 1:0:5e:0:0:fb on en0 ifscope permanent [ethernet]
? (239.255.255.250) at 1:0:5e:7f:ff:fa on en9 ifscope permanent [ethernet]
? (239.255.255.250) at 1:0:5e:7f:ff:fa on en0 ifscope permanent [ethernet]
```

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
