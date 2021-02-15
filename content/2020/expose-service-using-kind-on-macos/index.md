---
title: "Expose Service Using Kind on Macos"
description: ""
date: 2020-07-21T10:07:11+02:00
url: /expose-service-using-kind-on-macos
images:
  [
    expose-service-using-kind-on-macos/cover-expose-service-using-kind-on-macos.png,
  ]
draft: true
tags: [networking, sshuttle]
author: Maël Valais
devtoId: 0
devtoPublished: false
---

[Sshuttle](https://github.com/sshuttle/sshuttle) is a tool that lets you proxy traffic for an IP range (using iptables on Linux or pf on BSD derivatives including macOS) to a host through ssh.

I use Telepresence which makes sshuttle easier to use with Kubernetes. But I wondered: what happens under the hood? How come I am able to expose a service type=LoadBalancer using MetalLB on macOS using [Kind](https://github.com/kubernetes-sigs/kind)?

The idea is to have a Kind cluster (right box) in which we run a service type=LoadBalancer controller (MetalLB) and a simple pod. Making an HTTP request from the host should hit the pod thanks to sshuttle:

```plain
                                      +------------------------------+
                                      |                              |
 curl http://172.17.0.200             |  +-----------------------+   |
            |                         |  | metallb speaker (arp) |   |
            |                         |  +-----------------------+   |
            |172.17.0.0/16            |   range 172.17.0.{200-250}   |
            |                         |                              |
            v                         |                              |
      +------------+                  |    +------------------+      |
      |host's      |                  |    | pod nginx:alpine |      |
      |iptables/pf |                  |    +------------------+      |
      +------------+                  |             ^                |
            |                         |             |new tcp         |
            |                         |             |connection      |
            |                         |             |172.17.0.200:80 |
            v                         |             |                |
        tcp + udp     (raw over ssh)  |    +------------------+      |
       proxy server ---------------------> |   sshuttle pod   |      |
        (sshuttle)                    |    +------------------+      |
                                      +------------------------------+
```

<!-- https://textik.com/#19fa729cd40cd953 -->

Let's create a Kind cluster and setup MetalLB in L2 mode. MetalLB will advertize services of type LoadBalancer by picking up an IP on the range `172.17.0.200-172.17.0.250` and will advertize this IP using the ARP protocol.

```sh
export KUBECONFIG=/tmp/kind-kubeconfig
kind create cluster
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/metallb.yaml
kubectl -n metallb-system create secret generic memberlist --from-literal=secretkey="$(openssl rand -base64 128)"
kubectl -n metallb-system delete configmap config
NET=$(docker network inspect bridge | jq '.[].IPAM.Config[0] | select(.Gateway) | .Subnet' -r | sed 's|/16||')
kubectl -n metallb-system create configmap config --from-file=config=/dev/stdin <<EOF
address-pools:
- name: default
  protocol: layer2
  addresses:
  - ${NET/%.0/.200}-${NET/%.0/.250}
EOF
```

Now, let's create a pod and a service type=LoadBalancer:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx:alpine
      ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 80
```

Finally, let's run sshuttle:

```sh
brew install sshuttle
```

> ⚠️ I'm stuck here since `ssh` on the host can't reach any sshd... What telepresence does is to have an sshd running in some pod, run `kubectl port-forward` so that `ssh` can work from the host.

## Packet tunneling using sshuttle

Sshuttle runs a TCP/UDP proxy on the host on port 12300. The traffic that goes to `172.17.0.0/16` is then forwarded to that port (src and dst headers are left unchanged). Note that the 12300 port deals with both TCP and UDP packets at the same time!

- On macOS:

  ```sh
  pfctl -a sshuttle6-12300 -f /dev/stdin
  pfctl -a sshuttle-12300 -f /dev/stdin
  pfctl -E
  ```

- On Linux:

  ```sh
  iptables -t nat -N sshuttle-12300
  iptables -t nat -F sshuttle-12300
  iptables -t nat -I OUTPUT 1 -j sshuttle-12300
  iptables -t nat -I PREROUTING 1 -j sshuttle-12300
  iptables -t nat -A sshuttle-12300 -j RETURN --dest 172.17.0.0/16 -p tcp
  iptables -t nat -A sshuttle-12300 -j REDIRECT --dest 172.17.0.0/16 -p tcp --to-ports 12300 -m ttl '!' --ttl 42
  iptables -t nat -A sshuttle-12300 -j REDIRECT --dest 172.17.0.0/16 -p udp -dport 53 --to-ports 12300 -m ttl '!' --ttl 42
  ```
