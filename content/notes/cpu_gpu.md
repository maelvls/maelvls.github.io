---
title: Checking info about CPU and Nvidia GPU on Linux
date: 2018-11-01
tags: []
author: Maël Valais
---

If there is a Nvidia card:

    nvidia-smi

Otherwise:

     cat /proc/cpuinfo
     lscpu
