---
id: "01a1d2c5-07cf-73e8-891a-07fe24ff9c22"
title: "ISSUE-004: Staged and git-history scanning"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-4"
  - "git"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Milestone 4: --staged pre-commit mode and --git history mode with a streaming diff parser, commit attribution and fingerprint deduplication."
---

# ISSUE-004: Staged and git-history scanning

**Status/Milestone:** Open · Milestone 4

## Acceptance Criteria
- [ ] `whac scan --staged`: added-line ranges from `git diff --cached -U0`; staged blobs scanned in full; only added-line findings reported.
- [ ] `whac scan --git [range]`: streaming `git log -p -U0` parser; commit/author/date attribution; dedup by (fingerprint, path) to the earliest commit.
- [ ] Handles renames, deletions, binary diffs, CRLF and very large histories without buffering everything.
- [ ] Tests build throwaway repos in `t.TempDir()`.
- [ ] `.pre-commit-hooks.yaml` shipped.
- [ ] Docs: *Pre-commit Hook*, *Scanning Git History*.
