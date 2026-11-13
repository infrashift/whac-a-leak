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
description: "The #SecretRule CUE schema, rule kinds, rule tests and how user .whac/ layers unify with the embedded defaults."
---

# Rule schema

**Status:** Draft (implemented in Milestone 1 as `internal/rules/schema/schema.cue`)

## Shape
```cue
package whac

#SecretRule: {
	id:          =~"^[A-Z0-9_]+$"
	description: string
	kind:        *"pattern" | "assignment"
	severity:    *"HIGH" | "LOW" | "MEDIUM" | "CRITICAL"
	enabled:     *true | bool

	keywords:    [...string]        // Aho-Corasick prefilter; non-empty for kind=pattern
	regex?:      string             // RE2 (Go regexp) syntax
	secretGroup: *1 | int & >=0     // capture group holding the secret
	keyContext?: string             // regex on the key name (assignment rules)
	paths?:      [...string]        // sensitive key-path globs, e.g. "**.credentials.password"

	entropy?: {
		min?:     number & >=0 & <=8
		max?:     number & >=0 & <=8
		minLen:   *16 | int & >0
		charset?: "hex" | "base64" | "base64url" | "alnum"
	}
	stopwords:    [...string]
	allowRegexes: [...string]

	tests: {
		mustMatch:  [string, ...string]   // at least one positive sample
		mustIgnore: [...string]
	}
	if kind == "pattern" {
		keywords: [string, ...string]
		regex:    string
		tests: mustMatch: [...=~regex]    // cheap author-side check under `cue vet`
	}
}

rules:     [ID=string]: #SecretRule & {id: ID}
config:    #Config     // maxFileSize, exclude globs, sensitivePaths, ...
allowlist: #Allowlist  // fingerprints: [...{fp, reason}], paths: [...glob], regexes: [...]
```

## Agent-context configuration (Milestone 6)
```cue
#CredentialPath: {
	glob:        string                         // relative to $HOME or a --root, e.g. ".aws/sso/cache/*.json"
	os?:         "linux" | "darwin" | "windows" // omit for all
	description: string
	maxSize:     *(1 << 20) | int & >0          // shell histories can be large
}
#ExposureCheck: {
	severity: *"HIGH" | "LOW" | "MEDIUM" | "CRITICAL"
	enabled:  *true | bool
}
config: credentialPaths: [ID=string]: #CredentialPath  // keyed by ID so .whac/ layers can extend or disable
config: exposure: {
	sshAgent:   #ExposureCheck
	dockerSock: #ExposureCheck
	imds:       #ExposureCheck & {enabled: *false | bool}  // also requires --probe-imds
}
```
Environment variables become assignments with key path `env.<NAME>`, so ordinary `assignment` rules (`keyContext: "(?i)_(token|secret|password)$"`) apply to them. See [ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md).

## Rule kinds
- **pattern**: keyword-gated regex over raw content. Example: a provider token with a known prefix.
- **assignment**: evaluated against `(keyPath, value)` pairs from structured files, or from the generic `key = value` line extractor. Matches on `keyContext` and/or `paths`, then checks the value (entropy, placeholders, safe formats). Covers unprefixed secrets and the "zero-escape" entropy trap.

## Layering
The embedded schema and defaults load together with the user's `.whac/*.cue` files (`package whac`) as one CUE instance. CUE unification is monotonic, so a layer can only **add** rules or **tighten** them. Loosening is done through default markers:
- `rules: GENERIC_API_KEY: enabled: false` disables a rule.
- `rules: AWS_ACCESS_KEY: severity: "CRITICAL"` overrides a default.
- To loosen a regex, disable the rule and add a replacement under a new ID.

## Rule tests
Every rule ships `tests.mustMatch` (non-empty) and `tests.mustIgnore`. `cue vet` checks positives against the raw regex. At load time, and in `whac rules verify`, the Go side runs every sample through the **real engine pipeline**. A failing rule aborts loading with an error that names the rule and the sample index. Sample values must be obviously fake.
