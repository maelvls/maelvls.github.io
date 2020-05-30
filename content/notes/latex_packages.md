---
title: Managing Basic Texlive instead of Full Texlive/Mactex
date: 2017-05-25
---

```shell
tlmgr install texliveonfly
texliveonfly main.tex
```

will install automatically all packages called using `\usepackage{}`

To find a specific file that latex yells it cannot find:

```shell
tlmgr search --global --file fullpage.sty
```
