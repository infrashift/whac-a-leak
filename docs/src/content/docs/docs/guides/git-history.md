---
title: Scanning Git History
description: Find secrets that commits added, including secrets that later commits removed.
---

A secret that a commit removes stays in the git history. `whac scan --git`
scans the lines that each commit adds.

## Scan the history

1. Run `whac scan --git` in the repository.
2. Read the `commit:` line of each finding. It shows the short hash, the
   author and the date.

* **Expected result:** whac reports each secret once per path, at the
  earliest commit that added it.

To limit the commits, give a revision range with an equals sign:

```sh
whac scan --git=main..HEAD          # commits on this branch only
whac scan --git=v1.2.0..v1.3.0      # commits between two tags
```

Without `=`, whac reads the next word as the path to scan.

## How it works

* whac runs `git log -p -U0 --reverse` and reads the output as a stream. It
  keeps one file of one commit in memory at a time.
* whac scans only added lines. It does not scan deleted lines or binary
  files.
* Rename detection is on. A commit that only renames a file adds no lines.
* whac scans each hunk as unstructured text, because a hunk is part of a
  file. Assignment rules without keywords, such as `SENSITIVE_PATH`, do not
  run in history mode.
* Merge commits add no lines of their own.

## After you find a secret

1. Rotate the secret. Everyone with a clone can read a secret in the
   history.
2. Decide if you must rewrite the history. A rewrite does not undo the
   exposure.
3. For a false positive, allowlist the fingerprint. The fingerprint does not
   depend on the commit, so one entry covers every commit.
