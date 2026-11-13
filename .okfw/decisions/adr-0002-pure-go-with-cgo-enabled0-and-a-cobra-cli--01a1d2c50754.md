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
description: "whac is a single static pure-Go binary (CGO_ENABLED=0) with a command surface built on spf13/cobra."
---

# ADR-0002: Pure Go with CGO_ENABLED=0 and a Cobra CLI

**Status:** Accepted · Amendment 1 (2026-11-12): binary name
**Date:** 2026-11-12
**Source:** Planning session

## Context
A secret scanner runs on developer laptops, in pre-commit hooks, on CI runners and in minimal containers. It must ship as one download with no shared libraries. The command surface is small but has nested subcommands (`rules verify`, `allow add`). Users expect help and shell completion.

## Decision
- Build with `CGO_ENABLED=0`. The `Makefile` sets it. `make test` is the only exception, because the race detector needs cgo.
- Use `github.com/spf13/cobra` for the CLI in `internal/cli`. Keep `cmd/whac/main.go` a thin `main`.
- Use the standard layout: `cmd/` and `internal/`. Add no `pkg/` until a public API exists by design.
- Run git through the `git` binary (`os/exec`), not a Go git library. A machine with a repo to scan has git.

## Consequences
- Static binaries cross-compile for every target.
- Cobra adds `spf13/pflag` to the dependency tree. The project accepts this cost for help, completion and a familiar interface.
- Without cgo, hyperscan and vectorscan are not available. The Aho-Corasick prefilter supplies the speed (ADR-0003).

## Alternatives considered
- **stdlib `flag` with a custom dispatcher.** Lighter, but no completion and no consistent help. The maintainer rejected it.
- **go-git.** A large dependency, slower than the git CLI on large histories. Rejected.

See [CLI contract](/specs/cli-contract--01a1d2c50744.md).

## Amendment 1 (2026-11-12): the binary is `whac`
The first plan named the CLI `wal`. pywal, a popular Linux theming tool, already installs a `wal` binary. The two names collide on `$PATH`. The binary is **`whac`** (`cmd/whac`). Every user-facing identifier uses the same name:
- config dir: `.whac/`
- CUE package: `package whac`
- fingerprint prefix: `whac1:`
- inline suppression: `whac:ignore`
