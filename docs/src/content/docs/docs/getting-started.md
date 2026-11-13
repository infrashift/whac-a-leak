---
title: Getting Started
description: Install Whac-a-Leak and run a first scan.
---

The current release is **v0.1.0**. Release archives exist for Linux and
macOS, each on amd64 and arm64 (AArch64, including Apple silicon).

## Prerequisites

- **git**. whac runs the `git` binary for file lists that obey
  `.gitignore`, for staged changes and for history.
- **cue** (optional). Rule authors use it to run `cue vet` on their `.whac/`
  layers.

## Install a release

1. Download the archive for your platform and `SHA256SUMS` from the
   [v0.1.0 release](https://github.com/infrashift/whac-a-leak/releases/tag/v0.1.0).
   The archive name is `whac_0.1.0_<os>_<arch>.tar.gz`, for example
   `whac_0.1.0_darwin_arm64.tar.gz`.
2. Verify the checksum: `sha256sum --check --ignore-missing SHA256SUMS`.
   On macOS, use `shasum -a 256 --check --ignore-missing SHA256SUMS`.
3. Verify the build provenance:
   `gh attestation verify whac_0.1.0_darwin_arm64.tar.gz --repo infrashift/whac-a-leak`.
4. Extract the archive, and put `whac` on your `PATH`.
5. Run `whac version`.

* **Expected result:** whac prints `whac v0.1.0` with the commit and the
  build date.

## Install with Go

Run `go install github.com/infrashift/whac-a-leak/cmd/whac@v0.1.0`. This
needs Go 1.25 or newer.

## Build from source

```sh
git clone https://github.com/infrashift/whac-a-leak
cd whac-a-leak
make build        # → bin/whac
make release      # → dist/: archives for every platform and SHA256SUMS
```

## Scan a directory

1. Run `whac scan .` in the directory to scan.
2. Read the findings. Each finding shows the severity, the rule ID, the
   location, the redacted secret and the fingerprint.
3. Read the exit code: `0` for no findings, `1` for findings.

* **Expected result:** whac prints the findings and a summary line.

To check commits before they happen, read
[Pre-commit Hook](../guides/pre-commit/). To scan old commits, read
[Scanning Git History](../guides/git-history/). To add your own rules, read
[Writing Rules](../guides/writing-rules/). To
suppress a false positive, read [Allowlists & Fingerprints](../guides/allowlists/).

## Audit an agent's context

Run `whac scan --agent-context` before an AI agent starts. See
[Agent-Context Preflight](../guides/agent-context/).

Exit codes:

| Code | Meaning |
|---|---|
| `0` | No findings |
| `1` | One or more findings |
| `2` | Error: bad configuration, failed rule sample, or I/O error |
