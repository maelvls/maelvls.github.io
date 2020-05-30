---
title: Notes on OAuth
date: 00-00-00
---

I know two main ways of using OAuth2

- password-based client grant (2-leg oauth flow: on the project I worked on,
  the OAuth client was not third party server, but instead, it was the front-end.)

- code-based grant (3-legs oauth flow)


## Terminology

Two secrets:
- signing key = the private key