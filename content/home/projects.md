+++
# Projects widget.
# This widget displays all projects from `content/project/`.

date = "2016-04-20T00:00:00"
draft = false

title = "Projects"
subtitle = ""
widget = "projects"

# Order that this section will appear in.
weight = 50

# View.
# Customize how projects are displayed.
# Legend: 0 = list, 1 = cards.
view = 1

# Filter toolbar.

# Default filter index (e.g. 0 corresponds to the first `[[filter]]` instance below).
filter_default = 0

# Add or remove as many filters (`[[filter]]` instances) as you like.
# Use "*" tag to show all projects or an existing tag prefixed with "." to filter by specific tag.
# To remove toolbar, delete/comment all instances of `[[filter]]` below.
[[filter]]
  name = "All"
  tag = "*"

[[filter]]
  name = "Logic"
  tag = ".logic"

+++

### Minor contributions to open-source software

- [boost graph] is a part of the large Boost C++ library. I mainly contributed
  through [bug fixes][boost-contributions] to max flow algorithms.
- [gitlab] is a competitor to Github; I proposed multiple
  [bug fixes][gitlab-contribution] (as merge requests) to enhance the
  referencing behavior in issue tickets.

[boost graph]: http://www.boost.org/doc/libs/release/libs/graph/doc/index.html
[boost-contributions]: https://github.com/boostorg/graph/pulls?q=author%3Amvalaisalais
[gitlab]: https://about.gitlab.com
[gitlab-contribution]: https://gitlab.com/gitlab-org/gitlab-ce/merge_requests/1150

### Projects contributed