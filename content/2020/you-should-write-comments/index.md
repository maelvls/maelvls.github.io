---
title: You should write comments
description: "We often talk about avoiding unecessary comments that needlessly paraphrase what the code does. In this article, I gathered some thoughts about why writing comments is as important as writing the code itself."
date: 2021-04-15
url: "/you-should-write-comments"
images:
  - you-should-write-comments/cover-you-should-write-comments.png
tags:
  - go
  - software-engineering
aliases:
  - /you-should-write-comments
author: Maël Valais
devtoId: 313912
devtoPublished: false
draft: true
---

In his book [Clean Code](https://www.oreilly.com/library/view/clean-code-a/9780136083238/), Robert C. Martin makes a strong case against comments:

> You don't need comments if you write clean code.

This short sentence seems to have created a "no comment" movement; on [stackoverflow](https://softwareengineering.stackexchange.com/questions/1/comments-are-a-code-smell), we can read that comments are a "code smell", that they are adding "noise" and are often out of sync with the code. This movement seems to have grown from the misinterpration of Robert's sentence, which started as a way of saying that some comments (the "what", as we will see below), are not useful to the reader. To some extent, the sentence has become:

> If you write any kind of comment, it means you code isn't clean.

I think that this confusion around what comments are good for is unfortunate. Comments can be extremely valuable, and I want to prove it.

We usually write comments as a way to guide our mind through the intense process of writing code. These "initial comments" often serve as a to-do list, and naturally reflect the way we program. "Initial comments" often start with "if" and are focused on the "what".

The concerning part about these "initial comments" is that they often end up commited as-is.

In this article, using four different examples, I will show how to refactor these "initial comments" into useful comments.

#### Example 1

In the following example taken from [Google's Technical Writing](https://developers.google.com/tech-writing/two/code-comments) book, the developer had to shuffle some integers:

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

The comment focus on the "what" and paraphrases what immediately follows. The comment does not help the reader, since the logic itself is trivial. Refactoring this comment consists in focusing on the "why":

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

The "why" brings more value to the reader because the first instinctive question the reader would have:

> Why on earth are we returning now?

Starting comments with the "why" helps glanceability: the reader only needs to read the first few words to get the idea. Here is how I would diffenciate the "why" from the "what":

| Content in a comment | Description                                               |
| -------------------- | --------------------------------------------------------- |
| The "why"            | Helps you understand how this piece of code came to life. |
| The "what"           | Paraphrases what code does in a human-readable format.\*  |

\*sometimes, commenting on the "what" can still be valuable for readability purposes. I see three reasons to comment on the "what":

1. When the code is not self-explanatory, a "what" comment may avoid the reader the struggle of googling;
2. When a block of code is lenghy, adding "what" comments may help create "sections", helping the reader quicly find the part they are interested in.

#### Example 2

The next example [comes](https://github.com/jetstack/cert-manager/pull/3872) from cert-manager. The original comment focuses on the "what":

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

The first few words are paraphrasing what the code does, and the "why" is buried at the 6th line. We often naturally end up with this form of comment while programming: it helps us lay out the logic.

To refactor this comment, we want to enphasise on the "why". Notice that the "what" has not totally disappeared, but has been moved to the end of the comment. In this example, the code itself is not trivial due to the nesting of two conditions:

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

#### Example 3

The next [example](https://github.com/jetstack/cert-manager/blob/1dad685e/pkg/controller/ingress-shim/sync.go#L145) comes from another part of the cert-manager codebase:

```go
errs := validateIngressTLSBlock(tls)
// if this tls entry is invalid, record an error event on Ingress object and continue to the next tls entry
if len(errs) > 0 {
    errMsg := utilerrors.NewAggregate(errs).Error()
    c.recorder.Eventf(ing, corev1.EventTypeWarning, "BadConfig", fmt.Sprintf("TLS entry %d is invalid: %s", i, errMsg))
    continue
}
```

Let's refactor the comment by focusing on the "why":

```go
// Let the user know that an TLS entry has been skipped due to being invalid.
errs := validateIngressTLSBlock(tls)
if len(errs) > 0 {
    errMsg := utilerrors.NewAggregate(errs).Error()
    c.recorder.Eventf(ing, corev1.EventTypeWarning, "BadConfig", fmt.Sprintf("TLS entry %d is invalid: %s", i, errMsg))
    continue
}
```

#### Example 4

The next example [comes](https://github.com/jetstack/cert-manager/blob/1dad685e4/pkg/controller/ingress-shim/sync.go#L180-L209) from the same file as the previous example:

```go
// check if a Certificate for this TLS entry already exists, and if it
// does then skip this entry
if existingCrt != nil {
    log := logs.WithRelatedResource(log, existingCrt)
    log.V(logf.DebugLevel).Info("certificate already exists for ingress resource, ensuring it is up to date")

    if metav1.GetControllerOf(existingCrt) == nil {
        log.V(logf.InfoLevel).Info("certificate resource has no owner. refusing to update non-owned certificate resource for ingress")
        continue
    }

    if !metav1.IsControlledBy(existingCrt, ing) {
        log.V(logf.InfoLevel).Info("certificate resource is not owned by this ingress. refusing to update non-owned certificate resource for ingress")
        continue
    }

    if !certNeedsUpdate(existingCrt, crt) {
        log.V(logf.DebugLevel).Info("certificate resource is already up to date for ingress")
        continue
    }

    updateCrt := existingCrt.DeepCopy()

    updateCrt.Spec = crt.Spec
    updateCrt.Labels = crt.Labels
    setIssuerSpecificConfig(updateCrt, ing)
    updateCrts = append(updateCrts, updateCrt)
} else {
    newCrts = append(newCrts, crt)
}
```

As in the previous examples, refactoring this example is about moving the "why" to the start of the comment:

```go
// The "secretName" field may already refer to an existing Certificate object.
// We still want this existing Certificate to be .
//
// For example, imagine that the certificate named "example-tls" already exists.
//
//     kind: Ingress
//     spec:
//       tls:
//         - secretName: example-tls
//           hosts: [example.com]
//
if existingCrt != nil {
    log := logs.WithRelatedResource(log, existingCrt)
    log.V(logf.DebugLevel).Info("certificate already exists for ingress resource, ensuring it is up to date")

    if metav1.GetControllerOf(existingCrt) == nil {
        log.V(logf.InfoLevel).Info("certificate resource has no owner. refusing to update non-owned certificate resource for ingress")
        continue
    }

    if !metav1.IsControlledBy(existingCrt, ing) {
        log.V(logf.InfoLevel).Info("certificate resource is not owned by this ingress. refusing to update non-owned certificate resource for ingress")
        continue
    }

    if !certNeedsUpdate(existingCrt, crt) {
        log.V(logf.DebugLevel).Info("certificate resource is already up to date for ingress")
        continue
    }

    updateCrt := existingCrt.DeepCopy()

    updateCrt.Spec = crt.Spec
    updateCrt.Labels = crt.Labels
    setIssuerSpecificConfig(updateCrt, ing)
    updateCrts = append(updateCrts, updateCrt)
}

newCrts = append(newCrts, crt)
```

Note that I removed the `else` statement for the purpose of readability. Since "creating a certificate" seems to be the happy path of this function, it makes sense to unindent the code that relates to this happy path.

<!--
### Example 5

Let us study a case of lack of comments in the same file as above. The following [snippet](https://github.com/jetstack/cert-manager/blob/1dad685e/pkg/controller/ingress-shim/sync.go#L84-L90) does an update of a Certificate object, but this object does not seem to have been changed:

```go
for _, crt := range updateCrts {
    _, err := c.cmClient.CertmanagerV1().Certificates(crt.Namespace).Update(ctx, crt, metav1.UpdateOptions{})
    if err != nil {
        return err
    }
    c.recorder.Eventf(ing, corev1.EventTypeNormal, "UpdateCertificate", "Successfully updated Certificate %q", crt.Name)
}
```

From my experience, we often do this when we want to trigger a re-sync of the object, so that's my hypothesis on the "what". But I have no idea "why": why do we need to re-sync the Certificate immediately? What change are we expecting to happen when we do?

A comment would have helped me realize that

```go
for _, crt := range updateCrts {
    _, err := c.cmClient.CertmanagerV1().Certificates(crt.Namespace).Update(ctx, crt, metav1.UpdateOptions{})
    if err != nil {
        return err
    }
    c.recorder.Eventf(ing, corev1.EventTypeNormal, "UpdateCertificate", "Successfully updated Certificate %q", crt.Name)
}
```

-->

### From design philosophy to useful comments

Bill Kennedy [shared](https://changelog.com/gotime/172) the importance of making a concious effort towards having a common set of values that help a team write good software. Like anything, the team probably needs to find their own set of values, but the two values that Bill talked about in the Gotime episode seem valuable to me:

1. Be precise. Letting uncertainty and inprecision creep into variable names, function names and package mames greatly decreases the chance for someone to understand the concepts captured by your code.
2. Don’t make things easy to do; instead, make things easy to understand. That's an other
3. Early use of abstractions and other sorts of indirection hurt.

In my mind, being precise also applies to comments. Precision in comments avoid the reader having to wonder "is that what they really meant?". I think that the values and ideas contained in a design philosophy not only apply to code, but also to comments.

Comments are one of these topics that we seem to never agree on ([stackoverflow](https://softwareengineering.stackexchange.com/questions/1/comments-are-a-code-smell) says so). I often hear that comments add noise to the code and that comments never get properly updated.

The solution seems to be to properly self-document code. And I love Go for that; one of the Go proverbs even [says](https://go-proverbs.github.io/):

> Clear is better than clever.

It is true that Go favors easy-to-read code rather than smart-but-hard-to-parse code, which really helps to keep the level of comments low. But there are still many reasons comments are still needed.

### Fighting tribal knowledge

Excellent naming and structure cannot help with the "why" of a piece of code. Complex math operations is a typical example of code that needs to be thougouly commented. Teams want to keep the amount of "tribal knowledge" as low as possible, and the only way in this kind of situation is to write comments. Take a look at [this](https://github.com/haproxy/haproxy/blob/530408f976e5fe2f2f2b4b733b39da36770b566f/include/proto/freq_ctr.h#L138-L248) for example.

### Good naming takes time

Finding a good name that properly carries the exact intent and help self-documenting takes time. It always takes a few iterations before the code becomes self-documenting enough to be able to remove comments.

![Number of comments lowers with time](chart-comments-over-time.svg)

As shown by the diagram, the amount of comments for a given codebase decreases thanks to PR reviews and refactorings. The more we learn and understand our codebase, the better we become at self-documenting.

### Comments are disposable, don't copy them over

Now, let's talk about the issue of comments becoming outdated. Over time, comments start lying. That's where my second point comes into play: deleting and rewriting comments is part of our job. I would even say that it takes around 40% of my time spent coding.

During code reviews, I pay extra attention to these comments. And yes, very often, comments don't make sense anymore because of some copy-paste of code. We must delete any copy-pasted comment and rewrite it. Spreading copy-pasted comments that we don't really know why they were added in the first place is a plague.

---

As a final word, I want us to remember that yes, maintaining comments is a pain. Comments will eventually start lying if you don't delete and rewrite them. But would you rather have no comments at all and let the amount of tribal knowledge creep in every part of your codebase, making it harder and harder for new engineers to join the team?

I think we all enjoy reading good comments. Well commented code is an art. Some projects like HAProxy, Git or the Linux kernel have done an amazing job at keeping knowledge accessible as opposed to knowledge locked in and scattered across many brains. Just take a look at [`ebtree/ebtree.h`](https://github.com/haproxy/haproxy/blob/530408f976e5fe2f2f2b4b733b39da36770b566f/ebtree/ebtree.h#L23) (HAProxy), [`unpack-trees.c`](https://github.com/git/git/blob/2d2118b814c11f509e1aa76cb07110f7231668dc/unpack-trees.c#L821-L836) (Git) and [`kernel/sched/core.c`](https://github.com/torvalds/linux/blob/bfdc6d91a25f4545bcd1b12e3219af4838142ef1/kernel/sched/core.c#L157-L171) (Linux kernel). Absolute delights.

<!--
Join the discussion on Twitter:

{{< twitter 1233140530017644544 >}}
\-->
