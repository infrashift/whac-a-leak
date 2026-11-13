---
title: CI & SARIF
description: Run whac in GitHub Actions, Azure DevOps, GitLab CI/CD or Forgejo Actions, fail the build on a finding, and keep the SARIF report.
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

## Azure DevOps

1. Add this file to the repository as `azure-pipelines.yml`:

   ```yaml
   trigger:
     - main
   pr:
     - main

   pool:
     vmImage: ubuntu-latest

   variables:
     WHAC_VERSION: 0.2.0

   steps:
     - checkout: self
       fetchDepth: 0
       persistCredentials: false
     - script: |
         set -eu
         base=https://github.com/infrashift/whac-a-leak/releases/download/v$(WHAC_VERSION)
         archive=whac_$(WHAC_VERSION)_linux_amd64.tar.gz
         cd "$(Agent.TempDirectory)"
         curl -fsSLO "$base/$archive"
         curl -fsSLO "$base/SHA256SUMS"
         grep " $archive\$" SHA256SUMS | sha256sum --check -
         tar -xzf "$archive"
         echo "##vso[task.prependpath]$(Agent.TempDirectory)/whac_$(WHAC_VERSION)_linux_amd64"
       displayName: Install whac
     - script: whac scan --git .
       displayName: Scan the history
     - script: |
         mkdir -p "$(Build.ArtifactStagingDirectory)/CodeAnalysisLogs"
         whac scan --output sarif="$(Build.ArtifactStagingDirectory)/CodeAnalysisLogs/whac.sarif" . || test $? -eq 1
       displayName: Write the SARIF report
       condition: always()
     - publish: $(Build.ArtifactStagingDirectory)/CodeAnalysisLogs
       artifact: CodeAnalysisLogs
       displayName: Publish the SARIF report
       condition: always()
   ```

2. Create a pipeline from the file in *Pipelines → New pipeline*.

* **Expected result:** a finding fails the run. The run keeps `whac.sarif`
  in the artifact `CodeAnalysisLogs`.

`fetchDepth: 0` gives whac the full history. New Azure pipelines use a
shallow fetch by default. The *SARIF SAST Scans Tab* extension reads the
`CodeAnalysisLogs` artifact and shows each finding in a *Scans* tab.

## GitLab CI/CD

1. Add this job to `.gitlab-ci.yml`:

   ```yaml
   whac:
     stage: test
     image: golang:1.25
     variables:
       GIT_DEPTH: "0"
     script:
       - go install github.com/infrashift/whac-a-leak/cmd/whac@v0.2.0
       - whac scan --git --output sarif=whac.sarif .
     artifacts:
       when: always
       paths:
         - whac.sarif
       expire_in: 30 days
   ```

2. Push the file.

* **Expected result:** a finding fails the job. The job keeps `whac.sarif`
  as a job artifact.

`GIT_DEPTH: "0"` gives whac the full history. The `golang` image contains
`git`, and whac needs it for `--git`. The GitLab security dashboard does not
read SARIF, so the job keeps the report as a plain artifact.

## Forgejo Actions

1. Add this workflow as `.forgejo/workflows/secrets.yml`:

   ```yaml
   on:
     push:
       branches: [main]
     pull_request:

   jobs:
     secrets:
       runs-on: docker
       env:
         WHAC_VERSION: 0.2.0
       steps:
         - uses: actions/checkout@v4
           with:
             fetch-depth: 0
             persist-credentials: false
         - name: Install whac
           run: |
             set -eu
             base=https://github.com/infrashift/whac-a-leak/releases/download/v$WHAC_VERSION
             archive=whac_${WHAC_VERSION}_linux_amd64.tar.gz
             cd "$RUNNER_TEMP"
             curl -fsSLO "$base/$archive"
             curl -fsSLO "$base/SHA256SUMS"
             grep " $archive\$" SHA256SUMS | sha256sum --check -
             tar -xzf "$archive"
             echo "$RUNNER_TEMP/whac_${WHAC_VERSION}_linux_amd64" >> "$GITHUB_PATH"
         - name: Scan the history
           run: whac scan --git .
         - name: Write the SARIF report
           if: always()
           run: whac scan --output sarif=whac.sarif . || test $? -eq 1
         - uses: https://code.forgejo.org/forgejo/upload-artifact@v4
           if: always()
           with:
             name: whac-sarif
             path: whac.sarif
   ```

2. Push the workflow.

* **Expected result:** a finding fails the job. The run keeps `whac.sarif`
  in the artifact `whac-sarif`.

The runner label `docker` and its default image come from the runner
configuration. The job image must contain `node`, `git` and `curl`. Forgejo
has no code scanning view, so the job keeps the report as an artifact. The
Forgejo fork of `upload-artifact` works with the Forgejo artifact service.

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
