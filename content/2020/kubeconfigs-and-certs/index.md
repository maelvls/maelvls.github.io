---
title: "Kubeconfigs and Certs"
description: ""
date: 2020-04-18T22:14:07+02:00
url: /kubeconfigs-and-certs
images: [kubeconfigs-and-certs/cover-kubeconfigs-and-certs.png]
draft: true
author: Maël Valais
devtoId: 0
devtoPublished: false
---

Let's go through the client-certificate in my kube config. This kube config contains one user, "kind-kind". Instead of inspecting the certificate PEM with openssl, I use [certigo](https://github.com/square/certigo) (much nicer experience!)

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

The common name (`CN=`) is "kubernetes-admin", that's the name of the user. The organization (`O=`) is "system:master".

> Quick recap of what a certificate is. For us, "certificate" means a [X.509-v3](https://tools.ietf.org/html/rfc5280#section-4.1)-formated binary chunk. This binary chunk is, most of the time, presented in the PEM format (the binary chunk is encoded in base 64 and wrapped with some decorations:
>
> ```
> -----BEGIN RSA PRIVATE KEY-----
> <the base64-encoded X.509 binary chunk>
> -----END RSA PRIVATE KEY-----
> ```
>
> This chunk of binary data contains three important parts:
>
> 1. the public key of the certificate,
> 2. a list of attributes, e.g. `O=Google LLC,CN=*.google.com`,
> 3. a signature of (1)+(2) created using the certificate authority's private key.
>
> The CA certificate uses its private key to sign the concatenation of the attributes with the public key:
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
> When someone connects to foo.google.com and wants to make sure this server can be trusted, they would use the CA certificates that are securely stored on their disk in order to verify the signature of the untrusted certificate that foo.google.com just sent them.
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

{{< youtube gXz4cq3PKdg >}}

---

## See the ACME HTTP-01 protocol using Pebble, Lego CLI and mitmproxy

First, let us create a CA key pair using [minica](https://github.com/jsha/minica).

```sh
go get -u github.com/jsha/minica
minica --domains me -ip-addresses 127.0.0.1,0.0.0.0
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain minica.pem
```

We are going to use the domain `me` instead of `locahost` since Go programs [are ignoring the proxy settings](https://maelvls.dev/go-ignores-proxy-localhost/) when the host of the request is `localhost`. Thus, we need to make sure we can resolve the `me` name by adding a line in `/etc/hosts`:

```sh
# /etc/hosts
127.0.0.1   me
```

Now, let us run [mitmproxy](https://mitmproxy.org/):

```sh
mitmproxy -p 9090 --set ssl_verify_upstream_trusted_ca=minica.pem
```

> The `minica.pem` has to be given to mitmproxy since it does not know about the system keychain (on macOS). Even if we added this root CA to the keychain and set it as trusted, mitmproxy uses its own certificate bundle (handled by [certifi](https://github.com/certifi/python-certifi)). So we have to give mitmproxy the root CA of the upstream server that mitmproxy will talk to (i.e., pebble).

At this point, let's make sure `lego` will trust mitmproxy's man-in-the-middle CA:

```sh
sudo security add-trusted-cert -d -r trustAsRoot -k /Library/Keychains/System.keychain ~/.mitmproxy/mitmproxy-ca-cert.pem
```

Then, let us run [pebble](https://github.com/letsencrypt/pebble), the simple ACME server meant for testing:

```sh
go get -u github.com/letsencrypt/pebble/cmd/pebble
HTTP_PROXY=localhost:9090 pebble -config /dev/stdin <<EOF
{
  "pebble": {
    "listenAddress": "0.0.0.0:14000",
    "managementListenAddress": "0.0.0.0:15000",
    "certificate": "me/cert.pem",
    "privateKey": "me/key.pem",
    "httpPort": 5002,
    "tlsPort": 5001,
    "externalAccountBindingRequired": false
  }
}
EOF
```

Finally, let us run the ACME client [lego](https://github.com/go-acme/lego):

```sh
go get -u github.com/go-acme/lego/v4/cmd/lego
lego -server=https://me:14000/dir -domains me -email foo@bar.org -http.port :5002 -http run
```

And there we go!

![Screeshot of mitmproxy running interactively in the terminal and capturing the input and output from Pebble and Lego](mitmproxy-with-lego-and-pebble.png)

If you would like to dig into these traces, you can do so by downloading the mitmproxy flow here:

```sh
curl -O {{< ref path="kubeconfigs-and-certs/" >}}pebble-and-lego.mitmproxy
mitmproxy -r pebble-and-lego.mitmproxy
```
