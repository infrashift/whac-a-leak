---
id: "01a1d2c5-0736-7cb0-8108-544663871f42"
title: "Rule schema"
type: "Spec"
tags:
  - "spec"
  - "rules"
  - "cue"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "The #SecretRule CUE schema, the rule kinds, rule samples, and how layers in the config dir unify with the builtin rules."
---

# Rule schema

**Status:** Draft. Milestone 1 implements it in `internal/rules/builtin/schema.cue`.

## Shape
The source is `internal/rules/builtin/schema.cue`. The docs site shows the file in full on the *Rule Schema* page. [ADR-0012](/decisions/adr-0012-load-layers-as-one-in-memory-cue-build-instance--01a1d2e388a9.md) describes the loader.

| Top-level field | Type | Purpose |
|---|---|---|
| `rules` | `[ID=string]: #SecretRule & {id: ID}` | Rules. The field name is the rule ID. |
| `overrides` | `[ID=string]: #Override` | Re-grade (`severity`) or disable (`enabled`) a rule by ID. |
| `config` | `#Config` | `maxFileSize` (default 2 MiB), `exclude: [glob]: reason`, `sensitivePaths: [key-path glob]: reason`. |
| `allowlist` | `#Allowlist` | `fingerprints`, `paths` and `regexes`, each a map from key to reason. |

`#SecretRule` fields and defaults:

| Field | Default |
|---|---|
| `id`, `description` | required |
| `kind` | `pattern` |
| `severity` | `HIGH` |
| `enabled` | `true` |
| `keywords`, `regex` | required for pattern rules |
| `secretGroup` | `1` |
| `multiline` | `false`: the regex runs on the keyword's line only ([ADR-0013](/decisions/adr-0013-regexes-run-on-keyword-line-windows--01a1d2f04316.md)) |
| `keyContext`, `paths` | assignment rules (Milestone 3) |
| `entropy` | optional: `min`, `max`, `minLen` (default 16), `charset` |
| `stopwords`, `allowRegexes` | empty |
| `tests.mustMatch`, `tests.mustIgnore` | one or more, empty. Each sample is a string or `{file, content}`; the file name selects the extractor. |

Constraints in CUE:
- A pattern rule needs a `regex` and at least one keyword.
- Each `mustMatch` sample must match the `regex`.
- `entropy.max` must be at least `entropy.min`.
- A fingerprint key must match `^whac1:[0-9a-f]{64}$`. Every allowlist reason must be non-empty.

The loader also rejects unknown top-level fields and overrides for unknown rule IDs.

## Agent-context configuration (Milestone 6)
```cue
#CredentialPath: {
	glob:        string & !=""   // relative to $HOME or a --root, e.g. ".aws/sso/cache/*.json"
	description: string & !=""
	os?:         "linux" | "darwin" | "windows"
	format?:     "yaml" | "json" | "dotenv" | "properties" | "ini" | "hcl"
	maxSize:     *1048576 | (int & >0)
	enabled:     *true | bool
}
config: credentialPaths: [ID=string]: #CredentialPath
config: exposure: {
	sshAgent:      {severity: *"HIGH" | #Severity,     enabled: *true | bool}
	dockerSocket:  {severity: *"CRITICAL" | #Severity, enabled: *true | bool}
	cloudMetadata: {severity: *"HIGH" | #Severity,     enabled: *true | bool} // also needs --probe-imds
}
```
The environment becomes a dotenv document named `env`. Findings in it get the key path `env.<NAME>`. See [ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md).

## Rule kinds
- **pattern**: a regex runs over raw content after a keyword match. Example: a provider token with a known prefix.
- **assignment**: the rule runs on `(keyPath, value)` pairs. The structural extractors or the generic `key = value` line extractor produce these pairs. The rule matches on `keyContext`, `paths`, or both. Then it checks the value: entropy, placeholders, safe formats. This kind covers secrets without a prefix and the "zero-escape" entropy trap.

## Layering
The loader unifies the embedded schema, the builtin rules and the layers in `.whac/*.cue` (`package whac`) into one CUE instance. CUE unification is monotonic. A layer can only **add** rules or **tighten** rules. The `overrides` map makes two more changes possible ([ADR-0004 amendment 1](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md)):
- `overrides: GENERIC_API_KEY: enabled: false` disables a rule.
- `overrides: AWS_ACCESS_KEY_ID: severity: "LOW"` changes the severity of a rule.
- To loosen a regex, disable the rule and add a replacement rule with a new ID.
- A layer that sets `rules: X: severity` on a builtin rule conflicts with the concrete builtin value.

## Rule samples
Each rule has `tests.mustMatch` (one or more samples) and `tests.mustIgnore`. `cue vet` checks the positive samples against the raw regex. The Go engine runs every sample through the **real engine pipeline** at load time and in `whac rules verify`. A failed sample stops the load. The error names the rule and the sample index. Use sample values that are clearly fake.
