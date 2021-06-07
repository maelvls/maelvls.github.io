---
title: "Writing useful comments"
description: "We often talk about avoiding unecessary comments that needlessly paraphrase what the code does. In this article, I gathered some thoughts about why writing comments is as important as writing the code itself, and how to spot comments that should be refactored using the 'what' and the 'why'."
date: 2021-06-05
url: "/writing-useful-comments"
images:
  - writing-useful-comments/cover-writing-useful-comments.png
tags:
  - go
  - software-engineering
aliases:
  - /you-should-write-comments
author: Maël Valais
devtoId: 313912
devtoPublished: true
draft: false
devtoUrl: https://dev.to/maelvls/you-should-write-comments-36fd
---

In his book _[Clean Code](https://www.oreilly.com/library/view/clean-code-a/9780136083238/)_, Robert C. Martin makes a strong case against comments:

> You don't need comments if you write clean code.

Like any blanket statement, Martin is partially true. Although code clarity increases readability and decreases the need for comments, code has never been able to convey the whole context to the reader.

Programming patterns are often cited as a way to convey this context on a codebase scale. Well named functions and variables also convey some form of context. But most of the time, the context (the "why") never gets written anywhere.

In this article, I present a few examples meant to showcase how comments may be written to convey the context around code.

**Contents:**

1. [Code example 1](#code-example-1)
2. [Code example 2](#code-example-2)
3. [Code example 3](#code-example-3)
4. [Code example 4](#code-example-4)
5. [Conclusion](#conclusion)

## Code example 1

I call "in-flight comments" comments that we write to ease the process of writing code. In-flight comments help us articulate what we are trying to achieve.

In the following example taken from the [Google's Technical Writing](https://developers.google.com/tech-writing/two/code-comments), the developer wrote a comment as they were writing an algorithm for randomly shuffling a slice of integers:

```go
func Shuffle(slice []int) {
  curIndex := len(slice)

    // If the current index is 0, return directly.
    if (curIndex == 0) {
        return
    }

    ...
}
```

This comment is a typical in-flight comment: it only focuses on what the code does. This comment does not give context as to why this `if` statement exists. We can refactor this comment by focusing on the "why":

```go
func Shuffle(slice []int) {
  curIndex := len(slice)

    // No need to shuffle an array with zero elements.
    if (curIndex == 0) {
      return
    }

    ...
}
```

The reader now understand why this `if` exists, and does not have to dig further. Starting with the "why" helps glanceability: the reader only needs to read the first few words to get the idea.

Here is how I would diffenciate the "why" from the "what":

| Content in a comment | Description                                                  |
| -------------------- | ------------------------------------------------------------ |
| The "why"            | Helps you understand how this piece of code came to life.    |
| The "what"           | Paraphrases\* what is being done in a human-readable format. |

\*Sometimes, the "what" may be valuable for readability purposes. I see three reasons to comment on the "what":

1. When the code is not self-explanatory, a "what" comment may avoid the reader the struggle of googling;
2. When a block of code is lenghy, adding "what" comments may help create "sections", helping the reader quicly find the part they are interested in.

## Code example 2

Our next example comes from the cert-manager codebase:

```go
// If the certificate request has been denied, set the last failure time to
// now, and set the Issuing status condition to False with reason. We only
// perform this check if the request also doesn't have a Ready condition,
// since some issuers may not honor a Denied condition, and will sign and
// set the Ready condition to True anyway. We would still want to complete
// issuance for requests where the issuer doesn't respect approval.
cond := apiutil.GetCertificateRequestCondition(req, cmapi.CertificateRequestConditionReady)
if cond == nil {
    if apiutil.CertificateRequestIsDenied(req) {
        return c.failIssueCertificate(ctx, log, crt, apiutil.GetCertificateRequestCondition(req, cmapi.CertificateRequestConditionDenied))
    }
    return nil
}
```

After looking at the `if` statement, the reader wonders: why do we need this `if` statement?

Their first reaction will probably be to take a look at the comment right above. Unfortunately, the comment starts with even more confusing implementation details. The reader has to keep reading until the 6th line to find out why the code is doing what it is doing.

Let us rate each individual information that the comment conveys:

| Paragraph                                                                                                                                        | Usefulness |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ---------- |
| **(A)** If the certificate request has been denied, set the last failure time to now, and set the Issuing status condition to False with reason. | ⭐         |
| **(B)** We only perform this check if the request also doesn't have a Ready condition.                                                           | ⭐⭐       |
| **(C)** Some issuers may not honor a Denied condition, and will sign and set the Ready condition to True anyway.                                 | ⭐⭐⭐     |
| **(D)** We would still want to complete issuance for requests where the issuer doesn't respect approval.                                         | ⭐⭐⭐⭐   |

From the reader's perspective, the paragraphs (D) and (C) are the most useful. They give a high-level overview of why this `if` statement exists. We can merge the two paragraphs into one:

> **(C, D)** External issuers may ignore the approval API. An issuer ignores the approval API when it proceeds with the issuance even though the "Denied=True" condition is present on the CertificateRequest. To avoid breaking the issuers that are ignoring the approval API, we want to detect when CertificateRequests has ignored the Denied=True condition; when it is the case, we skip bubbling-up the (possible) Denied condition.

Paragraphs (B) and (A) may also be useful: they give the context about paragraph (C). The sentence lacks precision though, and "this check" may confuse the reader. Let us rephrase it:

> **(A, B)** To know when the Denied=True condition was ignored by the issuer, we look at the CertificateRequest's Ready condition. If both the Ready and Denied

```go
// Some issuers won't honor the "Denied=True" condition, and we don't want
// to break these issuers. To avoid breaking these issuers, we skip bubbling
// up the "Denied=True" condition from the certificate request object to the
// certificate object when the issuer ignores the "Denied" state.
//
// To know whether or not an issuer ignores the "Denied" state, we pay
// attention to the "Ready" condition on the certificate request. If a
// certificate request is "Denied=True" and that the issuer still proceeds
// to adding the "Ready" condition (to either true or false), then we
// consider that this issuer has ignored the "Denied" state.
cond := apiutil.GetCertificateRequestCondition(req, cmapi.CertificateRequestConditionReady)
if cond == nil {
    if apiutil.CertificateRequestIsDenied(req) {
        return c.failIssueCertificate(ctx, log, crt, apiutil.GetCertificateRequestCondition(req, cmapi.CertificateRequestConditionDenied))
    }
    return nil
}
```

## Code example 3

Another example [inspired](https://github.com/jetstack/cert-manager/blob/1dad685e/pkg/controller/ingress-shim/sync.go#L145) by the cert-manager codebase:

```go
errs := validateIngressTLSBlock(tls)
// if this tls entry is invalid, record an error event on Ingress object and continue to the next tls entry
if len(errs) > 0 {
    rec.Eventf(ingress, "Warning", "BadConfig", fmt.Sprintf("TLS entry %d is invalid: %s", i, errs))
    continue
}
```

Once again, the in-flight comment can be rewritten to focus on the "why":

```go
// Let the user know that an TLS entry has been skipped due to being invalid.
errs := validateIngressTLSBlock(tls)
if len(errs) > 0 {
    rec.Eventf(ingress, "Warning", "BadConfig", fmt.Sprintf("TLS entry %d is invalid: %s", i, errs))
    continue
}
```

## Code example 4

The next example [is inspired](https://github.com/jetstack/cert-manager/blob/1dad685e4/pkg/controller/ingress-shim/sync.go#L180-L209) by an other place in cert-manager.

This is another example of in-flight comment focusing on the "what". The context around this block of code is not obvious, which means this comment should refactored to focus on the "why".

```go
// check if a Certificate for this secret name exists, and if it
// does then skip this secret name.
expectedCrt := expectedCrt(ingress)
existingCrt, _ := client.Certificates(namespace).Get(ingress.secretName)
if existingCrt != nil {
    if !certMatchesUpdate(existingCrt, crt) {
        continue
    }

    toBeUpdated = append(toBeUpdated, updateToMatchExpected(expectedCrt))
} else {
    toBeCreated = append(toBeCreated, crt)
}
```

In this case, the code itself would benefit from a bit of refactoring. Regarding the comment, we start with the "why":

```go
// The secretName has an associated Certificate that has the same name. We want
// to make sure that this Certificate exists and that it matches the expected
// Certificate spec.
existingCrt, _ := client.Certificates(namespace).Get(secretName)
expectedCrt := expectedCrt()

if existingCrt == nil {
    toBeCreated = append(toBeCreated, expectedCrt)
    continue
}

if certMatchesExpected(existingCrt, expectedCrt) {
    continue
}

toBeUpdated = append(toBeUpdated, updateToMatchExpected(expectedCrt))
```

Note that we removed the `else` statement for the purpose of readability. The happy path is now clearly "updating the certificate".

## Conclusion

As a developer, I want to write code that is readable and maintainable, and keeping track of the "why" (i.e., the context) in a codebase is essential for readability and maintainability. I found that putting myself in the place of the future reader helps me write comments that focus on the "why".

HAProxy, Git or the Linux kernel are great examples of projects where there is a great focus on well-documented code:

- [`ebtree/ebtree.h`](https://github.com/haproxy/haproxy/blob/530408f976e5fe2f2f2b4b733b39da36770b566f/ebtree/ebtree.h#L23) (HAProxy)
- [`unpack-trees.c`](https://github.com/git/git/blob/2d2118b814c11f509e1aa76cb07110f7231668dc/unpack-trees.c#L821-L836) (Git)
- [`kernel/sched/core.c`](https://github.com/torvalds/linux/blob/bfdc6d91a25f4545bcd1b12e3219af4838142ef1/kernel/sched/core.c#L157-L171) (Linux kernel).

<!--

## Discussion

Bill Kennedy [shared](https://changelog.com/gotime/172) the importance of making a concious effort towards having a common set of values that help a team write good software. Like anything, the team probably needs to find their own set of values, but the two values that Bill talked about in the Gotime episode seem valuable to me:

1. Be precise. Letting uncertainty and inprecision creep into variable names, function names and package mames greatly decreases the chance for someone to understand the concepts captured by your code.
2. Don’t make things easy to do; instead, make things easy to understand. That's an other
3. Early use of abstractions and other sorts of indirection hurt.

In my mind, being precise also applies to comments. Precision in comments avoid the reader having to wonder "is that what they really meant?". I think that the values and ideas contained in a design philosophy not only apply to code, but also to comments.

### Good naming takes time

Finding a good name that properly carries the exact intent and help self-documenting takes time. It always takes a few iterations before the code becomes self-documenting enough to be able to remove comments.

![Number of comments lowers with time](chart-comments-over-time.svg)

As shown by the diagram, the amount of comments for a given codebase decreases thanks to PR reviews and refactorings. The more we learn and understand our codebase, the better we become at self-documenting.

### Comments are disposable, don't copy them over

Now, let's talk about the issue of comments becoming outdated. Over time, comments start lying. That's where my second point comes into play: deleting and rewriting comments is part of our job. I would even say that it takes around 40% of my time spent coding.

During code reviews, I pay extra attention to these comments. And yes, very often, comments don't make sense anymore because of some copy-paste of code. We must delete any copy-pasted comment and rewrite it. Spreading copy-pasted comments that we don't really know why they were added in the first place is a plague.

## Conclusion

As a final word, I want us to remember that yes, maintaining comments is a pain. Comments will eventually start lying if you don't delete and rewrite them. But would you rather have no comments at all and let the amount of tribal knowledge creep in every part of your codebase, making it harder and harder for new engineers to join the team?

-->

<!--
Join the discussion on Twitter:

{{< twitter 1233140530017644544 >}}
\-->
