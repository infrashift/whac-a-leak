---
id: "01a1d2c5-075d-77c0-982a-78e95b88f55b"
title: "ADR-0003: CUE at load time only; the scan engine is plain Go"
type: "ADR"
tags:
  - "adr"
  - "cue"
  - "engine"
  - "performance"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "CUE unifies, validates and decodes rules at startup; scanning never touches the CUE runtime and runs as an Aho-Corasick prefilter followed by Go regexp and entropy checks."
---

# ADR-0003: CUE at load time only; the scan engine is plain Go

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
The project pitch is "context-aware constraint evaluation" with CUE. Evaluating a CUE value for each candidate match would cost orders of magnitude more than a compiled regexp. Throughput (hundreds of MB/s on keyword-sparse content) is a core goal.

## Decision
- **Load time:** embedded schema + defaults + user `.whac/` layers → `cue/load` with an Overlay → unify → validate → `Decode` into Go structs → compile into `CompiledRule` (`*regexp.Regexp`, keyword set, entropy and filter parameters). Then the rule tests run (ADR-0009).
- **Scan time:** one Aho-Corasick pass over the lowercased content gives the candidate rules. Only those run their regex, then the secretGroup → entropy/charset → stopwords/allowRegexes/placeholders checks, then the allowlist.
- The CUE runtime is never called per file, line or match. `CLAUDE.md` rule 3 enforces this.

## Consequences
- CUE is in the binary but costs time only at startup. Budget: under ~150 ms with ~30 rules. If that's exceeded, `go generate` precompiles the embedded defaults to JSON and CUE runs only when user layers exist.
- Rule semantics are defined by the Go engine, so the [Rule schema](/specs/rule-schema--01a1d2c50736.md) spec must say exactly what each field means at scan time.
- Pattern rules need keywords (enforced by the schema); otherwise they would bypass the prefilter.

## Alternatives considered
- **Unifying each candidate against a CUE constraint at scan time.** Too slow. Rejected.
- **A flat TOML/YAML regex list (gitleaks style).** Loses typing, layering and rule self-tests. Rejected.
