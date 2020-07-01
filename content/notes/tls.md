---
title: Nice diagram about TLS, certificates and keys
date: 2020-01-20
tags: []
author: Maël Valais
---

```plain
 cert = (1) pub key, (2) information (3) signature of (1)+(2)
 self-signed cert = only (1) and (2)

   +--------------------------+ +-------------------------+
   |      CA CERTIFICATE      | |       ca cert key       |
   |+------------------------+| |+-----------------------+|
   ||      information       || ||      private key      ||
   |+------------------------+| |+-----------------------+|
   |+------------------------+| +-------------------------+
   ||       public key       ||               |
   ||                        ||               |
   |+------------------------+|               |
   +--------------------------+               |
                                              |
                                              |
   +--------------------------+               |
   |       CERTIFICATE        |               |
   |+------------------------+|               |
   ||1     information       ||               |
   |+------------------------+|               |
   |+------------------------+|               |
   ||2      public key       ||               |
   ||                        ||               |
   |+------------------------+|               |
   |+------------------------+|               |
   ||    signature of 1+2    ||<--------------+
   |+------------------------+|
   +--------------------------+
```

<!--
https://textik.com/#d85b4624473ca862
-->

## What's a "certificate" in the openssl jargon

For example, mitmproxy relies on a specific dir structure and filenames:

```sh
/Users/mvalais/.mitmproxy
├── mitmproxy-ca-cert.pem            # ca-cert = cert
└── mitmproxy-ca.pem                 # ca      = cert + key
```

That's the kind of directory structure that `--set client_dir=...` would
expect (in mitmproxy).
