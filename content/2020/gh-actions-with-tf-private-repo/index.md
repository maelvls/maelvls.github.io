---
title: "Github Actions with a private Terraform module"
description: |
  Terraform makes it easy to manage infrastructure at scale; you
  might want to share code between modules, and that's where it becomes tricky.
  In this post, I try to give some clues on how to use terraform across private
  Github repos.
date: 2020-05-09T16:02:26+02:00
url: /gh-actions-with-tf-private-repo
images: [gh-actions-with-tf-private-repo/cover-gh-actions-with-tf-private-repo.png]
tags: [terraform, github-actions]
---

A common way of sharing terraform modules is to move them in a separate
repo. And for companies, that means a private repo. When `terraform init`
is run, the terraform module is fetched and if this module is stored on a
Github private repo, you will need to work around the authentication.

Imagine that these shared modules are stored on the private Github repo
"github.com/your-org/terraform-modules". Importing this module from a
different repo would look something like:

```hcl
module "some_instance_of_this_module" {
  source = "git@github.com:your-org/terraform-modules.git//path/to/module?ref=master"
}
```

Using `git+ssh` as a way of fetching this private module will work great
locally since you might probably have a private key that Github knows
about. Locally, `terraform init` will work.

But what about CI, should I create a key pair and store the private key as
a secret and have the public key known by Github (or Gitlab)?

This method is not great: this key pair is tied to an individual and can't
be tied to a Github App like `github-bot`. A better way of doing is using
`git+https` by relying on a token. The ssh key pair mechanistm doesn't
offer access to a specific repo either but a Github App token can.

To use HTTPS instead of git over SSH, we start by changing the way we
import these modules:

```diff
 module "some_instance_of_this_module" {
-   source = "git@github.com:your-org/terraform-modules.git//path/to/module?ref=master"
+   source = "git::https://github.com/your-org/terraform-modules.git//path/to/module?ref=master"
 }
```

## Local development & git over HTTPS

Locally, you will have to make sure you can `git clone` this private repo,
for example, the following should work:

```sh
git clone https://github.com/your-org/terraform-modules.git
```

If it doesn't work, Github has a helpful page "[Caching your GitHub
password in
Git](https://help.github.com/en/github/using-git/caching-your-github-password-in-git)".

## Continous integration & git over HTTPS

It is a bit trickier to get HTTPS working on the CI. In the following, I'll
take the example of Github Actions but that will work for any CI provider.
There are two main solutions:

1. using `url.insteadof`,
2. or using `credential.helper`.

> ⚠️ NOTE: around April 2020, Github decided to prevent users from using
> Github Secrets names that begin with `GITHUB_`. We used to use the name
> `GITHUB_PAT` frequently in Github Actions readmes, I guess we will all
> have to update everything!

### Solution 1: `url.insteadof`

1. actions/checkout@v2 sets the 'Authorization' for any git command issued
   from the checked out repo. But since this uses the Github Actions
   `GITHUB_TOKEN` which is limited to the current repository. And since we
   want to access another private repo, we have to disable that.
2. instead, you can use `git config --global url.insteadof`. The `GH_TOKEN`
   Github Secret is a Github personal token
   (<https://github.com/settings/tokens>) that has the 'repo' scope (full
   control of private repositories).

> Note: when using git over HTTPS with a token on `https://github.com`, the
username doesn't matter, that's why we put `foo` here.

```yaml
- run: |
    git config --local --remove-section http."https://github.com/"
    git config --global url."https://foo:${GH_TOKEN}@github.com/your-org".insteadOf "https://github.com/your-org"
  env:
    GH_TOKEN: ${{ secrets.GH_TOKEN }}
```

### Solution 2: `credential.helper`

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

### Solution 2 bis: `credential.helper` with a Github Action

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
