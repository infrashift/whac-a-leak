---
id: "01a1d2c5-0754-781f-8f5f-2299deb95a26"
title: "ADR-0002: Pure Go with CGO_ENABLED=0 and a Cobra CLI"
type: "ADR"
tags:
  - "adr"
  - "go"
  - "cli"
  - "build"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "whac is a single static pure-Go binary (CGO_ENABLED=0) whose command surface is built with spf13/cobra."
---

# ADR-0002: Pure Go with CGO_ENABLED=0 and a Cobra CLI

**Status:** Accepted · Amendment 1 (2026-11-12): binary name
**Date:** 2026-11-12
**Source:** Planning session

## Context
A secret scanner runs everywhere code lives: developer laptops, pre-commit hooks, CI runners, minimal containers. It has to be one download with no shared libraries. The command surface is small, but it has nested subcommands (`rules verify`, `allow add`), and users expect help and shell completion.

## Decision
- Build with `CGO_ENABLED=0` (set in the `Makefile`). The only exception is `make test`, because the race detector needs cgo.
- Use `github.com/spf13/cobra` for the CLI. It lives in `internal/cli`; `cmd/whac/main.go` is a thin `main`.
- Standard layout: `cmd/` + `internal/`, and no `pkg/` until a public API is deliberately designed.
- Git is driven through the `git` binary (`os/exec`), not a Go git library. Any machine with a repo to scan has git.

## Consequences
- Static binaries cross-compile easily for every target.
- Cobra adds `spf13/pflag` to the dependency tree. That's accepted in exchange for help, completion and a familiar UX.
- No cgo means no hyperscan/vectorscan. Speed comes from the Aho-Corasick prefilter (ADR-0003).

## Alternatives considered
- **stdlib `flag` with a hand-rolled dispatcher.** Lighter, but loses completion and consistent help. Rejected by the maintainer.
- **go-git.** A large dependency that's slower than the git CLI on big histories. Rejected.

See [CLI contract](/specs/cli-contract--01a1d2c50744.md).

## Amendment 1 (2026-11-12): the binary is `whac`
The CLI was first planned as `wal`, but pywal, a popular Linux theming tool, already installs a `wal` binary, and the two would clash on `$PATH`. The binary is **`whac`** (`cmd/whac`), and every user-facing identifier follows it so the names stay consistent:
- config directory: `.whac/`
- CUE package: `package whac`
- fingerprint prefix: `whac1:`
- inline suppression: `whac:ignore`
