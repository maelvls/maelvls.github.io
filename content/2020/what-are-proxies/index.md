---
title: "What Are Proxies"
description: ""
date: 2020-07-13T14:18:40+02:00
url: /what-are-proxies
images: [what-are-proxies/cover-what-are-proxies.png]
draft: true
tags: []
author: Maël Valais
devtoId: 0
devtoPublished: false
---

A proxy is something (most of the time software, but sometimes also hardware like a dedicated FPGA) that forwards traffic from a place to another.

Most of the time, when someone talks about a "proxy", they mean an HTTP proxy that would allow this person to access the internet behind a corporate firewall.

| Layer | Proxies                                                      |
| ----- | ------------------------------------------------------------ |
| L7    | plain HTTP, HTTP `CONNECT`, SOCKS, SOCKS5, HTTPS/SNI proxies |
| L4    | TCP proxies                                                  |
| L3    | VPNs                                                         |

Terms:

- tunnel
- transparent proxy = doesn't temper with packets; only redirects them
- VPN
- forwarding
- reverse proxy

## Layer 7: HTTP CONNECT (TCP tunnel)

```http
CONNECT streamline.t-mobile.com:22 HTTP/1.1
Proxy-Authorization: Basic encoded-credentials
```

## Layer 3: the VPN

A VPN often acts at L3. For example, Telepresence sets a packet filter on BSD (or iptable rule on Linux) which redirects all packets to a given TCP proxy which then forwards traffic to into the pod.

Some VPNs have to use a higher protocol in order to work around network limitations (e.g., corporate firewall or country-wide censorship), such as OpenVPN which opens a TLS connection (L7) and then use the ciphered TCP connection to forward all TCP traffic; that's L3-into-L7 tunneling (watch out for the MTU!)

## Layer 4: the TCP proxy

I am particularly interested in two different kinds of TCP proxies.

## Layer 7: the standard HTTP proxy

There are many available HTTP proxies like Squid, tinyproxy, cntlm and mitmproxy. Let's try mitmproxy and handcraft an HTTP request:

```sh
# Open that in one shell session:
mitmproxy -p 9000

# And this in another shell session:
nc 127.0.0.1 9000 <<< $'\
GET http://httpbin.org/ HTTP/1.1\r
Host: httpbin.org\r
\r
'
```

<details>

<summary>Note on escaping <code>\r\n</code> in your shell</summary>

> The tricky part when crafting an HTTP request for `nc` in the shell is the backslash-escaped charaters expansion. Using a here-document does not work since backslash-escaped characters like `\r` and `\n` are not replaced with their actual value as per the ANSI C standard ([ISO/IEC > 9899:201x][] § 5.2.2 Character display semantics, p. 23)
>
> I use `$'text'` which is a feature of bash that replaces `\` sequences with their byte representation. Another possibility is to use `printf` or `echo -e`:
>
> ```sh
> printf 'GET http://httpbin.org/ HTTP/1.1\r
> Host: httpbin.org\r
> \r
> ' | nc 127.0.0.1 9000
> ```

[iso/iec 9899:201x]: http://www.open-std.org/JTC1/SC22/WG14/www/docs/n1539.pdf#page=42

</details>

The HTTP proxy uses the exact same protocol as HTTP; you can pass some special headers like `Proxy-Authorization: Basic ...` which are not going to be forwarded.

## Layer 7: the HTTPS + SNI proxy

tbd

## Layer 7: the reverse HTTP proxy

## Examples of proxies

By remote dialer, I mean something that exposes a port on a local machine, and everything that goes through this port is forwarded to a distant host; the distant host "dials" the destination host (we could say "proxies" the TCP connection, but I like the "dial" word 😅)

<!-- https://textik.com/#216b639e0aaa6953 -->

```plain
                   local host                    remote host
                      |                               |
 PHASE 1:             |                               |
 tunnel setup         |   create tcp connection       |
                      |   (e.g. ssh port forwarding,  |
                      |   websockets...)              |
                      |------------------------------>|
                      |                               |
                      |   tcp connection established  |
                      |<----------------------------->|
                      |                               |
                      |                               |
                      |    configure 80 -> 8000       |
                      |-----------------------------> |
                      |                        +-------------+
                      |                        |listen on 80 |
                      |                        +-------------+
                      |                               |
                      |                               |
                      |                               |
 PHASE 2:             |                               |
 remote port          |                               |  remote:80
 forwarding           |                               |<---------
                      |      forwards tcp packets     |
                      |<------------------------------|
 dials localhost:8000 |                               |
       <--------------|                               |
                      |                               |
                      |                               |
```

And by reverse remote dialer,

By tunnel we mean one local port that a process binds to and every TCP opened to that port gets forwarded to a remote dialer.

TCP tunnel vs. TCP proxy vs. reverse proxy = they are all the same in the end? Except for how things are captured on the local host (simple port binding) and to where the thing is forwarded to.

[Rpivot](https://github.com/klsecservices/rpivot) is a python tool used for penetration testing. It acts as a "remote dialer":

## What is socat

[socat](https://book.hacktricks.xyz/tunneling-and-port-forwarding#socat)

From [rancher/k3s](https://github.com/rancher/k3s/blob/fe7337937155af41f1aebeb87d1acd07091b71de/scripts/provision/generic/alpine310/vagrant#L25):

```sh
docker run -d -v /var/run/docker.sock:/var/run/docker.sock -p 127.0.0.1:2375:2375 alpine/socat TCP-LISTEN:2375,fork UNIX-CONNECT:/var/run/docker.sock
```

Probably a massive workaround for exposing `docker.sock` on `localhost:2375` when the only thing you have in hand is docker & you can’t run `socat` locally (or don’t want to bother installing it lol).

## Reverse/remote TCP connection (remote dialer)

Let's expose port 80 on the remote and that will forward traffic to localhost:8080 on my local machine:

```sh
# From local
ssh -R 42345:22 remote -p 22
ssh -R 80:8080 localhost -p 42345      # the reverse port-forwarding
```

Diagram of that:

```plain
      REMOTE TCP TUNNEL
      Forwards TCP connections from the remote to the
      local host, but the initial TCP connection is
      made from local to remote.

           local host             remote host
                |                       |
                |                       |
                |    ssh port-forward   |
                |---------------------->|
                |                       |
                |    ssh port-forward   |
                |<------(80:8080)-------|
                |                       |
                |                 +-----|------+
                |                 |listen on 80|
                |                 +-----|------+
                |                       |
                |                       |    host:80
                |   dial backend:8080   |<----------
          dials |<----------------------|
 fixedhost:8080 |                       |
 <--------------|                       |
```
