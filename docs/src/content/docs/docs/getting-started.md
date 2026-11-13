---
title: Getting Started
description: Build Whac-a-Leak from source and see which features work now.
---

Whac-a-Leak has no tagged release. Build it from source.

## Prerequisites

- **Go 1.25+**. The build uses `CGO_ENABLED=0`, so you need no C toolchain.
- **git**. whac runs the `git` binary for file lists that obey
  `.gitignore`, for staged changes and for history.
- **cue** (optional). Rule authors use it to run `cue vet` on their `.whac/`
  layers.

## Build

```sh
git clone https://github.com/infrashift/whac-a-leak
cd whac-a-leak
make build        # → bin/whac
```

## Scan a directory

1. Run `bin/whac scan .` in the directory to scan.
2. Read the findings. Each finding shows the severity, the rule ID, the
   location, the redacted secret and the fingerprint.
3. Read the exit code: `0` for no findings, `1` for findings.

* **Expected result:** whac prints the findings and a summary line.

To add your own rules, read [Writing Rules](../guides/writing-rules/). To
suppress a false positive, read [Allowlists & Fingerprints](../guides/allowlists/).

## Planned commands

The project CLI spec defines these commands. They do **not exist yet**:

```sh
whac scan --staged               # pre-commit: scan only the lines you commit
whac scan --git main..HEAD       # scan the history of a branch
whac scan . --format sarif       # output for GitHub code scanning
whac scan --agent-context        # preflight: env vars, credential files, sockets
```

Exit codes:

| Code | Meaning |
|---|---|
| `0` | No findings |
| `1` | One or more findings |
| `2` | Error: bad configuration, failed rule sample, or I/O error |
