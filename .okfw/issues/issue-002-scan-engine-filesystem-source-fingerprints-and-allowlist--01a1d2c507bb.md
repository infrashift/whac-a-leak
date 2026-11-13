---
id: "01a1d2c5-07bb-77c2-babe-38b0a30814b1"
title: "ISSUE-002: Scan engine, filesystem source, fingerprints and allowlist"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-2"
  - "engine"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Milestone 2: Aho-Corasick prefilter, regex, entropy and charset checks, filters, whac1 fingerprints, allowlists, filesystem source and text/JSON output."
---

# ISSUE-002: Scan engine, filesystem source, fingerprints and allowlist

**Status/Milestone:** Open · Milestone 2

## Notes from Milestone 1
- `internal/engine` already has Shannon entropy, charset classification, the length-aware default threshold and `matchRule`. `matchRule` runs every stage after the prefilter for one rule. The Milestone 2 scanner reuses it: one Aho-Corasick pass selects the candidate rules, and `matchRule` runs for each candidate.
- `engine.Verify` already runs the samples through `matchRule`. Add the new stages to `matchRule` so that verification stays identical to a scan.

## Acceptance Criteria
- [ ] `internal/engine`: AC prefilter, candidate regex, `secretGroup`, charset classification and Shannon entropy ([ADR-0006](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md)), stopwords, allowRegexes, safe formats.
- [ ] `whac rules verify` uses the full engine pipeline ([ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md)).
- [ ] `internal/fingerprint` implements the [Fingerprint format](/specs/fingerprint-format--01a1d2c5073d.md), with test vectors.
- [ ] `internal/allow`: fingerprints, path globs (doublestar), inline `whac:ignore`. Add the `whac allow add` command.
- [ ] `internal/source` filesystem: `git ls-files -co --exclude-standard` in a repo, else a directory walk. Binary and size gates.
- [ ] Worker pool. Sorted, deterministic output. Text and JSON reports, redacted by default. Exit codes per the [CLI contract](/specs/cli-contract--01a1d2c50744.md).
- [ ] A `testdata/` corpus with golden outputs: positive samples and false-positive traps.
- [ ] A throughput benchmark shows ≥100 MB/s on content with few keywords.
- [ ] Docs: *How Scanning Works*, *Writing Rules*, *Allowlists & Fingerprints*, *CLI*, *Rule Schema*, *Output Formats*.
