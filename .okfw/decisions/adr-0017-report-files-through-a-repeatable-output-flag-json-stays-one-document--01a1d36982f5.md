---
id: "01a1d369-82f5-7294-8cf6-be0af47c28c8"
title: "ADR-0017: Report files through a repeatable --output flag; JSON stays one document"
type: "ADR"
tags:
  - "adr"
  - "output"
  - "cli"
created: "2026-11-12T23:37:49-06:00"
timestamp: "2026-11-12T23:37:49-06:00"
description: "whac scan writes extra reports to files with --output FORMAT=PATH while --format sets stdout; the JSON report stays one document and whac does not add JSON Lines."
---

# ADR-0017: Report files through a repeatable --output flag; JSON stays one document

# ADR-0017: Report files through a repeatable --output flag; JSON stays one document

**Status:** Accepted
**Date:** 2026-11-12
**Source:** User request: text on the screen and a JSON report from one scan

## Context
`whac scan --format text|json|sarif` wrote every report to stdout ([CLI contract](/specs/cli-contract--01a1d2c50744.md)). A user who wanted readable text and a machine-readable report had to scan two times. The question also came up: JSON Lines, one JSON document, or both?

## Decision
- **`--output FORMAT=PATH`**, repeatable. `--format` keeps its meaning: the report on stdout. Example: `whac scan . --output json=whac.json --output sarif=whac.sarif`.
- whac checks each value before rules load. These values give exit code 2: a missing `=`, an empty path, an unknown format, a duplicate path (after `filepath.Clean`).
- Each file has mode `0600` and no ANSI colors. `--no-redact` applies to files and stdout the same way. SARIF always redacts.
- whac writes a temporary file in the target directory and renames it. A failed write leaves no partial report and gives exit code 2.
- **JSON stays one document.** whac adds no `jsonl` format.

## Reasons for one JSON document
- The pipeline collects and sorts all findings before output (`scan.Sort`, then `scan.Filter`). JSON Lines gives no streaming.
- The envelope (`apiVersion`, `kind`, `summary`) is the versioned contract. JSON Lines needs a second contract for the summary line.
- SARIF, the CI format, is also one document.
- `jq -c '.findings[]' whac.json` gives one finding per line.

## Rejected
- **JSON Lines as well as JSON:** two contracts to version for no streaming gain. Revisit when the engine streams findings, or when a log or SIEM consumer needs it.
- **`-o PATH` with the format from the file extension:** implicit, and `.json` cannot tell JSON from SARIF.
- **One flag per format (`--json-output`, `--sarif-output`):** a new flag for each new format.

## Consequences
- The CI self-scan uses `--output sarif=whac.sarif`, so the job log shows the text report.
- Tracking: [ISSUE-008](/issues/issue-008-report-files-with-output--01a1d3699d8b.md).
