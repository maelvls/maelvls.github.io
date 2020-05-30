---
title: Stuff about Wireshark
date: 2019-11-06
---

```sh
tcpdump -i eth0 -U -w - 'not port 22' | wireshark -k -i -

wireshark -i en9 -k -Y "ip.addr == 35.211.248.124 && tcp.port == 22"

nc -v 35.211.248.124 22
```
