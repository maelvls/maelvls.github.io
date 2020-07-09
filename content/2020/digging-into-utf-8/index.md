---
title: "Digging Into Utf 8"
description: ""
date: 2020-07-08T10:16:58+02:00
url: /digging-into-utf-8
images: [digging-into-utf-8/cover-digging-into-utf-8.png]
draft: true
tags: []
author: Maël Valais
---

Not sure if this addresses your issue, but symbol characters (https://copychar.cc/symbols/) are useful where emojis aren't always compatible, and will match your color:
- ✓
- ✗

that's an excellent point, ✗ and ✓ (which are UTF-8 "symbols" as opposed to "emojis") can be displayed in pretty much all the fonts

What I could do:
- red color + `✗` (✗ = U+2717 : BALLOT X, I looked it up [here](https://www.babelstone.co.uk/Unicode/whatisit.html))
- green color + `✓` (✓ = U+2713 = CHECK MARK)

But since today's terminal emulators all support UTF-8 emojis, I prefer keeping my emojis 😅


---

Using the macOS "emoji picker" (⌃⌘+space), I realized that Unicode and UTF-8
are different. For example, searching for "space" would yield:

|----------------|-----------------|-----------------|
| NO-BREAK SPACE | Unicode: U+00A0 | UTF-8: C2 A0    |
| EN SPACE       | Unicode: U+2002 | UTF-8: E2 80 82 |
| EM SPACE       | Unicode: U+2003 | UTF-8: E2 80 83 |

What is `U+XXXX`?? And what is `C2 A0`??

And why does the author of [this stackoverflow question][em-space] say that EM
SPACE is "8195 in UTF-8" since UTF-8 is usually given in hexadecimal format?

[em-space]: https://stackoverflow.com/a/58532995/3808537
