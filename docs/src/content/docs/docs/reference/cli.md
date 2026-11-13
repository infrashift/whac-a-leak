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

The scan commands arrive with Milestone 2 and later. See the
[roadmap](../../project/roadmap/).
