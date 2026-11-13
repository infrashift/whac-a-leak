---
id: "01a1d2c5-07a8-7302-a1ee-9ff30d338aaa"
title: "Whac-a-Leak architecture overview"
type: "Architecture"
tags:
  - "architecture"
  - "overview"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Package map, load phase and scan phase, sources and outputs of the whac secret scanner."
---

# Whac-a-Leak architecture overview

**Status:** Target architecture. Each milestone adds packages.

## Two phases
| Phase | Runs | Technology | Packages |
|---|---|---|---|
| **Load** | once at startup | CUE: unify, validate, `Decode` | `internal/rules` |
| **Scan** | per file or chunk | plain Go: Aho-Corasick, `regexp`, entropy | `internal/engine`, `internal/structural`, `internal/allow` |

The scan phase never runs CUE ([ADR-0003](/decisions/adr-0003-cue-at-load-time-only-the-scan-engine-is-plain-go--01a1d2c5075d.md)).

## Load phase
```
//go:embed builtin/*.cue (schema + builtin rules) ─┐
                                                   ├─ one CUE instance ─▶ unify ─▶ validate
layers: .whac/*.cue (package whac)                ─┘                                 │
                                                                                     ▼
                 CompiledRule[] ◀─ compile (regexp, AC keywords, filters) ◀─ Decode → Go structs
                      │
                      └─▶ verify: run every rule sample through the engine (ADR-0009) ─▶ stop on failure
```
Layering semantics: [ADR-0004](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md), [Rule schema](/specs/rule-schema--01a1d2c50736.md).

## Scan phase
1. **Source** (`internal/source`) produces chunks `{path, content, commit?, addedLines?}`:
   - filesystem: `git ls-files -co --exclude-standard` inside a repo, else a directory walk
   - staged: `git diff --cached -U0` supplies the added lines; the engine scans the staged blob
   - history: a streaming `git log -p -U0`, added lines only
   - env: the process environment → `env.<NAME>` assignments ([ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md))
   - home: the `config.credentialPaths` entries under `$HOME` and each `--root`
   - exposure checks (agent-context only): credential sockets and opt-in IMDS → **exposure** findings
2. **Gate**: skip binaries (a NUL byte in the first 8 KB) and files over `maxFileSize`. Build the newline index.
3. **Prefilter**: one Aho-Corasick pass selects the candidate pattern rules.
4. **Match**: candidate regexes → `secretGroup` → charset/entropy ([ADR-0006](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md)) → stopwords, allowRegexes, safe formats.
5. **Structural**: a format extractor or the generic line extractor → assignments → `assignment` rules and the placeholder filter ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md)).
6. **Allow**: fingerprint ([Fingerprint format](/specs/fingerprint-format--01a1d2c5073d.md)), path globs, inline `whac:ignore`.
7. **Report** (`internal/report`): text, JSON or SARIF. Sorted output. Redacted secrets ([CLI contract](/specs/cli-contract--01a1d2c50744.md)).

A worker pool of `GOMAXPROCS` workers processes the chunks. The engine sorts the results, so the output is deterministic.

## Package map
| Package | Responsibility |
|---|---|
| `cmd/whac` | thin `main` → `internal/cli` |
| `internal/cli` | Cobra commands, flag parsing, exit codes ([ADR-0002](/decisions/adr-0002-pure-go-with-cgo-enabled0-and-a-cobra-cli--01a1d2c50754.md)) |
| `internal/rules` | embedded CUE (`builtin/`), loader ([ADR-0012](/decisions/adr-0012-load-layers-as-one-in-memory-cue-build-instance--01a1d2e388a9.md)), decode, compile, overrides |
| `internal/engine` | prefilter, matcher, entropy, charset, filters, rule verification |
| `internal/structural` | YAML/JSON/dotenv/properties/HCL-lite extractors |
| `internal/source` | filesystem, staged, history, env and home sources; diff parser |
| `internal/allow` | allowlist evaluation |
| `internal/fingerprint` | `whac1:` fingerprints |
| `internal/report` | text, JSON and SARIF writers |

## Non-goals
Live credential verification, base64 blob decoding and process memory scans ([ADR-0008](/decisions/adr-0008-no-live-secret-verification-in-v1--01a1d2c5078c.md)).
