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
description: "A map keyed by rule ID holds the rules, with defaulted enabled and severity fields, so layers add, tighten, disable or re-grade rules without a fork."
---

# ADR-0004: Rules are keyed by ID and layered with CUE default markers

**Status:** Accepted · Amendment 1 (2026-11-12): `overrides` map
**Date:** 2026-11-12
**Source:** Planning session

## Context
Teams adapt the builtin rules. They add internal token formats, disable noisy rules and raise severities. CUE unification is monotonic. A layer can add or narrow constraints. A layer cannot remove a rule or widen a regex. A rule list also makes "the same rule" ambiguous across layers.

## Decision
- `rules: [ID=string]: #SecretRule & {id: ID}`: a map, so every layer addresses a rule by ID.
- CUE defaults mark the fields that a layer can override: `enabled: *true | bool`, `severity: *"HIGH" | ...`, `secretGroup: *1 | int`, `entropy.minLen: *16 | int`. A concrete value in a layer replaces the default.
- To loosen a regex or a keyword set, disable the rule and add a replacement rule with a new ID.
- `whac rules export-schema` writes the schema into `.whac/` for editors and `cue vet`.

## Consequences
- Upstream rule updates cause no merge conflicts in organization layers.
- Two layers that set different concrete values (for example, two severities) fail the load. The CUE conflict error names both sources. This behavior is intentional.

## Alternatives considered
- **Override by position or by file replacement.** Breaks without warning when upstream changes. Rejected.
- **A "patch" mechanism outside CUE.** Rejected: unification with defaults covers the need.

See [Rule schema](/specs/rule-schema--01a1d2c50736.md).

## Amendment 1 (2026-11-12): re-grade and disable through `overrides`
The Milestone 1 spike ([ISSUE-001](/issues/issue-001-rule-schema-cue-loader-and-rules-verify--01a1d2c507b1.md)) disproved one claim above. A builtin rule sets a concrete severity, for example `severity: "CRITICAL"`. A layer value such as `rules: AWS_ACCESS_KEY_ID: severity: "LOW"` then conflicts. The default marker applies only to rules that omit the field.

- The schema adds `overrides: [ID=string]: #Override`. `#Override` has the optional fields `severity` and `enabled`.
- The loader applies overrides after it decodes the rules. An override for an unknown rule ID stops the load.
- Two layers that set different override values conflict. The CUE error names both files.
- Allowlist entries and `config.exclude` are maps (key → reason), not lists. CUE unifies a list element by element, so two layers that each add a list entry conflict. Two layers can add map keys without a conflict.
- Lists inside a rule (`keywords`, `stopwords`, `allowRegexes`) stay atomic. A layer can set a list only if the rule does not set it.
