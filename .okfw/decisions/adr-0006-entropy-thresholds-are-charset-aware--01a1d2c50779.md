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
description: "Shannon entropy is compared against per-charset thresholds and minimum lengths, because the maximum entropy of a string is bounded by its alphabet."
---

# ADR-0006: Entropy thresholds are charset-aware

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
A single global threshold such as `minEntropy: 4.3` is wrong by construction. Shannon entropy per character is capped at log2 of the alphabet size: hex ≤ 4.0, base64 ≤ 6.0. A flat 4.3 can never flag a hex secret. Short strings are also capped at log2(len), so a 16-character string tops out at 4.0.

## Decision
- Classify each candidate secret as `hex`, `base64`, `base64url` or `alnum`, and compute Shannon entropy on the `secretGroup` capture only.
- A rule's `entropy` block sets `min`/`max`/`minLen`/`charset`. The engine has per-charset defaults for generic rules. Starting values: hex ≥ 3.0 at len ≥ 32, base64 ≥ 4.5 at len ≥ 20. They are tuned against the `testdata/` corpus.
- Safe-format exclusions run before entropy: UUIDs, hashes in checksum/integrity contexts, file paths, URLs without credentials.

## Consequences
- Hex API keys get caught, and natural-language strings are rejected.
- Thresholds are data. Changing one means updating the corpus goldens.
