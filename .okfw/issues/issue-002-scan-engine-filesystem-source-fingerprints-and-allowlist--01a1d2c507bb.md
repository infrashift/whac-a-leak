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
description: "Milestone 2: Aho-Corasick prefilter, regex/entropy/charset matching, filters, whac1 fingerprints, allowlists, filesystem source and text/json output."
---

# ISSUE-002: Scan engine, filesystem source, fingerprints and allowlist

**Status/Milestone:** Open · Milestone 2

## Acceptance Criteria
- [ ] `internal/engine`: AC prefilter, candidate regex, `secretGroup`, charset classification + Shannon entropy ([ADR-0006](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md)), stopwords / allowRegexes / safe formats.
- [ ] `whac rules verify` uses the full engine pipeline ([ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md)).
- [ ] `internal/fingerprint` implements [Fingerprint format](/specs/fingerprint-format--01a1d2c5073d.md) with test vectors.
- [ ] `internal/allow`: fingerprints, path globs (doublestar), inline `whac:ignore`; `whac allow add`.
- [ ] `internal/source` filesystem: `git ls-files -co --exclude-standard` in a repo, a walk otherwise; binary and size gates.
- [ ] Worker pool; deterministic sorted output; text + JSON reports, redacted by default; exit codes per [CLI contract](/specs/cli-contract--01a1d2c50744.md).
- [ ] `testdata/` corpus with golden outputs (positives and false-positive traps).
- [ ] Throughput benchmark ≥100 MB/s on keyword-sparse content.
- [ ] Docs: *How Scanning Works*, *Writing Rules*, *Allowlists & Fingerprints*, *CLI*, *Rule Schema*, *Output Formats*.
