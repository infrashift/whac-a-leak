---
title: Roadmap
description: The Whac-a-Leak milestones. The project OKF wiki tracks each milestone as an issue.
---

`PLAN.md` holds a checklist for each milestone. The project wiki
(`.okfw/issues/`) holds a mirror issue for each milestone. Each milestone lands
as one tested commit, together with the docs pages for its features.

| Milestone | Scope | Status |
|---|---|---|
| **0. Bootstrap** | Repository, Makefile, OKF wiki and MCP, ADRs, this site | Done |
| **1. CUE rules** | Rule schema, builtin rules, `.whac/` layers, `whac rules list` / `verify` | Done |
| **2. Scan engine** | Aho-Corasick prefilter, regex and charset-aware entropy, `whac1:` fingerprints, allowlists, filesystem source, text/JSON output | Done |
| **3. Structured files** | YAML/JSON/.env/properties/HCL-lite extractors, `assignment` rules, sensitive paths, placeholder filter, entropy trap | Next |
| **4. Git** | `--staged` pre-commit mode, `--git` history with commit attribution, `.pre-commit-hooks.yaml` | Planned |
| **5. Release** | SARIF, about 30 builtin rules, CI with self-scan, generated landing-page demo, benchmarks | Planned |
| **6. Agent-context preflight** | `whac scan --agent-context`: environment variables, well-known credential files (SSO caches, CLI configs, shell history), credential-socket exposure, opt-in IMDS probe | Planned |

## Outside the v1 scope

- **Live verification** of found credentials against provider APIs. whac
  stays offline ([ADR-0008](../decisions/#adr-0008-no-live-secret-verification-in-v1)).
- Recursive decoding of base64-encoded blobs.
- **Process memory scans.** They need ptrace or root. Security tools cannot
  tell them apart from credential-dumping malware. Agent-context mode audits
  what an agent can reach instead
  ([ADR-0011](../decisions/#adr-0011-agent-context-preflight)).
