---
title: "A proper prompt"
description: ""
date: 2020-04-20
url: /proper-prompt
draft: true
images: []
author: Maël Valais
---

- Go (powerline/bullettrain-like) https://github.com/jtyr/gbt
- Go (powerline-like): https://github.com/bullettrain-sh/bullettrain-go-core
- Rust: https://github.com/starship/starship
  (git not async: https://github.com/starship/starship/issues/301)

Specification:

- must be fast (pure zsh themes are just too slow)
- git should be [async](https://github.com/mafredri/zsh-async) like with
  [agkozak/agkozak-zsh-prompt](https://github.com/agkozak/agkozak-zsh-prompt).
- display the current kubernetes user & context & cluster
- display if I'm in a Go module or GOPATH
