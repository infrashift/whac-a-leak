---
title: Writing Rules
description: Write a whac rule in CUE, add samples, and verify the rule with the engine.
---

A rule is a CUE struct under `rules`. The rule ID is the field name. whac
loads the builtin rules and every layer in the config dir (`.whac/`).

## Write a rule

1. Create the file `.whac/acme.cue`.
2. Add this rule:

   ```cue
   package whac

   rules: ACME_API_TOKEN: {
   	description: "Acme API token"
   	severity:    "CRITICAL"
   	keywords: ["acme_"]
   	regex: #"\b(acme_[A-Za-z0-9]{32})\b"#
   	entropy: minLen: 37
   	tests: {
   		mustMatch: ["token = acme_EXAMPLE0123456789abcdefghijklmno"]
   		mustIgnore: ["token = acme_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"]
   	}
   }
   ```

3. Run `whac rules verify`.
4. Read the result. Each rule shows `PASS` or `FAIL`.

* **Expected result:** `ACME_API_TOKEN` shows `PASS`, and whac exits with code `0`.

## How the engine evaluates a pattern rule

The engine evaluates the stages in this order. A candidate secret that fails
one stage produces no finding.

1. **Keywords.** One keyword must occur in the content. The match ignores ASCII case.
2. **Window.** The regex runs on each line that holds a keyword. Set
   `multiline: true` for a secret that spans lines. The window then runs up to
   64 KiB from the start of the keyword's line.
3. **Regex.** The regex uses RE2 syntax (Go `regexp`).
4. **Secret group.** The capture group `secretGroup` holds the secret. The default is `1`. Use `0` for the whole match.
5. **Safe formats.** For a rule with an `entropy` block, the engine drops a UUID.
6. **Entropy.** The secret must meet the `entropy` constraints, if the rule has them.
7. **Stopwords.** The engine drops a secret that contains a stopword. The match ignores case.
8. **Allow regexes.** The engine drops a secret that matches an `allowRegexes` entry.

A keyword must occur on the first line of a multi-line secret. A `^` or `$`
anchor matches the start or end of the window, not of the file.

## Entropy

whac calculates Shannon entropy in bits per character. A string of `n`
characters from an alphabet of `k` symbols reaches at most `log2(min(n, k))`.

* Set `entropy.min` to use a fixed threshold.
* Omit `entropy.min` to use the default: `0.75 × log2(min(n, k))`.
* Set `entropy.charset` to fix `k`. Omit it, and whac classifies each secret.
* Set `entropy.minLen` to drop short secrets. The default is `16`.

| Charset | Alphabet size `k` |
|---|---|
| `hex` | 16 |
| `alnum` | 62 |
| `base64`, `base64url` | 64 |
| other text | 95 |

## Samples

Every rule has samples. `tests.mustMatch` must contain at least one sample.

* Each `mustMatch` sample must produce a finding.
* Each `mustIgnore` sample must produce no finding.

whac runs every sample through the engine each time it loads the rules. A
multi-line sample in a rule without `multiline: true` fails, and the error
tells you to set the field. A
failed sample stops the load, and whac exits with code `2`. The error names
the sample by position, for example `mustMatch[0]`. It does not print the
sample text.

Use values that cannot be real credentials. Add an `EXAMPLE` marker, or use a
fictional provider such as *Acme*.

## Check a rule without whac

`cue vet` checks that each `mustMatch` sample matches the regex.

1. Run `whac rules export-schema`. whac writes `.whac/schema.cue`.
2. Run `cue vet ./.whac`.

* **Expected result:** `cue vet` prints nothing and exits with code `0`.

The loader skips the exported `schema.cue`. Run `whac rules export-schema`
again after you upgrade whac.

## Rule kinds

`kind` is `"pattern"` (the default) or `"assignment"`. Assignment rules
evaluate key and value pairs from structured files. whac rejects assignment
rules until Milestone 3.
