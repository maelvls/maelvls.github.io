---
title: Memo for Tmux
date: 2018-11-03
tags: []
author: Maël Valais
devtoSkip: true
---

I installed <https://gist.github.com/cybic/6375457> in `~/.bash_completion` (file) to get the bash completion.

## General shortcuts

    tmux attach          -> opens tmux
    tmux ls              -> lists open windows
    tmux attach -d 0     -> reopens window 0

Inside tmux. prefix = ctrl+b

    prefix d         -> detach window (closes tmux)
    prefix w         -> lists windows
    prefix x         -> close window
    prefix c         -> create window
    prefix ,         -> rename window
    prefix [         -> scroll among history
    prefix arrow     -> move among panes
    prefix "         -> split vertically (up-down); % for horiz
    prefix maj+arrow -> bigger pane
    prefix alt+arrow -> even bigger pane
    prefix z         -> bigger current pane (redo that to resize like before)

I also created a script called `monitor.sh` to know which one of the 6 server I should use for my experiments. The thing is that I don't want the session with 6 htop open to run in background for days.

    prefix ctrl+z    -> suspend the current session

To scroll stdout history instead of command history with mouse wheel:

    set -g mouse on

If I want to put an open pane to its own window:

    break-pane           # current pane is sent to a new window
    join-pane -s 0       # -s = source window
    join-pane -t 0       # -t = destination window

## Gray not displaying (zsh)

    <can't remember how to do that>
