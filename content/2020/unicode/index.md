---
title: "Unicode"
description: ""
date: 2020-04-27T17:59:01+02:00
url: /unicode
images: [unicode/cover-unicode.png]
draft: true
tags: []
author: Maël Valais
---

I take a look at
"[whatisit](https://www.babelstone.co.uk/Unicode/whatisit.html)":

```plain
⚠️
U+26A0 : WARNING SIGN
U+FE0F : VARIATION SELECTOR-16 [VS16] {emoji variation selector}
```

But even better: [runes](https://github.com/mna/runes) is CLI for exploring
these emojis! Let's install it:

```sh
go get github.com/mna/runes
```

Now, let's see with an emoji that I know displays correctly in my terminal
(iTerm2):

```sh
% runes - ✅
[S So] U+2705 '✅'    [E2 9C 85]    [2705]      WHITE HEAVY CHECK MARK
[C Cc] U+0000         [00]          [0]         <control>
[C Cc] U+0000         [00]          [0]         <control>
```

Now, let's try with my multi-UTF-8 example "⚠️":

```sh
% runes - ⚠<fe0f>
[S So] U+26A0 '⚠'     [E2 9A A0]    [26A0]      WARNING SIGN
[C Cc] U+0000         [00]          [0]         <control>
[C Cc] U+0000         [00]          [0]         <control>
[M Mn] U+FE0F '️'     [EF B8 8F]    [FE0F]      VARIATION SELECTOR-16
[C Cc] U+0000         [00]          [0]         <control>
[C Cc] U+0000         [00]          [0]         <control>
```

My terminal (iTerm2) doesn't even allow me to paste "⚠️"?! It shows
`⚠<fe0f>` instead.


