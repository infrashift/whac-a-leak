---
id: "01a1d2c5-07c5-73e8-b714-22cb31976238"
title: "ISSUE-003: Structure-aware scanning and assignment rules"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-3"
  - "structural"
  - "closed"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T21:36:07-06:00"
description: "Milestone 3: YAML, JSON, dotenv, properties and HCL-lite extractors, assignment rules with keyContext and sensitive paths, placeholder filter and the zero-escape entropy trap."
---

# ISSUE-003: Structure-aware scanning and assignment rules

**Status/Milestone:** Closed · Milestone 3

## Acceptance Criteria
- [x] Remove the Milestone 1 compile error for `kind: "assignment"` in `internal/rules/compile.go`. Extend `engine.Verify` to assignment rules.
- [x] Extractors produce positioned `Assignment{KeyPath, Value, Line, Col}` records for YAML, JSON, dotenv, properties/ini and HCL-style files ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md)).
- [x] A generic line extractor covers all other text files.
- [x] `assignment` rules match `keyContext` and `paths` globs. `config.sensitivePaths` in `.whac/` works.
- [x] A placeholder filter drops `${VAR}`, `{{ }}`, `<...>`, `changeme`, `xxxx`, repeated characters and empty values.
- [x] A zero-escape entropy trap rule with safe-format exclusions: UUID, hashes, paths, URLs without credentials.
- [x] Corpus: whac finds a low-entropy `database.credentials.password` in YAML. whac ignores placeholders.
- [x] Docs: *Structured Files*, *Layering Org Policy*.

## Resolution
Done on 2026-11-12. `make lint vet-cue test build docs-style wiki docs-build` all pass.

- `internal/structural` holds the extractors ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md) amendment 1). A file that does not parse falls back to unstructured text.
- Two builtin assignment rules: `SENSITIVE_PATH` (paths from `config.sensitivePaths`, low entropy allowed) and `SECRET_ASSIGNMENT` (keyContext, entropy, safe formats).
- Samples can carry a file name, so each rule tests its behavior per format.
- Safe formats: UUID, file path, URL without credentials, hash under a hash-like key ([ADR-0006](/decisions/adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md) note).
- The scanner keeps one finding per span ([ADR-0013](/decisions/adr-0013-regexes-run-on-keyword-line-windows--01a1d2f04316.md) amendment 1).
- Corpus: `testdata/corpus/structured/` adds 8 positives (YAML, dotenv, JSON, HCL, properties, invalid YAML) and 9 traps. The golden report has 15 findings.
- A new test verifies every complete CUE layer in the docs, so docs examples cannot drift.
- Throughput: 195 MB/s on the benchmark corpus.
