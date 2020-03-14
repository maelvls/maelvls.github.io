---
title: "The Packet's-Eye View of a Kubernetes Service"
description: ""
date: 2020-02-17
url: /packets-eye-kubernetes-service
draft: true
---

<!--

1. service vs. ingress (L4 vs L7)

<https://github.com/kubernetes/ingress-gce/search?p=3&q=LoadBalancer&unscoped_q=LoadBalancer>

2. how service and ingress interact with their controllers
3. traffic flow with GKE's service LB and Traefik
4. using my own service controller
5. traffic flow with my own service controller
6. comparison, benchmark, recap

-->

A few weeks back, I had already written about [how to avoid the expensive
GKE load balancer](/avoid-gke-lb-with-hostport/). This time, I want to go a
bit deeper and detail how the Service object routes packets to a pod and
how the 'hostPort' method actually works under the hood.

In Kubernetes, the Service object holds the L4 endpoints that allow you to
expose a set of pods using selectors. Since it is L4-only, the service only
deals with TCP and UDP: only IPs and ports.

An Ingress is the L7 counterpart of the service: it deals with TLS,
hostnames, HTTP paths and virtual servers (called "backends").

In this article, I will focus on services and how the L4 traffic flows. You
might see some mentions of ingresses since Traefik is an ingress controller
but I will not describe the traffic handling at the L7 level.

An ingress controller, like Traefik, is a binary running as a simple
deployment that watches all Ingress objects and live-updates the L7 proxy.
Traefik has the nice property of embeding both the "ingress watcher" and
the L7 proxy in a single binary. HAProxy and Nginx both use separate
ingress controllers. And in order to get L4 traffic flowing in, the L7
proxy needs a service.

Now, we also want external traffic to be able to come in. That's where we
usually use an external load balancer like GKE's Network Load Balancer.
What happens is that Google runs a small closed-source binary that runs on
your master node and acts as a service-lb controller. And unfortunately you
cannot see any logs since it is not running as a pod.

I call "service-lb controller" a binary that watches service objects that
have the `LoadBalancer` type. Any time the Google's service-lb controller
sees a new `LoadBalancer` service, it spins up a load balancer and sets up
some firewall rules. Here is a diagram that represents how the service-lb
controller interacts with the ingress controller:

![](kubernetes-service-controllers-with-gke-service.svg)

So, how does external traffic make it to the pod? The following diagram
shows how a packet is forwarded to the right pod. Notice how many iptable
rewriting happen (which corresponds to one connection managed by
conntrack):

![](kubernetes-traffic-with-gke-lb.svg)

Now, let's see how it goes when using Akrobateo (I detailed that
[here](](/avoid-gke-lb-with-hostport/))). Instead of using an external
compute resource, we use the node's IP in order to let traffic in.

Note: Akrobateo is EOL, but K3s's service-lb and Metallb work in a very
similar way, setting the service's `status.loadBalancer` field with the
correct external IP.

![](kubernetes-service-controllers-with-akrobateo.svg)

You might wonder why the 'internal IP' is used in `status.loadBalancer`. We
might expect the external IP to be set there. But since Google's VPC NAT
swaps the node's external IP to the node's internal IP, the incoming
packets have the internal IP as source IP. So that's why 😁.

Here is the diagram of a packet being routed towards a pod:

![](kubernetes-traffic-with-akrobateo.svg)

With that method, we only rely on the VPC's firewall rules. But using the
node's IP is not perfect: it might be a seen as a security risk, and the
many IPs that end up in the service's `status.loadBalancer` isn't ideal
when it comes to setting your DNS `A` records:

- when a node disappears, the DNS might forward traffic to a unexisting node,
- offloading ingress traffic load balancing from a load balancer with a
  single IP to relying on DNS records isn't ideal since DNS records leave
  you very few options as to how to balance traffic.

K3s uses this approach of using the node IPs as service backends and does
not need any external load balancer.
