---
title: Nice diagram about TLS, certificates and keys
date: 2020-01-20
tags: []
author: Maël Valais
devtoSkip: true
---

<div class="nohighlight">

```plain
  ROLE OF THE CERTIFICATE AUTHORITY

      +------------------------------+  +-----------------+
     CERTIFICATE AUTHORITY (NOT A CERT) |     CA KEY      |
      |+----------------------------+|  |+---------------+|
      || O=GlobalSign,CN=GlobalSign ||  || private key   ||
      |+----------------------------+|  |+-------|-------+|
      |+----------------------------+|  +--------|--------+
      ||   public key (= VERIFY)    ||           |
      |+----------------------------+|           |
      +------------------------------+           |
                                                 |
 CA cert's pub key = signature                   |
 -----------------------------------------------------------
 Signed cert's pub key = encryption              |
                                                 |
      +------------------------------+           |
      |         CERTIFICATE          |           |
      |1----------------------------+|           |
      ||O=Google LLC,CN=*.google.com||           |
      |+----------------------------+|           |signs using
      |2----------------------------+|           |private key
      ||   public key (= DECRYPT)   ||           |
      |+----------------------------+|           |
      |+----------------------------+|           |
      ||      signature of 1+2      ||<----------+
      |+----------------------------+|
      +------------------------------+



   ROLE OF THE END-USER               +------------------------------+
                                      |   CA CERTIFICATE (trusted)   |
    Can I trust foo.google.com?       |+----------------------------+|
                |                     || O=GlobalSign,CN=GlobalSign ||
                |                     |+----------------------------+|
                |                     |+----------------------------+|
                |  +----------------->||    public key (= VERIFY)   ||
                |  |                  |+----------------------------+|
                |  |                  +------------------------------+
                |  |check signature   (this cert is secure on my disk)
                |  |using public key
                |  |
            (1) |  |(2)               +------------------------------+
                |  |                  |    CERTIFICATE (untrusted)   |
         verify |  |                  |1----------------------------+|
      signature |  |                  ||O=Google LLC,CN=*.google.com||
                |  |                  |+----------------------------+|
                |  |                  |2----------------------------+|
                |  |                  ||   public key (= ENCRYPT)   ||
                |  +------------------|+----------------------------+|
                |                     |+----------------------------+|
                +-------------------->||      signature of 1+2      ||
                                      |+----------------------------+|
                                      +------------------------------+
                                      (this cert is sent to me by the
                                                foo.google.com server)
```

</div>

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

That's the kind of directory structure that `--set client_dir=...` would expect (in mitmproxy).

## Using `openssl` to dig certificates

This [cheat sheet][cheatsheet] is nice. Here is a one-liner for fetching the certificate chain sitting on a server; for example for `api.github.com:443`:

```sh
% openssl x509 -in <(openssl s_client -connect api.github.com:443 -prexit <<< "") -noout -text

openssldepth=2 C = US, O = DigiCert Inc, OU = www.digicert.com, CN = DigiCert High Assurance EV Root CA
verify return:1
depth=1 C = US, O = DigiCert Inc, OU = www.digicert.com, CN = DigiCert SHA2 High Assurance Server CA
verify return:1
depth=0 C = US, ST = California, L = San Francisco, O = "GitHub, Inc.", CN = *.github.com
verify return:1
DONE
Certificate:
    Data:
        Version: 3 (0x2)
        Serial Number:
            0c:07:3b:67:6f:67:45:78:f9:99:81:48:52:84:46:51
    Signature Algorithm: sha256WithRSAEncryption
        Issuer: C=US, O=DigiCert Inc, OU=www.digicert.com, CN=DigiCert SHA2 High Assurance Server CA
        Validity
            Not Before: Jun 22 00:00:00 2020 GMT
            Not After : Aug 17 12:00:00 2022 GMT
        Subject: C=US, ST=California, L=San Francisco, O=GitHub, Inc., CN=*.github.com
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                Public-Key: (2048 bit)
                Modulus:
                    00:a0:e6:d3:87:ac:6f:4e:3b:29:75:60:4c:4a:1e:
                    fa:dd:af:81:0a:37:c4:89:ad:b5:8e:9d:1d:0c:55:
                    da:a4:b1:cb:ab:d3:14:bc:d2:6d:b8:d1:7c:36:1f:
                    45:a1:06:25:32:63:7f:94:4c:f4:d6:97:06:3f:24:
                    f2:85:f5:83:8c:27:8a:7f:6a:c8:46:e8:04:6f:4c:
                    5f:4c:48:9a:a6:80:c1:08:db:9c:6e:81:8b:54:59:
                    f0:c6:6d:58:2a:3d:42:ea:da:5d:aa:6b:90:7b:af:
                    12:34:30:1b:22:5c:af:cd:ee:f2:3c:08:90:99:91:
                    be:41:16:c6:e0:95:59:a9:d6:52:39:de:e9:a3:02:
                    e2:68:e3:f9:b5:56:ce:ae:62:27:5e:ff:a3:94:1f:
                    89:82:0f:5d:ea:82:4d:af:de:0f:3b:aa:04:4a:6f:
                    a4:85:43:80:11:35:f1:3b:d6:66:80:68:97:6e:0a:
                    e9:79:57:63:44:91:c1:e2:45:db:dd:2c:b9:2d:3d:
                    16:76:af:0b:a4:04:80:c0:10:35:26:f2:9f:38:43:
                    a8:1d:a5:19:79:f0:b0:60:98:40:f8:4c:54:53:5f:
                    32:0e:de:86:65:e7:4f:5c:21:7a:c7:35:7f:de:a3:
                    e0:8e:b2:d3:02:42:64:40:14:21:20:96:14:09:54:
                    1f:53
                Exponent: 65537 (0x10001)
        X509v3 extensions:
            X509v3 Authority Key Identifier:
                keyid:51:68:FF:90:AF:02:07:75:3C:CC:D9:65:64:62:A2:12:B8:59:72:3B

            X509v3 Subject Key Identifier:
                79:53:13:17:53:DA:1E:A9:73:A9:9E:88:8D:C2:53:D9:36:E2:E5:2A
            X509v3 Subject Alternative Name:
                DNS:*.github.com, DNS:github.com
            X509v3 Key Usage: critical
                Digital Signature, Key Encipherment
            X509v3 Extended Key Usage:
                TLS Web Server Authentication, TLS Web Client Authentication
            X509v3 CRL Distribution Points:

                Full Name:
                  URI:http://crl3.digicert.com/sha2-ha-server-g6.crl

                Full Name:
                  URI:http://crl4.digicert.com/sha2-ha-server-g6.crl

            X509v3 Certificate Policies:
                Policy: 2.16.840.1.114412.1.1
                  CPS: https://www.digicert.com/CPS
                Policy: 2.23.140.1.2.2

            Authority Information Access:
                OCSP - URI:http://ocsp.digicert.com
                CA Issuers - URI:http://cacerts.digicert.com/DigiCertSHA2HighAssuranceServerCA.crt

            X509v3 Basic Constraints: critical
                CA:FALSE
            1.3.6.1.4.1.11129.2.4.2:
                ...g.e.u.)y...99!.Vs.c.w..W}.`
..M]&\%].....r..*......F0D. f........
,cj..."~........M,..... U+.zza....WN9.V...A;U..?k.e%._.`.u.A...."FJ...:.B.^N1.....K.h..b......r..*......F0D. w.jXF...7K..v. %.A...;,.E.o}JN... 39..W..Z..Ff;.0.....u SZ
..&. ...u.F.U.u.. 0...i..}.,At..I.....p.mG...r..+6.....F0D. '....................O....Dmcpu@. A?W...o.d..o.!..e*....C(%......c
    Signature Algorithm: sha256WithRSAEncryption
         4e:0c:e9:56:59:75:8a:8c:af:75:d9:0e:57:58:fc:c0:8d:d9:
         35:67:84:61:a9:2c:0b:aa:ff:a6:62:31:31:53:6c:d6:50:44:
         1d:7e:0a:11:ec:40:00:d9:f4:d2:e0:f7:a3:f7:32:91:1b:ac:
         d0:4c:01:1e:03:f8:19:8b:1e:79:33:dd:38:c6:a7:f4:4b:3b:
         e9:31:11:b9:c6:a7:63:40:d6:02:2b:43:59:bb:98:4f:08:24:
         78:a0:67:1d:0e:35:bb:40:82:85:e0:d5:7e:76:7e:e4:94:d3:
         4e:a6:d1:31:c0:41:d2:b0:9a:b7:77:c9:36:b9:31:69:95:87:
         b0:50:d9:ef:e5:37:80:a9:f2:dc:0d:a0:82:c8:56:2e:af:85:
         1b:ef:0d:c2:eb:1c:33:d5:76:ae:cb:9d:f1:6e:b4:00:85:75:
         e2:b9:4c:3b:7f:16:41:47:a9:53:b0:c4:b1:7a:44:da:f6:5b:
         84:45:29:13:37:c3:aa:30:98:d2:67:8f:fb:1a:be:94:06:0e:
         b9:cf:e6:fc:5f:fc:e2:f1:1b:32:6a:8e:c2:16:a7:cd:45:c1:
         fe:75:8b:02:b2:f0:d9:3b:5d:7c:1b:a8:45:b8:b1:4d:d8:3c:
         66:2b:2b:33:1e:9f:dc:b8:5e:3b:fe:0a:e4:b2:5b:eb:41:48:
         59:10:6b:63
```

[cheatsheet]: https://www.sslshopper.com/article-most-common-openssl-commands.html

## Fun things about X.509

In [crypto/x509/verify.go](https://github.com/golang/go/blob/e6ac2df2/src/crypto/x509/verify.go#L700-L715):

```go
// KeyUsage status flags are ignored. From Engineering Security, Peter
// Gutmann: A European government CA marked its signing certificates as
// being valid for encryption only, but no-one noticed. Another
// European CA marked its signature keys as not being valid for
// signatures. A different CA marked its own trusted root certificate
// as being invalid for certificate signing. Another national CA
// distributed a certificate to be used to encrypt data for the
// country’s tax authority that was marked as only being usable for
// digital signatures but not for encryption. Yet another CA reversed
// the order of the bit flags in the keyUsage due to confusion over
// encoding endianness, essentially setting a random keyUsage in
// certificates that it issued. Another CA created a self-invalidating
// certificate by adding a certificate policy statement stipulating
// that the certificate had to be used strictly as specified in the
// keyUsage, and a keyUsage containing a flag indicating that the RSA
// encryption key could only be used for Diffie-Hellman key agreement.
```

## Encoding vs. format

Source: [James Roper, 2019](https://github.com/jetstack/cert-manager/issues/843#issuecomment-566054175)

There are multiple axes of encodings standards, there's PEM and DER, and there's
PKCS*. So you can have PEM encoded PKCS1, DER encoded PKCS8, PEM encoded PKCS8
etc. PKCS* is more commonly referred to as the format (in some places its called
the syntax), while PEM and DER are the ascii and binary encodings of whichever
format you've chosen, and so are more commonly referred to as the encoding. That
said, openssl just calls them all formats. It's really a mess.

DER = pure ASN.1, which is a binary format.
PEM = ASN.1 base64-encoded.