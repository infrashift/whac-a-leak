---
id: "01a1d369-9d8b-7b8e-ba6c-c262f684999f"
title: "ISSUE-008: Report files with --output"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-7"
  - "output"
  - "closed"
created: "2026-11-12T23:37:56-06:00"
timestamp: "2026-11-12T23:37:56-06:00"
description: "Milestone 7: whac scan writes reports to files with a repeatable --output FORMAT=PATH flag, so one scan gives text on the screen and JSON or SARIF files."
---

# ISSUE-008: Report files with --output

# ISSUE-008: Report files with --output

**Status/Milestone:** Closed · Milestone 7

## Description
One scan must give readable text on the terminal and a machine-readable report. See [ADR-0017](/decisions/adr-0017-report-files-through-a-repeatable-output-flag-json-stays-one-document--01a1d36982f5.md).

## Acceptance Criteria
- [x] `whac scan --output FORMAT=PATH`, repeatable; `--format` still sets stdout.
- [x] whac checks each value before the scan; bad values give exit code 2 and create no file.
- [x] Files have mode 0600, no colors, and the same content as `--format` on stdout. `--no-redact` applies.
- [x] Temporary file and rename: a failed write leaves no partial report.
- [x] Works with `--staged`, `--git` and the agent-context modes.
- [x] Tests in `internal/cli/output_test.go`.
- [x] Docs: CLI reference, Output Formats, CI & SARIF. The CI self-scan uses `--output`.
- [x] Release v0.2.0.

## Related
[ADR-0017](/decisions/adr-0017-report-files-through-a-repeatable-output-flag-json-stays-one-document--01a1d36982f5.md) · [CLI contract](/specs/cli-contract--01a1d2c50744.md)

## Resolution
Done on 2026-11-12. [PR #1](https://github.com/infrashift/whac-a-leak/pull/1) merged as `b17a45f` and `f327201`. The tag `v0.2.0` points to `f327201`, on which CI passed.

- `whac scan --output FORMAT=PATH` writes text, JSON or SARIF files from one scan. `--format` still sets stdout.
- The CI self-scan uses `--output sarif=whac.sarif`. The job log now shows the text summary.
- The Release workflow passed all nine jobs. The release https://github.com/infrashift/whac-a-leak/releases/tag/v0.2.0 has four archives and `SHA256SUMS`. All checksums match, and each archive has a provenance attestation.
- The JSON report stays one document ([ADR-0017](/decisions/adr-0017-report-files-through-a-repeatable-output-flag-json-stays-one-document--01a1d36982f5.md)).
