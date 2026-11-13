---
id: "01a1d2ff-1a65-7691-bf6b-e50734269eef"
title: "ADR-0014: Drive git with pinned options and count hunk lines"
type: "ADR"
tags:
  - "adr"
  - "git"
  - "source"
created: "2026-11-12T21:41:36-06:00"
timestamp: "2026-11-12T21:41:36-06:00"
description: "whac runs the git binary with options that override user config, and parses -U0 diffs by hunk line counts, so diff output always has one known shape."
---

# ADR-0014: Drive git with pinned options and count hunk lines

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Milestone 4 ([ISSUE-004](/issues/issue-004-staged-and-git-history-scanning--01a1d2c507cf.md))

## Context
[ADR-0002](/decisions/adr-0002-pure-go-with-cgo-enabled0-and-a-cobra-cli--01a1d2c50754.md) drives git through the `git` binary. User configuration changes the shape of git output. Examples: `diff.noprefix` drops the `a/` and `b/` prefixes, `log.showSignature` adds signature lines, `core.quotePath` escapes non-ASCII paths, and `diff.external` replaces the diff. An added line whose text starts with `++ ` appears in the diff as `+++ …`, the same prefix as a file header.

## Decision
- Every git call passes `-c core.quotePath=false -c log.showSignature=false -c diff.noprefix=false -c color.ui=never`.
- Diffs use `-U0 --no-color --no-ext-diff --no-textconv --no-relative --src-prefix=a/ --dst-prefix=b/ -M`.
- History uses `git log -p --reverse` with a commit header line that starts with `\x1eCOMMIT\x1f`. Diff content lines always start with ` `, `+`, `-` or `\\`, so the header cannot collide with content.
- The parser counts the old and new lines that each hunk header announces. While a hunk has lines left, every line is content, even `+++ …`.
- The parser streams. It keeps one file of one commit in memory.
- whac rejects a revision range that starts with `-`, so the range cannot act as a git option.
- Tests isolate git with `GIT_CONFIG_GLOBAL=/dev/null` and `GIT_CONFIG_NOSYSTEM=1`.

## Consequences
- Output parsing does not depend on the user's git configuration.
- A history scan of 1,033 commits (13.7 MiB of added lines) takes about 1 s.
- History mode scans hunks as unstructured text ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md)).

## Alternatives considered
- **go-git.** Rejected in ADR-0002.
- **`git log --numstat` plus `git show` per file.** Rejected: one process per file is slow on long histories.
