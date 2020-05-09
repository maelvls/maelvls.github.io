---
title: "ChatOps using Github Actions"
description: ""
date: 2020-05-08T14:40:31+02:00
url: /chatops-github-actions
images: [chatops-github-actions/cover-chatops-github-actions.png]
draft: true
tags: []
---

- ChatOps takes the GitOps to the next level and uses a chat interface
  (such as Slack or Github PR comments) to bring more transparency and
  retain knowledge about past actions.
- The Kubernetes project uses Prow, a chat bot.
- Chat UI = transparent, visible, avoids shadow operations (when a
  colleague has to do some `terraform apply` locally), workflow is public
  instead of in people's minds.
- Tekton = tasks & pipelines only but deals with inputs and outputs.
  API-only which means there is it kind of lacks user-friendliness. You
  only get the `tekton` CLI and a read-only web-based dashboard, but the
  nice thing is that it is Kubernetes-based so it integrates with many
  things like `kubectl`.
- Concourse CI = also input-output oriented. Has a very nice web UI that is
  mostly read-only (except for re-running jobs). It is API-only and the
  main way of configuring and running CI is to use the `fly` CLI.

Example: <https://github.com/maelvls/gh-actions-chatops>
