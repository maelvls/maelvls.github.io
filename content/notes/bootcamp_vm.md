---
title: Use the Bootcamp partition under VirtualBox
date: 2019-04-20
---

From: <https://apple.stackexchange.com/questions/220670/run-boot-camp-windows10-partition-inside-virtualbox>

First, find which disks are EFI and BOOTCAMP with `diskutil list`.

```sh
diskutil list
diskutil unmount /Volumes/BOOTCAMP
sudo chmod 777 /dev/disk0s1
sudo chmod 777 /dev/disk0s3
sudo VBoxManage internalcommands createrawvmdk -rawdisk /dev/disk0 -filename win10raw.vmdk -partitions 1,3
```

Uncheck "Enable EFI", uncheck "Enable VT-x/AMD-V"
