# Personal website

![Build and deploy to github pages](https://github.com/maelvls/maelvls.github.io/workflows/Build%20and%20deploy%20to%20github%20pages/badge.svg?branch=source)

Code is in the branch `source`, gp-pages are in `master`.

To create a new post:

```sh
hugo new -k bundle 2020/client-go-transitive-hell
```

The `-k` flag tells hugo to use `archetypes/bundle` ([see
doc](https://gohugo.io/content-management/archetypes/#directory-based-archetypes)).

Locally serve:

```sh
hugo serve --buildDrafts --buildFuture
```
