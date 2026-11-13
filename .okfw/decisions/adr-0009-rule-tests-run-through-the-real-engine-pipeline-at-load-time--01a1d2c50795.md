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
description: "The Go engine runs each rule's mustMatch and mustIgnore samples at every load, and a failure stops the load; CUE checks mustMatch against the regex for rule authors."
---

# ADR-0009: Rule tests run through the real engine pipeline at load time

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
Broken regexes and greedy patterns are the most common defects in custom rules. Samples in each rule (`tests.mustMatch`, `tests.mustIgnore`) let the rule set verify itself. A check in CUE alone is incomplete. `mustMatch: [...=~regex]` works. `mustIgnore: [...!~regex]` is **wrong**: many ignore samples match the regex on purpose. Entropy, stopwords or placeholders then reject them.

## Decision
- In CUE, pattern rules constrain `tests.mustMatch: [...=~regex]`. `cue vet` runs this fast check without the binary.
- In Go, every load and `whac rules verify` run each sample through the scan code path. A failure stops the load. The error names the rule ID, the sample index and the failed stage. It does not print the sample text.
- `mustMatch` must contain at least one sample (`[string, ...string]`).

## Consequences
- Every rule has at least one executable example. The example also documents the rule.
- The sample run adds a small startup cost, inside the ADR-0003 budget.
- The repository contains the samples, so they must be clearly fake. The repo's own `.whac/` allowlists `internal/rules/builtin/**`.

See [Rule schema](/specs/rule-schema--01a1d2c50736.md).

## Implementation note (Milestone 1)
- `engine.Verify` runs the samples. It lives in `internal/engine`, not `internal/rules`, because the engine imports the compiled rules. This placement avoids an import cycle.
- `rules.Load` only loads and compiles. Each command that needs rules calls `rules.Load` and then `engine.Verify`. A failed sample gives exit code 2.
- A failure names the rule, the sample position (`mustMatch[0]`) and the last stage the sample reached: no keyword, no regex match, or the rejection reason. The failure never contains the sample text.
