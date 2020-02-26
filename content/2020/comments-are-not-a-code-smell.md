---
title: Comments are a code smell but you should write some
description: ""
date: 2020-02-18
url: /you-should-write-comments
draft: true
---

So there was this [SO] question: "are comments a code smell".

[SO]:
https://softwareengineering.stackexchange.com/questions/1/comments-are-a-code-smell

Argumentation:

1. Writing code is an incremental process. At first, you end up with very
   complicated code with probably long functions that do not explain
   themselves. Through code review, refactoring and knowledge about the
   code, we can iteratively lower the amount of comments, e.g. by giving
   more self-explanatory names or by spliting complicated parts. Comments
   are proof that my code isn't perfect the first time, and I aknowledge it
   by commenting more.
2. We can't get names or proper function separation the first time around,
   and that's why commenting is important.
3. 5 seconds rule: any block that would take more than 5 seconds to get an
   idea what it is used for.
4. comments are disposable, it is important not to be afraid of deleting
   them. Remember that the end-goal is to eliminate all comments. We often
   ask for rephrasing, more details or deletion of certain comments during
   code review.
