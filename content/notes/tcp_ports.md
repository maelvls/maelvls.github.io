---
title: Everything I know about TCP
date: 00-00-00
---

## How is a TCP connection uniquely identified

Kill a process that is using a TCP port you need:

```sh
lsof -n -i :8001
# copy PID
kill 43903
```

A TCP connection is identified by four things:

- source port
- source ip
- dest port
- dest ip

This is why there is no confusion between two requests to a web server on
80 port.

## Stateful vs. Stateless Firewall

This only applies to appliances in between two computers, for example
a router using some kind of firewall. Is it???? Is it only for NATs or
gateways?

```plain
    A -----> Router -----> B
```

If the router manages connection in a stateful way, here is what is going
to happen:

1. A connects to B via a TCP connection identified by IP_A:PORT_A + IP_B:PORT_B
2. Router remembers the source port of A so that when B sends things back,
   it can

References:

- Using linux iptables to implement a stateless packet filtering firewall:
  <https://security.stackexchange.com/questions/74529>

Example:

```iptables
-A INPUT -p tcp -s 192.168.1.0/24 -m tcp --dport 80 -j ACCEPT
```

- `-A CHAIN RULE` tells iptables to append this RULE at the end of this CHAIN.
  A chain is identified by a name such as INPUT or OUTPUT.
- `-p TCP` means protocol
- `-s` is the source. Here, 192.168.1.0/24 means any IP adress that matches
  with `192.*.*.*`
- `-m TCP` is for extensions (see `man iptables-extensions`) but I didn't
  find it...
- `--dport 80` means the destination port must be 80
- `-j ACCEPT` means that the target is the action "ACCEPT"

QUESTIONS: what is the link between protocol and state?

## SG vs. ACLs

Security Groups = stateful firewall at the VM level (= iptables)
Access Control Lists = stateless firewall at the VPC level.

## Protocol stack

The protocol stack is the piece of software that implements (among others)
the TCP/IP layers. A socket is the internal representation of an instance
of this TCP stack. A socket is identified with a number (file descriptor in
Unix terms).

TCP connection = 2 sockets = from(ip:port) + to(ip:port)
TCP socket = internal representation of the 'from' (ip:port)

## The stack

| OSI     | Lv  | TCP/IP   | Keywords     |              |
| ------- | --- | -------- | ------------ | ------------ |
| Applica | L7  | HTTPS    | LB           | Deserialized |
| Present |     | HTTPS    |              | Serialized   |
| Session |     | HTTPS    |              |              |
| Transpo | L4  | TCP,ICMP | Firewall     | Segments     |
| Network | L3  | IP       | Router       | Packets      |
| Link    | L2  | MAC, ARP | VLAN, Switch | Frames       |
| Physic  | L1  |          |              | Bits         |

Warning: HTTP/1.1 and such have intertwined L5, L6 and L7 responsabilities.
As an example, HTTP/1.1 carries headers that can be thought of of
application data, but the payload (data) also spans over L6 (presentation
using json for example). Session is also part of HTTP/1.1.

**L5-L6-L7 (application) protocols** = HTTP/1.1, HTTP/2, WebSockets

**L6 serialization formats** = protobuf, json, avro, xml

**VLAN** = just like the good'old ethernet between multiple equipments.
Switches aliviates some of the issues with the one-cable-for-multiple-pc
with a table (forwarding table, NOT an ARP table) that allows a one-to-one
link with less noise on each link overall.

**TCP** = handles loss of packets using ACKs and SYN with sequence numbers.
SYN = please open a connection. ACK = acknowledge.

**Segments vs packets** = ??

gRPC

## CIDR, network masks, subnets

10.0.2.3/16 = 16 first bits (10.0) for subnet.

The idea of subnets: inside of a subnet, we pretty much share the same
VLAN. For example, let's say that we have two subnets A and B.

|       | Subnet A    | Subnet B    |
| ----- | ----------- | ----------- |
| CIDR  | 10.0.1.0/24 | 10.0.2.0/24 |
| Count | 254         | 254         |
| VLAN  | VLAN1       | VLAN2       |

### Link between Subnet (L3) and VLAN (L2)

First, let us remember that every machine in a VLAN can communicate with
each other at any time. Now, we want to control what goes from/to A and B.

In a VLAN, everything sees everything but it's now really helpful unless we
can use protocols like TCP. But TCP needs IPs. And IPs

A1 (10.0.1.7) ----> A2 (10.0.1.13)
Same subnet, ARP table says I know the MAC address so no need to route.

The subnet allows a machine to know what is 'internal' traffic and what is
routed traffic.

- A1 -----> IP inside subnet = A1 asks which MAC has this IP using ARP.
- A1 -----> IP outside subnet = NAT to the gateway (= a router). See 'Routing'.

## Routing

The gateway has routing tables such as

## AWS terms

IGW = internet gateway
VPC = virtual private cloud. ne VPC is equivalent to a VLAN: everyone can
talk to everyone on L2 (MAC level).

The routing in a VPC:

- each subnet has an ACL (network access control list)
- each EC2 has security groups (L4 stateful firewall)
