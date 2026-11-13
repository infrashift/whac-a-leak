---
id: "01a1d369-9d8b-7b8e-ba6c-c262f684999f"
title: "ISSUE-008: Report files with --output"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-7"
  - "output"
  - "in-progress"
created: "2026-11-12T23:37:56-06:00"
timestamp: "2026-11-12T23:37:56-06:00"
description: "Milestone 7: whac scan writes reports to files with a repeatable --output FORMAT=PATH flag, so one scan gives text on the screen and JSON or SARIF files."
---

# ISSUE-008: Report files with --output

# ISSUE-008: Report files with --output

**Status/Milestone:** In progress · Milestone 7

## Description
One scan must give readable text on the terminal and a machine-readable report. See [ADR-0017](/decisions/adr-0017-report-files-through-a-repeatable-output-flag-json-stays-one-document--01a1d36982f5.md).

## Acceptance Criteria
- [ ] `whac scan --output FORMAT=PATH`, repeatable; `--format` still sets stdout.
- [ ] whac checks each value before the scan; bad values give exit code 2 and create no file.
- [ ] Files have mode 0600, no colors, and the same content as `--format` on stdout. `--no-redact` applies.
- [ ] Temporary file and rename: a failed write leaves no partial report.
- [ ] Works with `--staged`, `--git` and the agent-context modes.
- [ ] Tests in `internal/cli/output_test.go`.
- [ ] Docs: CLI reference, Output Formats, CI & SARIF. The CI self-scan uses `--output`.
- [ ] Release v0.2.0.

## Related
[ADR-0017](/decisions/adr-0017-report-files-through-a-repeatable-output-flag-json-stays-one-document--01a1d36982f5.md) · [CLI contract](/specs/cli-contract--01a1d2c50744.md)
