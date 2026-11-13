---
id: "01a1d2c5-07cf-73e8-891a-07fe24ff9c22"
title: "ISSUE-004: Staged and git-history scanning"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-4"
  - "git"
  - "closed"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T21:41:54-06:00"
description: "Milestone 4: --staged pre-commit mode and --git history mode with a streaming diff parser, commit attribution and fingerprint deduplication."
---

# ISSUE-004: Staged and git-history scanning

**Status/Milestone:** Closed · Milestone 4

## Acceptance Criteria
- [x] `whac scan --staged`: `git diff --cached -U0` supplies the added-line ranges. whac scans the full staged blobs and reports only findings on added lines.
- [x] `whac scan --git [range]`: a streaming `git log -p -U0` parser. Commit, author and date attribution. Deduplication by (fingerprint, path), with the earliest commit kept.
- [x] The parser handles renames, deletions, binary diffs, CRLF and large histories. It does not buffer the full history.
- [x] Tests build temporary repos in `t.TempDir()`.
- [x] The repo contains a `.pre-commit-hooks.yaml`.
- [x] Docs: *Pre-commit Hook*, *Scanning Git History*.

## Resolution
Done on 2026-11-12. `make lint vet-cue test build docs-style wiki docs-build` all pass.

- `internal/source`: `RepoRoot`, `Staged` (staged blobs through `git cat-file blob :path`) and `History` (streaming `git log -p -U0 --reverse`). [ADR-0014](/decisions/adr-0014-drive-git-with-pinned-options-and-count-hunk-lines--01a1d2ff1a65.md) records the pinned git options and the hunk-count parser.
- `internal/scan`: `Pipeline.Staged` keeps findings that touch an added line. `Pipeline.History` maps chunk lines back to file lines, attributes the commit and keeps the earliest commit per (fingerprint, path).
- Parser tests cover a `+++ …` content line, CRLF, binary files, deleted files, quoted paths and several hunks. Git tests cover a rename-only commit, an unstaged change, a removed and re-added secret, and an empty repository.
- `.pre-commit-hooks.yaml` defines the hook `whac` (`whac scan --staged`, language `golang`).
- Measured: 1,033 commits in about 1 s and 132 MB maximum RSS; 97 commits in 0.15 s.
- Known limit: history mode skips structured parsing, so `SENSITIVE_PATH` does not run there.
