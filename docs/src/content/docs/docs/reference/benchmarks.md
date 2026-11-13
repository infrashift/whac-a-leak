---
title: Benchmarks
description: Measured speed of rule loading, rule verification and scanning.
---

Run `make bench` to repeat the measurements. The values below come from an
Intel Core i7-9700K (8 cores) with Go 1.27 and the 30 builtin rules.

| Measurement | Result |
|---|---|
| Load and compile the builtin rules (`rules.Load`) | 8.9 ms |
| Verify the 84 builtin samples (`engine.Verify`) | 2.7 ms |
| `whac rules list`, whole process | 10 ms, 22 MB RSS |
| Scan keyword-sparse source code, one core | 183 MB/s |
| `whac scan --git`, 97 commits | 0.15 s |
| `whac scan --git`, 1,033 commits, 13.7 MiB of added lines | 1.0 s, 132 MB RSS |

## Why the scan is fast

* CUE runs only at load time. The scan is plain Go.
* One Aho-Corasick pass finds the keywords of all rules.
* A rule's regex runs only on the lines that hold one of its keywords. The
  first version ran each regex over the whole file and reached 24 MB/s.
* whac scans files in parallel, one worker per CPU.
