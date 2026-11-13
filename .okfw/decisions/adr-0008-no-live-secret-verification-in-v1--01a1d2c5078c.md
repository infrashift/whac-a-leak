---
id: "01a1d2c5-078c-70e5-91fd-91c56f1fbc4a"
title: "ADR-0008: No live secret verification in v1"
type: "ADR"
tags:
  - "adr"
  - "scope"
  - "security"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "whac never makes network calls to check whether a found credential is live; detection is fully offline."
---

# ADR-0008: No live secret verification in v1

**Status:** Accepted · Amended by [ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md) (opt-in IMDS reachability probe)
**Date:** 2026-11-12
**Source:** Planning session

## Context
Some tools (e.g. TruffleHog) call provider APIs to check whether a found credential works. That needs network access, per-provider code, rate limiting, and sending the secret to a third party.

## Decision
whac is fully offline and never transmits a candidate secret. Live verification and recursive decoding of base64 blobs are out of scope for v1.

## Consequences
- Runs in air-gapped CI and pre-commit without surprises.
- Precision has to come from rule quality: multi-signal rules, charset-aware entropy and placeholder filters.

## Amendment (2026-11-12)
[ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md) adds one opt-in network call: `--probe-imds` checks whether the cloud metadata endpoint is reachable from an agent's context. It's off by default, never requests credential paths, and never transmits a secret. Everything else stays offline.
