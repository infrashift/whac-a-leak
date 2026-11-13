---
title: Pre-commit Hook
description: Stop a commit that adds a secret, with whac scan --staged.
---

`whac scan --staged` scans the staged version of each changed file. It
reports only findings on the lines that the commit adds. A secret that is
already in the repository does not block the commit.

## Use the pre-commit framework

1. Add this entry to `.pre-commit-config.yaml`:

   ```yaml
   repos:
     - repo: https://github.com/infrashift/whac-a-leak
       rev: v0.1.0
       hooks:
         - id: whac
   ```

2. Run `pre-commit install`.
3. Stage a file that contains a fake token, then run `git commit`.

* **Expected result:** the hook prints the finding, and git stops the commit.

The hook builds whac with Go 1.25 or newer. Update `rev` with
`pre-commit autoupdate` when a new release exists.

## Use a plain git hook

1. Build whac with `make build`, and put `bin/whac` on your `PATH`.
2. Create the file `.git/hooks/pre-commit` with this content:

   ```sh
   #!/bin/sh
   exec whac scan --staged
   ```

3. Run `chmod +x .git/hooks/pre-commit`.

* **Expected result:** `git commit` runs whac first. A finding stops the
  commit with exit code `1`.

## Handle a finding

* Remove the secret from the staged file, then rotate the secret.
* For a false positive, add `whac:ignore` to the line, or run
  `whac allow add <fingerprint> --reason "…"`.

Do not use `git commit --no-verify` to skip the hook for a real secret.
