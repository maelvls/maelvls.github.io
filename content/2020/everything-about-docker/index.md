---
title: "Everything About Docker"
description: ""
date: 2020-04-30T08:59:06+02:00
url: /everything-about-docker
images: [everything-about-docker/cover-everything-about-docker.png]
draft: true
tags: []
author: Maël Valais
---

- Linux kernel 3.10 or higher required
-


```sh
% cat /etc/containerd/config.toml
# explicitly use v2 config format
version = 2

# set default runtime handler to v2, which has a per-pod shim
[plugins."io.containerd.grpc.v1.cri".containerd]
default_runtime_name = "runc"
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
runtime_type = "io.containerd.runc.v2"

# Setup a runtime with the magic name ("test-handler") used for Kubernetes
# runtime class tests ...
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.test-handler]
runtime_type = "io.containerd.runc.v2"

# ensure the sandbox image matches kubeadm
# TODO: probably we should instead just use the containerd default
# Implementing the pod sandbox is a CRI implementation detail ..
[plugins."io.containerd.grpc.v1.cri"]
sandbox_image = "k8s.gcr.io/pause:3.2"
```

```sh
% cat /etc/crictl.yaml
runtime-endpoint: unix:///var/run/containerd/containerd.sock
```

## Containerd vs. RunC

As detailed [here](https://stackoverflow.com/questions/41645665/how-containerd-compares-to-runc#:~:text=runc%20is%20used%20by%20containerd,specification%20for%20runtime%20and%20images.):

> - [`containerd`](https://github.com/containerd/containerd) is a container
> runtime which can manage a complete container lifecycle -- from image
> transfer/storage (locally and from/to registries) to container
> execution, supervision and networking. Containerd abides by the
> client-side of the [OCI Distribution
> spec](https://github.com/opencontainers/distribution-spec/blob/master/spec.md).
> - [`containerd-shim`](http://alexander.holbreich.org/docker-components-explained#containerdshim)
> handles headless containers, meaning once `runc` initializes the
> containers, it exits handing the containers over to the container-shim
> which acts as some middleman.
> - [`runc`](https://github.com/opencontainers/runc) is a lightweight
> universal container runtime, which abides by the OCI specification.
> runc is used by containerd for spawning and running containers
> according to OCI spec. It is also the repackaging of Docker's
> libcontainer.
