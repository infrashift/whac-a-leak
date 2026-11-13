---
id: "01a1d321-f35a-70ee-b73c-735a9bc06ebe"
title: "ADR-0016: Release cross-compiled archives with checksums and provenance under Apache 2.0"
type: "ADR"
tags:
  - "adr"
  - "release"
  - "license"
  - "supply-chain"
created: "2026-11-12T22:19:39-06:00"
timestamp: "2026-11-12T22:19:39-06:00"
description: "A v* tag builds tar.gz archives for linux and darwin on amd64 and arm64, smoke-tests each on its platform, and publishes them with SHA256SUMS and a SLSA provenance attestation; the project uses Apache 2.0."
---

# ADR-0016: Release cross-compiled archives with checksums and provenance under Apache 2.0

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Request for the first release, v0.1.0

## Context
whac is a static pure-Go binary ([ADR-0002](/decisions/adr-0002-pure-go-with-cgo-enabled0-and-a-cobra-cli--01a1d2c50754.md)). Users need it on Linux CI runners and on developer machines with Intel and Apple-silicon Macs. A security tool must let users check that a download is the binary that CI built. The repository had no license.

## Decision
- **License:** Apache License 2.0, the same as `cuix` and `garmr-policy-agent`. Each archive contains `LICENSE` and `README.md`.
- **Platforms:** `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. AArch64 and ARM64 name the same architecture.
- **Build:** `make dist` cross-compiles one platform with `CGO_ENABLED=0`, `-trimpath` and version ldflags. `make release` builds all platforms and `SHA256SUMS`. The same targets run locally and in CI.
- **Workflow** (`.github/workflows/release.yml`, on a `v*` tag):
  1. A matrix job builds the four archives on Linux.
  2. A matrix job runs each binary on its own platform (Ubuntu amd64, Ubuntu arm64, macOS arm64; macOS amd64 through Rosetta 2). It checks `whac version`, `whac rules verify` and a scan of a fake token.
  3. A release job writes `SHA256SUMS`, creates a SLSA build provenance attestation (`actions/attest-build-provenance`), and creates the GitHub release with generated notes.
- **Versions:** tags use the `v` prefix that Go modules need. A 0.x release is a normal release, not a prerelease, so "latest release" works.

## Consequences
- Users verify a download with `SHA256SUMS` and `gh attestation verify`.
- `go install github.com/infrashift/whac-a-leak/cmd/whac@<tag>` works for every tag.
- The macOS amd64 smoke test depends on Rosetta 2 on the arm64 runner. If Rosetta is missing, the job fails fast.

## Alternatives considered
- **GoReleaser.** Rejected for now: one more tool and config file for four `go build` calls. Reconsider it when the project adds packages such as Homebrew or deb.
- **Native builds on each runner.** Rejected: pure Go cross-compiles. The native runners run the smoke tests instead.
