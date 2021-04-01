---
title: Using dd
date: 2018-10-08
tags: []
author: Maël Valais
devtoSkip: true
---

```shell
sudo dd if=$HOME/Downloads/elementaryos-0.4.1-stable.20170814.iso of=/dev/disk3s1 status=progress
df -h
diskutil unmountDisk /dev/...
```
