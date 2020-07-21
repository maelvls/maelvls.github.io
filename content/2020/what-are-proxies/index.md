---
title: "What Are Proxies"
description: ""
date: 2020-07-13T14:18:40+02:00
url: /what-are-proxies
images: [what-are-proxies/cover-what-are-proxies.png]
draft: true
tags: []
author: Maël Valais
---

A proxy is something (most of the time software, but sometimes also
hardware like a dedicated FPGA) that forwards traffic from a place to
another.

Most of the time, when someone talks about a "proxy", they mean an HTTP
proxy that would allow this person to access the internet behind a
corporate firewall.

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

A VPN often acts at L3. For example, Telepresence sets a packet filter on
BSD (or iptable rule on Linux) which redirects all packets to a given TCP
proxy which then forwards traffic to into the pod.

Some VPNs have to use a higher protocol in order to work around network
limitations (e.g., corporate firewall or country-wide censorship), such as
OpenVPN which opens a TLS connection (L7) and then use the ciphered TCP
connection to forward all TCP traffic; that's L3-into-L7 tunneling (watch
out for the MTU!)

## Layer 4: the TCP proxy

I am particularly interested in two different kinds of TCP proxies.

## Layer 7: the standard HTTP proxy

There are many available HTTP proxies like Squid, tinyproxy, cntlm and
mitmproxy. Let's try mitmproxy and handcraft an HTTP request:

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

> The tricky part when crafting an HTTP request for `nc` in the shell is
> the backslash-escaped charaters expansion. Using a here-document does not
> work since backslash-escaped characters like `\r` and `\n` are not
> replaced with their actual value as per the ANSI C standard ([ISO/IEC
> 9899:201x][] § 5.2.2 Character display semantics, p. 23)
>
> I use `$'text'` which is a feature of bash that replaces `\` sequences
> with their byte representation. Another possibility is to use `printf` or
> `echo -e`:
>
> ```sh
> printf 'GET http://httpbin.org/ HTTP/1.1\r
> Host: httpbin.org\r
> \r
> ' | nc 127.0.0.1 9000

[ISO/IEC 9899:201x]: http://www.open-std.org/JTC1/SC22/WG14/www/docs/n1539.pdf#page=42

</details>

The HTTP proxy uses the exact same protocol as HTTP; you can pass some
special headers like `Proxy-Authorization: Basic ...` which are not going
to be forwarded.

## Layer 7: the HTTPS + SNI proxy

tbd

## Layer 7: the reverse HTTP proxy

