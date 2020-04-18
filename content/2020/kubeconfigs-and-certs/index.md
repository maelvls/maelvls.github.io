---
title: "Kubeconfigs and Certs"
description: ""
date: 2020-04-18T22:14:07+02:00
url: /kubeconfigs-and-certs
images: [kubeconfigs-and-certs/cover-kubeconfigs-and-certs.png]
draft: true
---

Let's go through the client-certificate in my kube config. This kube config
contains one user, "kind-kind". Instead of inspecting the certificate PEM
with openssl, I use [certigo](https://github.com/square/certigo) (much
nicer experience!)

```sh
kubectl config view -ojson --flatten \
  | jq '.users[] | .user."client-certificate-data" | select(.!=null)' -r \
  | base64 -d \
  | certigo dump
```

That gives us:

```plain
** CERTIFICATE 1 **
Valid: 2020-03-03 12:31 UTC to 2021-03-03 12:31 UTC
Subject:
    O=system:masters, CN=kubernetes-admin
Issuer:
    CN=kubernetes
Warnings:
    Certificate doesn't have any valid DNS/URI names or IP addresses set
```

The common name (`CN=`) is "kubernetes-admin", that's the name of the user.
The organization (`O=`) is "system:master".

> Quick recap of what a certificate is. For us, "certificate" means a
> [X.509-v3](https://tools.ietf.org/html/rfc5280#section-4.1)-formated
> binary chunk. This binary chunk is, most of the time, presented in the
> PEM format (the binary chunk is encoded in base 64 and wrapped with some
> decorations:
>
> ```
> -----BEGIN RSA PRIVATE KEY-----
> <the base64-encoded X.509 binary chunk>
> -----END RSA PRIVATE KEY-----
> ```
>
> This chunk of binary data contains three important parts:
>
> 1) the public key of the certificate,
> 2) a list of attributes, e.g. `O=Google LLC,CN=*.google.com`,
> 3) a signature of (1)+(2) created using the certificate authority's
>    private key.
>
> The CA certificate uses its private key to sign the concatenation of the
> attributes with the public key:
>
> <div class="nohighlight">
>
> ```plain
>  +------------------------------+  +-------------------------+
>  |        CA CERTIFICATE        |  |         CA KEY          |
>  |+----------------------------+|  |+-----------------------+|
>  || O=GlobalSign,CN=GlobalSign ||  ||      private key      ||
>  |+----------------------------+|  |+-----------------------+|
>  |+----------------------------+|  +-------------------------+
>  ||         public key         ||               |
>  |+----------------------------+|               |
>  +------------------------------+               |
>                                                 |
>                                                 |
>                                                 |
>  +------------------------------+               |signs using
>  |         CERTIFICATE          |               |private key
>  |1----------------------------+|               |
>  ||O=Google LLC,CN=*.google.com||               |
>  |+----------------------------+|               |
>  |2----------------------------+|               |
>  ||         public key         ||               |
>  |+----------------------------+|               |
>  |+----------------------------+|               |
>  ||      signature of 1+2      ||<--------------+
>  |+----------------------------+|
>  +------------------------------+
> ```
>
> </div>
>
> When someone connects to foo.google.com and wants to make sure this
> server can be trusted, they would use the CA certificates that are
> securely stored on their disk in order to verify the signature of the
> untrusted certificate that foo.google.com just sent them.
>
> <div class="nohighlight">
>
> ```plain
>                                   +------------------------------+
>                                   |   CA CERTIFICATE (trusted)   |
> Can I trust foo.google.com?       |+----------------------------+|
>             |                     || O=GlobalSign,CN=GlobalSign ||
>             |                     |+----------------------------+|
>             |                     |+----------------------------+|
>             |  +----------------->||         public key         ||
>             |  |                  |+----------------------------+|
>             |  |                  +------------------------------+
>             |  |check signature   (this cert is secure on my disk)
>             |  |using public key
>             |  |
>         (1) |  |(2)               +------------------------------+
>             |  |                  |    CERTIFICATE (untrusted)   |
>      verify |  |                  |1----------------------------+|
>   signature |  |                  ||O=Google LLC,CN=*.google.com||
>             |  |                  |+----------------------------+|
>             |  |                  |2----------------------------+|
>             |  |                  ||         public key         ||
>             |  +------------------|+----------------------------+|
>             |                     |+----------------------------+|
>             +-------------------->||      signature of 1+2      ||
>                                   |+----------------------------+|
>                                   +------------------------------+
>                                   (this cert is sent to me by the
>                                             foo.google.com server)
> ```
>
> </div>
<!--
https://textik.com/#d85b4624473ca862
-->

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
