---
title: Getting Started
description: Build Whac-a-Leak from source and see which features work now.
---

Whac-a-Leak has no tagged release. Build it from source.

## Prerequisites

- **Go 1.25+**. The build uses `CGO_ENABLED=0`, so you need no C toolchain.
- **git**. whac runs the `git` binary for file lists that obey
  `.gitignore`, for staged changes and for history.
- **cue** (optional). Rule authors use it to run `cue vet` on their `.whac/`
  layers.

## Build

```sh
git clone https://github.com/infrashift/whac-a-leak
cd whac-a-leak
make build        # → bin/whac
```

## Current state

The repository is at **Milestone 1**. The `whac` CLI loads, layers and verifies
rules. It does not scan files yet. Milestone 2 adds `whac scan`. See the
[roadmap](../project/roadmap/).

1. Run `bin/whac rules list`.
2. Run `bin/whac rules verify`.

* **Expected result:** whac lists the builtin rules. Every rule shows `PASS`,
  and whac exits with code `0`.

To add your own rules, read [Writing Rules](../guides/writing-rules/).

## Planned commands

The project CLI spec defines these commands. They do **not exist yet**:

```sh
whac scan .                      # scan the working tree (obeys .gitignore)
whac scan --staged               # pre-commit: scan only the lines you commit
whac scan --git main..HEAD       # scan the history of a branch
whac scan . --format sarif       # output for GitHub code scanning
whac scan --agent-context       # preflight: env vars, credential files, sockets
whac rules verify                # run every rule sample
whac allow add whac1:… --reason "test fixture"
```

Exit codes:

| Code | Meaning |
|---|---|
| `0` | No findings |
| `1` | One or more findings |
| `2` | Error: bad configuration, failed rule sample, or I/O error |
