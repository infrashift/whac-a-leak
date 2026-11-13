---
id: "01a1d2c5-07d9-76e0-ac5d-8f4d906b49c8"
title: "ISSUE-005: SARIF, default ruleset, CI, docs demo and release polish"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-5"
  - "release"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Milestone 5: SARIF output, about 30 builtin rules, CI workflow with self-scan, generated landing-page demo, remaining docs and benchmarks."
---

# ISSUE-005: SARIF, default ruleset, CI, docs demo and release polish

**Status/Milestone:** Open · Milestone 5

## Acceptance Criteria
- [ ] SARIF 2.1.0 output with `partialFingerprints`. A test validates it against the schema.
- [ ] About 30 builtin rules: AWS, GCP service-account JSON, Azure, GitHub, GitLab, Slack, Stripe, Twilio, SendGrid, npm, PyPI, Docker auth, PEM private keys, JWT, generic.
- [ ] `.github/workflows/ci.yml` runs `make lint vet-cue test build` and `whac scan --git` on its own repo.
- [x] `docs/src/demo/gen.go` generates the landing-page demo with whac itself. The demo replaces the placeholder panel. Done early in Milestone 2 as `make docs-demo` (no Go generator needed).
- [ ] The remaining docs pages (*CI & SARIF*, *Default Rules*) and the README.
- [ ] The docs publish the benchmark results.
