---
title: Output Formats
description: The text and JSON report formats of whac scan.
---

Set the format with `--format text` (the default) or `--format json`. whac
redacts each secret. It keeps the first four and the last four characters of
a single-line secret with 16 or more characters. It prints `****` for other
secrets. Add `--no-redact` to print each secret in full.

## Text

```text
CRITICAL  GITHUB_TOKEN  app.env:4:10
          GitHub personal access, OAuth, app or refresh token
          secret: ghp_****oPqr
          fingerprint: whac1:ceaecd45ce7e62d41df233cda112cf1e4aed182b36f61607fec15d2826878fec

1 finding(s): 1 CRITICAL. Scanned 1 file(s), 187 B.
```

Line and column numbers start at 1. The column counts characters, not bytes.
whac colors the severity when the output is a terminal and `NO_COLOR` is not
set.

## JSON

```json
{
  "apiVersion": "whac.infrashift.io/v1alpha1",
  "kind": "ScanReport",
  "summary": {
    "findings": 1,
    "bySeverity": { "CRITICAL": 1 },
    "files": 1,
    "bytes": 187,
    "skippedBinary": 0,
    "skippedLarge": 0,
    "skippedExcluded": 0,
    "skippedUnreadable": 0,
    "suppressed": 0
  },
  "findings": [
    {
      "ruleId": "GITHUB_TOKEN",
      "description": "GitHub personal access, OAuth, app or refresh token",
      "severity": "CRITICAL",
      "path": "app.env",
      "line": 4,
      "column": 10,
      "endLine": 4,
      "fingerprint": "whac1:ceaecd45ce7e62d41df233cda112cf1e4aed182b36f61607fec15d2826878fec",
      "secret": "ghp_****oPqr",
      "redacted": true
    }
  ]
}
```

A finding from `--git` also has `commit`, `author` and `date`. The text
format prints them on a `commit:` line. The summary then counts commits and
file changes.

An assignment finding also has `keyPath`, for example
`"database.credentials.password"`. The text format prints it as `key:`.

`findings` is an empty array when whac finds nothing. `endLine` differs from
`line` for a secret that spans lines, such as a PEM key.

SARIF output arrives in Milestone 5.
