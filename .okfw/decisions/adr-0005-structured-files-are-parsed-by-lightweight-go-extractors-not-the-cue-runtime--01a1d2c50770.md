---
id: "01a1d2c5-0770-7512-a10d-c10febde71d9"
title: "ADR-0005: Structured files are parsed by lightweight Go extractors, not the CUE runtime"
type: "ADR"
tags:
  - "adr"
  - "structural"
  - "performance"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "YAML, JSON, dotenv, properties/ini and HCL-style files are turned into positioned (keyPath, value) assignments by small Go extractors; CUE only declares which key paths are sensitive."
---

# ADR-0005: Structured files are parsed by lightweight Go extractors, not the CUE runtime

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
Line-based regex loses the key/value relationship in structured config. `database.credentials.password: hunter2` has no prefix and low entropy. The original idea was to ingest every YAML/JSON file into CUE and unify it with a "sensitive path" schema. The CUE decoders are much heavier than the scan budget allows, and a unification failure is a poor way to report a positioned finding.

## Decision
- `internal/structural` extractors emit `Assignment{KeyPath, Value, Line, Col}`:
  - YAML via yaml.v3 `Node` (positions included; the same module CUE already depends on)
  - JSON via a `json.Decoder` token walker using `InputOffset`
  - hand-written lexers for dotenv, properties/ini and HCL-style `k = "v"` lines (no HCL dependency)
- `assignment` rules match on `keyContext` (a regex on the last key segment) and/or `paths` (doublestar globs over the dotted key path), then check the value.
- Files without an extractor go through a generic `key = value` / `key: value` line regex, which feeds the same assignment rules.
- A placeholder filter (`${VAR}`, `{{ }}`, `<...>`, `changeme`, `xxxx`, repeated characters, empty values) runs before any assignment finding.

## Consequences
- Low-entropy secrets in sensitive paths are caught without flooding the output.
- In git-history mode, hunks are partial, so structured parsing is skipped there and only the line-level fallback runs.

## Alternatives considered
- **`cuelang.org/go/encoding/{yaml,json}` + unify per file.** Too slow and memory-heavy. Rejected.
