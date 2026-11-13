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
description: "Small Go extractors turn YAML, JSON, dotenv, properties/ini and HCL-style files into positioned key-path and value pairs; CUE only declares sensitive key paths."
---

# ADR-0005: Structured files are parsed by lightweight Go extractors, not the CUE runtime

**Status:** Accepted · Amendment 1 (2026-11-12): implementation details
**Date:** 2026-11-12
**Source:** Planning session

## Context
A line-based regex loses the key/value relationship in structured config. `database.credentials.password: hunter2` has no prefix and low entropy. The first idea was to load every YAML/JSON file into CUE and unify it with a "sensitive path" schema. The CUE decoders cost far more than the scan budget allows. A unification failure also reports a positioned finding badly.

## Decision
- `internal/structural` extractors produce `Assignment{KeyPath, Value, Line, Col}`:
  - YAML through yaml.v3 `Node`, which keeps positions. CUE already depends on this module.
  - JSON through a `json.Decoder` token walker that uses `InputOffset`.
  - Small custom lexers for dotenv, properties/ini and HCL-style `k = "v"` lines. No HCL dependency.
- `assignment` rules match on `keyContext` (a regex on the last key segment), on `paths` (doublestar globs over the dotted key path), or on both. Then they check the value.
- Files without an extractor go through a generic `key = value` / `key: value` line regex. Its output feeds the same assignment rules.
- A placeholder filter runs before any assignment finding. It drops `${VAR}`, `{{ }}`, `<...>`, `changeme`, `xxxx`, repeated characters and empty values.

## Consequences
- whac finds low-entropy secrets in sensitive paths without noise.
- Git-history mode has only partial hunks. It skips structured parsing and runs only the line-level fallback.

## Alternatives considered
- **`cuelang.org/go/encoding/{yaml,json}` with a unification per file.** Too slow and too much memory. Rejected.

## Amendment 1 (2026-11-12): implementation in Milestone 3
- Extractors: YAML (yaml.v3 `Node`), JSON (token walker), dotenv, Java properties, INI (with sections), HCL-lite (quoted strings in nested blocks). `structural.Detect` picks the extractor from the file name.
- A file that does not parse falls back to unstructured text.
- On unstructured text, the generic extractor runs only on the lines that hold a keyword of an assignment rule. This keeps the line-window speed of [ADR-0013](/decisions/adr-0013-regexes-run-on-keyword-line-windows--01a1d2f04316.md). An assignment rule without keywords runs on structured files only.
- The generic extractor skips comparisons (`==`, `===`, `!=`).
- A rule with both `keyContext` and `paths` needs both to match. Path globs use `.` separators and ignore case.
- Samples may carry a file name (`{file, content}`), so a rule can test its behavior on each format.
- `config.sensitivePaths` is a map that layers extend. The builtin `SENSITIVE_PATH` rule builds its `paths` from it with a CUE comprehension.
- The benchmark drops from 245 to 195 MB/s with the two builtin assignment rules.
