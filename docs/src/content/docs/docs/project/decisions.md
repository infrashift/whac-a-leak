---
title: Design Decisions
description: The architecture decisions behind Whac-a-Leak, summarized from its ADRs.
---

Each decision is recorded in full as an ADR in the project's OKF wiki
(`.okfw/decisions/`). This page summarizes them.

## ADR-0001: An OKF wiki for project knowledge

Architecture notes, specs, ADRs, issues and daily notes live in an
[okfw](https://github.com/infrashift/okf-wiki)-managed wiki that is committed
with the code. Agents reach it through the `okfw mcp` server registered in
`.mcp.json`. The wiki is for the people building whac; this site is for the
people using it.

## ADR-0002: Pure Go, `CGO_ENABLED=0`, a Cobra CLI

A secret scanner runs on laptops, in pre-commit hooks, on CI runners and in
minimal containers, so it ships as one static binary. The CLI uses Cobra for
consistent help and shell completion. Git is driven through the `git` binary
rather than a Go git library, because anything with a repository to scan
already has git.

The binary is `whac`, not `wal`: pywal, a popular Linux theming tool, already
installs a `wal` binary. The config directory (`.whac/`), CUE package,
fingerprint prefix (`whac1:`) and inline marker (`whac:ignore`) follow it.

## ADR-0003: CUE at load time only

Rules are CUE, but scanning is plain Go. At startup, the embedded schema,
default rules and your `.whac/` layers are unified, validated and decoded into
Go structs, then compiled into regexes and an Aho-Corasick keyword automaton.
CUE is never evaluated per file, line or match: that would cost orders of
magnitude in throughput.

## ADR-0004: Rules keyed by ID, layered with defaults

CUE unification can only *add* information, never remove it. Rules therefore
live in a map keyed by ID, and the fields you would want to override use CUE
defaults:

```cue
rules: GENERIC_API_KEY: enabled: false      // turn a rule off
rules: AWS_ACCESS_KEY: severity: "CRITICAL"  // re-grade it
```

To loosen a regex, disable the rule and add your own under a new ID. Two
layers that set conflicting values fail loudly, naming both sources.

## ADR-0005: Lightweight extractors for structured files

YAML, JSON, `.env`, properties/ini and HCL-style files are turned into
positioned `(key path, value)` pairs by small Go parsers. `assignment` rules
match key names and sensitive paths such as `**.credentials.password`. A
placeholder filter keeps `${DB_PASSWORD}`, `changeme` and `<your-token>` out of
your results. Parsing every config file through the CUE runtime was considered
and rejected as too slow.

## ADR-0006: Charset-aware entropy

Shannon entropy per character can never exceed log₂ of the alphabet size: 4.0
bits for hex, 6.0 for base64. A single global threshold such as 4.3 can
therefore never flag a hex key. whac classifies each candidate's character set
and applies a matching threshold and minimum length.

## ADR-0007: Fingerprints and allowlists

A finding's identity is `whac1:` + SHA-256 of `rule_id ‖ 0x00 ‖ secret`, with
no path or line number, so allowlist entries survive rebases. SHA-256 needs no
extra dependency. An HMAC key would have to be shared with every developer and
CI job, so it would add nothing. **Allowlists are for false positives only**:
a committed hash of a short real password could be brute-forced. Real secrets
get rotated, never allowlisted.

## ADR-0008: No live secret verification in v1

whac never calls a provider API to check whether a credential works, and never
transmits a candidate secret. It runs the same in air-gapped CI as on a laptop.

## ADR-0009: Rule tests run through the real engine

`cue vet` checks each rule's `mustMatch` samples against its regex, which
gives rule authors a quick check without the binary. The authoritative check
runs in Go: every sample goes through the same pipeline as a real scan, every
time rules load. A pure-CUE `mustIgnore` check would be wrong, because many
ignore samples are *meant* to match the regex and then be rejected by entropy
or placeholder filters.

## ADR-0010: Starlight docs on GitHub Pages

This site is Astro Starlight, built with bun and deployed to GitHub Pages by
`.github/workflows/docs-release.yml`. Pull requests build the site, and pushes
to `main` deploy it. Each milestone ships the docs for the features it
delivers.

## ADR-0011: Agent-context preflight

Coding agents can read whatever their process can reach. `whac scan
--agent-context` audits exactly that:

- **Environment variables**, scanned as `env.<NAME>` key paths.
- **Well-known credential files** under `$HOME` and mounted roots: AWS SSO
  cache, gcloud ADC, Azure MSAL, `gh`, Docker, kube, `.npmrc`, `.netrc`,
  `.git-credentials` and shell history. The list is declared in CUE, so
  organizations can extend it.
- **Exposure checks**: a connectable `SSH_AUTH_SOCK` or an accessible
  `docker.sock`. Probing the cloud metadata endpoint is opt-in
  (`--probe-imds`), and it is the only network call whac ever makes.

Values are never printed. It runs as a preflight gate at devcontainer or
microVM startup and exits 1 if anything is reachable. Process memory is
deliberately out of scope: reading it needs ptrace or root and looks
exactly like credential-dumping malware.
