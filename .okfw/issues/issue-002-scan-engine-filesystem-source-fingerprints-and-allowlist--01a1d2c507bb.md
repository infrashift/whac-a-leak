---
id: "01a1d2c5-07bb-77c2-babe-38b0a30814b1"
title: "ISSUE-002: Scan engine, filesystem source, fingerprints and allowlist"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-2"
  - "engine"
  - "closed"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T21:25:46-06:00"
description: "Milestone 2: Aho-Corasick prefilter, regex, entropy and charset checks, filters, whac1 fingerprints, allowlists, filesystem source and text/JSON output."
---

# ISSUE-002: Scan engine, filesystem source, fingerprints and allowlist

**Status/Milestone:** Closed · Milestone 2

## Notes from Milestone 1
- `internal/engine` already has Shannon entropy, charset classification, the length-aware default threshold and `matchRule`. `matchRule` runs every stage after the prefilter for one rule. The Milestone 2 scanner reuses it: one Aho-Corasick pass selects the candidate rules, and `matchRule` runs for each candidate.
- `engine.Verify` already runs the samples through `matchRule`. Add the new stages to `matchRule` so that verification stays identical to a scan.

## Acceptance Criteria
- [x] `internal/engine`: AC prefilter, candidate regex, `secretGroup`, charset classification and Shannon entropy ([ADR-0006](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md)), stopwords, allowRegexes, safe formats.
- [x] `whac rules verify` uses the full engine pipeline ([ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md)).
- [x] `internal/fingerprint` implements the [Fingerprint format](/specs/fingerprint-format--01a1d2c5073d.md), with test vectors.
- [x] `internal/allow`: fingerprints, path globs (doublestar), inline `whac:ignore`. Add the `whac allow add` command.
- [x] `internal/source` filesystem: `git ls-files -co --exclude-standard` in a repo, else a directory walk. Binary and size gates.
- [x] Worker pool. Sorted, deterministic output. Text and JSON reports, redacted by default. Exit codes per the [CLI contract](/specs/cli-contract--01a1d2c50744.md).
- [x] A `testdata/` corpus with golden outputs: positive samples and false-positive traps.
- [x] A throughput benchmark shows ≥100 MB/s on content with few keywords.
- [x] Docs: *How Scanning Works*, *Writing Rules*, *Allowlists & Fingerprints*, *CLI*, *Rule Schema*, *Output Formats*.

## Resolution
Done on 2026-11-12. `make lint vet-cue test build docs-style wiki docs-build` all pass.

- **Throughput:** the first benchmark gave 24 MB/s. Regexes now run on keyword line windows ([ADR-0013](/decisions/adr-0013-regexes-run-on-keyword-line-windows--01a1d2f04316.md)), and the benchmark gives about 245 MB/s on one core. The schema gained `multiline`.
- **Prefilter:** overlapping `StandardMatch` iteration. A test guards against a leftmost-longest prefilter that hides an overlapping keyword.
- **New package:** `internal/scan` holds the pipeline, so `internal/cli` stays thin. Milestone 4 reuses `Pipeline.Content` for staged blobs and history hunks.
- **Safe formats:** UUID only, and only for rules with an `entropy` block ([ADR-0006](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md) note).
- **Robustness:** an unreadable or vanished file counts as skipped. It does not stop the scan.
- **allow add:** writes `.whac/allowlist.cue`, reloads every layer, and restores the file if the entry breaks the config. It creates a missing config dir.
- **Landing demo:** `make docs-demo` runs whac on `docs/src/demo/app.env`. `TestDocsDemoOutputIsCurrent` fails when the committed output is stale. This item moved forward from ISSUE-005.
- **Corpus:** `testdata/corpus` holds 7 positives and 9 traps. `testdata/golden/corpus.json` is the expected JSON report.
