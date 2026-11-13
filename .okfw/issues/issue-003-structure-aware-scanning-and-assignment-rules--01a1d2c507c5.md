---
id: "01a1d2c5-07c5-73e8-b714-22cb31976238"
title: "ISSUE-003: Structure-aware scanning and assignment rules"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-3"
  - "structural"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Milestone 3: YAML, JSON, dotenv, properties and HCL-lite extractors, assignment rules with keyContext and sensitive paths, placeholder filter and the zero-escape entropy trap."
---

# ISSUE-003: Structure-aware scanning and assignment rules

**Status/Milestone:** Open · Milestone 3

## Acceptance Criteria
- [ ] Remove the Milestone 1 compile error for `kind: "assignment"` in `internal/rules/compile.go`. Extend `engine.Verify` to assignment rules.
- [ ] Extractors produce positioned `Assignment{KeyPath, Value, Line, Col}` records for YAML, JSON, dotenv, properties/ini and HCL-style files ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md)).
- [ ] A generic line extractor covers all other text files.
- [ ] `assignment` rules match `keyContext` and `paths` globs. `config.sensitivePaths` in `.whac/` works.
- [ ] A placeholder filter drops `${VAR}`, `{{ }}`, `<...>`, `changeme`, `xxxx`, repeated characters and empty values.
- [ ] A zero-escape entropy trap rule with safe-format exclusions: UUID, hashes, paths, URLs without credentials.
- [ ] Corpus: whac finds a low-entropy `database.credentials.password` in YAML. whac ignores placeholders.
- [ ] Docs: *Structured Files*, *Layering Org Policy*.
