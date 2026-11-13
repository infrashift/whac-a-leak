---
id: "01a1d2c5-07b1-7b8a-a5ec-68806dcc0ce9"
title: "ISSUE-001: Rule schema, CUE loader and rules verify"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-1"
  - "rules"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Milestone 1: Cobra skeleton, embedded CUE schema and seed rules, overlay loader with user layers, decode/compile, and the rules list/verify commands."
---

# ISSUE-001: Rule schema, CUE loader and rules verify

**Status/Milestone:** Open · Milestone 1

## Description
Stand up `cmd/whac` + `internal/cli` (Cobra) and `internal/rules`. Start with a spike that confirms `cue/load` Overlay can merge embedded files with a user `.whac/` directory, and that `Decode` handles default markers.

## Acceptance Criteria
- [ ] `cmd/whac/main.go` is a thin main; `whac version` works (ldflags version/commit/date).
- [ ] `internal/rules/schema/schema.cue` implements the [Rule schema](/specs/rule-schema--01a1d2c50736.md) (`#SecretRule`, `#Config`, `#Allowlist`); `make vet-cue` passes.
- [ ] About 5 seed rules in `internal/rules/defaults/` (e.g. AWS access key, GitHub PAT, Slack token, PEM private key, generic assignment), each with `mustMatch` / `mustIgnore`.
- [ ] Loader: embedded files + optional `--config-dir` layers → unified instance → Go structs; CUE errors are reported with file:line.
- [ ] Compile step builds `CompiledRule`s (regexp compiled, keywords lowercased).
- [ ] `whac rules list` and `whac rules verify` work. Verify runs samples through a minimal matcher (regex + secretGroup); the full pipeline arrives in ISSUE-002.
- [ ] Tests: layering (disable, severity override, add, a conflict produces an error), a broken regex fails to load, a failing sample fails to load.
- [ ] Startup benchmark recorded (budget ~150 ms).

## Related
[ADR-0003](/decisions/adr-0003-cue-at-load-time-only-the-scan-engine-is-plain-go--01a1d2c5075d.md) · [ADR-0004](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md) · [ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md) · [Architecture](/architecture/whac-a-leak-architecture-overview--01a1d2c507a8.md)
