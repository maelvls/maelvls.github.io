---
title: You should write comments
description: |
   We often talk about avoiding unecessary comments that needlessly paraphrase
   what the code does. In this article, I gathered some thoughts about why
   writing comments is as important as writing the code itself.
date: 2020-02-27
url: /you-should-write-comments
images: [you-should-write-comments/cover-you-should-write-comments.png]
tags: [go, software-engineering]
---

Comments is one of these topics that we seem to never agree on
([stackoverflow](https://softwareengineering.stackexchange.com/questions/1/comments-are-a-code-smell)
says so). I often hear that comments add noise to the code and that
comments never get properly updated.

The solution seems to be to properly self-document code. And I love Go for
that; one of the Go proverbs even [says](https://go-proverbs.github.io/):

> Clear is better than clever.

It is true that Go favors easy-to-read code rather than
smart-but-hard-to-parse code, which really helps keeping the level of
comments low. But there are still many reasons comments are still needed.

### Fighting tribal knowledge

Excellent naming and structure cannot help with the "why" of a piece of
code. Complex math operations is a typical example of code that needs to be
thougouly commented. Teams want to keep the amount of "tribal knowledge" as
low as possible, and the only way in this kind of situation is to write
comments. Take a look at
[this](https://github.com/haproxy/haproxy/blob/530408f976e5fe2f2f2b4b733b39da36770b566f/include/proto/freq_ctr.h#L138-L248)
for example.

### Good naming takes time

Finding a good name that properly carries the
exact intent and help self-documenting takes time. It always takes a few
iterations before the code becomes self-documenting enough to be able to
remove comments.

![Number of comments lowers with time](chart-comments-over-time.svg)

As shown by the diagram, the amount of comments for a given code base
decreases thanks to PR reviews and refactorings. The more we learn and
understand our code base, the better we become at self-documenting.

### Comments are disposable, don't copy them over

Now, let's talk about the issue of comments becoming outdated. Over time,
comments start lying. That's where my second point comes into play:
deleting and rewriting comments is part of our job. I would even say that
it takes around 40% of my time spent coding.

During code reviews, I pay extra attention to these comments. And yes, very
often, comments don't make sense anymore because of some copy-paste of
code. We must delete any copy-pasted comment and rewrite it. Spreading
copy-pasted comments that we don't really know why they were added in the
first place is a plague.

---

As a final work, I want us to remember that yes, maintaining comments is a
pain. Comments will eventually start lying if you don't delete and rewrite
them. But would you rather have no comments at all and let the amount of
tribal knowledge creep in every part of your code base, making it harder
and harder for new engineers to join the team?

I think we all enjoy reading good comments. Well commented code is an art.
Some project like HAProxy, Git or the Linux kernel have done an amazing job
at keeping knowledge accessible as opposed to knowledge locked in and
scattered across many brains. Just take a look at
[`ebtree/ebtree.h`](https://github.com/haproxy/haproxy/blob/530408f976e5fe2f2f2b4b733b39da36770b566f/ebtree/ebtree.h#L23),
[`unpack-trees.c`](https://github.com/git/git/blob/2d2118b814c11f509e1aa76cb07110f7231668dc/unpack-trees.c#L821-L836)
and
[`kernel/sched/core.c`](https://github.com/torvalds/linux/blob/bfdc6d91a25f4545bcd1b12e3219af4838142ef1/kernel/sched/core.c#L157-L171).
Absolute delight.

---

Join the discussion on Twitter:

{{< twitter 1233140530017644544 >}}
