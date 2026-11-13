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
description: "whac scan --agent-context audits what an agent process can reach (environment, credential files, credential sockets, optional IMDS) as a preflight gate; process memory is out of scope."
---

# ADR-0011: Agent-context preflight scans the environment and credential files, never process memory

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning discussion on environment variables, SSO tokens and process memory

## Context
Coding agents run in devcontainers and microVMs. An agent can read everything its process can reach: environment variables, files of its user, mounted volumes, credential sockets and the cloud metadata endpoint. An `AWS_SECRET_ACCESS_KEY` in the environment gives an agent too much access. A fresh AWS SSO token in `~/.aws/sso/cache` does the same. The discussion considered three targets: environment variables, SSO/CLI tokens and process memory.

## Decision
- **Environment variables: in scope.** An `env` source turns the process environment into `Assignment{KeyPath: "env.<NAME>", Value}` records for the engine ([ADR-0005](/decisions/adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md)):
  - `keyContext` rules find `*_TOKEN`, `*_SECRET` and `*_PASSWORD`.
  - Pattern rules find tokens with a prefix.
  - The placeholder filter drops values such as `${…}`.
- **SSO and CLI tokens: in scope, as well-known file locations plus rules.** A `home` source reads `config.credentialPaths`, a list in CUE. Each entry has a glob, an optional OS and a description. The source resolves each glob under `$HOME` and under each `--root`, such as a mounted volume. Default entries:
  - AWS: `~/.aws/sso/cache/*.json`, `~/.aws/credentials`
  - gcloud: `~/.config/gcloud/application_default_credentials.json`
  - Azure: the MSAL token cache
  - `~/.config/gh/hosts.yml`, `~/.docker/config.json`, `~/.kube/config`, `~/.npmrc`, `~/.pypirc`, `~/.netrc`, `~/.git-credentials`
  - shell histories (bash, zsh, fish)

  The Milestone 3 extractors parse these files. New `paths` rules (`**.accessToken`, `**.refreshToken`, `**.refresh_token`) and a generic JWT rule cover them. Organizations extend the list in `.whac/`, the same way they layer rules ([ADR-0004](/decisions/adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md)).
- **Exposure findings: a second finding category.** An exposure finding records reachable access to credentials, not a secret string. Examples: `SSH_AUTH_SOCK` points at a socket that accepts connections; `/var/run/docker.sock` is accessible. CUE declares each exposure check with a default severity. Organizations can re-grade or disable each check.
- **IMDS: opt-in only.** `--probe-imds` sends one GET to the metadata endpoint root with a short timeout. It never requests credential paths and never transmits a secret. It is the only opt-in exception to the offline rule in [ADR-0008](/decisions/adr-0008-no-live-secret-verification-in-v1--01a1d2c5078c.md).
- **Process memory: out of scope.**
  - Access to the memory of another process needs ptrace or root. The method differs on each OS. It is slow and noisy.
  - Endpoint security tools cannot tell it apart from credential-dumping malware.
  - It does not match the threat model. An agent without ptrace rights cannot read that memory either. The agent's own process holds the credentials it legitimately uses.
- **Preflight gate.** `whac scan --agent-context` runs `env`, `home` and the exposure checks. Run it at devcontainer or microVM startup, as the agent's user, with the agent's environment and mounts. It exits 1 when anything at or above the threshold is reachable. Agent-context findings name the variable, file or key path, the rule and the `whac1:` fingerprint. They **never** contain the value. `--no-redact` has no effect in this mode, because preflight output often goes to logs that the agent can read.

## Consequences
- whac **audits** a context. The environment builder makes the context clean: an allowlist of environment variables (`env -i`), scoped mounts, and short-lived credentials with narrow scope. This fits the trusted-devcontainers work.
- The feature needs the Milestone 3 extractors. It is Milestone 6.
- whac reads sensitive files by design. It writes nothing and transmits nothing. The opt-in IMDS probe is the only network call.
- Credential paths differ on Linux, macOS and Windows. Each entry has an optional `os` field.

## Alternatives considered
- **Process memory scan.** Rejected for the reasons above.
- **IMDS probe on by default.** Rejected: it is a network call, and it adds noise on laptops.
- **A separate tool.** Rejected: the `env` and `home` sources reuse the engine, the rules and the allowlists.
