---
id: "01a1d2cc-210e-70b6-98bd-0a17f95aa83d"
title: "ADR-0011: Agent-context preflight scans the environment and credential files, never process memory"
type: "ADR"
tags:
  - "adr"
  - "agent-context"
  - "scope"
  - "security"
created: "2026-11-12T20:45:55-06:00"
timestamp: "2026-11-12T20:45:55-06:00"
description: "whac scan --agent-context audits what an agent's process can reach (environment variables, well-known credential files, credential sockets, optionally IMDS) and reports it as a preflight gate; process memory is out of scope."
---

# ADR-0011: Agent-context preflight scans the environment and credential files, never process memory

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning discussion on scanning environment variables, SSO tokens and process memory

## Context
Coding agents run in devcontainers and microVMs, and anything their process can read is exposed to them: environment variables, files their user can read, mounted volumes, credential sockets and the cloud metadata endpoint. Leaving an `AWS_SECRET_ACCESS_KEY` in the environment or a fresh AWS SSO token in `~/.aws/sso/cache` gives an agent far more access than intended. We considered three targets: environment variables, SSO/CLI tokens and process memory.

## Decision
- **Environment variables: in scope.** An `env` source turns the scanning process's environment into `Assignment{KeyPath: "env.<NAME>", Value}` records for the existing engine ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md)):
  - `keyContext` rules catch `*_TOKEN`, `*_SECRET` and `*_PASSWORD`.
  - pattern rules catch prefixed tokens.
  - the placeholder filter drops values like `${…}`.
- **SSO and CLI tokens: in scope, as well-known file locations plus rules.** A `home` source reads a CUE-declared list, `config.credentialPaths` (glob, optional OS, description), resolved under `$HOME` and any extra `--root` such as a mounted volume. Defaults:
  - AWS: `~/.aws/sso/cache/*.json`, `~/.aws/credentials`
  - gcloud: `~/.config/gcloud/application_default_credentials.json`
  - Azure: the MSAL token cache
  - `~/.config/gh/hosts.yml`, `~/.docker/config.json`, `~/.kube/config`, `~/.npmrc`, `~/.pypirc`, `~/.netrc`, `~/.git-credentials`
  - shell histories (bash, zsh, fish)

  The files go through the Milestone 3 extractors. New `paths` rules (`**.accessToken`, `**.refreshToken`, `**.refresh_token`) and a generic JWT rule cover them. Organizations extend the list in `.whac/` the same way they layer rules ([ADR-0004](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md)).
- **Exposure findings: a second finding category.** These record reachable access to credentials, not secret strings: `SSH_AUTH_SOCK` pointing at a socket that accepts connections, and an accessible `/var/run/docker.sock`. Exposure checks are declared in CUE with default severities, so organizations can re-grade or disable them.
- **IMDS: opt-in only.** `--probe-imds` sends one GET to the metadata endpoint root with a short timeout. It never requests credential paths and never sends a secret. This is the single, opt-in exception to the offline rule in [ADR-0008](/decisions/adr-0008-no-live-secret-verification-in-v1--01a1d2c5078c.md).
- **Process memory: out of scope.**
  - Reading another process's memory needs ptrace or root, behaves differently on each OS, and is slow and noisy.
  - It looks the same as credential-dumping malware to endpoint security tools.
  - It doesn't match the threat model: an agent without ptrace rights can't read that memory either, and the agent's own process has to hold whatever credentials it legitimately uses.
- **Preflight gate.** `whac scan --agent-context` combines `env`, `home` and the exposure checks. It is meant to run at devcontainer or microVM startup, as the agent's user with the agent's environment and mounts, and it exits 1 if anything at or above the threshold is reachable. Agent-context findings name the variable, file or key path, the rule and the `whac1:` fingerprint, and **never** the value. `--no-redact` has no effect in this mode, because preflight output usually ends up in logs the agent can read.

## Consequences
- whac **audits** whether a context is clean. Making it clean is the environment builder's job: an allowlist of environment variables (`env -i`), scoped mounts, and short-lived, narrowly scoped credentials. This fits the trusted-devcontainers work.
- Depends on the Milestone 3 extractors, so it lands as Milestone 6.
- whac reads sensitive files by design, but it writes and transmits nothing. The only network call is the opt-in IMDS probe.
- Credential paths differ between Linux, macOS and Windows, so entries carry an optional `os`.

## Alternatives considered
- **Scanning process memory.** Rejected for the reasons above.
- **Always probing IMDS.** Rejected: it is a network call, and noisy on laptops.
- **A separate tool.** Rejected: the env and home sources reuse the whole engine, rule set and allowlists.
