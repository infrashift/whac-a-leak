---
id: "01a1d2c5-0795-75ca-afba-71e9f9893a21"
title: "ADR-0009: Rule tests run through the real engine pipeline at load time"
type: "ADR"
tags:
  - "adr"
  - "rules"
  - "testing"
  - "cue"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Each rule's mustMatch and mustIgnore samples are executed by the Go engine whenever rules load, and a failure aborts loading; CUE checks mustMatch against the regex for rule authors."
---

# ADR-0009: Rule tests run through the real engine pipeline at load time

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
Broken regexes and overly greedy patterns are the most common failures in custom rules. Embedding `tests.mustMatch` / `tests.mustIgnore` in each rule lets the ruleset verify itself. A pure-CUE check is incomplete. `mustMatch: [...=~regex]` works, but `mustIgnore: [...!~regex]` would be **wrong**: many ignore samples are meant to match the regex and then be rejected by entropy, stopwords or placeholders.

## Decision
- In CUE, pattern rules constrain `tests.mustMatch: [...=~regex]`. That's a cheap check under `cue vet` with no binary.
- In Go, every load (and `whac rules verify`) runs each sample through the same code path as a scan. A failure aborts loading with the rule ID, sample index and failing stage. The sample text is not printed.
- `mustMatch` must be non-empty (`[string, ...string]`).

## Consequences
- Every rule has at least one executable example, which doubles as documentation.
- Rule loading costs a little more at startup (counted in the ADR-0003 budget).
- Sample secrets live in the repo, so they must be obviously fake. The repo's own `.whac/` allowlists `internal/rules/defaults/**`.

See [Rule schema](/specs/rule-schema--01a1d2c50736.md).
