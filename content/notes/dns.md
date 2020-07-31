---
title: It's always the DNS' fault
date: 2020-07-25
tags: [networking]
author: Maël Valais
---

## Terms

A domain name (or just "domain") is a string the form `bar.foo.com.`. Not
all domains refer to physical machines; for example, one of my domains,
`k.maelvls.dev.`, do not point to any physical machine.

We often represent the domain name space using a tree. Each node is a
domain. Leaves and nodes may have `A` records. Here is a simple example of
the domain space represented as a tree:

```sh
.
├── com.
├── dev.
│  └── maelvls.dev.
│     └── k.maelvls.dev.
└── io.
```

A [zone][rfc1034-terms] is a subtree of the domain space that is under the
authority of a given name server. In the following, I will use "subtree"
and "zone" interchangeably and I identify a zone (or subtree) by its apex
domain; the apex is domain name at the root of a zone.

> I decided to put domain names directly in each node; note that the [RFC
> 1034][rfc1034-space] represents the domain space as a tree of labels (a
> label is of the form `foo` or `foo-bar`). In this tree, the domain name
> of a node has to be reconstructed by concatenating the labels starting
> from the node and ending with the root of the domain space tree.

From the above example, my zone is:

```plain
maelvls.dev.
└── k.maelvls.dev.
```

A zone has "authority" over its subtree as long as one of its parent
domains has an `NS` record to a name server that has the records for that
zone. I told my registrar (Google Domains) to use the Google DNS name
servers, which means that the `dev.` name servers now have `NS` records
that point to Google DNS' name servers. With this `NS` record, the `dev.`
zone "delegates" the `maelvls.dev.` zone to Google DNS' name servers.

For example, let us use `dig` to see all intermediate DNS queries at once:

```sh
% dig +trace minio.k.maelvls.dev
# I omitted the DNSSEC-related records (RRSIG, DS and NSEC3 records).
# I also omitted some NS records when there was too many of them.
.                              342831  IN  NS     a.root-servers.net.
.                              342831  IN  NS     b.root-servers.net.
dev.                           172800  IN  NS     ns-tld1.charlestonroadregistry.com.
dev.                           172800  IN  NS     ns-tld2.charlestonroadregistry.com.
maelvls.dev.                   10800   IN  NS     ns-cloud-a1.googledomains.com.
maelvls.dev.                   10800   IN  NS     ns-cloud-a2.googledomains.com.
maelvls.dev.                   300     IN  SOA   ns-cloud-a1.googledomains.com. ...
minio.k.maelvls.dev.           300     IN  A       91.211.152.190
```

## Client-side name guessing

In some cases, the DNS query may omit the right-most part of the domain.
For example, in Kubernetes, you may either use the fully-qualified domain
name (FQDN, which is a domain that contains the `.` root domain), or just
use a part of the domain name:

```sh
nslookup kubernetes.default.svc.cluster.local    # works (FQDN)
nslookup kubernetes.default.svc.cluster          # doesn't work
nslookup kubernetes.default.svc                  # works
nslookup kubernetes.default                      # works
nslookup kubernetes                              # works (guesses namespace)
```

In CoreDNS, you can give the apex domain name (by default, in Kubernetes,
the apex is `cluster.local.`). The Corefile needs to load the [kubernetes
plugin][kubernetes-plugin]:

```sh
.:53 {
    kubernetes cluster.local
}
```

> Note that `kubernetes.default` is not a subdomain; a [subdomain][rfc7719]
> of a domain is a domain that contains the domain itself on it right-most
> side.
>
> ```plain
> minio.k.maelvls.dev   is subdomain of     k.maelvls.dev
> minio.k.maelvls.dev   is subdomain of       maelvls.dev
> k.maelvls.dev         is subdomain of       maelvls.dev
> ```

Whenever a container is given to resolve any kind of name (even one that
looks like an FQDN; the client doesn't known, really), it will go through
the "search domain names" list configured in `/etc/resolv.conf`:

```sh
% cat /etc/resolv.conf
search default.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.96.0.10
options ndots:5
```

Let's try with `kubernetes.default`:

```sh
% tcpdump "udp port 53"
% nslookup kubernetes.default
12:01:17.430931 IP foo.39435 > kube-dns.kube-system.svc.cluster.local.53: 32862+ A? kubernetes.default.default.svc.cluster.local. (62)
12:01:17.431573 IP kube-dns.kube-system.svc.cluster.local.53 > foo.39435: 32862 NXDomain*- 0/1/0 (155)
12:01:17.431987 IP foo.35023 > kube-dns.kube-system.svc.cluster.local.53: 39314+ A? kubernetes.default.svc.cluster.local. (54)
12:01:17.433258 IP kube-dns.kube-system.svc.cluster.local.53 > foo.35023: 39314*- 1/0/0 A 10.96.0.1 (106)
Server:		10.96.0.10
Address:	10.96.0.10#53

Name:	kubernetes.default.svc.cluster.local
Address: 10.96.0.1
```

The client (the container `foo`) has to do two consecutive queries in order
to get an answer; as you can see, the way partial domain name guessing is
done is quite dumb and relies on many client-side queries:

| queried name                         | queries count |
| ------------------------------------ | ------------- |
| kubernetes                           | 1             |
| kubernetes.default                   | 2             |
| kubernetes.default.svc               | 3             |
| kubernetes.default.svc.cluster       | (*)           |
| kubernetes.default.svc.cluster.local | 4             |
| foo (the container's name)           | 4             |
| google.com                           | 4             |

> (*) since the search domain name `cluster` doesn't exist in
> `/etc/resolv.conf`, this name can't be resolved.

I noticed something unexpected when querying `foo` (which is the
container's name). I was expecting the name to be picked from `/etc/hosts`:

```sh
% cat /etc/hosts
# Kubernetes-managed hosts file.
127.0.0.1	localhost
::1	localhost ip6-localhost ip6-loopback
fe00::0	ip6-localnet
fe00::0	ip6-mcastprefix
fe00::1	ip6-allnodes
fe00::2	ip6-allrouters
10.244.0.14	foo                     # this!
```

But weirdly enough, `foo` is resolved by the name server:

```sh
% hostname
foo
% tcpdump "udp port 53"
% nslookup foo
10:35:09.052394 IP foo.58416 > kube-dns.kube-system.svc.cluster.local.53: 39460+ A? foo.default.svc.cluster.local. (47)
10:35:09.052993 IP kube-dns.kube-system.svc.cluster.local.53 > foo.58416: 39460 NXDomain*- 0/1/0 (140)
10:35:09.053708 IP foo.53273 > kube-dns.kube-system.svc.cluster.local.53: 3313+ A? foo.svc.cluster.local. (39)
10:35:09.054315 IP kube-dns.kube-system.svc.cluster.local.53 > foo.53273: 3313 NXDomain*- 0/1/0 (132)
10:35:09.055023 IP foo.44989 > kube-dns.kube-system.svc.cluster.local.53: 19225+ A? foo.cluster.local. (35)
10:35:09.056578 IP kube-dns.kube-system.svc.cluster.local.53 > foo.44989: 19225 NXDomain*- 0/1/0 (128)
10:35:09.057185 IP foo.39326 > kube-dns.kube-system.svc.cluster.local.53: 2052+ A? foo. (21)
10:35:09.060754 IP kube-dns.kube-system.svc.cluster.local.53 > foo.39326: 2052 1/0/0 A 127.0.0.1 (40)
Server:		10.96.0.10
Address:	10.96.0.10#53

Non-authoritative answer:
Name:	foo
Address: 127.0.0.1
```

Whenever a container queries a name that is outside of Kubernetes, it has
to go through all these four   NSDOMA  ` before actually getting a response:

```sh
% tcpdump "udp port 53"
% nslookup google.com
10:40:12.569439 IP foo.49028 > kube-dns.kube-system.svc.cluster.local.53: 33319+ A? google.com.default.svc.cluster.local. (54)
10:40:12.572349 IP kube-dns.kube-system.svc.cluster.local.53 > foo.49028: 33319 NXDomain*- 0/1/0 (147)
10:40:12.572970 IP foo.54948 > kube-dns.kube-system.svc.cluster.local.53: 48254+ A? google.com.svc.cluster.local. (46)
10:40:12.573828 IP kube-dns.kube-system.svc.cluster.local.53 > foo.54948: 48254 NXDomain*- 0/1/0 (139)
10:40:12.574303 IP foo.56236 > kube-dns.kube-system.svc.cluster.local.53: 26722+ A? google.com.cluster.local. (42)
10:40:12.574865 IP kube-dns.kube-system.svc.cluster.local.53 > foo.56236: 26722 NXDomain*- 0/1/0 (135)
10:40:12.576272 IP foo.48021 > kube-dns.kube-system.svc.cluster.local.53: 2652+ A? google.com. (28)
10:40:12.609035 IP kube-dns.kube-system.svc.cluster.local.53 > foo.48021: 2652 1/0/0 A 172.217.18.206 (54)
Server:		10.96.0.10
Address:	10.96.0.10#53

Non-authoritative answer:
Name:	google.com
Address: 172.217.18.206
```


The DHCP answer from my router contains the [option header
15](https://tools.ietf.org/html/rfc2132#section-3.17) "Domain Name"; in my
case, the only domain name returned is `home`. Which means that anytime the
client (my machine) wants to query a name, say `macbook-pro`, it will try
in this order:

```sh
macbook-pro.home.
macbook-pro.
```

Here is a capture of my machine trying to figure out the name
`macbook-pro`:

```sh
% tcpdump -ien0 '(udp port 53 && (src 192.168.1.1 || dst 192.168.1.1))'
listening on en0, link-type EN10MB (Ethernet), capture size 262144 bytes
12:04:34.873178 IP 192.168.1.14.58108 > 192.168.1.1.domain: 1698+ A? macbook-pro.home. (41)
12:04:34.924824 IP 192.168.1.1.domain > 192.168.1.14.58108: 1698 NXDomain 0/1/0 (116)
12:04:34.925217 IP 192.168.1.14.53531 > 192.168.1.1.domain: 9166+ A? macbook-pro. (39)
12:04:34.948013 IP 192.168.1.1.domain > 192.168.1.14.53531: 9166* 2/0/0 A 192.168.1.14, A 192.168.1.21 (71)
```

- a hierarchical DNS (also called DNS chaining or DNS delegation) is when a
  DNS has a record `NS` to another DNS server.

[kubernetes-plugin]: https://github.com/coredns/coredns/tree/master/plugin/kubernetes
[rfc1123]: https://tools.ietf.org/html/rfc1123
[rfc1035]: https://tools.ietf.org/html/rfc1035#section-2.3.1 "Label and subdomain"
[rfc1034-terms]: https://tools.ietf.org/html/rfc1034#section-2.4 "Elements of a DNS"
[rfc1034-space]: https://tools.ietf.org/html/rfc1034#section-3.1 "Domain space"
[rfc7719]: https://tools.ietf.org/html/rfc7719 "DNS terminology"

One single [SOA record](https://simpledns.plus/help/soa-records) (start of
authority) exists for every given zone. As you can see here, my zone is
`maelvls.dev.`.

## Playing with `k8s_gateway`

The Corefile is available
[here](https://github.com/maelvls/k.maelvls.dev/blob/0e77a838251b209646f433a2b1d5e1a440f8e856/helm/ext-coredns.yaml#L13-L31).

Before, my top-level DNS would be littered with records created by
ExternalDNS:

```sh
% gcloud dns record-sets list --zone=maelvls
NAME                      TYPE   TTL    DATA
maelvls.dev.              A      300    185.199.108.153,185.199.109.153,185.199.110.153,185.199.111.153
maelvls.dev.              MX     300    1 aspmx.l.google.com.,5 alt1.aspmx.l.google.com.,5 alt2.aspmx.l.google.com.,10 alt3.aspmx.l.google.com.,10 alt4.aspmx.l.google.com.,15 pdquboxtbnqki2zinxaksc3jnnluefibfdbqhi7ghbhfbg7ef47q.mx-verification.google.com.
maelvls.dev.              NS     21600  ns-cloud-a1.googledomains.com.,ns-cloud-a2.googledomains.com.,ns-cloud-a3.googledomains.com.,ns-cloud-a4.googledomains.com.
maelvls.dev.              SOA    21600  ns-cloud-a1.googledomains.com. cloud-dns-hostmaster.google.com. 12 21600 3600 259200 300
maelvls.dev.              TXT    300    "keybase-site-verification=PnIWsZlbzCGwYrc5J_VCVphBOMHCVjcIx6nMSkeCZzI"
concourse.k.maelvls.dev.  A      300    91.211.152.190
concourse.k.maelvls.dev.  TXT    300    "heritage=external-dns,external-dns/owner=k8s,external-dns/resource=ingress/concourse/cm-acme-http-solver-xlvtk"
drone.k.maelvls.dev.      A      300    91.211.152.190
drone.k.maelvls.dev.      TXT    300    "heritage=external-dns,external-dns/owner=k8s,external-dns/resource=ingress/drone/cm-acme-http-solver-xv5cs"
minio.k.maelvls.dev.      A      300    91.211.152.190
minio.k.maelvls.dev.      TXT    300    "heritage=external-dns,external-dns/owner=k8s,external-dns/resource=ingress/minio/cm-acme-http-solver-82slp"
*.minio.k.maelvls.dev.    A      300    91.211.152.190
*.minio.k.maelvls.dev.    TXT    300    "heritage=external-dns,external-dns/owner=k8s,external-dns/resource=ingress/minio/minio"
ns.k.maelvls.dev.         A      300    91.211.152.190
ns.k.maelvls.dev.         TXT    300    "heritage=external-dns,external-dns/owner=k8s,external-dns/resource=service/ext-coredns/ext-coredns"
```

After:

```plain
% gcloud dns record-sets list --zone=maelvls
NAME               TYPE  TTL    DATA
maelvls.dev.       A     300    185.199.108.153,185.199.109.153,185.199.110.153,185.199.111.153
maelvls.dev.       MX    300    1 aspmx.l.google.com.,5 alt1.aspmx.l.google.com.,5 alt2.aspmx.l.google.com.,10 alt3.aspmx.l.google.com.,10 alt4.aspmx.l.google.com.,15 pdquboxtbnqki2zinxaksc3jnnluefibfdbqhi7ghbhfbg7ef47q.mx-verification.google.com.
maelvls.dev.       NS    21600  ns-cloud-a1.googledomains.com.,ns-cloud-a2.googledomains.com.,ns-cloud-a3.googledomains.com.,ns-cloud-a4.googledomains.com.
maelvls.dev.       SOA   21600  ns-cloud-a1.googledomains.com. cloud-dns-hostmaster.google.com. 13 21600 3600 259200 300
maelvls.dev.       TXT   300    "keybase-site-verification=PnIWsZlbzCGwYrc5J_VCVphBOMHCVjcIx6nMSkeCZzI"
k.maelvls.dev.     NS    300    ns.k.maelvls.dev.
ns.k.maelvls.dev.  A     300    91.211.152.190
ns.k.maelvls.dev.  TXT   300    "heritage=external-dns,external-dns/owner=k8s,external-dns/resource=service/ext-coredns/ext-coredns"
```

It works!

```sh
% dig +trace minio.k.maelvls.dev
.                     407578  IN  NS  a.root-servers.net.
.                     407578  IN  NS  b.root-servers.net.
.                     407578  IN  NS  c.root-servers.net.
dev.                  172800  IN  NS  ns-tld1.charlestonroadregistry.com.
dev.                  172800  IN  NS  ns-tld2.charlestonroadregistry.com.
maelvls.dev.          10800   IN  NS  ns-cloud-a1.googledomains.com.
maelvls.dev.          10800   IN  NS  ns-cloud-a2.googledomains.com.
k.maelvls.dev.        300     IN  NS  ns.k.maelvls.dev.
minio.k.maelvls.dev.  5       IN  A   91.211.152.190
```
