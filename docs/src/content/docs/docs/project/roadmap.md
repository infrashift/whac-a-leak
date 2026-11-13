---
title: Roadmap
description: Whac-a-Leak's milestones, tracked as issues in the project's OKF wiki.
---

Each milestone has a checklist in `PLAN.md` and a mirror issue in the project
wiki (`.okfw/issues/`). Each one lands as its own tested commit, together with
the docs pages for the features it delivers.

| Milestone | Scope | Status |
|---|---|---|
| **0. Bootstrap** | Repository, Makefile, OKF wiki + MCP, ADRs, this site | Done |
| **1. CUE rules** | Rule schema, embedded defaults, `.whac/` layering, `whac rules list` / `verify` | Next |
| **2. Scan engine** | Aho-Corasick prefilter, regex + charset-aware entropy, `whac1:` fingerprints, allowlists, filesystem source, text/JSON output | Planned |
| **3. Structure-aware** | YAML/JSON/.env/properties/HCL-lite extractors, `assignment` rules, sensitive paths, placeholder filter, entropy trap | Planned |
| **4. Git** | `--staged` pre-commit mode, `--git` history with commit attribution, `.pre-commit-hooks.yaml` | Planned |
| **5. Release** | SARIF, ~30 default rules, CI with self-scan, generated landing-page demo, benchmarks | Planned |
| **6. Agent-context preflight** | `whac scan --agent-context`: environment variables, well-known credential files (SSO caches, CLI configs, shell history), credential-socket exposure, opt-in IMDS probe | Planned |

## Not planned for v1

- **Live verification** of found credentials against provider APIs: whac stays
  offline ([ADR-0008](../decisions/#adr-0008-no-live-secret-verification-in-v1)).
- Recursive decoding of base64-encoded blobs.
- **Scanning process memory**: it needs ptrace or root and is indistinguishable
  from credential-dumping malware. Agent-context mode audits what an agent can
  actually reach instead ([ADR-0011](../decisions/#adr-0011-agent-context-preflight)).
