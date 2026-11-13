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
description: "The location-independent whac1: fingerprint that identifies a finding in allowlists, deduplication and reports."
---

# Fingerprint format

**Status:** Draft. Milestone 2 implements it in `internal/fingerprint`.

```
whac1:<hex(SHA-256(rule_id || 0x00 || secret))>
```

- `rule_id` is the rule ID string. The `0x00` byte separates the two fields. Thus `("AB", "C…")` and `("A", "BC…")` give different fingerprints.
- `secret` is the bytes of the `secretGroup` capture or of the assignment value. The engine removes surrounding whitespace and matching quote characters. It does no other normalization: no case folding, no Unicode normalization.
- The fingerprint does not include the path, line or commit. It stays the same after a rebase, a move or a reformat. Allowlist entries do not go stale on line numbers.
- The `whac1:` prefix carries the scheme version. A new algorithm gets `whac2:`. The loader accepts both prefixes for one release.

## Security caveat
An attacker can brute-force a committed fingerprint of a **low-entropy** real secret, such as a short password. Allowlists exist only for false positives: fixtures, documented examples, test keys. Rotate a real leaked secret. Do not allowlist it. The docs and the `whac allow add` help text state this rule.
