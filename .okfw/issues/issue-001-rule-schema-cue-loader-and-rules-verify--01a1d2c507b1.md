---
id: "01a1d2c5-07b1-7b8a-a5ec-68806dcc0ce9"
title: "ISSUE-001: Rule schema, CUE loader and rules verify"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-1"
  - "rules"
  - "closed"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T21:12:24-06:00"
description: "Milestone 1: Cobra skeleton, embedded CUE schema and builtin rules, loader for layers, decode and compile, and the rules list and verify commands."
---

# ISSUE-001: Rule schema, CUE loader and rules verify

**Status/Milestone:** Closed · Milestone 1

## Description
Build `cmd/whac`, `internal/cli` (Cobra), `internal/rules` and the first part of `internal/engine`. Start with a spike that compares loader designs and tests layering with default markers.

## Acceptance Criteria
- [x] `cmd/whac/main.go` is a thin main. `whac version` prints the ldflags version, commit and date. Without ldflags, it reads the Go build information.
- [x] `internal/rules/builtin/schema.cue` implements the [Rule schema](/specs/rule-schema--01a1d2c50736.md): `#SecretRule`, `#Override`, `#Config`, `#Allowlist`. `make vet-cue` passes.
- [x] Five builtin rules in `internal/rules/builtin/`: `AWS_ACCESS_KEY_ID`, `GITHUB_TOKEN`, `PRIVATE_KEY`, `JWT`, `GENERIC_API_KEY`. Each rule has `mustMatch` and `mustIgnore` samples.
- [x] Loader: builtin files and optional `--config-dir` layers → one in-memory instance → Go structs. CUE errors show file:line.
- [x] The compile step compiles each regex, lowercases keywords and stopwords, and checks `secretGroup` against the capture groups.
- [x] `whac rules list` and `whac rules verify` work. Verify runs the samples through the engine: keywords, regex, secret group, entropy, stopwords, allow regexes.
- [x] Tests: layering (disable, re-grade, add, tighten), conflicts, typos, broken regex, failed sample, export-schema, the landing-page demo rule.
- [x] A benchmark records the startup time (budget about 150 ms).

## Resolution
Done on 2026-11-12. `make lint vet-cue test build docs-style wiki docs-build` all pass.

- **Loader:** an in-memory `build.Instance` replaces the planned `cue/load` Overlay ([ADR-0012](/decisions/adr-0012-load-layers-as-one-in-memory-cue-build-instance--01a1d2e388a9.md)). The schema and the builtin rules share `internal/rules/builtin/`, so `cue vet -c` checks the rules against the schema.
- **Overrides:** the spike disproved re-grading through `rules: X: severity`. A builtin severity is concrete, so a layer value conflicts. The schema adds an `overrides` map ([ADR-0004 amendment 1](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md)). Allowlist entries and `config.exclude` became maps, so two layers can add entries without a conflict.
- **Entropy:** the planned thresholds were not reachable for short secrets. The default threshold is now `0.75 × log2(min(n, k))` ([ADR-0006 amendment 1](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md)). Shannon entropy and charset classification moved forward from Milestone 2.
- **Verify:** `engine.Verify` runs the samples. It sits in `internal/engine` to avoid an import cycle ([ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md) implementation note).
- **Assignment rules:** the compile step rejects `kind: "assignment"` until Milestone 3.
- **Performance:** `rules.Load` takes about 2 ms for 5 rules. `engine.Verify` takes about 24 µs. `whac rules list` runs in under 10 ms. The stripped binary is 17.6 MB, mostly the CUE runtime.
- **Tooling:** the installed golangci-lint was built with go1.26 and cannot type-check the go1.27 standard library. The Makefile now runs golangci-lint v2.15.0 through `go run`, so the current toolchain builds it.

## Related
[ADR-0003](/decisions/adr-0003-cue-at-load-time-only-the-scan-engine-is-plain-go--01a1d2c5075d.md) · [ADR-0004](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md) · [ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md) · [Architecture](/architecture/whac-a-leak-architecture-overview--01a1d2c507a8.md)
