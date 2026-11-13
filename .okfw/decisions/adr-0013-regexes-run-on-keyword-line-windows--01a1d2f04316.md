---
id: "01a1d2f0-4316-7893-a5bd-d75cb4d95e73"
title: "ADR-0013: Regexes run on keyword line windows"
type: "ADR"
tags:
  - "adr"
  - "engine"
  - "performance"
created: "2026-11-12T21:25:23-06:00"
timestamp: "2026-11-12T21:25:23-06:00"
description: "The engine runs a rule's regex only on the lines that hold one of its keywords; a multiline rule widens the window to 64 KiB, and verification uses the same windows."
---

# ADR-0013: Regexes run on keyword line windows

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Milestone 2 benchmark ([ISSUE-002](/issues/issue-002-scan-engine-filesystem-source-fingerprints-and-allowlist--01a1d2c507bb.md))

## Context
[ADR-0003](/decisions/adr-0003-cue-at-load-time-only-the-scan-engine-is-plain-go--01a1d2c5075d.md) runs a candidate rule's regex over the whole content. The first Milestone 2 benchmark measured 24 MB/s on keyword-sparse code. The target is 100 MB/s. One occurrence of `token` made `GENERIC_API_KEY` a candidate, and its regex then ran over 4 MiB. The regex has no literal prefix, so Go `regexp` checks every byte. The Aho-Corasick pass alone runs at 266 MB/s.

## Decision
- The prefilter records every keyword occurrence. It uses `StandardMatch` with overlapping iteration. A leftmost match kind hides a keyword that overlaps a longer keyword, so it can hide a candidate rule.
- For each hit, the engine takes the whole line that holds it. The engine merges overlapping windows and runs the regex on each window.
- The schema adds `multiline: *false | bool`. A multiline rule's window runs up to 64 KiB from the start of the keyword's line, rounded to a line end. `PRIVATE_KEY` sets it.
- Keyword search folds ASCII case only. Byte offsets then stay valid, and the single-rule path (`keywordHits`) agrees with the automaton.
- `engine.Verify` and `engine.MatchRule` use the same windows as a scan. A multi-line sample in a rule without `multiline` fails verification with a hint to set the field.

## Consequences
- Throughput on the benchmark corpus is about 245 MB/s on one core.
- A keyword must occur on the first line of a multi-line secret.
- `^` and `$` match the start and end of a window, not of the file.
- A single-line rule cannot match across a line break, even if its regex allows `\s`. Structured files get assignment rules in Milestone 3.

## Alternatives considered
- **Regex over the whole content.** Rejected: 24 MB/s.
- **Leftmost-longest prefilter (the gitleaks setting).** Rejected: an overlapping keyword can hide a rule.
- **Detect multi-line regexes automatically.** Rejected: `\s` makes almost every regex look multi-line.

## Amendment 1 (2026-11-12): one finding per span
A pattern rule and an assignment rule can report the same value, for example `GENERIC_API_KEY` and `SECRET_ASSIGNMENT`. The scanner keeps one match per span: the higher severity, then the lower rule ID. A test compares the spans of the multi-rule scanner with the union of single-rule scans.
