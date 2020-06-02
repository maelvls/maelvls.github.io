---
title: "Slack: don't miss any channels!"
description: ""
date: 2020-05-08T10:37:20+02:00
url: /slack-dont-miss-channels
images: [slack-dont-miss-channels/cover-slack-dont-miss-channels.png]
draft: true
tags: []
author: Maël Valais
---

Given $U$ the set of users and a pair of users $(u, v) \in U^2$, we denote
by $f$ the function that tells us when two users belong to the same channel
$c \in C$ where $C$ is the set of channels. It is defined as

$$
f(u,v,c) = \begin{cases}1 & \text{ if } u \text{ and } v \text{ both in channel } c \\ 0 &\text{ otherwise. }\end{cases}
$$

Let us define the distance between two users, denoted $dist$, as

$$
dist(u, v) = \sum_{c \in C} f(u,v,c)\text{.}
$$

The optimization problem can be formulated as:

> Given a user $u$, we want to find the channel $c$ in which he does not
> already belong and which maximizes the chances of subscribing to an
> interesting channel. This assumes that the user has already subscribed to
> some channels and that some of these channels are 

For a given user $v$:

$$\max_{c \in C} \sum_{u \in U, u \ne v} dist(v, u)$$

