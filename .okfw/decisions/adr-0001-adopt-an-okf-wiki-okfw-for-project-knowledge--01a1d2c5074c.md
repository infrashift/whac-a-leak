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
description: "Project knowledge (architecture, specs, ADRs, issues, daily notes) lives in an okfw-managed OKF wiki at .okfw/."
---

# ADR-0001: Adopt an OKF wiki (okfw) for project knowledge

**Status:** Accepted
**Date:** 2026-11-12
**Source:** `CLAUDE.md`, `PLAN.md`

## Context
Whac-a-Leak starts as an idea and a logo. Several of its design decisions go against how other secret scanners work: CUE only at load time, monotonic layering, charset-aware entropy, location-free fingerprints. If the reasons for them live only in chat history and commit messages, agents and contributors can't find them. Sibling infrashift repos (`vizzini-pr-reviewer`, `tenet-cac`, `cuix`, `garmr-policy-agent`) already use okfw.

## Decision
- An OKF wiki lives at `.okfw/` in the repo root, managed by the `okfw` CLI, and is committed with the code.
- `okfw install --platform claude` provides the Claude Code skill and the managed `CLAUDE.md` block. The project `.mcp.json` registers `okfw mcp`.
- `PLAN.md` remains the milestone checklist. Each milestone has a mirror issue in `issues/`, and the two link to each other.
- Conventions: [Wiki conventions](/wiki-conventions--01a1d2c5072f.md). Loop: [Development workflow](/development-workflow--01a1d2c50728.md).

## Consequences
- Decisions and open work can be searched (`okfw search`) and are linked into a graph.
- `PLAN.md` and the issues can drift apart. The workflow guide requires updating both.
- The wiki is committed, so it must never contain secrets (see the secrets hygiene section of the conventions).

## Alternatives considered
- **Plain `docs/` markdown.** Rejected: no frontmatter, no typed search, no generated indexes. The public site in `docs/` is for users, the wiki is for builders.
- **An external issue tracker only.** Rejected: it separates the reasoning from the code, and agents running offline can't reach it.
