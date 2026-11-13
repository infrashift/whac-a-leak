---
title: CLI
description: The whac commands, flags and exit codes.
---

## Global flags

| Flag | Default | Description |
|---|---|---|
| `--config-dir` | `.whac` | The directory of layers (`*.cue`, package `whac`). whac ignores a missing default directory. |

## Commands

| Command | Description |
|---|---|
| `whac version` | Print the version, the commit and the build date. |
| `whac rules list [--format text\|json]` | Print each rule after all layers apply. |
| `whac rules verify` | Run the samples of each rule through the engine. Print `PASS` or `FAIL` per rule. |
| `whac rules export-schema [--force]` | Write the schema to `<config-dir>/schema.cue` for editors and `cue vet`. |
| `whac scan [path]` | Scan a directory or a file. The default path is `.`. |
| `whac allow add <fingerprint> --reason "…"` | Add an allowlist entry to `<config-dir>/allowlist.cue`. |

## `whac scan` flags

| Flag | Default | Description |
|---|---|---|
| `--format` | `text` | `text` or `json`. See [Output Formats](../output-formats/). |
| `--no-redact` | off | Print each secret in full. |
| `--severity-threshold` | `LOW` | Report findings at or above `LOW`, `MEDIUM`, `HIGH` or `CRITICAL`. |
| `--staged` | off | Scan staged changes. Report only added lines. See [Pre-commit Hook](../../guides/pre-commit/). |
| `--git[=<range>]` | off | Scan the lines that each commit adds. The default range is `HEAD`. See [Scanning Git History](../../guides/git-history/). |

`--staged` and `--git` exclude each other. With either flag, the path is a
directory inside the git work tree.

`whac rules list` verifies the samples first. A failed sample stops the
command with exit code `2`.

`whac rules export-schema` overwrites a schema copy that whac wrote before. It
refuses to overwrite another file unless you add `--force`.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success. No findings at or above the threshold. |
| `1` | whac reported findings. |
| `2` | Usage error, invalid layer, failed sample, or I/O error. |

## Planned commands

SARIF output arrives with Milestone 5. `whac scan --agent-context` arrives
with Milestone 6. See the [roadmap](../../project/roadmap/).
