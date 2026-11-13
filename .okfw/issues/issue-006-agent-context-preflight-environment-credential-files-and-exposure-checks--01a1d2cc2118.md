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
description: "Milestone 6: env and home sources, credential paths in CUE, exposure findings for credential sockets and opt-in IMDS, and the whac scan --agent-context preflight gate."
---

# ISSUE-006: Agent-context preflight: environment, credential files and exposure checks

**Status/Milestone:** Open · Milestone 6 (needs the ISSUE-003 extractors)

## Description
Implement [ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md): audit what an agent process can reach. Fail a devcontainer or microVM preflight when a credential is reachable.

## Acceptance Criteria
- [ ] `internal/source` `env` source: the process environment → `env.<NAME>` assignments. No output contains a value.
- [ ] `internal/source` `home` source: resolves `config.credentialPaths` (glob, optional `os`, description) under `$HOME` and each `--root`. The structural extractors parse the files. Shell histories go through the line path with their own size limit.
- [ ] Builtin credential paths in `internal/rules/builtin/`. The list covers the AWS SSO cache and credentials, gcloud ADC and the Azure MSAL cache. It also covers gh hosts, docker config, kubeconfig, npmrc, pypirc, netrc, git-credentials and bash/zsh/fish history. Organizations extend the list from `.whac/`.
- [ ] Rules: `**.accessToken` / `**.refreshToken` / `**.refresh_token` paths, generic JWT, `*_TOKEN` / `*_SECRET` / `*_PASSWORD` environment names. Each rule has fake samples.
- [ ] Exposure findings (a separate category in text, JSON and SARIF): a connectable `SSH_AUTH_SOCK`, an accessible `docker.sock`. CUE declares them with default severities.
- [ ] `--probe-imds`: opt-in, off by default. One GET to the endpoint root with a short timeout. It never requests credential paths.
- [ ] `whac scan --agent-context` = env + home + exposure. Exit 1 on findings at or above the threshold. `--no-redact` has no effect in this mode.
- [ ] Tests use a fake `$HOME` in `t.TempDir()`, a controlled environment (`t.Setenv`) and a temporary Unix socket. No test reads the real home directory, real secrets in the environment, or the network.
- [ ] Docs: an *Agent-Context Preflight* guide (devcontainer `postStartCommand` and microVM init examples) and the CLI reference.

## Notes
- No process memory scan (see ADR-0011).
- This milestone pairs with the trusted-devcontainers work. whac audits a context. The environment builder makes it clean: `env -i` allowlists, scoped mounts, short-lived credentials.

## Related
[ADR-0011](/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md) · [ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md) · [ADR-0008](/decisions/adr-0008-no-live-secret-verification-in-v1--01a1d2c5078c.md) · [CLI contract](/specs/cli-contract--01a1d2c50744.md) · [Architecture](/architecture/whac-a-leak-architecture-overview--01a1d2c507a8.md)
