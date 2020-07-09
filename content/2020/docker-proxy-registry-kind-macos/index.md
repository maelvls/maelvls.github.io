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
  means the "registry" hostname won't be resolved (but we could work around
  that).

[docker-provider]: https://github.com/kubernetes-sigs/cluster-api/tree/master/test/infrastructure/docker

---

[Kind](https://kind.sigs.k8s.io/) is an awesome tool that allows you to
spin up local Kubernetes clusters locally in seconds. It is perfect for
Kubernetes developers or anyone who wants to play with controllers.

One thing I hate about Kind is that images are not cached between two Kind
containers. Even worse: when deleting and re-creating a cluster, all the
downloaded images disappear.

And since I install ClusterAPI on my Kind cluster, it means all the (quite
heavy) images have to be re-downloaded every single time. Just take a look
at all the images that get re-downloaded:

```sh
% docker exec -it helix-control-plane crictl images
IMAGE                                                                      TAG      SIZE
quay.io/jetstack/cert-manager-cainjector                                   v0.11.0  11.1MB
quay.io/jetstack/cert-manager-controller                                   v0.11.0  14MB
quay.io/jetstack/cert-manager-webhook                                      v0.11.0  14.3MB
us.gcr.io/k8s-staging-capi-docker/capd-manager/capd-manager-amd64          dev      53.5MB
us.gcr.io/k8s-artifacts-prod/cluster-api/cluster-api-controller            v0.3.0   20.3MB
us.gcr.io/k8s-artifacts-prod/cluster-api/kubeadm-bootstrap-controller      v0.3.0   19.6MB
us.gcr.io/k8s-artifacts-prod/cluster-api/kubeadm-control-plane-controller  v0.3.0   21.1MB
```

One solution to this problem is to [spin up an intermediary Docker
registry](https://kind.sigs.k8s.io/docs/user/local-registry/) in a side
container; as long as this container exists, all the images that have
already been downloaded once can be served from cache. Let's spin up this
registry:

```sh
docker run -d --net=other --restart=always --name registry registry:2
```

Now, let's tell `kind` that we want the created node to be using this
cache:

```sh
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry"]
      endpoint = ["http://registry:5000"]
EOF
```

And... it doesn't work! The first thing to notice is that the registry runs
on the default network (named [`bridge`][default-bridge]) but my cluster
was created with a separate network (named `kind`):

```sh
% docker inspect registry --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
bridge 172.17.0.2

% docker inspect kind-control-plane --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
kind 172.18.0.2

% docker network ls
NETWORK ID          NAME                DRIVER              SCOPE
a6ceea984c68        bridge              bridge              local
6f6a9618d746        host                host                local
4927dc2eba9b        kind                bridge              local
```

My first attempt was to move the kind cluster to the default network (why
does it use another network anyway??) and kind has an experimental option
for that:

```sh
% KIND_EXPERIMENTAL_DOCKER_NETWORK=bridge kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry"]
      endpoint = ["http://registry:5000"]
EOF

Creating cluster "kind" ...
WARNING: Overriding docker network due to KIND_EXPERIMENTAL_DOCKER_NETWORK
WARNING: Here be dragons! This is not supported currently.
 ✓ Ensuring node image (kindest/node:v1.18.2) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
```

Perfect! My local cluster is now on the same network as my tiny registry
cache:

```sh
% docker inspect registry --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
bridge 172.17.0.2

% docker inspect kind-control-plane --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
bridge 172.17.0.6
```

But... it doesn't seem to work... The registry is still empty, no cached
blobs:

```sh
% docker exec -it registry ls /var/lib/registry
# Nothing!
```

After digging a bit more, I realized that Kind had chosen to create it own
network due to the fact that the default "bridge" has limitations and
[doesn't allow you][dns-services] to use container names as DNS names:

> By default, a container inherits the DNS settings of the host, as defined
> in the `/etc/resolv.conf` configuration file. Containers that use the
> default bridge network get a copy of this file, whereas containers that
> use a custom network use Docker’s embedded DNS server, which forwards
> external DNS lookups to the DNS servers configured on the host.

[default-bridge]: https://docs.docker.com/network/bridge/#use-the-default-bridge-network
[dns-services]: https://docs.docker.com/config/containers/container-networking/#dns-services

Let's try to reproduce this issue. I first run a registry with the default
network, and I then try to connect to it from a second container using the
hostname `registry`:

```sh
% docker run -d --rm --name registry registry:2
% docker run -it --rm alpine nslookup registry
Server:         127.0.0.11
Address:        127.0.0.11:53
Non-authoritative answer:
Name:   registry
Address: 172.18.0.7

# Let's cleanup:
% docker kill registry
```

Now, let's do the same but a custom network named "other" instead of the
default network:

```sh
% docker network create other
% docker run -d --rm --net=other --name registry registry:2
% docker run -it --rm --net=other alpine nslookup registry
Server:         127.0.0.11
Address:        127.0.0.11:53
Non-authoritative answer:
Name:   registry
Address: 172.18.0.7

# Let's cleanup:
% docker kill registry
```

I thought I could first start the container on the default network and then
move the containers to the "other" network (so that DNS with container
names works) but it does not seem to work either:

```sh
% docker run -d --rm --name registry registry:2
% docker run -d --rm --name alpine alpine sleep 1d

% docker network create other
% docker network disconnect bridge registry
% docker network disconnect bridge alpine
% docker network connect other registry
% docker network connect other alpine

# Now, let's see if 'alpine' can resolve 'registry':
% docker exec -it alpine nslookup registry
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
docker run -d --name registry --restart=always --net=kind registry:2
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry"]
      endpoint = ["http://registry:5000"]
EOF
```

Both are now on the same subnet:

```sh
% docker inspect registry --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
kind 172.18.0.2

% docker inspect kind-control-plane --format '{{range $net, $cfg := .NetworkSettings.Networks}}{{$net}} {{$cfg.IPAddress}}{{end}}'
kind 172.18.0.3
```

But creating a simple deployment doesn't seem to work! By looking at the
registry logs, we see no activity at all. Let's see what `containerd` is up
to:

```log
% docker exec -i kind-control-plane journalctl -u containerd | grep 'registry:5000'
containerd[129]: Start cri plugin with config {Registry:{Mirrors:map[
  docker.io: {Endpoints:[https://registry-1.docker.io]}
  registry:  {Endpoints:[http://registry:5000]}
]}}
```

> ✅ I had to remove `-t` (tty) from the above command. That's because
> `journalctl` was using a pager because this terminal was TTY (which means
> it had /dev/stdin open? not sure). To disable the pager I removed `-t`.

I guess contained picks the first mirror (`docker.io`); so let's override
the `docker.io` key:

```sh
docker rm -f registry
docker run -d --name registry --restart=always --net=kind registry:2
kind create cluster --config /dev/stdin <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://registry:5000"]
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
registry"]
          endpoint = ["http://registry:5000"]
```

We can see the registry serving images by creating a deployment:

```sh
% kubectl create deployment test --image alpine
% docker logs registry | tail
172.18.0.3 - - [03/Jul/2020:12:18:21 +0000] "HEAD /v2/library/alpine/manifests/latest HTTP/1.1" 404 96 "" "containerd/v1.3.3-14-g449e9269"
172.18.0.3 - - [03/Jul/2020:12:18:50 +0000] "HEAD /v2/library/alpine/manifests/latest HTTP/1.1" 404 96 "" "containerd/v1.3.3-14-g449e9269"
172.18.0.3 - - [03/Jul/2020:12:19:40 +0000] "HEAD /v2/library/alpine/manifests/latest HTTP/1.1" 404 96 "" "containerd/v1.3.3-14-g449e9269"
```

Nope, not yet... All the GET requests are answered with a `404 Not Found`.
It's probably because I forgot to enable [the pull-through
feature](https://docs.docker.com/registry/configuration/#proxy) which turns
the registry into an "image proxy":

```sh
docker rm -f registry
docker run -d --name registry --restart=always --net=kind -e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io registry:2
```

And... It works!!

```sh
% docker logs registry | tail
172.18.0.3 - - [03/Jul/2020:12:51:24 +0000] "HEAD /v2/library/alpine/manifests/latest HTTP/1.1" 200 1638 "" "containerd/v1.3.3-14-g449e9269"
time="2020-07-03T12:51:38.728711481Z" level=info msg="response completed" go.version=go1.11.2 http.request.host="registry:5000" http.request.id=89341304-f38b-4ce3-9364-429709311783 http.request.method=HEAD http.request.remoteaddr="172.18.0.3:54188" http.request.uri="/v2/library/alpine/manifests/latest" http.request.useragent="containerd/v1.3.3-14-g449e9269" http.response.contenttype="application/vnd.docker.distribution.manifest.list.v2+json" http.response.duration=271.043848ms http.response.status=200 http.response.written=1638
172.18.0.3 - - [03/Jul/2020:12:51:38 +0000] "HEAD /v2/library/alpine/manifests/latest HTTP/1.1" 200 1638 "" "containerd/v1.3.3-14-g449e9269"
```

> Note: the [Docker
> docs](https://docs.docker.com/registry/configuration/#proxy) indicates
> YAML fields; to pass the same configuration as env variables, you can use
> `REGISTRY_` and then the path to the field you want. For example,
>
> ```yaml
> proxy:
>   remoteurl: https://registry-1.docker.io
> ```
>
> becomes
>
> ```sh
> REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io
> ```

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