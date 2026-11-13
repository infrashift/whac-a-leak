---
id: "01a1d2c5-0779-7ac4-9488-aa24bc724d91"
title: "ADR-0006: Entropy thresholds are charset-aware"
type: "ADR"
tags:
  - "adr"
  - "entropy"
  - "engine"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "The engine compares Shannon entropy with thresholds for each charset and a minimum length, because the alphabet limits the maximum entropy of a string."
---

# ADR-0006: Entropy thresholds are charset-aware

**Status:** Accepted · Amendment 1 (2026-11-12): length-aware default
**Date:** 2026-11-12
**Source:** Planning session

## Context
One global threshold such as `minEntropy: 4.3` is wrong by design. The alphabet size limits Shannon entropy per character to log2 of that size: hex ≤ 4.0, base64 ≤ 6.0. A threshold of 4.3 never flags a hex secret. Length also limits entropy to log2(len). A 16-character string reaches at most 4.0.

## Decision
- Classify each candidate secret as `hex`, `base64`, `base64url` or `alnum`. Calculate the Shannon entropy on the `secretGroup` capture only.
- A rule's `entropy` block sets `min`, `max`, `minLen` and `charset`. The engine has defaults for each charset for generic rules. Initial values: hex ≥ 3.0 at len ≥ 32, base64 ≥ 4.5 at len ≥ 20. The `testdata/` corpus tunes them.
- Safe-format exclusions run before the entropy check: UUIDs, hashes in checksum/integrity contexts, file paths, URLs without credentials.

## Consequences
- whac finds hex API keys and rejects natural-language strings.
- Thresholds are data. A threshold change requires an update to the corpus golden files.

## Amendment 1 (2026-11-12): the default threshold scales with length
The initial values above are not reachable for short secrets. A 20-character string reaches at most log2(20) ≈ 4.32, so "base64 ≥ 4.5 at len ≥ 20" never passes.

- When a rule omits `entropy.min`, the engine uses `0.75 × log2(min(n, k))`. `n` is the secret length. `k` is the alphabet size of the charset.
- Alphabet sizes: hex 16, alnum 62, base64 and base64url 64, other text 95.
- `entropy.charset` fixes `k`. Without it, the engine classifies each secret as the smallest charset that contains it.
- Real tokens in the builtin samples score 0.85–0.97 of the maximum. Placeholders such as `xxxx…` score below 0.2.
- Milestone 1 implements Shannon entropy and charset classification in `internal/engine/entropy.go`. The `testdata/` corpus tunes the 0.75 fraction in Milestone 2.
