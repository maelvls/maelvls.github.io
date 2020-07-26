---
title: Pull-through Docker registry on Kind clusters on macOS
description: "Kind offers an excellent UX to Kubernetes developers but lacks support for caching images; each time you recreate a new cluster, all the previous downloaded images are gone. In this post, I explain why the default Docker network is a trap and how to set up a registry & make sure that it actually works."
date: 2020-07-03T15:13:39+02:00
url: /docker-proxy-registry-kind-macos
images: [docker-proxy-registry-kind-macos/cover-docker-proxy-registry-kind-macos.png]
tags: [kubernetes, kind, docker, networking]
author: Maël Valais
---

<!--
Diagram on macOS + Docker: https://textik.com/#b185c1a72a6e782d
-->

**TL;DR:**

- to create a local pull-through registry to speed up image pulling in a
  [Kind](https://kind.sigs.k8s.io/) cluster, run:

  ```sh
  docker run -d --name proxy --restart=always --net=kind -e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io registry:2
  kind create cluster --config /dev/stdin <<EOF
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  containerdConfigPatches:
    - |-
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
        endpoint = ["http://proxy:5000"]
  EOF
  ```

- [you can't](https://docs.docker.com/registry/configuration/#proxy) use
  this pull-through proxy registry to push your own images (e.g. to [speed
  up Tilt builds](https://github.com/tilt-dev/kind-local)), but you can
  create two registries (one for caching, the other for local images). See
  [this section](#docker-proxy-vs-local-registry) for more context; the
  lines are:

  ```sh
  docker run -d --name proxy --restart=always --net=kind -e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io registry:2
  docker run -d --name registry --restart=always -p 5000:5000 --net=kind registry:2
  kind create cluster --config /dev/stdin <<EOF
  kind: Cluster
  apiVersion: kind.x-k8s.io/v1alpha4
  containerdConfigPatches:
    - |-
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
        endpoint = ["http://proxy:5000"]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
        endpoint = ["http://registry:5000"]
  EOF
  ```

- in case you often create & delete Kind clusters, using a local registry
  that serves as a proxy avoids redundant downloads
- `KIND_EXPERIMENTAL_DOCKER_NETWORK` is useful but remember that the
  default network (`bridge`) doesn't have DNS resolution for container
  hostnames
- the Docker default network (`bridge`) [has
  limitations](https://stackoverflow.com/questions/41400603/dockers-embedded-dns-on-the-default-bridged-network)
  as [detailed by Docker][default-bridge].
- If you play with [ClusterAPI](https://cluster-api.sigs.k8s.io/) with its
  [Docker provider][docker-provider], you might not be able to use a local
  registry due to the clusters being created on the default network, which
  means the "proxy" hostname won't be resolved (but we could work around
  that).

[docker-provider]: https://github.com/kubernetes-sigs/cluster-api/tree/master/test/infrastructure/docker

---

[Kind](https://kind.sigs.k8s.io/) is an awesome tool that allows you to
spin up local Kubernetes clusters locally in seconds. It is perfect for
Kubernetes developers or anyone who wants to play with controllers.

One thing I hate about Kind is that images are not cached between two Kind
containers. Even worse: when deleting and re-creating a cluster, all the
downloaded images disappear.

In this post, I detail my discoveries around local registries and why the
default Docker network is a trap.

1. [Kind has no image caching mechanism](#kind-has-no-image-caching-mechanism)
2. [The DNS mysteries of the Docker default network (`bridge`)](#the-dns-mysteries-of-the-docker-default-network-bridge)
3. [Docker proxy vs. local registry](#docker-proxy-vs-local-registry)
4. [Improving the ClusterAPI docker provider to use a given network](#improving-the-clusterapi-docker-provider-to-use-a-given-network)

---

## Kind has no image caching mechanism

Whenever I re-create a Kind cluster and try to install ClusterAPI, all the
(quite heavy) images have to be re-downloaded. Just take a look at all the
images that get re-downloaded:

```sh
# That's the cluster created using 'kind create cluster'
% docker exec -it kind-control-plane crictl images
IMAGE                                                                      TAG      SIZE
quay.io/jetstack/cert-manager-cainjector                                   v0.11.0  11.1MB
quay.io/jetstack/cert-manager-controller                                   v0.11.0  14MB
quay.io/jetstack/cert-manager-webhook                                      v0.11.0  14.3MB
us.gcr.io/k8s-staging-capi-docker/capd-manager/capd-manager-amd64          dev      53.5MB
us.gcr.io/k8s-artifacts-prod/cluster-api/cluster-api-controller            v0.3.0   20.3MB
us.gcr.io/k8s-artifacts-prod/cluster-api/kubeadm-bootstrap-controller      v0.3.0   19.6MB
us.gcr.io/k8s-artifacts-prod/cluster-api/kubeadm-control-plane-controller  v0.3.0   21.1MB

# I also use a ClusterAPI-created cluster (relying on CAPD):
% docker exec -it capd-capd-control-plane-l4tx7 crictl images ls
docker.io/calico/cni                  v3.12.2             8b42391a46731       77.5MB
docker.io/calico/kube-controllers     v3.12.2             5ca01eb356b9a       23.1MB
docker.io/calico/node                 v3.12.2             4d501404ee9fa       89.7MB
docker.io/calico/pod2daemon-flexvol   v3.12.2             2abcc890ae54f       37.5MB
docker.io/metallb/controller          v0.9.3              4715cbeb69289       17.1MB
docker.io/metallb/speaker             v0.9.3              f241be9dae666       19.2MB
```

That's a total of 418 MB that get re-downloaded every time I restart both
clusters!

Unfortunately, there is no way to re-use the image registry built into
Docker for Mac. One solution to this problem is to [spin up an intermediary
Docker registry](https://kind.sigs.k8s.io/docs/user/local-registry/) in a
side container; as long as this container exists, all the images that have
already been downloaded once can be served from cache.

## The DNS mysteries of the Docker default network (`bridge`)

We want to create a registry with a simple Kind cluster; let's start with
the registry:

```sh
docker run -d --name proxy --restart=always --net=kind -e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io registry:2
```

Details:

- `--net kind` is required because Kind creates its containers in a
  separate network; it does that the because the "bridge" has
  [limitations][default-bridge] and [doesn't allow you][dns-services] to
  use container names as DNS names:

  > By default, a container inherits the DNS settings of the host, as
  > defined in the `/etc/resolv.conf` configuration file. Containers that
  > use the default bridge network get a copy of this file, whereas
  > containers that use a custom network use Docker’s embedded DNS server,
  > which forwards external DNS lookups to the DNS servers configured on
  > the host.

  which means that the container runtime (containerd) that runs our Kind
  cluster won't be able to resove the address `proxy:5000`.

- `REGISTRY_PROXY_REMOTEURL` is required due to the fact that by default,
  the registry won't forward requests. It simply tries to find the image in
  `/var/lib/registry/docker/registry/v2/repositories` and returns 404 if it
  doesn't find it.

  > Using the
  > [pull-through](https://docs.docker.com/registry/configuration/#proxy)
  > feature (I call it "caching proxy"), the registry will proxy all
  > requests coming from all mirror prefixes and cache the blobs and
  > manifests locally. To enable this feature, we set
  > `REGISTRY_PROXY_REMOTEURL`.
  >
  > Other interesting bit about `REGISTRY_PROXY_REMOTEURL`: this
  > environement variable name is mapped from [the registry YAML config
  > API](https://docs.docker.com/registry/configuration/#proxy). The
  > variable
  >
  > ```sh
  > REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io
  > ```
  >
  > is equivalent to the following YAML config:
  >
  > ```yaml
  > proxy:
  >   remoteurl: https://registry-1.docker.io
  > ```

  ⚠️ The registry can't be both in normal mode ("local proxy") and in
  caching proxy mode at the same time, see
  [below](#docker-proxy-vs-local-registry).

[default-bridge]: https://docs.docker.com/network/bridge/#use-the-default-bridge-network
[dns-services]: https://docs.docker.com/config/containers/container-networking/#dns-services

The second step is to create a Kind cluster and tell the container runtime
to use a specific registry; here is the command to create it:

```sh
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://proxy:5000"]
EOF
```

Details:

- `containerdConfigPatches` is a way to semantically patch
  `/etc/containerd/config.conf`. By default, this file looks like:

  ```sh
  % docker exec -it kind-control-plane cat /etc/containerd/config.toml
  [plugins]
    [plugins."io.containerd.grpc.v1.cri"]
      [plugins."io.containerd.grpc.v1.cri".registry]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
          [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
            endpoint = ["https://registry-1.docker.io"]
  ```

  For information, the mirror prefix (`docker.io`) can be omitted for
  images stored on Docker Hub. For other registries such as `gcr.io`, this
  mirror prefix has to be given. Here is a table with some examples of
  image names that are first prepended with "docker.io" if the mirror
  prefix is not present, and we get the final address by mapping these
  mirror prefixes with mirror entries:

  | image name                  | "actual" image name         | registry address w.r.t. mirrors        |
  | --------------------------- | --------------------------- | -------------------------------------- |
  | alpine                      | docker.io/alpine            | https://registry-1.docker.io/v2/alpine |
  | gcr.io/istio-release/galley | gcr.io/istio-release/galley | https://gcr.io/v2/istio-release/galley |
  | something/someimage         | something/someimage         | https://something/v2/someimage         |

Let's see if the proxy registry works by running a pod:

```sh
% kubectl run foo -it --rm --image=nicolaka/netshoot
% docker exec -it proxy ls /var/lib/registry/docker/registry/v2/repositories
nicolaka
```

We can also see through the registry logs that everything is going well:

```log
# docker logs proxy | tail
time="2020-07-26T14:52:44.2624761Z" level=info msg="Challenge established with upstream : {https   registry-1.docker.io /v2/  %!s(bool=false)  } &{{{%!s(int32=0) %!s(uint32=0)} %!s(uint32=0) %!s(uint32=0) %!s(int32=0) %!s(int32=0)} map[https://registry-1.docker.io:443/v2/:[{bearer map[realm:https://auth.docker.io/token service:registry.docker.io]}]]}" go.version=go1.11.2 http.request.host="proxy:5000" http.request.id=15e9ac86-7d79-4883-a8ce-861a7484887c http.request.method=HEAD http.request.remoteaddr="172.18.0.2:57588" http.request.uri="/v2/nicolaka/netshoot/manifests/latest" http.request.useragent="containerd/v1.4.0-beta.1-34-g49b0743c" vars.name="nicolaka/netshoot" vars.reference=latest
time="2020-07-26T14:52:45.4195817Z" level=info msg="Adding new scheduler entry for nicolaka/netshoot@sha256:04786602e5a9463f40da65aea06fe5a825425c7df53b307daa21f828cfe40bf8 with ttl=167h59m59.9999793s" go.version=go1.11.2 instance.id=ba959eb9-2fa3-47c0-beb7-91480c8a31ee service=registry version=v2.7.1
172.18.0.2 - - [26/Jul/2020:14:52:43 +0000] "HEAD /v2/nicolaka/netshoot/manifests/latest HTTP/1.1" 200 1999 "" "containerd/v1.4.0-beta.1-34-g49b0743c"
time="2020-07-26T14:52:45.4204299Z" level=info msg="response completed" go.version=go1.11.2 http.request.host="proxy:5000" http.request.id=15e9ac86-7d79-4883-a8ce-861a7484887c http.request.method=HEAD http.request.remoteaddr="172.18.0.2:57588" http.request.uri="/v2/nicolaka/netshoot/manifests/latest" http.request.useragent="containerd/v1.4.0-beta.1-34-g49b0743c" http.response.contenttype="application/vnd.docker.distribution.manifest.v2+json" http.response.duration=1.6697112s http.response.status=200 http.response.written=1999
```

<!--

## Why "proxy" wasn't resolved inside containers

```sh
% docker inspect proxy --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
bridge 172.17.0.2

% docker inspect kind-control-plane --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
kind 172.18.0.2

% docker network ls
NETWORK ID          NAME                DRIVER              SCOPE
a6ceea984c68        bridge              bridge              local
6f6a9618d746        host                host                local
4927dc2eba9b        kind                bridge              local
```

Let's try to reproduce this issue. I first run a registry with the default
network, and I then try to connect to it from a second container using the
hostname `proxy`:

```sh
% docker run -d --rm --name proxy registry:2
% docker run -it --rm alpine nslookup proxy
Server:         127.0.0.11
Address:        127.0.0.11:53
Non-authoritative answer:
Name:   proxy
Address: 172.18.0.7

# Let's cleanup:
% docker kill proxy
```

Now, let's do the same but a custom network named "other" instead of the
default network:

```sh
% docker network create other
% docker run -d --rm --net=other --name proxy registry:2
% docker run -it --rm --net=other alpine nslookup proxy
Server:         127.0.0.11
Address:        127.0.0.11:53
Non-authoritative answer:
Name:   proxy
Address: 172.18.0.7

# Let's cleanup:
% docker kill proxy
```

I thought I could first start the container on the default network and then
move the containers to the "other" network (so that DNS with container
names works) but it does not seem to work either:

```sh
% docker run -d --rm --name proxy registry:2
% docker run -d --rm --name alpine alpine sleep 1d

% docker network create other
% docker network disconnect bridge proxy
% docker network disconnect bridge alpine
% docker network connect other proxy
% docker network connect other alpine

# Now, let's see if 'alpine' can resolve 'proxy':
% docker exec -it alpine nslookup proxy
```

I also tried to understand the difference between containers with the
default network and containers with the "other" network. As stated [in the
Docker documentation][dns-services], a container created on the default
network is setup slightly differently.

```sh
# 1️⃣ With default network:
% docker run -it --rm alpine cat /etc/resolv.conf
# This file is included on the metadata iso
nameserver 192.168.65.1

% docker run -it --rm alpine cat /etc/hosts
127.0.0.1       localhost
::1     localhost ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
172.17.0.4      8f4a5e6bad39

# 2️⃣ With the 'other' network:
% docker run -it --rm --net=other alpine cat /etc/resolv.conf
nameserver 127.0.0.11
options ndots:0

% docker run -it --rm --net=other alpine cat /etc/hosts
127.0.0.1       localhost
::1     localhost ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
172.19.0.4      fed220ca6bd0
```

So it definitely comes from `/etc/resolv.conf`! With the default network,
the `/etc/resolv.conf` that gets set up does not allow you to resolve
container names.

So in order to use container names as hostnames, I have to create your own
using `docker network create`.

Let's re-created my registry on the "kind" network and also re-create the
cluster without the `KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge` option:

```sh
docker network create kind
docker run -d --name proxy --restart=always --net=kind registry:2
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."proxy"]
      endpoint = ["http://proxy:5000"]
EOF
```

Both are now on the same subnet:

```sh
% docker inspect proxy --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
kind 172.18.0.2

% docker inspect kind-control-plane --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
kind 172.18.0.3
```

But creating a simple deployment doesn't seem to work! By looking at the
registry logs, we see no activity at all. Let's see what `containerd` is up
to:

```log
% docker exec -i kind-control-plane journalctl -u containerd | grep 'proxy:5000'
containerd[129]: Start cri plugin with config {Registry:{Mirrors:map[
  docker.io: {Endpoints:[https://registry-1.docker.io]}
  proxy:  {Endpoints:[http://proxy:5000]}
]}}
```

> ✅ I had to remove `-t` (tty) from the above command. That's because
> `journalctl` was using a pager because this terminal was TTY (which means
> it had /dev/stdin open? not sure). To disable the pager I removed `-t`.

I guess contained picks the first mirror (`docker.io`); so let's override
the `docker.io` key:

```sh
docker rm -f registry
docker run -d --name proxy --restart=always --net=kind registry:2
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://proxy:5000"]
EOF
```

And here is the actual configuration given to containerd:

```sh
% docker exec -it kind-control-plane cat /etc/containerd/config.toml
version = 2
[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    sandbox_image = "k8s.gcr.io/pause:3.2"
    [plugins."io.containerd.grpc.v1.cri".containerd]
      default_runtime_name = "runc"
      [plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
          runtime_type = "io.containerd.runc.v2"
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler]
          runtime_type = "io.containerd.runc.v2"
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = ["https://registry-1.docker.io"]
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."proxy"]
          endpoint = ["http://proxy:5000"]
```

-->

## Docker proxy vs. local registry

A bit later, I discovered that [you
can't](https://docs.docker.com/registry/configuration/#proxy) push to a
proxy registry. [Tilt](https://tilt.dev/) is a tool I use to ease the
process of developping in a containerized environment (and it works best
with Kubernetes); it [relies on a local
registry](https://github.com/tilt-dev/kind-local) in order to cache build
containers even when restarting the Kind cluster.

Either the registry is used as a "local registry" (where you can push
images), or it is used as a pull-through proxy. So instead of configuring
one single "proxy" registry, I configure two registries: one for local
images, one for caching.

```sh
docker run -d --name proxy --restart=always --net=kind -e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io registry:2
docker run -d --name registry --restart=always -p 5000:5000 --net=kind registry:2
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://proxy:5000"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
      endpoint = ["http://registry:5000"]
EOF
```

Note that we do use a port-forwarding proxy (`-p 5000:5000`) so that we can
push images "from the host", e.g.:

```sh
% docker tag alpine localhost:5000/alpine
% docker push localhost:5000/alpine
The push refers to repository [localhost:5000/alpine]
50644c29ef5a: Pushed
latest: digest: sha256:a15790640a6690aa1730c38cf0a440e2aa44aaca9b0e8931a9f2b0d7cc90fd65 size: 528

# Let's see if this image is also available from the cluster:
% docker exec -it kind-control-plane crictl pull localhost:5000/alpine
Image is up to date for sha256:a24bb4013296f61e89ba57005a7b3e52274d8edd3ae2077d04395f806b63d83e
```

If you use Tilt, you might also want to tell Tilt that it can use the local
registry. I find it a bit weird to have to set an annotation (hidden Tilt
API?) but whatever. If you set this:

```sh
kind get nodes | xargs -L1 -I% kubectl annotate node % tilt.dev/registry=localhost:5000 --overwrite
```

then Tilt [will use](legacy-annotation-based-registry-discovery) `docker
push localhost:5000/you-image` (from your host, not from the cluster
container) in order to speed up things. Note that there is a proposal ([KEP
1755](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry))
that aims at standardizing the discovery of local registries using a
configmap. Tilt already supports it, so you may use it!

## Improving the ClusterAPI docker provider to use a given network

When I play with ClusterAPI, I usually use the CAPD provider (ClusterAPI
Provider Docker). This provider [is kept
in-tree](https://github.com/kubernetes-sigs/cluster-api/blob/master/test/infrastructure/docker)
inside the cluster-api projet.

I want to use the caching mechanism presented above. But to do that, I need
to make sure the clusters created by CAPD are not created on the default
network ([current
implementation](https://sigs.k8s.io/cluster-api/test/infrastructure/docker/docker/kind_manager.go#L178)
creates CAPD clusters on the default "bridge" network).

I want to be able to customize the network on which the CAPD provider
creates the container that make up the cluster. Imagine that we could pass
the network name as part of a
[DockerMachineTemplate](https://github.com/kubernetes-sigs/cluster-api/blob/6821939410c37743b45c36ec91d94c37dba1998e/test/e2e/data/infrastructure-docker/cluster-template.yaml#L26-L35)
(the content of the `spec` is defined in code
[here](https://github.com/kubernetes-sigs/cluster-api/blob/2ac3728d26593f7c54520999477aad45934e1c59/test/infrastructure/docker/api/v1alpha3/dockermachine_types.go#L30-L55)):

```yaml
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: DockerMachineTemplate
metadata:
  name: capd-control-plane
  namespace: default
spec:
  template:
    spec:
      extraMounts:
        - containerPath: /var/run/docker.sock
          hostPath: /var/run/docker.sock

      network: kind        # 🔰 This field does not exist yet.
```

**Update 26 July 2020**: added a section about local registry vs. caching
proxy. Reworked the whole post (less noise, more useful information).
