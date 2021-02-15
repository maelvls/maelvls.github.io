---
title: Remote Docker using docker-machine
date: 2018-09-13
tags: []
author: Maël Valais
devtoId: 365839
devtoPublished: false
---

## Use docker-machine for remoting docker

```shell
docker-machine create --driver generic --generic-ip-address=141.115.74.15 --generic-ssh-key ~/.ssh/id_rsa --generic-ssh-user=mvalais polatouche-docker-host

eval $(docker-machine env polatouche-docker-host)
```

## Tunneling

```sh
docker-machine ssh default -L 0.0.0.0:8000:localhost:8000
```

## Namespaces, cgroups

1. namespaces
2. cgroup (systemd-cgls)

## Commands

```sh
docker run --rm -it ubuntu
```

- `-i` means interactive (stdin will be attached)
- `-t` (`--tty`) means that this is a pseudo-tty terminal instead of a non-tty (for example, without tty, bash has no prompt; with a tty, it gives a nice prompt such as `bash-4.4#`).
