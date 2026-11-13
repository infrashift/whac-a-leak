---
id: "01a1d2c5-074c-710d-96f4-d1b6fbdcc6de"
title: "ADR-0001: Adopt an OKF wiki (okfw) for project knowledge"
type: "ADR"
tags:
  - "adr"
  - "process"
  - "okf"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "An okfw-managed OKF wiki in .okfw/ holds project knowledge: architecture, specs, ADRs, issues and daily notes."
---

# ADR-0001: Adopt an OKF wiki (okfw) for project knowledge

**Status:** Accepted
**Date:** 2026-11-12
**Source:** `CLAUDE.md`, `PLAN.md`

## Context
Whac-a-Leak starts as an idea and a logo. Several design decisions differ from other secret scanners: CUE only at load time, monotonic layering, charset-aware entropy, fingerprints without location. Chat history and commit messages hide the reasons for these decisions from agents and contributors. Sibling infrashift repos (`vizzini-pr-reviewer`, `tenet-cac`, `cuix`, `garmr-policy-agent`) already use okfw.

## Decision
- An OKF wiki lives at `.okfw/` in the repo root. The `okfw` CLI manages it. The repository includes it.
- `okfw install --platform claude` adds the Claude Code skill and the managed `CLAUDE.md` block. The project `.mcp.json` registers `okfw mcp`.
- `PLAN.md` stays the milestone checklist. Each milestone has a mirror issue in `issues/`. The two link to each other.
- Conventions: [Wiki conventions](/wiki-conventions--01a1d2c5072f.md). Loop: [Development workflow](/development-workflow--01a1d2c50728.md).

## Consequences
- `okfw search` finds decisions and open work. Links connect them into a graph.
- `PLAN.md` and the issues can drift apart. The workflow guide requires an update to both.
- The repository includes the wiki, so the wiki must not contain secrets. See the secrets hygiene section of the conventions.

## Alternatives considered
- **Plain `docs/` markdown.** Rejected: no frontmatter, no typed search, no generated indexes. The public site in `docs/` serves users. The wiki serves contributors.
- **An external issue tracker only.** Rejected: it separates the reasons from the code, and offline agents cannot reach it.
