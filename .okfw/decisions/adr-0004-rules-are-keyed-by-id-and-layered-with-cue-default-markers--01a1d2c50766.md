---
id: "01a1d2c5-0766-7eeb-965d-2f4d16fd577f"
title: "ADR-0004: Rules are keyed by ID and layered with CUE default markers"
type: "ADR"
tags:
  - "adr"
  - "cue"
  - "rules"
  - "layering"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Rules live in a map keyed by ID with defaulted enabled/severity fields, so org layers can add, tighten, disable or re-grade rules through unification without forking."
---

# ADR-0004: Rules are keyed by ID and layered with CUE default markers

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session

## Context
Teams need to adapt the embedded ruleset: add internal token formats, turn off noisy rules, raise severities. CUE unification is monotonic. A layer can add or narrow constraints but can never remove a rule or widen a regex. A list of rules would also make "the same rule" ambiguous across layers.

## Decision
- `rules: [ID=string]: #SecretRule & {id: ID}`: a map, so every layer addresses a rule by ID.
- Overridable fields use CUE defaults: `enabled: *true | bool`, `severity: *"HIGH" | ...`, `secretGroup: *1 | int`, `entropy.minLen: *16 | int`. A concrete value in a user layer replaces the default.
- Loosening a regex or keyword set means disabling the rule and adding a replacement under a new ID.
- `whac rules export-schema` writes the schema into `.whac/` so editors and `cue vet` work there.

## Consequences
- Upstream rules can be updated without merge conflicts in org layers.
- A layer that sets a conflicting concrete value (say, two layers set different severities) fails to load with a CUE conflict error that names both sources. This is intended.

## Alternatives considered
- **Overriding by position or replacing whole files.** Breaks silently when upstream changes. Rejected.
- **A non-CUE "patch" mechanism.** Rejected: unification plus defaults already covers it.

See [Rule schema](/specs/rule-schema--01a1d2c50736.md).
