---
title: Allowlists & Fingerprints
description: Suppress false positives by fingerprint, path, regex or inline marker.
---

Use an allowlist only for false positives, such as test fixtures and
documented examples. Rotate a real secret. Do not allowlist it.

## Fingerprints

Each finding has a fingerprint:

```text
whac1:<hex(SHA-256(rule_id || 0x00 || secret))>
```

* The fingerprint contains no path and no line number. A rebase or a file
  move does not change it.
* whac trims spaces and one pair of matching quotes from the secret first.
* whac does not change the case of the secret.

A committed fingerprint of a short real secret is open to an offline guess.
This is one more reason to rotate real secrets.

## Allowlist a finding

1. Run `whac scan`.
2. Copy the `fingerprint` value of the false positive.
3. Run `whac allow add <fingerprint> --reason "test fixture"`.
4. Run `whac scan` again.

* **Expected result:** whac does not report the finding. The summary counts
  it as suppressed.

`whac allow add` writes the entry to `.whac/allowlist.cue`. It then loads all
layers again. If the new entry breaks a layer, whac restores the file and
exits with code `2`.

## Other suppressions

| Method | Example | Scope |
|---|---|---|
| Path | `allowlist: paths: "testdata/**": "fixtures"` | Every finding in matching files. |
| Regex | `allowlist: regexes: "^tok_EXAMPLE": "docs"` | Every secret that matches the regex. |
| Inline | `key = "…" # whac:ignore` | The findings on that line. |
| Exclude | `config: exclude: "vendor/**": "third-party code"` | whac does not read the files. |

Path and exclude globs use doublestar syntax. They match the path relative to
the scan root.
