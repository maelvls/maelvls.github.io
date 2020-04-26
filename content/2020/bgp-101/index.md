---
title: "BGP 101"
description: ""
date: 2020-04-19T17:51:55+02:00
url: /bgp-101
images: [bgp-101/cover-bgp-101.png]
draft: true
---

```sh
whois $(curl ifconfig.me/ip)
% Information related to '90.76.0.0/16AS3215'
route:          90.76.0.0/16
descr:          France Telecom IP2000-ADSL-BAS
origin:         AS3215                           # ✳
mnt-by:         FT-BRX
created:        2018-08-16T14:22:19Z
last-modified:  2018-08-16T14:22:19Z
source:         RIPE
```

> [AS3215](https://bgp.he.net/AS5511) is my ISP's autonomous system.

```sh
% sudo mtr google.fr
                                        My traceroute  [v0.93]
macbook-pro-de-mael-2.home (2a01:cb19:86a2:8c00:20db:756d:83b3:8891)         2020-04-19T17:55:26+0200
Keys:  Help   Display mode   Restart statistics   Order of fields   quit
                                                             Packets               Pings
 Host                                                      Loss%   Snt   Last   Avg  Best  Wrst StDev
 1. livebox.home                                            0.0%    69    5.6  41.7   3.2 700.4 128.3
 2. 2a01cb08a00402190193025300740181.ipv6.abo.wanadoo.fr    0.0%    69   18.5  61.2   9.8 1012. 158.1
 3. 2a01:cfc4:0:d00::a                                      0.0%    69   42.9  55.9  16.5 957.9 137.9
 4. bundle-ether103.pastr4.-.opentransit.net               76.5%    69   31.8  33.3  23.3  60.9  11.5
 5. et-15-0-2-0.pastr3.-.opentransit.net                    0.0%    69   59.6  70.8  23.4 1050. 153.0
 6. 2001:4860:1:1::336                                      0.0%    69   41.0  67.6  23.6 943.7 133.7
 7. 2001:4860:0:1017::1                                    49.3%    69   29.9  77.2  25.3 1046. 185.7
 8. 2001:4860:0:1::1e3b                                     0.0%    68   34.5  80.7  23.0 952.3 179.7
 9. par10s29-in-x03.1e100.net                               0.0%    68   29.1  72.1  23.7 836.8 154.7
```

```sh
% subnetcalc 80.10.236.253/26
Address       = 80.10.236.253
                   01010000 . 00001010 . 11101100 . 11111101
Network       = 80.10.236.192 / 26
Netmask       = 255.255.255.192
Broadcast     = 80.10.236.255
Wildcard Mask = 0.0.0.63
Hosts Bits    = 6
Max. Hosts    = 62   (2^6 - 2)
Host Range    = { 80.10.236.193 - 80.10.236.254 }
Properties    =
   - 80.10.236.253 is a HOST address in 80.10.236.192/26
   - Class A
DNS Hostname  = (nodename nor servname provided, or not known)
```

```sh
% nc route-server.opentransit.net 23

&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&
         route-server.opentransit.net -- & Opentransit
                        IPv4/IPv6 views


This router keeps peering sessions with all the Opentransit Backbone Routers,
throughout the Opentransit IP Backbone as follows:



[IPv4/IPv6 view]


IPv4: 193.251.245.1             IPv6: 2001:688:0:1::158 Dallas
IPv4: 193.251.245.3             IPv6: 2001:688:0:1::1c  Los Angeles
IPv4: 193.251.245.7             IPv6: 2001:688:0:1::55  London
IPV4: 193.251.245.9             IPv6: 2001:688:0:1::41  Palo Alto
IPv4: 193.251.245.10            IPv6: 2001:688:0:1::4e  Paris
IPv4: 193.251.245.16            IPv6: 2001:688:0:1::168 New York
IPv4: 193.251.245.19            IPv6: 2001:688:0:1::4b  Barcelona
IPv4: 193.251.245.28            IPv6: 2001:688:0:1::8   Frankfurt
IPv4: 193.251.245.37            IPv6: 2001:688:0:1::    Frankfurt
IPv4: 193.251.245.49            IPv6: 2001:688:0:1::19  London
IPv4: 193.251.245.53            IPv6: 2001:688:0:1::1e  Chicago
IPv4: 193.251.245.57            IPv6: 2001:688:0:1::45  Miami
IPv4: 193.251.245.66            IPv6: 2001:688:0:1::44  Geneva
IPV4: 193.251.245.69            IPV6: 2001:688:0:1::56  Singapore
IPv4: 193.251.245.76            IPv6: 2001:688:0:1::d   New York
IPv4: 193.251.245.78            IPv6: 2001:688:0:1::22  Madrid
IPv4: 193.251.245.81            IPv6: 2001:688:0:1::4   Barcelona
IPv4: 193.251.245.88            IPv6: 2001:688:0:1::f   London
IPv4: 193.251.254.92            IPv6: 2001:688:0:1::24  Paris
IPv4: 193.251.245.96            IPv6: 2001:688:0:1::3E  Brussels
IPv4: 193.251.245.134           IPv6: 2001:688:0:1::4a  Zurich
IPv4: 193.251.245.147           IPv6: 2001:688:0:1::2A  HongKong
IPv4: 193.251.245.163           IPv6: 2001:688:0:1::18  New York
IPv4: 193.251.245.170           IPv6: 2001:688:0:1::2f  Madrid
IPv4: 193.251.245.181           IPv6: 2001:688:0:1::3C  Singapore
IPv4: 193.251.245.196           IPv6: 2001:688:0:1::57  Frankfurt
IPv4: 193.251.245.216           IPv6: 2001:688:0:1::16  Miami
IPv4: 193.251.245.251           IPv6: 2001:688:0:1::30  Amsterdam
IPv4: 193.251.245.252           IPv6: 2001:688:0:1::12  Ashburn


For questions about this route-server, send email to: opentransit.iptac@orange.com


*** Log in with username 'rviews', password 'Rviews' ***


User Access Verification

Username: rviews
rviews
Password: Rviews

OAKRS1#show ?
show ?
  banner      Display banner information
  bgp         BGP information
  bootflash:  display information about bootflash: file system
  flash:      display information about flash: file system
  ip          IP information
  ipv6        IPv6 information
  webui:      display information about webui: file system
  <cr>        <cr>

OAKRS1#show
% Type "show ?" for a list of subcommands
OAKRS1#show bgp
show bgp
% Command accepted but obsolete, unreleased or unsupported; see documentation.

BGP table version is 183946817, local router ID is 204.59.3.38
Status codes: s suppressed, d damped, h history, * valid, > best, i - internal,
              r RIB-failure, S Stale, m multipath, b backup-path, f RT-Filter,
              x best-external, a additional-path, c RIB-compressed,
              t secondary path, L long-lived-stale,
Origin codes: i - IGP, e - EGP, ? - incomplete
RPKI validation codes: V valid, I invalid, N Not found

     Network          Next Hop            Metric LocPrf Weight Path
 r>i  0.0.0.0          193.251.245.7                 100      0 i
 r i                   193.251.245.252               100      0 i
 * i  1.0.0.0/24       4.68.70.169            100     85      0 3356 13335 i
 * i                   195.219.87.114         100     85      0 6453 13335 i
 * i                   4.68.63.97             100     85      0 3356 13335 i
 * i                   4.68.73.97             100     85      0 3356 13335 i
 * i                   193.251.249.46         100     85      0 2914 13335 i
 * i                   4.68.127.233           100     85      0 3356 13335 i
 * i                   129.250.66.141         100     85      0 2914 13335 i
 * i                   193.251.247.156        100     85      0 2914 13335 i
 * i                   4.68.70.173            100     85      0 3356 13335 i
 * i                   66.198.127.81          100     85      0 6453 13335 i
 * i                   4.68.73.153            100     85      0 3356 13335 i
```



<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>
