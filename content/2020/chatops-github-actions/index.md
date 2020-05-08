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

## Appendix: and what if you have modules in private repos?

Here are some ways to make `terraform init` work when the module source is
a private Github repo, which means `git clone` will fail.

> ⚠️ Warning: around April 2020, Github decided to prevent users from using
> Github Secrets names that begin with `GITHUB_`. We used to use the name
> `GITHUB_PAT` frequently in Github Actions readmes, I guess we will all
> have to update everything!

### Solution 1

1. actions/checkout@v2 sets the 'Authorization' for any git command issued
   from the checked out repo. But since this uses the Github Actions
   `GITHUB_TOKEN` which is limited to the current repository. And since we
   want to access another private repo, we have to disable that.
2. instead, you can use `git config --global url.insteadof`. The `GH_TOKEN`
   Github Secret is a Github personal token
   (<https://github.com/settings/tokens>) that has the 'repo' scope (full
   control of private repositories).

> Note: when using github over https with a token, the username doesn't
matter, that's why we put `foo` here.

```yaml
- run: |
    git config --local --remove-section http."https://github.com/"
    git config --global url."https://foo:${GH_TOKEN}@github.com/your-org".insteadOf "https://github.com/your-org"
  env:
    GH_TOKEN: ${{ secrets.GH_TOKEN }}
```

### Solution 2

When `git config url.insteadof` does not work, you can try using `git
credential.helper` instead. For example:

```yaml
- run: |
    cat <<EOF > creds
      #!/bin/sh
      echo protocol=https
      echo host=github.com
      echo username=foo
      echo password=${GH_TOKEN}
    EOF
    sudo install creds /usr/local/bin/creds
    git config --global credential.helper "creds"
  env:
    GH_TOKEN: ${{ secrets.GH_TOKEN }}
```

### Solution 3

Same as solution 2 but wrapped in a neat Github Action
[setup-git-credentials](https://github.com/marketplace/actions/setup-git-credentials).

This action stores the credentials in the file
`$XDG_CONFIG_HOME/git/credentials` and configures git to use it by calling:

```sh
git config --global credential.helper store/
```

The actions also makes sure all `ssh` calls are rewritten to `https`. The
'credentials' field must be of the form <https://foo:$GH_TOKEN@github.com/>

```yaml
- uses: fusion-engineering/setup-git-credentials@v2
  with:
    credentials: https://foo:{{secrets.GH_TOKEN}}@github.com
```
