---
title: Epic journey with statically and dynamically-linked libraries (.a, .so)
description: "Dynamic libraries and PIC (position-independant code) are great features of modern systems. But trying to get the right library built can become a nightmare as soon as you rely on other libraries that may or may not have these features in the first place... In this post, I detail the hacks I made to the ./configure-based build system of Yices, a C++ library."
date: 2020-05-30T20:45:06+02:00
url: /static-libraries-and-autoconf-hell
images: [static-libraries-and-autoconf-hell/cover-static-libraries-and-autoconf-hell.png]
tags: [autotools, c++, c, ocaml]
---

Between May and June 2016, I worked with
[ocamlyices2](https://github.com/polazarus/ocamlyices), an OCaml package
that binds to the [Yices](https://github.com/SRI-CSL/yices2) C++ library.
Both projects (as well as many Linux projects) are built using the
"Autotools" suite.

The Autotools suite includes tools like
[autoconf](https://www.gnu.org/software/autoconf/),
[automake](https://www.gnu.org/software/automake) and
[libtool](https://www.gnu.org/software/libtool/). These tools generate a
bunch of shell scripts and Makefiles using shell scripts and the
[M4](https://en.wikipedia.org/wiki/M4_(computer_language)) macro language.
The user of your projects ends up two simple commands to build your
project:

```sh
./configure
make
```

> Why did I bother with this? As part of my PhD, I worked on a tool,
> [touist](https://github.com/touist/touist), which uses SMT solvers like
> Yices. The `touist` CLI was written in OCaml (a popular language among
> academics, including my supervisor), which meant I had to go through
> hoops to interoperate with C/C++ solvers like Yices.

My first challenge with ocamlyices2 was the fact that I needed a
statically-linked `libyices.a`. And since Yices depends on
[GMP](https://gmplib.org/), I had to dive deep into Yices2's `configure.ac`
and find a way to select the static version of GMP.

But building a static library `libyices.a` was not enough. I was to build
in PIC mode (position-independant code, enabled with `-fPIC` in gcc). The
position-independant code is required when you want to embed a static
library into a dynamically-linked library. That's due to the fact that
OCaml requires both the `.so` and `.a` versions of the "stub" library (a
stub library is a C library that wraps another C library using OCaml C
primitives). And naturally, Yices' build system had not been written to
support building a static PIC `libyices.a`.

I remember these days as an epic struggle against old build systems. This
experience taught me everything about `autoconf`, Makefiles,
position-independant code, `gcc`, `ldd` and `libtool`. And in this post, I
want to share these discoveries and how I progressed into contributing to
the [ocamlyices2](https://github.com/polazarus/ocamlyices) project.

Here is a diagram showing the dependencies between libraries. Ocamlyices2
depends on Yices and Yices depends on GMP.

<div class="nohighlight">

```plain
                            build:   dune
             touist         lang:    ocaml
                |           output:  touist binary (statically linked)
                |
                |
                |depends on
                |
                |
                |
                v
           ocamlyices2      build:   autoconf + make
                |           lang:    ocaml + C
                |           output:  libyices2_stubs.a
                |
                |depends on
                |
                |
                |
                v           build:   autoconf + hacky make
              yices         lang:    C++
                |           output:  libyices.a (with PIC enabled)
                |
                |
                |depends on
                |
                |
                |
                v           build:   autoconf + automake + libtool
               gmp          lang:    C, assembly
                            output:  libgmp.a (with PIC enabled)
```

</div>

<!-- https://textik.com/#23689eecb0630a77 -->

The most important thing about this diagram is that since we need a
PIC-enabled static library `libyices.a` in order to build the "binding
libraries": the shared library `yices2.cmxs` and the static library
`libyices2_stubs.a`.

And in order to get this PIC-enabled `libyices.a`, I needed to make sure
that the GMP library picked by the Yices `./configure` would be static and
PIC-enabled.

In early 2016, the Yices build system was limited to building a shared
library with no support for cross-compilation (a requirement to build on
Windows) and no support for enforcing PIC in `libgmp.a` and `libyices.a`.

## Yices2 & autoconf: an attempt at fixing a limited build system

I remember the warmth that we had in May 2016. My daughter had turned three
and we were still living in a tiny appartment since I was technically still
a student.

My first patch to the ocamlyices2 project took over a month of intense
work. The stake was immense: I aimed at revamping the whole Yices2 build
system. The original build system didn't allow developers to statically
build `libyices.a`. More generally, it was a pain to work with: most
configuration was happening at `./configure` time, but a ton of things had
still to be passed at make time (e.g., `make VAR=value`).

This change
([25f5eb15](https://github.com/polazarus/ocamlyices2/commit/25f5eb15))
brought a ton of features. But the most important ones were:

1. **Better `./configure && make` experience**. Instead of having some
   parts of the build configuration being passed as Makefile variables, I
   moved everything to nice flags that you would pass to `./configure`.
2. **Proper PIC support**. Sometimes, Yices' `./configure` would pick up a
   `libgmp.a` that would not be "PIC". So I wrote a new
   [M4](https://en.wikipedia.org/wiki/M4_(computer_language)) macro,
   `CSL_CHECK_STATIC_GMP_HAS_PIC`, that would do the PIC check using
   libtool. For example, the developer would be able to ask for a static
   library with PIC support:

   ```sh
   ./configure --with-pic --enable-static --disable-shared --with-pic-gmp=$PWD/with-pic/libgmp.a
   ```

3. **Static `libyices.a`**. The original build system did not
4. **Using `libtool`**. Instead of hand-crafted targets for building the static
   & dynamic libraries,  `libtool` allowed me to build both with very
   little effort (at the cost of slightly longer build times). I remember
   trying to fiddle with the Makefile to be able to get PIC/non-PIC as well
   as dynamic/static `.o` units. Using `libtool` made it so easier to build
   simultanously the static and dynamic versions of `libyices`
   (`libyices.a` and `libyices.so`).
5. **Cross-compilation**. I wanted to be able to release binaries for
   Windows users, which meant that I needed to cross-compile from a Cygwin
   environment to a Win32 executable.


I also fixed weird bugs like a failure on Alpine 3.5 due to a race
condition in `ltmain.sh`. Yes, you heard well! A race condition in a build
system!

Not sure, but
[25f5eb15](https://github.com/polazarus/ocamlyices2/commit/25f5eb15) might
be my biggest commit ever:

```eml
Author: Maël Valais <mael.valais@gmail.com>
Date:   Fri May 5 10:31:37 2017 +0200
Commit: 25f5eb15

    libyices: build system: moved all config from Makefile to ./configure

    What I tried to fix:
    * Impossible to produces static archive libyices.a for the mingw32 host.
    * So much configuration done in Makefile instead of ./configure.
    * The 'make build OPTION= MODE=' is kind of unexpected and not really standard.
    * As an end-user builder, it is hard to guess that my system-installed libgmp.a
      is not PIC. The build system should help me with that.
    * CFLAGS and CPPFLAGS cannot be overritten by the end-user builder at 'make'
      time. See 'Preset Output Variables' in the autoconf manual.
    * `smt2_commands.c`, `smt2_parser.c` and `smt2_term_stack.c` are all including
      `smt2_commands.h`, and in `smt2_commands.h` there is a definition of a type,
      so there is a clash of symbols at link time. Solution 1: use 'extern' in .h
      and definition in .c. Solution 2: typedef the enum.
    * The build system should be fully parallel-proof. Note that 'make -j' is a
      thread bomb; prefer using 'make -j4' if you have 4 processors.
    * Why removing libyices.a when using 'dist'? (rm -f $(distdir)/lib/libyices.a)
      libyices.a should be distributed to the end-users along with the shared lib.
    * Why does libgmp.a contain more than libgmp.so? I guess it is because some
      function are needed by the tests and the tests are never dynamically linked
      to the shared libyices.so... For example, the test `test_type_matching` calls
      a function 'delete_pstore' that is not exported. This means we cannot use the
      created libyices.so! We must instead use the object files directly.

    Features of the new build system:
    * It is now possible to select what you want to build using --enable-static
      and --enable-shared. By default, both are built. You can speed up the build
      by disabling one of the modes, for example --disable-shared.
    * Running ./configure will tell you if the libgmp.a found/given with
      --with-static-gmp or --with-pic-gmp is PIC or not.
    * Libtool now handles all shared library naming with version number and symbolic
      links. on linux, .so, on mac os .dylib, on windows .a, .dll.a and .dll...
    * It is now possible to choose if you want PIC-only code in the static
      library libyices.a using --with-pic.
    * The ./configure now configures and the Makefile makes; moved all the
      configuration steps that were done in the Makefile.
    * It is not required anymore to pass OPTION when using make. OPTION is now
      handled by passing for example --host=i686-w64-mingw32 when running
      ./configure, instead of 'make OPTION=mingw32',
    * It is not required anymore to pass MODE in make. MODE is now handled by
      the argument --with-mode=release (for example) when running ./configure.
    * Removed the confusing mess of build/<host>/... and configs/make.<host>.
      Build objects are simply put in build/.
    * Merged the many Makefiles that were sharing a lot of code in common.
    * Standardized the 'make' target and experience:
      - make build for building binaries and library (no need for OPTION or MODE)
      - make dist (non-standard) for showing the results of the build as it would
        be distributed
      - make distclean for removing any files created by ./configure
      - make lib if you only want the library
      - make install and uninstall for installing/uninstalling (DESTDIR supported)
    * On Windows, we can now build a static library libyices.a.
    * On Windows, shared and static libraries can be built at once. The static
      version of libyices.a can be renamed using --with-static-name.
    * DESTDIR works as expected: it will reproduce the hierarchy using the prefix
      when running 'make install DESTDIR=/path'
    * A nice summary of the configuration is now printed when running ./configure.
      It allows to check if libgmp.a is PIC or not, and helps to have a clear
      view of what is going to happen.
    * Moved version number of Yices in configure.ac
    * If the user wants to use --with-pic-gmp but his libgmp.a is not PIC, give him
      an indication of what command to run to build the with-PIC libgmp.a.
    * Moved the gmaketest into configure; warn the user if 'make' is not gnu make.
    * Parallel build is now fully supported (make -j)
    * when using --with-static-gmp (and other similar flags), try to find gmp.h
      in . and ../include automatically, and fall back with the system-wide gmp.h
    * check for gmp.h even when no --with-static-gmp-include-dir is given
    * moved all csl_* functions into autoconf/m4/csl_check_libs.m4 so that they
      can be reused somewhere else
    * compute dependencies only if not in release mode and at compile time
      instead of ahead of time. This saves time during compilation, because deps
      are not necessary if the .c or .h files are not changed (i.e., if the builder
      is the end-user). If the builder is a developer, then he will set
      --with-mode={debug,profile...} and this will trigger the deps to be computed.
    * 'make test' compiles and runs all tests in tests/unit
    * 'make test_api12' will compile and run the test tests/unit/test_api12.c
    * removed version_*.c file as it is rebuilt at make time
    * gperf is now only necessary when changing the tokens.txt or keywords.txt files,
      the end-user builder does not have to have gperf installed.

    Side notes:
    * CPPFLAGS, CFLAGS, LDFLAGS and any other makefile variable can be overwritten
      using 'make LDFLAGS=...' (was already the case with the previous version of the
      build system)

    Known issues:
    * on Mac OS X, linking executables agains non-PIC libgmp.a will throw the
      following warning:
      ld: warning: PIE disabled. Absolute addressing (perhaps -mdynamic-no-pic)
      not allowed in code signed PIE, but used in ___gmpn_mul_1 from libgmp.a(mul_1.o).
      To fix this warning, don't compile with -mdynamic-no-pic or link with -Wl,-no_pie

    Todo:
    * produce libyices.def on Windows
    * make sure the test on libgmp-10.dll is future-proof (remove the '-10')
    * fix the 'echo summary'
    * I did not test the checks made configure.ac on mcsat and libpoly; we should
      do the same tests as done on libgmp.a (for checking that it is PIC) and
      add a helping message at the end of configure.ac.
    * For the tests, I read that two kind of tests were compiled:
      - 'tests' where the tests are linked to the non-PIC libgmp.a and static libgmp.a
      - 'tests-static' where the tests are linked to the PIC libgmp.a and shared
        GMP library.
      I changed the second one: it links to the shared version of libyices and shared
      version of GMP. But building the with-PIC libyices.a would be really easy
      (and it is still possible using the flag --with-pic).
    * 'make dist' is not a staging area (for now) for 'make install'. They both
      install from built objects.
    * it is not possible to compile the tests against the shared library
      because many symbols which are used in the tests are not exported ('export'
      in the C code). They exist if we do 'nm' but they are just not usable.

    * pstore issue:

    The visibility is T (visible) in static mode and t (hidden) in shared mode. This
    is because 'abstract_values.c' has not been 'exported' (=added in `yices_api.c`).
    Here is an example of the difference (functions from )
    ```
    # nm build/lib/libyices.a | grep pstore
    0000000000000090 T _delete_pstore
    0000000000000000 T _init_pstore
    # nm build/lib/.libs/libyices.2.dylib | grep pstore
    0000000000037e10 t _delete_pstore
    0000000000037d80 t _init_pstore
    ```

    I tweaked the ltmain.sh to be able to pass different CPPFLAGS for the compilation
    of static and shared objects by the `%.o: %.c` rule. CPPFLAGS must be different
    because of Windows dlls: `-DNOYICES_DLL -D__GMP_LIBGMP_DLL=0` for example.

    I added two variables LT_STATIC_CFLAGS and LT_SHARED_CFLAGS. They allow me
    to pass CFLAGS to libtool for .c -> .o targets. It allows me to pass things
    like -DYICES_STATIC, -DNOYICES_DLL and -D_LIBGMP_DLL.

    ``` diff
    diff --git a/ext/yices/autoconf/ltmain.sh b/ext/yices/autoconf/ltmain.sh
    index bf5d83b..7b7dd3a 100644
    --- a/ext/yices/autoconf/ltmain.sh
    +++ b/ext/yices/autoconf/ltmain.sh
    @@ -3500,10 +3500,10 @@ compiler."
           fbsd_hideous_sh_bug=$base_compile

           if test no != "$pic_mode"; then
    -       command="$base_compile $qsrcfile $pic_flag"
    +       command="$base_compile $LT_SHARED_CFLAGS $qsrcfile $pic_flag"
           else
            # Don't build PIC code
    -       command="$base_compile $qsrcfile"
    +       command="$base_compile $LT_SHARED_CFLAGS $qsrcfile"
           fi

           func_mkdir_p "$xdir$objdir"
    @@ -3552,9 +3552,9 @@ compiler."
         if test yes = "$build_old_libs"; then
           if test yes != "$pic_mode"; then
            # Don't build PIC code
    -       command="$base_compile $qsrcfile$pie_flag"
    +       command="$base_compile $LT_STATIC_CFLAGS $qsrcfile$pie_flag"
           else
    -       command="$base_compile $qsrcfile $pic_flag"
    +       command="$base_compile $LT_STATIC_CFLAGS $qsrcfile $pic_flag"
           fi
           if test yes = "$compiler_c_o"; then
            func_append command " -o $obj"
    ```
```

What a massive commit message, right?!

Oviously, I still needed to use all the new features that I had just addeed
to the Yices `./configure`. So I proposed a second patch
([ccb5a563](https://github.com/polazarus/ocamlyices2/commit/ccb5a563)) with
changes to the ocamlyices2's own `./configure`; that took the form of new
flags so that I would be able to build a static `libyices.a` by specifying
the static version of the GMP library `libgmp.a`.

I also used the new cross-compilation capability of the Yices `./configure`
so that it would be possible to build ocamlyices2 on Windows. The
compilation would rely on the POSIX-compliant
[Cygwin](https://www.cygwin.com/) suite and cross-compile to a native Win32
executable using [MinGW32](http://www.mingw.org/) (which a port of GCC).

```eml
Author: Maël Valais <mael.valais@gmail.com>
Date:   Fri May 5 16:18:46 2017 +0200
Commit: ccb5a563

    ocamlyices2: only build the static stub using static libgmp.a.

    The library libgmp.a will be searched in system dirs or you can use the flag
    --with-static-gmp= when running ./configure for setting your own libgmp.a. If
    no libgmp.a is found, the shared library is used. You can force the use of
    shared gmp library with --with-shared-gmp.

    If --with-shared-gmp is not given, the libgmp.a that has been found
    will be included in the list of installed files. The reason is because
    if we want to build a shared-gmp-free binary, zarith will sometimes pick the
    shared library (with -lgmp) over the static lbirary libgmp.a.

    Including libgmp.a in the distribution of ocamlyices2 is a convenience for
    creating gmp-shared-free binaries.

    Why do we prefer using a static version of libgmp.a?
    ===================================================

    This is because we build a non-PIC static version of libyices.a. If
    we wanted to build both static and shared stubs, we should either
    - build a PIC libyices.a but it would conflict with the non-PIC one
    - build a shared library libyices.so.

    For now, I chose to just skip the shared stubs (dllyices2_stubs.so).

    Also:

    * turn on -fPIC (in configure.ac) only if non-static gmp
    * added a way to link statically to libgmp.a (--with-static-gmp)
    * use -package instead of -I/lib for compiling *.c in ocamlc

    This option uses the change I made to the build system of libyices.
    Why? Because I want the possibility of producing binaries that do
    not need any dll alongside.

    Guess the host system and pass it to libyices ./configure
    =========================================================
    It is now possible to use ./configure for mingw32 cross-compilation.
```

A month past and I soon realized that GMP was not embedded at all in
`libyices.a`. I could see the symbols as undefined (`U`) when running `nm
libyices.a`. It took me a while to figure this out... back to tweaking the
fragile Yices `./configure`!

Along the way, I also realized how different Linux distributions are. Arch
Linux is notably lacking support for partial linking (`ld -r`). Which meant
I had to add a flag for this specific purpose in commit
[38200b0a](https://github.com/polazarus/ocamlyices2/commit/38200b0a):

```eml
Author: Maël Valais <mael.valais@gmail.com>
Date:   Fri Jun 16 21:18:37 2017 +0200
Commit: 38200b0a

    libyices: added option --without-gmp-embedded

    This option will disable the partial linking that allows to embed
    GMP into libyices.a (only with --enable-static).

    For example, on Arch linux, the partial linking command

        ld -r -lgmp *.o -o libyices.o

    would fail even if libgmp.so is correclty installed. It seems that 'ld -r'
    would only work with a static libgmp.a (but the arch linux repo only installs
    the shared gmp library).

    Why this `ld -r`? This command is the only way I found to compile a
    static libyices.a from either a shared or a static libgmp and produce
    a gmp-depend-free libyices.a.

    Two solutions:

    1. Drop the necessity for building a libyices.a free of gmp dependency.
       In this case, I could remove the `ld -r`.
       It would then create a libyices.a that depends on libgmp.a/so.
    2. Separate the gmp-dependency-free libyices.a from the normal
       gmp-dependent libyices.a. For example, I could use the option
       `--without-gmp-embedded`

    So I went with solution (2).

    This option disables the embedding of GMP inside libyices.a. This
    'embedding' is made using partial linking (ld -r) which seems to
    be failing on Arch Linux when using the shared GMP library.
```

Although I had already added a check (`CSL_CHECK_STATIC_GMP_HAS_PIC` in
[25f5eb15](https://github.com/polazarus/ocamlyices2/commit/25f5eb15)) to
make sure that the user-provided libgmp was PIC when using `--with-pic`, I
realized that it was trickier that what I thought in
[55c8e92a](https://github.com/polazarus/ocamlyices2/commit/55c8e92a)...

```eml
Author: Maël Valais <mael.valais@gmail.com>
Date:   Sat Jun 17 14:09:12 2017 +0200
Commit: 55c8e92a

    libyices: with --enable-static and --with-pic, enforce PIC libgmp.a

    One problem I came across was that most of the time, when I was doing

        ./configure --enable-static --with-pic

    the produced libyices.a would still contain non-PIC libgmp.a, although
    being itself PIC. In this commit, I enforce that if --with-pic is given,
    then:

    1) either --with-pic-gmp has been given, in this case we use that for
       creating the PIC libyices.a;
    2) or --with-pic-gmp has not been given and thus we try to simply use the
       shared gmp through -lgmp.

    Reminder: we also check that the system libgmp.a or the libgmp.a given with
    --with-static-gmp is PIC. If it is the case, the PIC libgmp.a will be used
    for --with-pic.
```

Now, I also had to fix ocamlyices2's `./configure` since `ld -r` (partial
linking) was not supported on Arch Linux (see [PR's CI failure](pull
request](https://github.com/ocaml/opam-repository/pull/9086). I remember
waiting for hours for the CI to run on all imaginable systems: Debian,
Ubuntu, Suse, Arch Linux, Alpine, CentOS, Fedora, macOS and Windows...

Commits
[df7c89a1](https://github.com/polazarus/ocamlyices2/commit/df7c89a1) and
[70dc5de5](https://github.com/polazarus/ocamlyices2/commit/70dc5de5)) fix
the partial linking issue on Arch Linux:

```eml
Author: Maël Valais <mael.valais@gmail.com>
Date:   Sat Jun 17 16:52:59 2017 +0200
Commit: df7c89a1

    ocamlyices2: added --with-libyices and --with-libyices-include-dir

    These options allow to give your own libyices.a and the include directory
    where the libyices headers are.

Author: Maël Valais <mael.valais@gmail.com>
Date:   Sat Jun 17 17:01:58 2017 +0200
Commit: 70dc5de5

    ocamlyices2: use --without-gmp-embedded by default

    After giving it some thoughts, the need for having a self-contained libyices.a
    (which would only need -lyices, no -lgmp needed) in ocamlyices2 is pointless
    as 'zarith' will still need '-lgmp' anyway.

    The Makefile will still put libgmp.a and libyices.a inside src/ so that
    the static version of gmp is used (with -L.) instead of the shared version.

    Rationale: disabling the partial linking fixes the build on Arch Linux, which
    (I re-tested on a docker image) cannot accept partial linking with -lgmp when
    only libgmp.so is available. Here is the failing command:

        ld -r *.o -lgmp -o libyices.o
```

## Final result of the Yices2 build system

The experience with the re-written `./configure`
([here](https://github.com/polazarus/ocamlyices2/tree/master/ext/yices)) is
very different from the original one. When the user wants to compile the
library with PIC, they get a warning if one of the dependencies is not PIC.
There is a much finer control over what the user wants: dynamic vs. static,
PIC vs. non-PIC. But also more control over dependencies like GMP, since
the user must be able to pass a static or dynamic version of libgmp.

The `./configure` that you can see
[here](https://github.com/polazarus/ocamlyices2/tree/master/ext/yices)
gained many features like `--with-shared-gmp` or `--with-pic`.

```sh
% ./configure --help
`configure' configures Yices 2.5.2 to adapt to many kinds of systems.

Usage: ./configure [OPTION]... [VAR=VALUE]...

System types:
  --build=BUILD     configure for building on BUILD [guessed]
  --host=HOST       cross-compile to build programs to run on HOST [BUILD]

Optional Features:
  --enable-shared[=PKGS]  build shared libraries [default=yes]
  --enable-static[=PKGS]  build static libraries [default=yes]
  --disable-libtool-lock  avoid locking (might break parallel builds)
  --enable-mcsat          Enable support for MCSAT. This requires the libpoly
                          library.

Optional Packages:
  --with-pic[=PKGS]       try to use only PIC/non-PIC objects [default=use
                          both]
  --with-static-gmp=<path>
                          Full path to a static GMP library (e.g., libgmp.a)
  --with-static-gmp-include-dir=<directory>
                          Directory of include file "gmp.h" compatible with
                          static GMP library
  --with-pic-gmp=<path>   Full path to a relocatable GMP library (e.g.,
                          libgmp.a)
  --with-pic-gmp-include-dir=<directory>
                          Directory of include file "gmp.h" compatible with
                          relocatable GMP library
  --with-static-libpoly=<path>
                          Full path to libpoly.a
  --with-static-libpoly-include-dir=<directory>
                          Path to include files compatible with libpoly.a
                          (e.g., /usr/local/include)
  --with-pic-libpoly=<path>
                          Full path to a relocatable libpoly.a
  --with-pic-libpoly-include-dir=<directory>
                          Path to include files compatible with the
                          relocatable libpoly.a
  --with-shared-gmp       By default, a static version of the GMP library will
                          be searched. This option forces the use of the
                          shared version. This applies for both shared and
                          static libraries.
  --without-gmp-embedded  (Only when --enable-static) By default, the static
                          library libyices.a created will be partially linked
                          (ld -r) so that the GMP library is not needed
                          afterwards (i.e., only -lyices is needed). If you
                          want to disable the partial linking (and thus -lgmp
                          and -lyices will be needed), you can use this flag.
  --with-mode=MODE        The mode used during compilation/distribution. It
                          can be one of release, debug, devel, profile, gcov,
                          valgrind, purify, quantify or gperftools. (default:
                          release)
  --with-static-name=name (Windows only) when building simultanously shared
                          and static libraries, allows you to give a different
                          name for the static version of libyices.a.
```

I also added a ton of diagnostic information that appears at the end of
`./configure`. That's very useful when you want to make sure that
`./configure` has picked up the right version of `libgmp`:

```plain
configure: Summary of the configuration:
EXEEXT:
SED:                        /usr/bin/sed
LN_S:                       ln -s
MKDIR_P:                    /usr/local/opt/coreutils/libexec/gnubin/mkdir -p
CC:                         gcc
LD:                         /Library/Developer/CommandLineTools/usr/bin/ld
AR:                         ar
RANLIB:                     ranlib
STRIP:                      strip
GPERF:                      gperf
NO_STACK_PROTECTOR:         -fno-stack-protector
STATIC_GMP:                 /usr/local/lib/libgmp.a
STATIC_GMP_INCLUDE_DIR:
PIC_GMP:                    /usr/local/lib/libgmp.a
PIC_GMP_INCLUDE_DIR:
ENABLE_MCSAT:               no
STATIC_LIBPOLY:
STATIC_LIBPOLY_INCLUDE_DIR:
PIC_LIBPOLY:
PIC_LIBPOLY_INCLUDE_DIR:

Version:                    Yices 2.5.2
Host type:                  x86_64-apple-darwin19.4.0
Install prefix:             /Users/mvalais/code/ocamlyices2/ext/yices
Build mode:                 release

For both static and shared library:
  CPPFLAGS:                  -DMACOSX -DNDEBUG
  CFLAGS:                    -fvisibility=hidden -Wall -Wredundant-decls -O3 -fomit-frame-pointer -fno-stack-protector
  LDFLAGS:

For static library          libyices.a:
  Enable:                   yes
  STATIC_CPPFLAGS:           -DYICES_STATIC
  STATIC_LIBS:
  Libgmp.a found:           yes
  Libgmp.a path:            /usr/local/lib/libgmp.a
  Libgmp.a is pic:          yes     (non-PIC is faster for the static library)
  PIC mode for libyices.a:  default
  Use shared gmp instead of libgmp.a:  no
  Embed gmp in libyices.a:  yes

For shared library:
  Enable:                   yes
  SHARED_CPPFLAGS:
  SHARED_LIBS:
  Libgmp.a with PIC found:  yes
  Libgmp.a path:            /usr/local/lib/libgmp.a
  Use shared gmp instead of libgmp.a: no
```

A final word about the `Makefile`: since all the build configuration is
handled by `./configure`, you don't have to pass any variables at `make`
time anymore, which really helps when you need to `make` multiple times in
a row and don't want to type the variables every single time.

## Inpecting static and dynamic libraries

Here are two tips that I learned along the way. First, I very often need to
know what libraries an executable is depending on:

```sh
# macOS
% otool -L /bin/ls
/bin/ls:
    /usr/lib/libutil.dylib (compatibility version 1.0.0, current version 1.0.0)
    /usr/lib/libncurses.5.4.dylib (compatibility version 5.4.0, current version 5.4.0)
    /usr/lib/libSystem.B.dylib (compatibility version 1.0.0, current version 1281.100.1)

# Linux (Alpine Linux 3.9)
ldd /bin/ls
    /lib/ld-musl-x86_64.so.1 (0x7fc9b2c84000)
    libc.musl-x86_64.so.1 => /lib/ld-musl-x86_64.so.1 (0x7fc9b2c84000)
```

I also had to dig into the symbols of libraries in order to make sure that
static libraries contained all the needed symbols:

```sh
% nm ext/libyices_pic_no_gmp/lib/libyices.a
ext/libyices_pic_no_gmp/lib/libyices.a(libyices.o):
00000000000ef6c0 t _convert_rba_tree
0000000000049a20 t _convert_simple_value
00000000000c2f70 t _convert_term_to_bit
00000000000d97e0 t _convert_term_to_conditional
0000000000049700 t _convert_term_to_val
0000000000049b30 t _convert_val
00000000000028b0 T _yices_or
0000000000002fa0 T _yices_or2
0000000000002ca0 T _yices_or3
0000000000003480 T _yices_pair
000000000002ad30 t _yices_parse
0000000000006420 T _yices_parse_bvbin
00000000000064b0 T _yices_parse_bvhex
0000000000004200 T _yices_parse_float
0000000000004180 T _yices_parse_rational
000000000000b150 T _yices_parse_term
000000000000b090 T _yices_parse_type
                 U _memcpy
                 U _memset
                 U _memset_pattern16
                 U ___error
0000000000145280 S ___gmp_0
000000000017dcf0 D ___gmp_allocate_func
00000000001099e0 T ___gmp_assert_fail
0000000000109980 T ___gmp_assert_header
0000000000145460 S ___gmp_binvert_limb_table
000000000014527c S ___gmp_bits_per_limb
0000000000109a60 T ___gmp_default_allocate
0000000000109ae0 T ___gmp_default_free
0000000000109a90 T ___gmp_default_reallocate
0000000000145290 S ___gmp_digit_value_tab
```

The letter before the symbol is the "symbol type" (from `man nm`):

> Each symbol name is preceded by its value (blanks if undefined). This
> value is followed by one of the following characters, representing the
> symbol type:
>
> - U = undefined,
> - T (text section symbol),
> - D (data section symbol),
> - S (symbol in a section other than those above).
>
> If the symbol is local (non-external), the symbol's type is instead
> represented by the corresponding lowercase letter. A lower case u in a
> dynamic shared library indicates a undefined reference to a private
> external in another module in the same library.

For example, the symbol `_yices_parse_float` is an external symbol, meaning
that this symbol isn't static to `libyices.a`. On the other side,
`_convert_simple_value` is statically defined (`t`).

## Contributing the new Yices build system to upstream

On 16 June 2016, I sent an email to Bruno Dutertre, one of the developers
at SRI (the company behind the Yices SMT solver). I proposed all these
changes with links to the various patches on GitHub. Unfortunately, it
didn't work out, and the reason might be that the whole patch was enormous
and very hard to review.

> Hi Maël,
>
> Thanks for the message and for your efforts. We'll look into your updates.
>
>We know that the Yices build system is unconventional because most of the
>work is done in the Makefiles rather that in the configure script. There
>are historical reason for this (and it should be able to build PIC
>libraries without problems).
>
>By the way, Yices is now open-source (GPL) on github:
>https://github.com/SRI-CSL/yices2. Take a look when you have time,
>
>Thanks again,
>
>Bruno

I wish we had a unit-test framework for `autoconf`. The `autoconf`
ecosystem generates very fragile scripts and the only way to test them is
to run them with all possible flag combinations, which is pretty much
impossible.

- **Update 31 May 2020**: added the ascii "dependency" diagram to give a
  better sense of what the challenge was.
