---
title: "Deep Dive into Go Assembly"
description: ""
date: 2020-04-16T15:42:35+02:00
url: /deep-dive-go-assembly
images: []
draft: true
author: Maël Valais
devtoId: 0
devtoPublished: false
---

Last month, Sylvain Wallez, author of the excellent "[Go: the Good, the Bad and the Ugly](https://bluxte.net/musings/2018/04/10/go-good-bad-ugly/)", wrote:

{{< twitter user="bluxte" id="1192945301604184065" >}}

Once again, my beloved language takes a hit. As time passes, I come to realize Go has its strenghts (ease of adoption, no overly abstract REST framework, readability, fast compilation) but also its weaknesses. I can think of the weirdness of variable scoping ([scopelint](https://github.com/kyoh86/scopelint) really helps) and also variable shadowing, weird syntax around variable assignment, `go help` is unreadable and just not up to the standard (compared to `git help` for example), the ever-changing behaviour of GO111MODULE (that's a temporary issue), `go get` that do too many things and changes `go.mod` when you just want to install some tool, but also `go install` & `go run` that don't support versions (e.g. ). Wait, also the "semantic import versioning" that everybody seems to avoid (including Google itself in [protobuf](https://github.com/golang/protobuf/issues/1049)).

But the issue Sylvain raised really bothered me: it felt like a bug.

> Given `v` a pointer to an interface, `v != nil` when `v` is `nil`.

Oooh, let's see.

> Boxed type, fat pointer:
>
> Do Go use boxed types for map[string]string? No! It doesn't use interface{} boxing (boxed = fat pointer = one more level of indirection, less locality – same problem as linked lists).

Here is a short program:

```go
package main

import (
	"fmt"
	"strconv"
)

type abstract interface {
	f()
}

type concrete struct{}

func (concrete) f() {
	i := 0
	i++
	fmt.Print(strconv.Itoa(i))
}

func main() {
	var v1 abstract = &concrete{}
	v1.f()

	v2 := concrete{}
	v2.f()
}
```

Let's go over what `f1` and `f2` do.

```s
# go build . -o ./main
# go tool objdump -S ./main > main.s


TEXT main.concrete.f(SB) /Users/mvalais/code/go-atomic-issue-wallez-tweet/main.go
func (concrete) f() {
  0x10994c0		65488b0c2530000000	MOVQ GS:0x30, CX
  0x10994c9		483b6110		CMPQ 0x10(CX), SP
  0x10994cd		0f869c000000		JBE 0x109956f
  0x10994d3		4883ec58		SUBQ $0x58, SP
  0x10994d7		48896c2450		MOVQ BP, 0x50(SP)
  0x10994dc		488d6c2450		LEAQ 0x50(SP), BP
	fmt.Print(strconv.Itoa(i))
  0x10994e1		48c7042401000000	MOVQ $0x1, 0(SP)
	return FormatInt(int64(i), 10)
  0x10994e9		48c74424080a000000	MOVQ $0xa, 0x8(SP)
  0x10994f2		e8097ffcff		CALL strconv.FormatInt(SB)
  0x10994f7		488b442410		MOVQ 0x10(SP), AX
  0x10994fc		488b4c2418		MOVQ 0x18(SP), CX
	fmt.Print(strconv.Itoa(i))
  0x1099501		48890424		MOVQ AX, 0(SP)
  0x1099505		48894c2408		MOVQ CX, 0x8(SP)
  0x109950a		e8a1f7f6ff		CALL runtime.convTstring(SB)
  0x109950f		488b442410		MOVQ 0x10(SP), AX
  0x1099514		0f57c0			XORPS X0, X0
  0x1099517		0f11442440		MOVUPS X0, 0x40(SP)
  0x109951c		488d0dfd1a0100		LEAQ runtime.types+72160(SB), CX
  0x1099523		48894c2440		MOVQ CX, 0x40(SP)
  0x1099528		4889442448		MOVQ AX, 0x48(SP)
	return Fprint(os.Stdout, a...)
  0x109952d		488b05c4ea0d00		MOVQ os.Stdout(SB), AX
  0x1099534		488d0d251b0500		LEAQ go.itab.*os.File,io.Writer(SB), CX
  0x109953b		48890c24		MOVQ CX, 0(SP)
  0x109953f		4889442408		MOVQ AX, 0x8(SP)
  0x1099544		488d442440		LEAQ 0x40(SP), AX
  0x1099549		4889442410		MOVQ AX, 0x10(SP)
  0x109954e		48c744241801000000	MOVQ $0x1, 0x18(SP)
  0x1099557		48c744242001000000	MOVQ $0x1, 0x20(SP)
  0x1099560		e8cb97ffff		CALL fmt.Fprint(SB)
  0x1099565		488b6c2450		MOVQ 0x50(SP), BP
  0x109956a		4883c458		ADDQ $0x58, SP
  0x109956e		c3			RET
func (concrete) f() {
  0x109956f		e8ac7ffbff		CALL runtime.morestack_noctxt(SB)
  0x1099574		e947ffffff		JMP main.concrete.f(SB)

  0x1099579		cc			INT $0x3
  0x109957a		cc			INT $0x3
  0x109957b		cc			INT $0x3
  0x109957c		cc			INT $0x3
  0x109957d		cc			INT $0x3
  0x109957e		cc			INT $0x3
  0x109957f		cc			INT $0x3

TEXT main.main(SB) /Users/mvalais/code/go-atomic-issue-wallez-tweet/main.go
func main() {
  0x1099580		65488b0c2530000000	MOVQ GS:0x30, CX
  0x1099589		483b6110		CMPQ 0x10(CX), SP
  0x109958d		7631			JBE 0x10995c0
  0x109958f		4883ec10		SUBQ $0x10, SP
  0x1099593		48896c2408		MOVQ BP, 0x8(SP)
  0x1099598		488d6c2408		LEAQ 0x8(SP), BP
	v.f()
  0x109959d		488d059c1a0500		LEAQ go.itab.*main.concrete,main.abstract(SB), AX
  0x10995a4		8400			TESTB AL, 0(AX)
  0x10995a6		488d05eba40f00		LEAQ runtime.zerobase(SB), AX
  0x10995ad		48890424		MOVQ AX, 0(SP)
  0x10995b1		e81a000000		CALL main.(*concrete).f(SB)
}
  0x10995b6		488b6c2408		MOVQ 0x8(SP), BP
  0x10995bb		4883c410		ADDQ $0x10, SP
  0x10995bf		c3			RET
func main() {
  0x10995c0		e85b7ffbff		CALL runtime.morestack_noctxt(SB)
  0x10995c5		ebb9			JMP main.main(SB)

  0x10995c7		cc			INT $0x3
  0x10995c8		cc			INT $0x3
  0x10995c9		cc			INT $0x3
  0x10995ca		cc			INT $0x3
  0x10995cb		cc			INT $0x3
  0x10995cc		cc			INT $0x3
  0x10995cd		cc			INT $0x3
  0x10995ce		cc			INT $0x3
  0x10995cf		cc			INT $0x3

TEXT main.(*concrete).f(SB) <autogenerated>

  0x10995d0		65488b0c2530000000	MOVQ GS:0x30, CX
  0x10995d9		483b6110		CMPQ 0x10(CX), SP
  0x10995dd		7631			JBE 0x1099610
  0x10995df		4883ec08		SUBQ $0x8, SP
  0x10995e3		48892c24		MOVQ BP, 0(SP)
  0x10995e7		488d2c24		LEAQ 0(SP), BP
  0x10995eb		488b5920		MOVQ 0x20(CX), BX
  0x10995ef		4885db			TESTQ BX, BX
  0x10995f2		7523			JNE 0x1099617
  0x10995f4		48837c241000		CMPQ $0x0, 0x10(SP)
  0x10995fa		740e			JE 0x109960a
  0x10995fc		e8bffeffff		CALL main.concrete.f(SB)
  0x1099601		488b2c24		MOVQ 0(SP), BP
  0x1099605		4883c408		ADDQ $0x8, SP
  0x1099609		c3			RET
  0x109960a		e8c1ddf6ff		CALL runtime.panicwrap(SB)
  0x109960f		90			NOPL
  0x1099610		e80b7ffbff		CALL runtime.morestack_noctxt(SB)
  0x1099615		ebb9			JMP main.(*concrete).f(SB)
  0x1099617		488d7c2410		LEAQ 0x10(SP), DI
  0x109961c		48393b			CMPQ DI, 0(BX)
  0x109961f		75d3			JNE 0x10995f4
  0x1099621		488923			MOVQ SP, 0(BX)
  0x1099624		ebce			JMP 0x10995f4
```

```go
package main

import (
	"fmt"
	"strconv"
)

type abstract interface {
	f()
}

type concrete struct{}

func (concrete) f() {
	i := 0
	i++
	fmt.Print(strconv.Itoa(i))
}

func main() {
	var v1 abstract = &concrete{}
	v1.f()

	v2 := concrete{}
	v2.f()
}
```

## Review of what happens in `v1.f()`

```s
  v1.f()
  0x109959d		488d059c1a0500		LEAQ go.itab.*main.concrete,main.abstract(SB), AX
  0x10995a4		8400			TESTB AL, 0(AX)
  0x10995a6		488d05eba40f00		LEAQ runtime.zerobase(SB), AX
  0x10995ad		48890424		MOVQ AX, 0(SP)
  0x10995b1		e81a000000		CALL main.(*concrete).f(SB)
```

Let's read that line by line.

```s
  0x109959d		488d059c1a0500		LEAQ go.itab.*main.concrete,main.abstract(SB), AX
```

We first load the `tab` field of the `itab` struct into the address pointed by the stack pointer (SP). This `itab` refers to the field 'itab' of the iface struct (see [iface][] in runtime2.go):

```go
type iface struct {
  tab  *itab
  data unsafe.Pointer
}
```

[iface]: https://github.com/golang/go/blob/bf86aec25972f3a100c3aa58a6abcbcc35bdea49/src/runtime/runtime2.go#L143-L146

```s
0x10995a4		8400			TESTB AL, 0(AX)
```

Here, we test that AL, which contains

```s
0x10995a6		488d05eba40f00		LEAQ runtime.zerobase(SB), AX
0x10995ad		48890424		MOVQ AX, 0(SP)
```

Now that the `tab` address is contained in the address pointed by the SP, it's time to find out which concrete implementation of `f()` should be called:

```s
0x10995b1		e81a000000		CALL main.(*concrete).f(SB)
```

And here is the dispath function `main.(*concrete).f`:

```s
TEXT main.(*concrete).f(SB) <autogenerated>

0x10995d0		65488b0c2530000000	MOVQ GS:0x30, CX
0x10995d9		483b6110		CMPQ 0x10(CX), SP
0x10995dd		7631			JBE 0x1099610
```

That's the [stack-split prologue](https://cmc.gitbook.io/go-internals/chapter-i-go-assembly#splits) that deals with growing the stack when the goroutine doesn't have enough stack space. Basicall, `JBE` means 'jump to the stack-split epilogue if the value contained in `GS:0x30` is lower or equal to the SP (stack pointer). Since the stack grows backwards, meaning that SP will decrease as the stack grows).

```s
0x10995df		4883ec08		SUBQ $0x8, SP
0x10995e3		48892c24		MOVQ BP, 0(SP)
0x10995e7		488d2c24		LEAQ 0(SP), BP
0x10995eb		488b5920		MOVQ 0x20(CX), BX
0x10995ef		4885db			TESTQ BX, BX
0x10995f2		7523			JNE 0x1099617
0x10995f4		48837c241000		CMPQ $0x0, 0x10(SP)
0x10995fa		740e			JE 0x109960a
0x10995fc		e8bffeffff		CALL main.concrete.f(SB)
0x1099601		488b2c24		MOVQ 0(SP), BP
0x1099605		4883c408		ADDQ $0x8, SP
0x1099609		c3			RET
0x109960a		e8c1ddf6ff		CALL runtime.panicwrap(SB)
0x109960f		90			NOPL
0x1099610		e80b7ffbff		CALL runtime.morestack_noctxt(SB)
0x1099615		ebb9			JMP main.(*concrete).f(SB)
0x1099617		488d7c2410		LEAQ 0x10(SP), DI
0x109961c		48393b			CMPQ DI, 0(BX)
0x109961f		75d3			JNE 0x10995f4
0x1099621		488923			MOVQ SP, 0(BX)
0x1099624		ebce			JMP 0x10995f4
```

## Review of what happens in `v2.f()`

```s
	v2.f()
  0x10995b6		e805ffffff		CALL main.concrete.f(SB)
```

## Notes

### The AX and AL registers

From [intel64][] (Figure 3-5. Alternate General-Purpose Register Names, p. 3-12 Vol. 1):

```plain
| 0000 0001 0010 0011 0100 0101 0110 0111 | ------> EAX

|                     0100 0101 0110 0111 | ------> AX

|                               0110 0111 | ------> AL

|                     0100 0101           | ------> AH
```

(from a Stackoverflow [post](https://stackoverflow.com/a/52892696/3808537>))

From [intel64][] (p. 3-16, Vol. 1):

> AF (bit 4) Auxiliary Carry flag — Set if an arithmetic operation generates a carry or a borrow out of bit 3 of the result; cleared otherwise. This flag is used in binary-coded decimal (BCD) arithmetic.

### The GS register

In the x86-64 [intel64][] (p. 3-14, Vol. 1):

> The DS, ES, FS, and GS registers point to four data segments. The availability of four data segments permits efficient and secure access to different types of data structures. For example, four separate data segments might be created: one for the data structures of the current module, another for the data exported from a higher-level module, a third for a dynamically created data structure, and a fourth for data shared with another program. To access additional data segments, the application program must load segment selectors for these segments into the DS, ES, FS, and GS registers, as needed.

### The 'Q' suffix in MOVQ and LEAQ

I could not find that in the [Intel 64 PDF][intel64]. Apparently, these prefixes were introduced with the assembler (`as`). See binutils's as [i386mnemonics]:

> Instruction mnemonics are suffixed with one character modifiers which specify the size of operands. The letters ‘b’, ‘w’, ‘l’ and ‘q’ specify byte, word, long and quadruple word operands.

[i386mnemonics]: https://sourceware.org/binutils/docs/as/i386_002dMnemonics.html#Instruction-Naming

## What are SP and SB

From <https://golang.org/doc/asm#symbols>:

- FP: Frame pointer: arguments and locals.
- PC: Program counter: jumps and branches.
- SB: Static base pointer: global symbols.
- SP: Stack pointer: top of stack.

[plan9-assembler-manual]: https://9p.io/sys/doc/asm.html

## References

- go-internals book: <https://github.com/teh-cmc/go-internals>
- x86 Programming course: <https://courses.cs.washington.edu/courses/cse351/17wi/lectures/CSE351-L09-x86-II_17wi.pdf>

[intel64]: https://software.intel.com/sites/default/files/managed/39/c5/325462-sdm-vol-1-2abcd-3abcd.pdf "Intel 64's x86_64 specification"

```s
	v2.f()
  0x10995b6		e805ffffff		CALL main.concrete.f(SB)
```
