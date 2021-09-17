

packet was coming as

Incoming packet:

    src: 10.8.8.10
    dst: 10.4.0.2

The Windows VM receives the packet on `nic2`. IIS, which is listening on all
interfaces, handles the connection and sends the response to 10.8.8.10.

The Windows IP stack looks at its routes:

```
C:\> route print
IPv4 Route Table
===========================================================================
Active Routes:
Network Destination        Netmask          Gateway       Interface  Metric
          0.0.0.0          0.0.0.0       10.132.0.1      10.132.0.13      5
          0.0.0.0          0.0.0.0         10.4.0.1         10.4.0.2    261
         10.4.0.0    255.255.240.0         On-link          10.4.0.2    261
         10.4.0.2  255.255.255.255         On-link          10.4.0.2    261
      10.4.15.255  255.255.255.255         On-link          10.4.0.2    261
       10.132.0.0    255.255.240.0       10.132.0.1      10.132.0.13      6
       10.132.0.1  255.255.255.255         On-link       10.132.0.13      6
      10.132.0.13  255.255.255.255         On-link       10.132.0.13    261
===========================================================================
Persistent Routes:
  Network Address          Netmask  Gateway Address  Metric
  169.254.169.254  255.255.255.255         On-link        1
          0.0.0.0          0.0.0.0         10.4.0.1  Default
===========================================================================
```

![](remmina_Quick%20Connect_localhost_2021812-2141.png)

The response IP headers:

    src: 10.4.0.2
    dst: 10.8.8.10

Since the address `10.8.8.10` is

When the nic 2 has an empty gateway:

```
C:\> route print
IPv4 Route Table
===========================================================================
Active Routes:
Network Destination        Netmask          Gateway       Interface  Metric
          0.0.0.0          0.0.0.0       10.132.0.1      10.132.0.13      5
         10.4.0.0    255.255.240.0         On-link          10.4.0.2    261
         10.4.0.2  255.255.255.255         On-link          10.4.0.2    261
      10.4.15.255  255.255.255.255         On-link          10.4.0.2    261
       10.132.0.0    255.255.240.0       10.132.0.1      10.132.0.13      6
       10.132.0.1  255.255.255.255         On-link       10.132.0.13      6
      10.132.0.13  255.255.255.255         On-link       10.132.0.13    261
        127.0.0.0        255.0.0.0         On-link         127.0.0.1    331
        127.0.0.1  255.255.255.255         On-link         127.0.0.1    331
  127.255.255.255  255.255.255.255         On-link         127.0.0.1    331
  169.254.169.254  255.255.255.255         On-link       10.132.0.13      6
  169.254.169.254  255.255.255.255         On-link          10.4.0.2    261
        224.0.0.0        240.0.0.0         On-link         127.0.0.1    331
        224.0.0.0        240.0.0.0         On-link          10.4.0.2    261
        224.0.0.0        240.0.0.0         On-link       10.132.0.13    261
  255.255.255.255  255.255.255.255         On-link         127.0.0.1    331
  255.255.255.255  255.255.255.255         On-link          10.4.0.2    261
  255.255.255.255  255.255.255.255         On-link       10.132.0.13    261
===========================================================================
Persistent Routes:
  Network Address          Netmask  Gateway Address  Metric
  169.254.169.254  255.255.255.255         On-link        1
===========================================================================
```

![](remmina_Quick%20Connect_localhost_2021812-21538.png)

Hypotheses:

1. The packet arrives into the VM but the TCP connection cannot be initiated.
2. The reason the TCP connection cannot be initiated is that Windows fails sending an ACK packet because it does not have a gateway IP set for this network card.
3. The ACK cannot be sent because Windows does not find an entry for 10.8.8.10 in its ARP table, and since no gateway is set, it cannot figure out to which MAC address the ACK packet should be addressed.

Is it the same on Linux?
