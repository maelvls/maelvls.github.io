+++
# About/Biography widget.
#     date --iso-8601=minutes
date = "2017-10-09"
draft = false

widget = "about"

# Order that this section will appear in.
weight = 5

# List your academic interests.
[interests]
  interests = [
    "Golang",
    "Developer tooling",
    "Functional programming",
    "CI/CD and Kubernetes"
  ]

# List your qualifications (such as academic degrees).
[[education.courses]]
  course = "PhD in Artificial Intelligence"
  institution = "Université Toulouse 3 – Paul Sabatier"
  year = 2019

[[education.courses]]
  course = "MSc in Operations Research"
  institution = "Université Toulouse 3 – Paul Sabatier"
  year = 2016

[[education.courses]]
  course = "BSc in Computer Science"
  institution = "Université Toulouse 1 – Capitole"
  year = 2014
+++

- <i class="fa fa-linkedin" style="margin-right:0.5em"></i> [LinkedIn] profile (most up-to-date information)
- <i class="fa fa-file" style="margin-right:0.5em"></i> [Resume/CV PDF](resume_mael_valais.pdf) (updated on June 7th, 2019)
- Contact: [mael.valais@gmail.com](#), +33 7 86 48 43 91

<!--
  [<img src="img/irit2018.svg" style="max-width:30%;min-width:2cm;float:right;margin:1em;margin-top:1cm">][irit]
-->

## About me

> **Note:** if you want the most up-to-date information about me, I
> recommend taking a look at my [LinkedIn][] page.

- I speak and write English fluently. Come try my delicious and refreshing
  french-flavoured accent! 😄
- I enjoy contributing to open-source projects; when I do, it is mostly a
  way of scratching a developer itch on one of my tools. I sent pull
  requests to [ocaml-minisat][], [ocaml-qbf][], [ocamlyices2][] and [opam][]
  (OCaml), [gitlab-ce][] (Ruby on rails, Rspec), [boost-graph][] (C++).
- I authored and am the maintainer of various
  projects: [homebrew-amc][] (Ruby, Travis CI), [touist][] (OCaml) and
  four [vscode-extensions][] (they use Typescript; one of them has 29k
  download! 😊).
- From my experience, open-source promotes tolerance when it comes to
  discussing and accepting other's patches as well as leading people to
  give friendly and constructive critics (most of the time though; mileage
  may vary, see [this email Linus wrote][linus-fuck-kay] to one of his
  contributors).
- I love functional programming; I have 3 years of experience using OCaml
  building a compiler and solver for propositional logic. I also did some
  ReasonML. I discovered Elixir and Erlang a couple of days ago and so far, I
  love it. Still a lot to learn though, but I start to get the hang of it.
- I am very interested by micro-service architectures. I did some
  side-projects using Go, gRPC (e.g. [maelvls/users-grpc][]) with
  Kubernetes. I had a lot of fun with it (thanks to Go's excellent 'dev
  experience'). I also played with Elixir which also benefits from a very
  polished 'dev experience'. I also played a lot with Rust and compared it
  to Go ([rust-chat][], [touist-server][]). Rust is is by far the fastest
  'modern' language, but not the easiest to learn: borrow checker,
  lifetimes, traits...
- Throughout my work, I like to improve the 'developer experience' (DX) by
  improving the tooling as well as the overall DevOps workflow. I think that a
  good developer experience keeps retaining and gaining good developers. If
  given the opportunity, I wish to contribute in that regard.
- I worked with multiple automation and continuous integration tools (Drone.io,
  Travis CI, Gitlab CI, Circle CI and Appveyor; pull request lifecycle using
  bots and Slack ChatOps integration with Slack). On the CI/CD side, I also
  collaborated with the Homebrew 'tap' people: [How to automate the build of
  bottles on your Homebrew tap][homebrew-tap-auto-bottles].
- I experimented with Docker, Kubernetes using Terraform, Helm, Traefik and
  Prometheus/Grafana; both on AWS EKS and GCP GKE
  (see [maelvls/awx-gke-terraform][] and [maelvls/terraform-touist][]).
- I have some knowledge on machine learning (more specifically, deep
  learning) as it was one of the topics during the first 6 months of my PhD
  (see my [masters-thesis][]).
- I can bring some knowledge about routing problems, more specifically shortest
  path algorithms. During an internship at Mobigis, I developed a shortest-path
  algorithm based on Dijkstra for carpooling on actual geographic data; I also
  contributed to the open-source [boost-graph][] library (mainly written in
  C++11). I also worked on vehicle routing ([vehicule-routing-report][] in
  French).
- On the teamwork side, I worked in various project agile setups (Scrum and
  Kanban) in teams ranging from 2 to 6 people. I enjoy sharing ideas around
  team workflows and ways of shipping smoother and faster.
- As a last note, I really think pair-programming and code reviews can make us
  developers grow and learn from others, not only about code but also finding
  the best tooling and shortcuts (Emacs, IDEs, command line tools...) and such.

[ocamlyices2]: https://github.com/polazarus/ocamlyices2/pulls?utf8=%E2%9C%93&q=author%3Amaelvls+
[ocaml-minisat]: https://github.com/c-cube/ocaml-minisat/pulls?utf8=%E2%9C%93&q=author%3Amaelvls
[ocaml-qbf]: https://github.com/c-cube/ocaml-qbf/issues?utf8=%E2%9C%93&q=author%3Amaelvls
[opam]: https://github.com/ocaml/opam-repository/pulls?utf8=%E2%9C%93&q=author%3Amaelvls
[gitlab-ce]: https://gitlab.com/gitlab-org/gitlab-ce/merge_requests/1150
[boost-graph]: https://github.com/boostorg/graph/issues?utf8=%E2%9C%93&q=author%3Amaelvls
[homebrew-amc]: https://github.com/maelvls/homebrew-amc
[touist]: https://github.com/touist/touist
[maelvls/awx-gke-terraform]: https://github.com/maelvls/awx-gke-terraform
[maelvls/terraform-touist]: https://github.com/maelvls/terraform-touist
[masters-thesis]: https://drive.google.com/file/d/0B5mz8k-t6PT0N2lINEZYX2duOFU/view
[vehicule-routing-report]: http://homepages.laas.fr/sungueve/Docs/Etu/rapport-ter-aide-humanitaire.pdf
[homebrew-tap-auto-bottles]: https://gist.github.com/maelvls/068af21911c7debc4655cdaa41bbf092
[maelvls/users-grpc]: https://github.com/maelvls/users-grpc
[rust-chat]: https://github.com/maelvls/rust-chat
[touist-server]: https://github.com/maelvls/touist-editor/blob/master/touist-server/src/main.rs
[maelvls.github.io]: https://maelvls.github.io/
[mael.valais@gmail.com]: mailto:mael.valais@gmail.com
[vscode-extensions-github]: https://github.com/maelvls?utf8=%E2%9C%93&tab=repositories&q=vscode&type=&language=
[vscode-extensions]: https://marketplace.visualstudio.com/search?term=maelvalais&target=VSCode&category=All%20categories&sortBy=Relevance
[linus-fuck-kay]: http://lkml.iu.edu/hypermail/linux/kernel/1404.0/01331.html

## Past Experiences

I am currently employed by [SQUAD] and am currently contracted to La Banque
Postale (Toulouse, France). I work on tools for network administrators,
including a web front-end to Ansible Tower APIs.

Before that, I was contracted to Orange, where I worked on tools for
developers, examples and documentation in order to improve the developer
experience when it comes to using the Orange IT's private cloud. I also try
to promote and find new ways to guide people into sharing their own code as
well getting the most out of Orange's internal GitLab.

[squad]: https://www.squad.fr

On April, 10th 2019, I defended my PhD thesis at [IRIT] \(Institut de
Recherche en Informatique de Toulouse, France – [location]) in the [LILaC]
and [ADRIA] team. My PhD work was supervised by Olivier Gasquet (IRIT),
Dominique Longin (CNRS), Frédéric Maris (IRIT) and Andreas Herzig (CNRS).
My goal was to develop a tool and a language, [TouIST] \(pronounced
_twist_, standing for _Toulouse Integrated Satisfiability Tool_\), that
will allow us to express and solve real-world problems through the use of
multiple logic theories: SAT, SMT and QBF for now.

[touist]: https://www.irit.fr/touist
[github]: https://github.com/touist/touist
[irit]: https://www.irit.fr
[lilac]: https://www.irit.fr/-Equipe-LILaC-
[adria]: https://www.irit.fr/-Equipe-ADRIA-
[linkedin]: https://www.linkedin.com/in/maelvalais/
[location]: https://goo.gl/maps/nuxdSM6P65J2
[twitter]: https://twitter.com/maelvls
[profile]: https://www.irit.fr/spip.php?page=annuaire&code=10566

[^touist-meaning]:

  _**Tou**louse **i**ntegrated **s**atisfiability **t**ool_.
  it is prononced _twist_. we were looking for a memorable and
  pronounceable name that had no homonym on google. and it
  had to sound like fun, too!
