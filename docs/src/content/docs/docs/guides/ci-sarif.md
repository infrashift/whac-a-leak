---
title: CI & SARIF
description: Run whac in CI, fail the build on a finding, and send findings to GitHub code scanning.
---

whac exits with code `1` when it reports a finding. A CI step that runs whac
therefore fails on a secret.

## GitHub Actions

1. Add this job to a workflow file in `.github/workflows/`:

   ```yaml
   jobs:
     secrets:
       runs-on: ubuntu-latest
       permissions:
         contents: read
         security-events: write
       steps:
         - uses: actions/checkout@v7
           with:
             fetch-depth: 0
             persist-credentials: false
         - uses: actions/setup-go@v7
           with:
             go-version: "1.25"
         - run: go install github.com/infrashift/whac-a-leak/cmd/whac@v0.2.0
         - name: Scan the history
           run: whac scan --git .
         - name: Scan the work tree for code scanning
           if: always()
           run: whac scan --output sarif=whac.sarif . || test $? -eq 1
         - uses: github/codeql-action/upload-sarif@v4
           if: always()
           with:
             sarif_file: whac.sarif
             category: whac
   ```

2. Push the workflow.

* **Expected result:** a finding fails the job. Code scanning shows each
  finding under *Security → Code scanning*.

`fetch-depth: 0` gives whac the full history. To skip the Go build, download
the release archive instead, and verify `SHA256SUMS` first.

## SARIF output

`whac scan --format sarif` writes SARIF 2.1.0 to stdout.
`whac scan --output sarif=whac.sarif` writes it to a file and prints the text
report in the job log.

* Each rule appears in `tool.driver.rules` with a `security-severity` value
  for GitHub: `CRITICAL` 9.5, `HIGH` 8.0, `MEDIUM` 5.5, `LOW` 3.0.
* The SARIF level is `error` for `CRITICAL` and `HIGH`, `warning` for
  `MEDIUM` and `note` for `LOW`.
* `partialFingerprints.whacFingerprint/v1` holds the `whac1:` fingerprint.
  Code scanning uses it to track a finding across commits.
* SARIF output always redacts secrets. `--no-redact` does not apply, because
  whac sends SARIF files to a code scanning service.
* A finding from `--git` has `commit`, `author` and `date` in `properties`.

A test validates the SARIF output against the official OASIS SARIF 2.1.0
JSON Schema.

## This repository

whac scans its own repository in CI. The file `.whac/repo.cue` allowlists
the fixture directories. The file `.whac/allowlist.cue` allowlists the fake
values in test files by fingerprint.
