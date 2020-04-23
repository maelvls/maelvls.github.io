---
title: "Setting Up Concourse"
description: ""
date: 2020-04-22T17:37:12+02:00
url: /setting-up-concourse
images: [setting-up-concourse/cover-setting-up-concourse.png]
draft: true
tags: []
---

```sh
# Make sure I use the same version as the server.
% fly -t main sync

# Login.
% fly login -t main -c https://concourse.k.maelvls.dev/
logging in to team 'main'

navigate to the following URL in your browser:

  https://concourse.k.maelvls.dev/login?fly_port=53932

or enter token manually:
target saved
```

Can `-t main` be remembered by default like contexts and namespaces with
`kubectl config set-context`? See:
- [issue 213](https://github.com/concourse/fly/issues/213):
-
- [issue 1933](https://github.com/concourse/concourse/issues/1933):
  > Could `fly` remember the target? Either with `FLY_TARGET` env var or
  > setting a default in ~/.flyrc?
  >
  > Response: The `-t` flag is intentionally stateless and must be
  > explicitly added to each command. This reduces the risk of accidentally
  > running a command against the wrong environment when you have multiple
  > targets defined.

OK. I hate that kind of stubborness.

```sh
% cat <<EOF > pipeline.yml
---
resources:
  - name: concourse-docs-git
    type: git
    icon: github-circle
    source:
      uri: https://github.com/concourse/docs
jobs:
  - name: job
    public: true
    plan:
      - get: concourse-docs-git
        trigger: true
      - task: list-files
        config:
          inputs:
            - name: concourse-docs-git
          platform: linux
          image_resource:
            type: registry-image
            source: { repository: busybox }
          run:
            path: ls
            args: ["-la", "./concourse-docs-git"]
EOF
```

```sh
fly -t main set-pipeline -pmaster -c pipeline.yml
```

<script src="https://utteranc.es/client.js"
        repo="maelvls/maelvls.github.io"
        issue-term="pathname"
        label="💬"
        theme="github-light"
        crossorigin="anonymous"
        async>
</script>

