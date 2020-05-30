---
title: Stuff about managing Ruby & rvm
date: 00-00-00
---

## Warning 'Ignoring... because its extensions are not built'

The message is:

    Ignoring atomic-1.1.101 because its extensions are not built.  Try: gem pristine atomic --version 1.1.101

Try:

    gem 2>&1 | perl -ne '/Try: gem pristine ([^ ]+)/ && print $1 . " "' | xargs gem pristine
