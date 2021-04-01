---
title: Using Linyxbrew on university servers
date: 2016-12-15
tags: []
author: Maël Valais
devtoSkip: true
---

On the university server azteca or inca (must use vpn or sassh through-pass)

    export http_proxy=proxy.univ-tlse3.fr:3128
    export https_proxy=proxy.univ-tlse3.fr:3128

Then,

    ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Linuxbrew/install/master/install)"

But on linux with gcc-4.4.7 only, I had also to follow the instructions on <https://github.com/Linuxbrew/homebrew-core/issues/4077>. It will

1. first install glibc 2.20
2. install gcc-5 and others
3. then upgrade to glibc 2.23

   ```shell
   brew install --only-dependencies glibc
   brew install --ignore-dependencies https://raw.githubusercontent.com/Linuxbrew/homebrew-core/6fb5dfd50895416bea3d00628b8d3b41fa1f4f32/Formula/glibc.rb # 2.20
   brew install --ignore-dependencies xz gmp mpfr libmpc isl gcc
   brew upgrade glibc
   ```

Then I also got on the last `brew upgrade glibc` an 'Illegal instruction'; the issue is mentionned in

- <https://github.com/Linuxbrew/legacy-linuxbrew/issues/173>
- <https://github.com/Linuxbrew/homebrew-core/issues/4244>
- (gcc failing) <https://github.com/Linuxbrew/brew/issues/488>
- (gmp 6.1.2 issue) <https://github.com/Linuxbrew/homebrew-core/issues/4261>

The error was:

```plain
make[2]: *** [/tmp/glibc-20170921-55100-dli0th/glibc-2.23/build/csu/elf-init.oS] Erreur 1
libc-tls.c: In function '__libc_setup_tls':
libc-tls.c:105:1: internal compiler error: Illegal instruction
 __libc_setup_tls (size_t tcbsize, size_t tcbalign)
 ^
../sysdeps/x86/libc-start.c: In function '__libc_start_main':
../sysdeps/x86/libc-start.c:20:0: internal compiler error: Illegal instruction
 # else
 ^
../sysdeps/x86/libc-start.c: In function 'apply_irel':
../sysdeps/x86/libc-start.c:40:1: internal compiler error: Illegal instruction
 }
 ^
0x7f23a592972f ???
    /tmp/glibc-20170921-64874-12af6ru/glibc-2.20/signal/../sysdeps/unix/sysv/linux/x86_64/sigaction.c:0
0x7f23a59167cc __libc_start_main
    /tmp/glibc-20170921-64874-12af6ru/glibc-2.20/csu/libc-start.c:289
Please submit a full bug report,
with preprocessed source if appropriate.
Please include the complete backtrace with any bug report.
See <https://github.com/Homebrew/homebrew/issues> for instructions.
```

To install gmp 6.1.1:

    brew install --ignore-dependencies https://raw.githubusercontent.com/Linuxbrew/homebrew-core/58cb0879c8b423ccf7a7cfd8641abc13da5a0753/Formula/gmp.rb

To avoid 'brew updating...' all the time:

    export HOMEBREW_NO_AUTO_UPDATE=1

NOTE: to install on REHL 6.9 or CentOS6, just use the commands given by sjackman at <https://github.com/Linuxbrew/brew/wiki/CentOS6>
