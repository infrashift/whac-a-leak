---
id: "01a1d2e3-88a9-7757-a1bd-0c2f2d9c699f"
title: "ADR-0012: Load layers as one in-memory CUE build instance"
type: "ADR"
tags:
  - "adr"
  - "cue"
  - "rules"
  - "loader"
created: "2026-11-12T21:11:29-06:00"
timestamp: "2026-11-12T21:11:29-06:00"
description: "The loader parses the builtin files and each layer into one in-memory CUE package instance; it needs no cue.mod, no overlay and no files on disk."
---

# ADR-0012: Load layers as one in-memory CUE build instance

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Milestone 1 spike ([ISSUE-001](/issues/issue-001-rule-schema-cue-loader-and-rules-verify--01a1d2c507b1.md))

## Context
The plan proposed `cue/load` with an Overlay. An overlay maps the embedded files into a directory on disk. This approach needs a real directory and module rules, and the config dir is optional. The Milestone 1 spike compared it with an in-memory build instance.

## Decision
- The loader parses each source with `cue/parser` and adds the file to one `build.Instance` with `AddSyntax`. `cuecontext` then builds the instance.
- The sources are the builtin files in `internal/rules/builtin/` (embedded), then each `*.cue` file in the config dir, sorted by name.
- Builtin positions use the prefix `(builtin)/`. Layer positions use the file path. CUE errors show both.
- The schema and the builtin rules share one directory. `cue vet -c ./internal/rules/builtin/` then checks the rules against the schema.
- The loader skips a layer file that starts with the header of `whac rules export-schema`. A stale schema copy cannot conflict with the builtin schema.
- The loader rejects unknown top-level fields. Package scope in CUE is open, so a misspelt `rule:` passes CUE without this check.

## Consequences
- The config dir needs no `cue.mod`. A missing default `.whac/` means no layers.
- Layers can refer to schema definitions such as `#SecretRule` and can import CUE standard library packages. The spike confirmed both.
- A layer needs the package clause `package whac`. A different package name stops the load.

## Alternatives considered
- **`cue/load` with an Overlay.** Rejected: it needs a directory on disk and module handling for no gain.
- **Compile each file separately and unify the values.** Rejected: a layer cannot then refer to `#SecretRule`.
