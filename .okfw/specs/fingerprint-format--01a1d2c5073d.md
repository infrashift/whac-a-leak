---
id: "01a1d2c5-073d-7e9a-8614-f535ba8a16fb"
title: "Fingerprint format"
type: "Spec"
tags:
  - "spec"
  - "fingerprint"
  - "allowlist"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "The location-independent whac1: finding fingerprint used by allowlists, deduplication and reports."
---

# Fingerprint format

**Status:** Draft (implemented in Milestone 2 as `internal/fingerprint`)

```
whac1:<hex(SHA-256(rule_id || 0x00 || secret))>
```

- `rule_id` is the rule's ID string. `0x00` is a domain separator, so `("AB", "C…")` and `("A", "BC…")` can't collide.
- `secret` is the bytes of the rule's `secretGroup` capture (or the assignment value). The only normalization is trimming surrounding whitespace and matching quote characters. There's no case folding or Unicode normalization.
- The fingerprint doesn't include the path, line or commit. It survives rebases, moves and reformatting, so `.whac/` allowlists never go stale over line numbers.
- The `whac1:` prefix versions the scheme. A future change of algorithm gets `whac2:`, and loaders accept both for one release.

## Security caveat
A committed fingerprint of a **low-entropy** real secret (such as a short password) can be brute-forced offline. Allowlists exist for false positives: fixtures, documented examples, test keys. A real leaked secret gets rotated, never allowlisted. The docs and `whac allow add` help text both say so.
