---
id: "01a1d2c5-0782-7f2e-98fa-35ac58d71ec4"
title: "ADR-0007: Domain-separated SHA-256 fingerprints; allowlists are for false positives only"
type: "ADR"
tags:
  - "adr"
  - "fingerprint"
  - "allowlist"
  - "security"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "A finding's identity is whac1:SHA-256(rule_id || 0x00 || secret). It survives rebases. Allowlists hold only false positives, so committed fingerprints are acceptable."
---

# ADR-0007: Domain-separated SHA-256 fingerprints; allowlists are for false positives only

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
A baseline keyed by line number breaks on every rebase. A baseline of redacted strings leaks part of each secret. The proposal was a Blake2b or HMAC-SHA256 hash of `(rule_id + normalized_secret)`.

## Decision
- Fingerprint = `whac1:` + hex SHA-256 of `rule_id || 0x00 || secret`. See [Fingerprint format](/specs/fingerprint-format--01a1d2c5073d.md).
- SHA-256, not BLAKE2b: the standard library has it with hardware acceleration. BLAKE2b adds a dependency with no gain.
- No HMAC: every developer and CI job needs the key to keep allowlists portable, so the key is effectively public.
- The allowlist (`allowlist.fingerprints` in `.whac/*.cue`) holds only false positives. Rotate real secrets.

## Consequences
- An attacker can brute-force a committed hash of a low-entropy real secret. The docs state this risk. The project accepts it, because an allowlist must never hold a real secret.
- whac deduplicates findings by fingerprint across files and history.

## Alternatives considered
- **Baselines keyed by line number.** Break on rebase. Rejected.
- **HMAC with a key per repo.** No security gain. Rejected.
- **A slow KDF (argon2).** Slows scans and adds a dependency for a non-goal. Rejected.
