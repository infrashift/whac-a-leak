---
id: "01a1d2c5-0744-7c2a-b546-061453b7dfd4"
title: "CLI contract"
type: "Spec"
tags:
  - "spec"
  - "cli"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "The whac command surface: subcommands, flags, output formats and exit codes."
---

# CLI contract

**Status:** Draft (built up over Milestones 1–5 in `internal/cli`)

## Commands
| Command | Purpose |
|---|---|
| `whac scan [path]` | Scan a directory or file. Inside a git repo the file list comes from `git ls-files -co --exclude-standard`, so `.gitignore` is respected. |
| `whac scan --staged` | Pre-commit mode: scan staged blobs and report only findings on added lines. |
| `whac scan --git [rev-range]` | Scan the history's added lines with commit attribution, deduplicated by (fingerprint, path) to the earliest commit. |
| `whac scan --agent-context [--root DIR]… [--probe-imds]` | Preflight gate ([ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md)): environment variables + well-known credential files under `$HOME` and each `--root` + exposure checks (credential sockets; IMDS only when `--probe-imds` is set). Values are never printed; `--no-redact` has no effect. |
| `whac scan --env` / `whac scan --home` | The agent-context sources on their own. |
| `whac rules list` | Effective rules after layering: ID, kind, severity, enabled. |
| `whac rules verify` | Run every rule's tests through the engine and print a pass/fail table. |
| `whac rules export-schema` | Write the CUE schema into `.whac/` so `cue vet` and editors work there. |
| `whac allow add <fingerprint> --reason "…"` | Append an allowlist entry to `.whac/allowlist.cue`. |
| `whac version` | Version, commit, build date. |

## Common flags
`--format text|json|sarif` (default `text`) · `--config-dir .whac` · `--no-redact` (secrets are redacted by default, e.g. `ghp_****abcd`) · `--severity-threshold LOW|MEDIUM|HIGH|CRITICAL`

## Exit codes
| Code | Meaning |
|---|---|
| 0 | No findings at or above the threshold |
| 1 | Findings reported |
| 2 | Usage, configuration, rule-verification or I/O error |

## Finding categories
- **secret**: a matched secret string (every source).
- **exposure**: reachable access to credentials rather than a secret string, e.g. a connectable `SSH_AUTH_SOCK`, an accessible `docker.sock`, a reachable IMDS. Agent-context mode only.

## Outputs
- **text**: human-readable, grouped by file, with severity colour when stdout is a TTY.
- **json**: an object with an `apiVersion`/`kind` envelope and a `findings` array (rule, severity, path, line, column, commit?, fingerprint, redacted match).
- **sarif**: SARIF 2.1.0 for GitHub code scanning; `partialFingerprints` carries the `whac1:` fingerprint.
