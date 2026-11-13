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
description: "CUE unifies, validates and decodes rules at startup. The scan never calls the CUE runtime; it runs an Aho-Corasick prefilter, Go regexp and entropy checks."
---

# ADR-0003: CUE at load time only; the scan engine is plain Go

**Status:** Accepted · Refined by [ADR-0013](/decisions/adr-0013-regexes-run-on-keyword-line-windows--01a1d2f04316.md) (line windows)
**Date:** 2026-11-12
**Source:** Planning session

## Context
The project pitch is "context-aware constraint evaluation" with CUE. A CUE evaluation for each candidate match costs orders of magnitude more than a compiled regexp. High throughput is a core goal: hundreds of MB/s on content with few keywords.

## Decision
- **Load phase:** the loader takes the embedded schema, the builtin rules and the layers in `.whac/`. It unifies them as one CUE instance, validates the result and decodes it into Go structs. Then it compiles each rule into a `CompiledRule` (`*regexp.Regexp`, keyword set, entropy and filter parameters). Then the rule samples run (ADR-0009).
- **Scan phase:** one Aho-Corasick pass over the lowercased content selects the candidate pattern rules. Only those rules run their regex. Each match then passes the secretGroup, entropy/charset, stopwords/allowRegexes and allowlist checks.
- The scan never calls the CUE runtime per file, line or match. `CLAUDE.md` rule 3 enforces this.

## Consequences
- CUE is in the binary but costs time only at startup. Budget: about 150 ms with about 30 rules. If startup exceeds the budget, `go generate` precompiles the builtin rules to JSON. CUE then runs only when layers exist.
- The Go engine defines rule semantics. The [Rule schema](/specs/rule-schema--01a1d2c50736.md) spec must state what each field does at scan time.
- Pattern rules need keywords. The schema enforces this, so no pattern rule bypasses the prefilter.

## Alternatives considered
- **CUE unification for each candidate at scan time.** Too slow. Rejected.
- **A flat TOML/YAML regex list (gitleaks style).** No typing, no layering, no rule samples. Rejected.
