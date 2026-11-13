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
- [ ] `whac scan --staged`: `git diff --cached -U0` supplies the added-line ranges. whac scans the full staged blobs and reports only findings on added lines.
- [ ] `whac scan --git [range]`: a streaming `git log -p -U0` parser. Commit, author and date attribution. Deduplication by (fingerprint, path), with the earliest commit kept.
- [ ] The parser handles renames, deletions, binary diffs, CRLF and large histories. It does not buffer the full history.
- [ ] Tests build temporary repos in `t.TempDir()`.
- [ ] The repo contains a `.pre-commit-hooks.yaml`.
- [ ] Docs: *Pre-commit Hook*, *Scanning Git History*.
