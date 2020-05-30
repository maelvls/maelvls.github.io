---
title: All about libraries (.a, .so)
date: 00-00-00
---

To know what libraries an executable is depending on (mac, linux):

    otool -L binary
    ldd binary

To see the symbols of a library:

    nm lib.a
