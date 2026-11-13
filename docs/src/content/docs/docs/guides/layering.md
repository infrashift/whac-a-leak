---
title: Layering Org Policy
description: Add, tighten, disable and re-grade rules from the config dir without a fork of the builtin rules.
---

whac unifies three sources into one CUE package, `whac`:

1. The builtin schema.
2. The builtin rules.
3. Every `*.cue` file in the config dir. Each file is a layer.

The config dir is `.whac/` by default. Set another directory with
`--config-dir`. whac ignores a missing `.whac/`. whac exits with code `2` if
you set `--config-dir` to a directory that does not exist.

## What a layer can do

| Task | Layer content |
|---|---|
| Add a rule | `rules: MY_RULE: {...}` |
| Disable a rule | `overrides: JWT: enabled: false` |
| Change the severity of a rule | `overrides: GENERIC_API_KEY: severity: "LOW"` |
| Add stopwords to a rule without stopwords | `rules: GENERIC_API_KEY: stopwords: ["dummy"]` |
| Exclude paths | `config: exclude: "vendor/**": "third-party code"` |
| Add a sensitive key path | `config: sensitivePaths: "**.vault.unseal_key": "Vault unseal keys"` |
| Allowlist a false positive | `allowlist: fingerprints: "whac1:<hex>": "test fixture"` |

## What a layer cannot do

CUE unification only adds constraints. It never removes them.

* A layer cannot change the regex or keywords of a rule. Disable the rule with
  `overrides`, and add a new rule with a new ID.
* A layer cannot change a builtin severity through `rules`. The builtin value
  is concrete, so CUE reports a conflict. Use `overrides`.
* A layer cannot append to a list in a rule. A list unifies element by element.

## Conflicts

Two layers that set different values cause a conflict. whac prints both
positions and exits with code `2`. For example:

```text
overrides.JWT.severity: conflicting values "HIGH" and "LOW":
    .whac/a.cue:2:27
    .whac/b.cue:2:27
```

## Typing errors

The schema closes each rule. A misspelt field such as `severty` stops the
load. whac also rejects unknown top-level fields. The valid top-level fields
are `rules`, `overrides`, `config` and `allowlist`.

## Allowlist entries

Each allowlist entry maps a key to a reason. The reason must not be empty.

* `fingerprints`: a `whac1:` fingerprint. The key must have 64 hex digits.
* `paths`: a doublestar glob.
* `regexes`: a regex over the secret.

Allowlist entries are for false positives only. Rotate a real secret. Do not
allowlist it. See [Allowlists & Fingerprints](../allowlists/).

## Inspect the result

Run `whac rules list`. The table shows each rule after all layers apply. The
last line lists the layer files.
