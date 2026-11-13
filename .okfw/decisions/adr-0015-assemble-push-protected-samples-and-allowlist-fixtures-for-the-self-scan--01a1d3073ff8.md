---
id: "01a1d307-3ff8-7685-b537-a7284bda2bbf"
title: "ADR-0015: Assemble push-protected samples and allowlist fixtures for the self-scan"
type: "ADR"
tags:
  - "adr"
  - "rules"
  - "ci"
  - "security"
created: "2026-11-12T21:50:30-06:00"
timestamp: "2026-11-12T21:50:30-06:00"
description: "Rule samples for token formats that GitHub push protection detects are assembled from pieces at load time; the repository scans itself in CI with path and fingerprint allowlists for its fixtures."
---

# ADR-0015: Assemble push-protected samples and allowlist fixtures for the self-scan

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Milestone 5 ([ISSUE-005](/issues/issue-005-sarif-default-ruleset-ci-docs-demo-and-release-polish--01a1d2c507d9.md))

## Context
Every rule needs samples that the rule detects ([ADR-0009](/decisions/adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md)). A sample for Slack, Stripe, npm or a similar format looks like a real token. GitHub push protection can block a push that contains such a value. whac also scans its own repository in CI. Without an allowlist, every fixture is a finding.

## Decision
- For a format that push protection detects, the rule file builds the sample from pieces with CUE string concatenation, for example `"sk" + "_live_" + "EXAMPLE…"`. CUE joins the pieces at load time. No complete token appears in the file.
- Samples keep an `EXAMPLE` marker or an obviously fake value (an invalid checksum, a counting sequence).
- `.whac/repo.cue` allowlists three fixture directories by path: `testdata/**`, `internal/rules/builtin/**` and `docs/src/demo/**`.
- `.whac/allowlist.cue` allowlists each fake value in a test file or a docs page by fingerprint, through `whac allow add`. A new real secret in a test file is still a finding.
- CI runs `whac scan --git .` and fails on any finding. CI also uploads SARIF from the work-tree scan to GitHub code scanning.

## Consequences
- A push of the builtin rule files needs no push-protection bypass.
- A new fake value in a test file needs a `whac allow add` entry. This keeps the allowlist deliberate.

## Alternatives considered
- **Allowlist all test files by path.** Rejected: whac then misses a real secret in a test file.
- **Literal samples plus a push-protection bypass.** Rejected: push protection then blocks each contributor.
