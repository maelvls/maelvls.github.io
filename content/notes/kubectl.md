---
title: Some kubectl stuff
date: 2019-01-03
tags: []
author: Maël Valais
devtoSkip: true
---

```sh
kubectl run --generator=run-pod/v1 tmp-shell --rm -i --tty --image nicolaka/netshoot -- /bin/bash
```

```sh
gcloud ssh ...
docker ps --format='{{.ID}} {{.Names}}'
kubectl run --generator=run-pod/v1 tmp-shell --rm -i --tty --image nicolaka/netshoot -- /bin/bash
sudo iptables-save > a && vim a && sudo iptables-restore < a
```

```sh
% mvalais@gke-august-period-micro-g961 ~ $ vmstat -SM 1 10
procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
 r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
 0  0      0    104      7    218    0    0   160    57  273  760  2  1 96  1  0
```

```sh
k get pods -A --field-selector=spec.nodeName=gke-august-period-234610-worker-micro-1de60498-g961
gcloud compute ssh gke-august-period-234610-worker-micro-1de60498-g961 --zone=us-east1-c --command "docker run --rm --net=host corfr/tcpdump -i cbr0 -U -w - 'not port 22'" | wireshark -k -i -
```

```sh
dmsg
```

```s
# Host B is 10.142.15.201
# Doesn't work:
#   from the pod on host B, nslookup github.com (tcpdump on pod's eth0)
10.24.10.2      10.27.240.10    DNS 98 Standard query            A github.com.traefik.svc.cluster.local
#   same from host's eth0
10.24.10.2      10.24.9.3       DNS 96 Standard query            A github.com.traefik.svc.cluster.local

# Works:
#   from the host B, nslookup github.com 10.27.240.10 (tcpdump on host's eth0)
10.142.15.201   10.24.9.3       DNS 70 Standard query            A github.com
10.24.9.3       10.142.15.201   DNS 86 Standard query response   A github.com A 140.82.113.4
```

Dumping iptables on host B:

```sh
% gcloud compute ssh gke-august-period-234610-worker-micro-1de60498-g961 --zone=us-east1-c --command "docker run --rm --net=host --privileged praqma/network-multitool iptables-save" | pbcopy
-A KUBE-HP-JSWHQ3CPP5EMKPCA -s 10.24.10.2/32 -m comment --comment "akrobateo-traefik-pqkqf_traefik hostport 443" -j KUBE-MARK-MASQ
-A KUBE-HP-JSWHQ3CPP5EMKPCA -p tcp -m comment --comment "akrobateo-traefik-pqkqf_traefik hostport 443" -m tcp -j DNAT --to-destination 10.24.10.2:443
-A KUBE-HP-Q34NV5ACWTD24LGG -s 10.24.10.2/32 -m comment --comment "akrobateo-traefik-pqkqf_traefik hostport 80" -j KUBE-MARK-MASQ
-A KUBE-HP-Q34NV5ACWTD24LGG -p tcp -m comment --comment "akrobateo-traefik-pqkqf_traefik hostport 80" -m tcp -j DNAT --to-destination 10.24.10.2:80
```

Contrack:

```s
# From the container. Since orig dst != reply src, packet is DNAT.
% nslookup github.com
% mvalais@gke-august-period-234610-worker-micro-1de60498-g961 ~ $ docker run --net=host --privileged --rm missioncriticalkubernetes/conntrack -L | grep 10.24
conntrack v1.4.4 (conntrack-tools): 52 flow entries have been shown.
udp 17 26  src=10.24.10.2    dst=10.27.240.10 sport=57400 dport=53 [UNREPLIED] src=10.24.9.3 dst=10.24.10.2     sport=53 dport=57400           mark=0 use=1
# From host B. Since orig dst != reply src, packet is DNAT.
% nslookup github.com 10.27.240.10
% mvalais@gke-august-period-234610-worker-micro-1de60498-g961 ~ $ docker run --net=host --privileged --rm missioncriticalkubernetes/conntrack -L | grep 10.24
udp 17 175 src=10.142.15.201 dst=10.27.240.10 sport=37337 dport=53             src=10.24.9.26 dst=10.142.15.201 sport=53 dport=37337 [ASSURED] mark=0 use=1
```

```sh
mvalais@gke-august-period-234610-worker-micro-1de60498-g961 ~ $ docker run -it --rm --net=host praqma/network-multitool ip route
default via 10.142.0.1 dev eth0    proto dhcp                src 10.142.15.201 metric 1024
default via 10.142.0.1 dev eth0    proto dhcp                                  metric 1024
10.24.10.0/24          dev cbr0    proto kernel scope link   src 10.24.10.1
10.142.0.1             dev eth0    proto dhcp scope   link   src 10.142.15.201 metric 1024
10.142.0.1             dev eth0    proto dhcp scope   link                     metric 1024
169.254.123.0/24       dev docker0 proto kernel scope link   src 169.254.123.1 linkdown
```

What's very weird is that ICMP packets are properly routed:

```sh
% mvalais@gke-august-period-234610-worker-micro-1de60498-g961 ~ $ docker run --rm -it --net=container:9e557d74b06c praqma/network-multitool bash
% bash-5.0# ping 10.24.9.3
PING 10.24.9.3 (10.24.9.3) 56(84) bytes of data.
64 bytes from 10.24.9.3: icmp_seq=1 ttl=62 time=1.47 ms

% bash-5.0# nc -v 10.24.9.21 9000
```

Let's take another pod that is on host A. It's minio, which serves on port 10.24.9.21:9000:

```sh
bash-5.0# ping 10.24.9.21
PING 10.24.9.21 (10.24.9.21) 56(84) bytes of data.
64 bytes from 10.24.9.21: icmp_seq=1 ttl=62 time=1.16 ms

% bash-5.0# nc -v 10.24.9.21 9000
```

Then I found out that host B has a 'stale kube dns':

```sh
mvalais@gke-august-period-234610-worker-micro-1de60498-g961 ~ $ tail /var/log/kube-proxy.log
I0125 12:30:37.882376       1 service.go:334] Updating existing service port "traefik/traefik:http" at 10.27.244.111:80/TCP
I0125 12:31:35.093785       1 proxier.go:675] Stale udp service kube-system/kube-dns:dns -> 10.27.240.10
```

Now, let's trace in iptables on host B: <https://www.opsist.com/blog/2015/08/11/how-do-i-see-what-iptables-is-doing.html>

```sh
% sudo iptables -t nat -I PREROUTING 1 -j LOG
% sudo iptables -t nat -I POSTROUTING 1 -j LOG
% sudo iptables -t nat -I OUTPUT 1 -j LOG
% demsg -w
[89712.445268]                                                 IN=cbr0 OUT=                                                     PHYSILEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=38810 PROTO=UDP SPT=54407 DPT=53 LEN=62
[89712.463153]                                                 IN=cbr0 OUT=                                                     PHYSILEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=38810 PROTO=UDP SPT=54407 DPT=53 LEN=62
[89712.481021]                                                 IN=cbr0 OUT=                                                     PHYSILEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=38810 PROTO=UDP SPT=54407 DPT=53 LEN=62
[89712.498895] IN= OUT=eth0                                                     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.10    LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=38810 PROTO=UDP SPT=54407 DPT=53 LEN=62
```

Let's try again with the TRACE module:

```s
% sudo modprobe ipt_LOG
% sudo iptables -t raw -I PREROUTING -s 10.24.0.0/16 -j TRACE
% docker run --rm -it --net=container:9e557d74b06c --privileged praqma/network-multitool nslookup github.com
% dmesg -w
[90332.179725] TRACE: raw:PREROUTING:policy:2                  IN=cbr0 OUT=     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.27.240.10  LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.200327] TRACE: mangle:PREROUTING:policy:1               IN=cbr0 OUT=     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.27.240.10  LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.221171] TRACE: nat:PREROUTING:rule:1                    IN=cbr0 OUT=     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.27.240.10  LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.241576] TRACE: nat:KUBE-SERVICES:rule:14                IN=cbr0 OUT=     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.27.240.10  LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.262385] TRACE: nat:KUBE-SVC-TCOU7JCQXEZGVUNU:rule:2     IN=cbr0 OUT=     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.27.240.10  LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.284116] TRACE: nat:KUBE-SEP-MQYG7Z5PYI3N6YFY:rule:2     IN=cbr0 OUT=     PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.27.240.10  LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.305838] TRACE: mangle:FORWARD:policy:1                  IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.326491] TRACE: filter:FORWARD:rule:1                    IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.346973] TRACE: filter:KUBE-FORWARD:return:4             IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.368068] TRACE: filter:FORWARD:rule:2                    IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.389031] TRACE: filter:KUBE-SERVICES:return:1            IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.410253] TRACE: filter:FORWARD:rule:3                    IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.430735] TRACE: filter:DOCKER-USER:return:1              IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.451887] TRACE: filter:FORWARD:rule:4                    IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.472366] TRACE: filter:DOCKER-ISOLATION-STAGE-1:return:2 IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.494488] TRACE: filter:FORWARD:rule:10                   IN=cbr0 OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.515049] TRACE: mangle:POSTROUTING:policy:1              IN=     OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.531693] TRACE: nat:POSTROUTING:rule:1                   IN=     OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.547953] TRACE: nat:KUBE-POSTROUTING:return:2            IN=     OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.564769] TRACE: nat:POSTROUTING:rule:2                   IN=     OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.580978] TRACE: nat:IP-MASQ:rule:2                       IN=     OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
[90332.596840] TRACE: nat:POSTROUTING:policy:4                 IN=     OUT=eth0 PHYSIN=veth4a3719bc SRC=10.24.10.2 DST=10.24.9.3     LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=43592 PROTO=UDP SPT=36775 DPT=53 LEN=62
```

Note that using TRACE is extremely expensive. For example, using `-p icmp` would result in pings going from 1ms (without tracing) to 500ms. I also overloaded one of the nodes by adding a rule that was "too wide" and nearly killed the entire node. Use tracing filters (`-s`, `-p`, `-d`) as narrow as possible!

## Other experiment: icmp

At this point, I recreated both nodes which changed all the IP CIDRs.

- Host A CIDR is `10.24.9.0/24`,
- Host B CIDR is `10.24.11.0/24`,
- We use the traefik pod (`10.24.11.7`) on host B for reproducing the connectivity issue.
- KubeDNS is running on `10.24.9.14`.

Host B:

```s
% docker run --rm -it --net=container:$(docker ps | grep POD_traefik | head -1 | cut -f1 -d' ') --privileged praqma/network-multitool ping 10.24.9.14
% sudo iptables -A PREROUTING -p icmp -t raw -s 10.24.11.0/24 -j TRACE
% sudo iptables -A PREROUTING -p icmp -t raw -d 10.24.11.0/24 -j TRACE
% dmesg -w
[ 5365.769471] TRACE: raw:PREROUTING:policy:5     IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=64 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5365.790765] TRACE: mangle:PREROUTING:policy:1  IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=64 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5365.812684] TRACE: mangle:FORWARD:policy:1     IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=63 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5365.834225] TRACE: filter:FORWARD:rule:1       IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=63 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5365.855880] TRACE: filter:KUBE-FORWARD:rule:2  IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=63 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5365.877624] TRACE: mangle:POSTROUTING:policy:1 IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=63 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
```

Host A (that's where KubeDNS' `10.24.9.14` is):

```s
% sudo iptables -A PREROUTING -p icmp -t raw -s 10.24.11.0/24 -j TRACE
% sudo iptables -A PREROUTING -p icmp -t raw -d 10.24.11.0/24 -j TRACE
% dmesg -w
[ 5528.648942] TRACE: raw:PREROUTING:policy:5     IN=eth0 OUT=     SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=63 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5528.668115] TRACE: mangle:PREROUTING:policy:1  IN=eth0 OUT=     SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=63 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5528.687502] TRACE: mangle:FORWARD:policy:1     IN=eth0 OUT=cbr0 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=62 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5528.706928] TRACE: filter:FORWARD:rule:1       IN=eth0 OUT=cbr0 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=62 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5528.726471] TRACE: filter:KUBE-FORWARD:rule:2  IN=eth0 OUT=cbr0 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=62 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
[ 5528.746157] TRACE: mangle:POSTROUTING:policy:1 IN=     OUT=cbr0 SRC=10.24.11.7 DST=10.24.9.14 LEN=84 TOS=0x00 PREC=0x00 TTL=62 ID=10728 DF PROTO=ICMP TYPE=8 CODE=0 ID=1 SEQ=1
```

## Let's try again with udp

Host B:

```s
% docker run --rm -it --net=container:$(docker ps | grep POD_traefik | head -1 | cut -f1 -d' ') --privileged praqma/network-multitool nslookup github.com
% sudo iptables -A PREROUTING -p udp -t raw -d 10.24.11.0/24 -j TRACE
% sudo iptables -A PREROUTING -p udp -t raw -s 10.24.11.0/24 -j TRACE
% dmesg -w
[ 6160.660913] TRACE: raw:PREROUTING:policy:7                  IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.27.240.10 LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.681727] TRACE: mangle:PREROUTING:policy:1               IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.27.240.10 LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.702962] TRACE: nat:PREROUTING:rule:1                    IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.27.240.10 LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.723562] TRACE: nat:KUBE-SERVICES:rule:2                 IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.27.240.10 LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.744284] TRACE: nat:KUBE-SVC-TCOU7JCQXEZGVUNU:rule:2     IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.27.240.10 LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.766008] TRACE: nat:KUBE-SEP-FNEHYJGPMQEV2RPP:rule:2     IN=cbr0 OUT=     PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.27.240.10 LEN=82 TOS=0x00 PREC=0x00 TTL=64 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.788020] TRACE: mangle:FORWARD:policy:1                  IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.809069] TRACE: filter:FORWARD:rule:1                    IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.829669] TRACE: filter:KUBE-FORWARD:return:4             IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.850869] TRACE: filter:FORWARD:rule:2                    IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.871663] TRACE: filter:KUBE-SERVICES:return:1            IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.893472] TRACE: filter:FORWARD:rule:3                    IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.914059] TRACE: filter:DOCKER-USER:return:1              IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.935373] TRACE: filter:FORWARD:rule:4                    IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.956083] TRACE: filter:DOCKER-ISOLATION-STAGE-1:return:2 IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.978371] TRACE: filter:FORWARD:rule:10                   IN=cbr0 OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6160.999283] TRACE: mangle:POSTROUTING:policy:1              IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6161.016840] TRACE: nat:POSTROUTING:rule:1                   IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6161.033296] TRACE: nat:KUBE-POSTROUTING:return:2            IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6161.050267] TRACE: nat:POSTROUTING:rule:2                   IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6161.066717] TRACE: nat:IP-MASQ:rule:2                       IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
[ 6161.083318] TRACE: nat:POSTROUTING:policy:4                 IN=     OUT=eth0 PHYSIN=veth98b9f376 SRC=10.24.11.7 DST=10.24.9.25 LEN=82 TOS=0x00 PREC=0x00 TTL=63 ID=18335 PROTO=UDP SPT=57262 DPT=53 LEN=62
```

Host A:

```s
% sudo iptables -A PREROUTING -p udp -t raw -d 10.24.11.0/24 -j TRACE
% sudo iptables -A PREROUTING -p udp -t raw -s 10.24.11.0/24 -j TRACE
% dmesg -w
# Nothing!
```

## Communication between two pods across nodes

Host A (10.24.9.26):

```sh
% kubectl run --generator=run-pod/v1 tmp-host-a --overrides='{"apiVersion":"v1","spec":{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchFields":[{"key":"metadata.name","operator":"In","values":["gke-august-period-234610-worker-96e2b32f-6k31"]}]}]}}}}}' --rm -i --tty --image nicolaka/netshoot
% nc -v 10.24.11.21 80
Connection to 10.24.11.21 80 port [tcp/http] succeeded!
```

Host B (10.24.11.21):

```sh
% kubectl run --generator=run-pod/v1 tmp-host-b --overrides='{"apiVersion":"v1","spec":{"affinity":{"nodeAffinity":{"requiredDuringSchedulingIgnoredDuringExecution":{"nodeSelectorTerms":[{"matchFields":[{"key":"metadata.name","operator":"In","values":["gke-august-period-234610-worker-micro-2a7d2fc5-f85j"]}]}]}}}}}' --rm -i --tty --image nicolaka/netshoot
% nc -v -l 80
Listening on [0.0.0.0] (family 0, port 80)
Connection from 10.24.9.26 38274 received!
```

Now, using udp:

```sh
% nc -v -u 10.24.11.21 80
Connection to 10.24.11.21 80 port [udp/http] succeeded!
```

```sh
% tcpdump & nc -u -l 80
[1] 32
tcpdump: verbose output suppressed, use -v or -vv for full protocol decode
listening on eth0, link-type EN10MB (Ethernet), capture size 262144 bytes
# Nothing
```

Definitely something wrong with anything else than 80/tcp!

## OMG it was a missing firewall rule

I realised that only 80/tcp, 443/tcp and icmp packets would get routed. It's surprisingly very specific: what if the VPC firewall rules were off. Maybe these 80/tcp and 443/tcp are correspond to the rule "akrobateo-fw-traefik" I had added when I wanted to load-balance traffic direclty from the cluster (as opposed as external load balancing).

```sh
% gcloud compute firewall-rules list
NAME                                 NETWORK  DIRECTION  PRIORITY  ALLOW                         DENY  DISABLED
akrobateo-fw-traefik                 default  INGRESS    1000      tcp:80,tcp:443                      False
default-allow-icmp                   default  INGRESS    65534     icmp                                False
default-allow-internal               default  INGRESS    65534     tcp:0-65535,udp:0-65535,icmp        False
default-allow-rdp                    default  INGRESS    65534     tcp:3389                            False
default-allow-ssh                    default  INGRESS    65534     tcp:22                              False
```

No rule for traffic in `10.24.0.0/14`! This CIDR is the subnet used for distributing CIDRs to each node. So I added a new rule "all"; here is the new list of rules:

```sh
% gcloud compute firewall-rules list
NAME                                 NETWORK  DIRECTION  PRIORITY  ALLOW                         DENY  DISABLED
akrobateo-fw-traefik                 default  INGRESS    1000      tcp:80,tcp:443                      False
default-allow-icmp                   default  INGRESS    65534     icmp                                False
default-allow-internal               default  INGRESS    65534     tcp:0-65535,udp:0-65535,icmp        False
default-allow-rdp                    default  INGRESS    65534     tcp:3389                            False
default-allow-ssh                    default  INGRESS    65534     tcp:22                              False
gke-august-period-234610-all         default  INGRESS    1000      udp,icmp,esp,ah,sctp,tcp            False
```

Now, let's test again:

Pod "tmp-host-b" (10.24.11.26) on Host B:

```sh
% nc -v -u 10.24.9.27 8888
# Exits with 0 and no output
```

Pod "tmp-host-a" (10.24.9.27) on Host A:

```sh
% tcpdump
tcpdump: verbose output suppressed, use -v or -vv for full protocol decode
listening on eth0, link-type EN10MB (Ethernet), capture size 262144 bytes
21:54:39.787779 IP 10.24.11.26.40760 > tmp-host-a.8888: UDP, length 1
21:54:39.787802 IP tmp-host-a > 10.24.11.26: ICMP tmp-host-a udp port 8888 unreachable, length 37
21:54:39.787805 IP 10.24.11.26.40760 > tmp-host-a.8888: UDP, length 1
21:54:39.787808 IP tmp-host-a > 10.24.11.26: ICMP tmp-host-a udp port 8888 unreachable, length 37
```

Yay!
