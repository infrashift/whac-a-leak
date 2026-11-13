---
title: How Scanning Works
description: The stages of a whac scan, from the file list to the report.
---

whac runs the same stages for each file. CUE runs once, when whac loads the
rules. The scan itself is plain Go.

## Stages

1. **Load.** whac loads the builtin rules and each layer, then runs the
   samples of each rule. A failed sample stops the scan with exit code `2`.
2. **List files.** In a git work tree, whac runs
   `git ls-files --cached --others --exclude-standard`. `.gitignore` applies.
   Outside a git work tree, whac walks the directory and skips `.git`.
3. **Gate.** whac skips symlinks, files in `config.exclude`, files larger
   than `config.maxFileSize` (default 2 MiB), and binary files. A file with a
   NUL byte in its first 8000 bytes is binary.
4. **Prefilter.** One Aho-Corasick pass finds every keyword occurrence. The
   match ignores ASCII case. Only rules with a keyword hit continue.
5. **Windows.** For each keyword hit, the engine takes the whole line. A
   `multiline` rule takes up to 64 KiB from the start of that line.
6. **Match.** The regex of a pattern rule runs on each window. The engine then
   checks the secret group, safe formats, entropy, stopwords and allow regexes.
7. **Assignments.** For a structured file, whac extracts every key/value pair.
   For other text, whac extracts pairs on the keyword lines of each assignment
   rule. Assignment rules check each pair. See
   [Structured Files](../structured-files/).
8. **Deduplicate.** When two rules report the same span, whac keeps the higher
   severity.
9. **Suppress.** whac drops a finding when its line contains `whac:ignore`, or
   when an allowlist entry matches it.
10. **Report.** whac sorts the findings by path, line, column and rule ID. It
   writes text or JSON.

## Safe formats

A rule with an `entropy` block uses a heuristic. A UUID has high entropy, but
it is not a secret. whac drops a UUID for these rules. A rule without an
`entropy` block matches a precise token format and keeps every match.

## Performance

The prefilter keeps the regex work small. On keyword-sparse source code, the
builtin rules scan about 183 MB/s on one CPU core. See
[Benchmarks](../../reference/benchmarks/). whac scans files in
parallel, with one worker per CPU.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | No findings at or above `--severity-threshold`. |
| `1` | whac reported findings. |
| `2` | Usage error, invalid layer, failed sample, or I/O error. |
