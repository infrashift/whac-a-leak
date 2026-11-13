---
id: "01a1d321-f364-77fd-9ac4-8f4bb984d202"
title: "ISSUE-007: Release v0.1.0"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "in-progress"
  - "release"
created: "2026-11-12T22:19:40-06:00"
timestamp: "2026-11-12T22:19:40-06:00"
description: "Tag and publish v0.1.0: Apache 2.0 license, release archives for linux and darwin on amd64 and arm64, checksums, provenance and install docs."
---

# ISSUE-007: Release v0.1.0

**Status/Milestone:** In progress · Release v0.1.0

## Description
Publish the first release after Milestones 1–6. See [ADR-0016](/decisions/adr-0016-release-cross-compiled-archives-with-checksums-and-provenance-under-apache-2-0--01a1d321f35a.md).

## Acceptance Criteria
- [x] `LICENSE` holds the Apache License 2.0 text. README and Contributing name it.
- [x] `make dist` and `make release` build archives for linux/amd64, linux/arm64, darwin/amd64 and darwin/arm64, plus `SHA256SUMS`.
- [x] `.github/workflows/release.yml`: matrix build, native smoke tests, provenance attestation, GitHub release.
- [x] Docs: install from a release or with `go install`, verification steps, pinned versions in the pre-commit and CI examples.
- [ ] Tag `v0.1.0` on a commit where CI passes. The Release workflow passes.
- [ ] The release has four archives and `SHA256SUMS`. `gh attestation verify` accepts an archive.

## Related
[ADR-0016](/decisions/adr-0016-release-cross-compiled-archives-with-checksums-and-provenance-under-apache-2-0--01a1d321f35a.md) · [ADR-0002](/decisions/adr-0002-pure-go-with-cgo-enabled0-and-a-cobra-cli--01a1d2c50754.md)
