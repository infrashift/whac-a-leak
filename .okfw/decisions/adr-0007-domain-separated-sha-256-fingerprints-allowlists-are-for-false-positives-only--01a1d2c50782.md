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
description: "Findings are identified by whac1:SHA-256(rule_id || 0x00 || secret), which is stable across rebases; committed fingerprints are acceptable because allowlists only hold false positives."
---

# ADR-0007: Domain-separated SHA-256 fingerprints; allowlists are for false positives only

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
Baselines keyed by line number break on every rebase. Redacted-string baselines leak partial secrets. The proposal was a Blake2b or HMAC-SHA256 hash of `(rule_id + normalized_secret)`.

## Decision
- Fingerprint = `whac1:` + hex SHA-256 of `rule_id || 0x00 || secret`. See [Fingerprint format](/specs/fingerprint-format--01a1d2c5073d.md).
- SHA-256 instead of BLAKE2b: it's in the standard library and hardware-accelerated, while BLAKE2b would add a dependency for no gain.
- No HMAC: a key shared by every developer and CI job, so that allowlists stay portable, would be effectively public.
- The allowlist (`.whac/*.cue` → `allowlist.fingerprints`) holds only false positives. Real secrets are rotated.

## Consequences
- A committed hash of a low-entropy real secret could be brute-forced. Documented and accepted, because real secrets must never be allowlisted.
- Findings deduplicate by fingerprint across files and history.

## Alternatives considered
- **Line-number baselines.** Break on rebase. Rejected.
- **HMAC with a per-repo key.** No security gain. Rejected.
- **A slow KDF (argon2).** Slows scans and adds a dependency for a non-goal. Rejected.
