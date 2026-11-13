---
title: Getting Started
description: Build Whac-a-Leak from source and see what is available today.
---

Whac-a-Leak has no tagged release yet, so for now you build it from source.

## Prerequisites

- **Go 1.25+**. The binary is built with `CGO_ENABLED=0`, so no C toolchain is needed.
- **git**. whac drives the `git` binary directly for `.gitignore`-aware file
  lists, staged changes and history.
- **cue** (optional): for rule authors who want to `cue vet` their `.whac/` rules.

## Build

```sh
git clone https://github.com/infrashift/whac-a-leak
cd whac-a-leak
make build        # → bin/whac
```

## What works today

The repository is at **Milestone 0**: project scaffolding, design records and
this site. The `whac` CLI arrives in Milestone 1, starting with
`whac rules list` and `whac rules verify`. This page will grow with each
milestone. See the [roadmap](../project/roadmap/).

## Where it's going

When the scanner lands, day-to-day use will look like this. These commands are
**planned**, and their contract is fixed in the project's CLI spec:

```sh
whac scan .                      # scan the working tree (respects .gitignore)
whac scan --staged               # pre-commit: only lines you're about to commit
whac scan --git main..HEAD       # scan the history of a branch
whac scan . --format sarif       # for GitHub code scanning
whac scan --agent-context       # preflight: env vars, credential files, sockets
whac rules verify                # run every rule's own tests
whac allow add whac1:… --reason "test fixture"
```

Exit codes: `0` no findings, `1` findings reported, `2` an error (bad config,
a rule failing its tests, I/O).
