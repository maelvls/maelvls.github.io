---
title: git stash vs. git reset --hard
date: 2017-12-20
---

Many times, I used `git reset --hard` and found out later that all my changes had disappeared.
I think this is because I do not fully understand what `git reset --hard` does:

1. move the branch ref pointer to the given reference
2. set HEAD to this ref
3. Depending on the option:

- --soft: working dir and stage are not touched, useful when `git reset --hard HEAD^` to
  change the content of the commit without having to re-add everything
- --mixed: working dir is not touched, but the stage is set to match the HEAD
- --hard: stage + working dir are set to match the HEAD

On the other side, `git stash` will set stage + working dir to match the HEAD.

So instead of `git reset --hard REF` I should rather do `git stash && git reset --hard REF`.
This would avoid cases where I forgot that I had made changes... and end up losing them all.

IDEAS ON NEW COMMANDS

- I often have committed something like
  @~0: edit touist.rb
  @~1: edit .travis.yml

I want to add a command like `git ci --amend~2` that would amend the commit in @~1,
not the HEAD. Here is what I must do:

    vim .travis.yml
    git stash
    git rb -i @^^ # edit HEAD^
    git stash pop
    git add .travis.yml
    git ci --amend
    git rb --continue

Facebook's hg extensions have 'absorb' which does this but on @~1, @~2... and chooses
automatically in which commit it should be amended.
See:

- <https://stackoverflow.com/questions/24625411/amend-the-second-to-last-commit>
- <https://blog.filippo.io/git-fixup-amending-an-older-commit/>
