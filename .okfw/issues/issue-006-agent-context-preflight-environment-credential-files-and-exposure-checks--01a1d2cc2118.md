---
id: "01a1d2cc-2118-7c86-be9b-d99bea2fa539"
title: "ISSUE-006: Agent-context preflight: environment, credential files and exposure checks"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-6"
  - "agent-context"
created: "2026-11-12T20:45:55-06:00"
timestamp: "2026-11-12T20:45:55-06:00"
description: "Milestone 6: env and home sources, CUE-declared credential paths, exposure findings for credential sockets and opt-in IMDS, and the whac scan --agent-context preflight gate."
---

# ISSUE-006: Agent-context preflight: environment, credential files and exposure checks

**Status/Milestone:** Open · Milestone 6 (needs ISSUE-003's extractors)

## Description
Implement [ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md): audit what an agent's process can reach, and fail a devcontainer or microVM preflight if any credential is reachable.

## Acceptance Criteria
- [ ] `internal/source` `env` source: the process environment → `env.<NAME>` assignments; values never appear in any output.
- [ ] `internal/source` `home` source: `config.credentialPaths` (glob, optional `os`, description) resolved under `$HOME` and repeatable `--root`; files are routed through the structural extractors; shell histories go through the line path with their own size limit.
- [ ] Default credential paths in `internal/rules/defaults/`: AWS SSO cache and credentials, gcloud ADC, the Azure MSAL cache, gh hosts, docker config, kubeconfig, npmrc, pypirc, netrc, git-credentials, bash/zsh/fish history. Organizations can extend the list from `.whac/`.
- [ ] Rules: `**.accessToken` / `**.refreshToken` / `**.refresh_token` paths, generic JWT, `*_TOKEN` / `*_SECRET` / `*_PASSWORD` env names; each with fake test samples.
- [ ] Exposure findings (a separate category in text/JSON/SARIF): `SSH_AUTH_SOCK` connectable, `docker.sock` accessible; CUE-declared with default severities.
- [ ] `--probe-imds`: opt-in; one GET to the endpoint root with a short timeout; off by default; never requests credential paths.
- [ ] `whac scan --agent-context` = env + home + exposure; exit 1 on findings at or above the threshold; `--no-redact` ignored in this mode.
- [ ] Tests use a fake `$HOME` in `t.TempDir()`, a controlled environment (`t.Setenv`), and a temporary Unix socket; no test touches the real home directory, the environment's real secrets, or the network.
- [ ] Docs: an *Agent-Context Preflight* guide (devcontainer `postStartCommand` / microVM init example) and the CLI reference.

## Notes
- No process-memory scanning (see ADR-0011).
- Pairs with the trusted-devcontainers work: `whac` audits; the builder makes the context clean (`env -i` allowlists, scoped mounts, short-lived credentials).

## Related
[ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md) · [ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md) · [ADR-0008](/decisions/adr-0008-no-live-secret-verification-in-v1--01a1d2c5078c.md) · [CLI contract](/specs/cli-contract--01a1d2c50744.md) · [Architecture](/architecture/whac-a-leak-architecture-overview--01a1d2c507a8.md)
