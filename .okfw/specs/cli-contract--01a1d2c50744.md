---
id: "01a1d2c5-0744-7c2a-b546-061453b7dfd4"
title: "CLI contract"
type: "Spec"
tags:
  - "spec"
  - "cli"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "The whac command surface: subcommands, flags, finding categories, output formats and exit codes."
---

# CLI contract

**Status:** Draft. Milestones 1–6 build it in `internal/cli`.

## Commands
| Command | Purpose |
|---|---|
| `whac scan [path]` | Scan a directory or file. Inside a git repo, `git ls-files -co --exclude-standard` supplies the file list, so whac obeys `.gitignore`. |
| `whac scan --staged` | Pre-commit mode: scan the staged blobs. Report only findings on added lines. |
| `whac scan --git [rev-range]` | Scan the added lines in history with commit attribution. Deduplicate by (fingerprint, path) and keep the earliest commit. |
| `whac scan --agent-context [--root DIR]… [--probe-imds]` | Preflight gate ([ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md)): environment variables, well-known credential files under `$HOME` and each `--root`, and exposure checks. The IMDS check runs only with `--probe-imds`. whac never prints values. `--no-redact` has no effect. |
| `whac scan --env` / `whac scan --home` | Run one agent-context source alone. |
| `whac rules list` | Show the effective rules after layering: ID, kind, severity, enabled. |
| `whac rules verify` | Run every rule sample through the engine. Print a pass/fail table. |
| `whac rules export-schema [--force]` | Write the CUE schema to `<config-dir>/schema.cue` with a generated-file header. The loader skips that file. whac overwrites its own copy and refuses another file unless `--force` is set. |
| `whac allow add <fingerprint> --reason "…"` | Add an allowlist entry to `.whac/allowlist.cue`. |
| `whac version` | Show the version, commit and build date. |

## Common flags
`--format text|json|sarif` (default `text`) · `--config-dir .whac` · `--no-redact` (whac redacts secrets by default, for example `ghp_****abcd`) · `--severity-threshold LOW|MEDIUM|HIGH|CRITICAL`

## Exit codes
| Code | Meaning |
|---|---|
| 0 | No findings at or above the threshold |
| 1 | One or more findings |
| 2 | Usage, configuration, rule-verification or I/O error |

## Finding categories
- **secret**: a matched secret string. All sources produce this category.
- **exposure**: reachable access to credentials, not a secret string. Examples: a connectable `SSH_AUTH_SOCK`, an accessible `docker.sock`, a reachable IMDS. Only agent-context mode produces this category.

## Outputs
- **text**: readable output, grouped by file. Severity has color when stdout is a TTY.
- **json**: an object with an `apiVersion`/`kind` envelope and a `findings` array (rule, severity, path, line, column, commit?, fingerprint, redacted match).
- **sarif**: SARIF 2.1.0 for GitHub code scanning. `partialFingerprints` carries the `whac1:` fingerprint.

## Config dir
- `--config-dir` sets the directory of layers. The default is `.whac`.
- A missing default directory means no layers. A missing directory that the user sets with `--config-dir` gives exit code 2.
- `whac rules list` and every scan command verify the rule samples first. A failed sample gives exit code 2.

## Implemented
Milestone 1: `whac version`, `whac rules list [--format text|json]`, `whac rules verify`, `whac rules export-schema [--force]`.

Milestone 2: `whac scan [path]` with `--format text|json`, `--no-redact` and `--severity-threshold`; `whac allow add <fingerprint> --reason`.
- JSON envelope: `apiVersion: whac.infrashift.io/v1alpha1`, `kind: ScanReport`, `summary` (counts and skip statistics), `findings` (never null).
- Redaction: the first and last four characters of a single-line secret with 16 or more characters; `****` for other secrets.
- Columns count characters (runes) and start at 1.
- Unreadable files do not stop a scan. The summary counts them.

Milestone 3: JSON findings of assignment rules carry `keyPath`. The text report prints it as `key:`.

Milestone 4: `whac scan --staged` and `whac scan --git[=<range>]` ([ADR-0014](/decisions/adr-0014-drive-git-with-pinned-options-and-count-hunk-lines--01a1d2ff1a65.md)).
- `--git` has an optional value, so a range needs `=`: `--git=main..HEAD`. A bare `--git` scans `HEAD`. Without `=`, the next word is the path; whac then hints at `--git=<range>`.
- `--staged` and `--git` exclude each other. The path is a directory in the work tree. Paths in findings are relative to the repository root.
- History findings carry `commit`, `author` (name only, no email) and `date`. The summary counts commits and file changes.

Milestone 5: `--format sarif` (SARIF 2.1.0).
- `tool.driver.rules` lists the enabled rules with `security-severity` (9.5, 8.0, 5.5, 3.0) and the level `error`, `warning` or `note`.
- `partialFingerprints.whacFingerprint/v1` holds the `whac1:` fingerprint. History findings carry `commit`, `author` and `date` in `properties`.
- SARIF always redacts secrets. `--no-redact` does not apply.
- A test validates the output against the OASIS SARIF 2.1.0 JSON Schema, which `testdata/sarif/` vendors. CUE's `encoding/jsonschema` reads the schema, so the test needs no new dependency.

Milestone 6: `--agent-context`, `--env`, `--home`, `--root` (repeatable) and `--probe-imds` ([ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md)).
- The agent-context flags take no path and exclude `--staged` and `--git`. `--probe-imds` needs `--agent-context`. `--root` needs `--agent-context` or `--home`.
- Findings have a `category` (`secret` or `exposure`). Agent-context findings print `[masked]`; `--no-redact` has no effect.
- Environment findings have the path `env`, the key path `env.<NAME>` and no line. Exposure findings have the socket path or URL as their path. SARIF omits the region when there is no line.
