# PLAN.md — Whac-a-Leak (`whac`)

Work on one milestone at a time (`CLAUDE.md` rule 1). Each milestone has a mirror issue in the OKF wiki (`.okfw/issues/`). When a milestone is done, tick this checklist and the issue's acceptance criteria.

Loop: implement → `make lint vet-cue test` → repeat until green → update wiki and docs → `make wiki docs-style docs-build` → commit.

Architecture: [.okfw/architecture/whac-a-leak-architecture-overview--01a1d2c507a8.md](.okfw/architecture/whac-a-leak-architecture-overview--01a1d2c507a8.md)

## Milestone 0 — Project bootstrap
- [x] git repo, Go module, `Makefile`, `.golangci.yml`, `.gitignore`
- [x] `CLAUDE.md` (purpose, pinned libraries, rules) + okfw managed block
- [x] OKF wiki (`okfw init`, `okfw install --platform claude`), `.mcp.json` registering `okfw mcp`
- [x] Wiki seeded: guides, architecture, specs, ADR-0001…0011, ISSUE-001…006
- [x] Binary renamed `wal` → `whac` (pywal clash; ADR-0002 amendment)
- [x] Astro Starlight docs in `docs/` + `.github/workflows/docs-release.yml` (GitHub Pages)

## Milestone 1 — Rule schema, CUE loader, `rules verify`
Tracking: [ISSUE-001](.okfw/issues/issue-001-rule-schema-cue-loader-and-rules-verify--01a1d2c507b1.md)
- [x] Cobra skeleton: `cmd/whac` → `internal/cli`; `whac version`
- [x] Spike: an in-memory build instance replaces the `cue/load` Overlay (ADR-0012); `overrides` replaces re-grading through `rules` (ADR-0004 amendment)
- [x] `internal/rules/builtin/schema.cue` (`#SecretRule`, `#Override`, `#Config`, `#Allowlist`)
- [x] 5 builtin rules in `internal/rules/builtin/` with samples
- [x] Loader, decode, compile → `rules.Rule`; `engine.Verify` with entropy and charset checks
- [x] `whac rules list`, `whac rules verify`, `whac rules export-schema`
- [x] Layering / conflict / broken-rule / failed-sample tests; startup benchmark (about 2 ms)

## Milestone 2 — Scan engine, filesystem source, fingerprints, allowlist
Tracking: [ISSUE-002](.okfw/issues/issue-002-scan-engine-filesystem-source-fingerprints-and-allowlist--01a1d2c507bb.md)
- [ ] AC prefilter, regex + `secretGroup`, charset-aware entropy, stopwords / allowRegexes / safe formats
- [ ] `rules verify` through the full pipeline
- [ ] `whac1:` fingerprints; allowlist (fingerprints, path globs, inline `whac:ignore`); `whac allow add`
- [ ] Filesystem source (`git ls-files` / walk), binary + size gates, worker pool
- [ ] Text + JSON reports (redacted), exit codes
- [ ] `testdata/` corpus + goldens; throughput benchmark ≥100 MB/s
- [ ] Docs pages for the delivered features

## Milestone 3 — Structure-aware scanning
Tracking: [ISSUE-003](.okfw/issues/issue-003-structure-aware-scanning-and-assignment-rules--01a1d2c507c5.md)
- [ ] Extractors: YAML, JSON, dotenv, properties/ini, HCL-lite; generic line extractor
- [ ] `assignment` rules: `keyContext`, `paths`, `config.sensitivePaths`
- [ ] Placeholder filter; zero-escape entropy trap with safe-format exclusions
- [ ] Corpus additions; docs pages

## Milestone 4 — Staged and git-history scanning
Tracking: [ISSUE-004](.okfw/issues/issue-004-staged-and-git-history-scanning--01a1d2c507cf.md)
- [ ] `--staged` (added lines, staged blobs)
- [ ] `--git [range]` streaming diff parser, attribution, dedup
- [ ] Temp-repo tests; `.pre-commit-hooks.yaml`; docs pages

## Milestone 5 — SARIF, default ruleset, CI, docs demo, polish
Tracking: [ISSUE-005](.okfw/issues/issue-005-sarif-default-ruleset-ci-docs-demo-and-release-polish--01a1d2c507d9.md)
- [ ] SARIF 2.1.0
- [ ] About 30 builtin rules
- [ ] `.github/workflows/ci.yml` with self-scan
- [ ] `docs/src/demo/gen.go` landing-page demo; remaining docs; README; benchmarks

## Milestone 6 — Agent-context preflight
Tracking: [ISSUE-006](.okfw/issues/issue-006-agent-context-preflight-environment-credential-files-and-exposure-checks--01a1d2cc2118.md) · Decision: [ADR-0011](.okfw/decisions/adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md) · Needs Milestone 3's extractors
- [ ] `env` source: process environment → `env.<NAME>` assignments (values never printed)
- [ ] `home` source: CUE-declared `config.credentialPaths` under `$HOME` + `--root` (AWS SSO cache, gcloud ADC, Azure MSAL, gh, docker, kube, npmrc, pypirc, netrc, git-credentials, shell histories)
- [ ] Rules: `**.accessToken` / `**.refreshToken` paths, generic JWT, `*_TOKEN|_SECRET|_PASSWORD` env names
- [ ] Exposure findings: `SSH_AUTH_SOCK`, `docker.sock`; opt-in `--probe-imds` (the one network call)
- [ ] `whac scan --agent-context` preflight gate (exit 1, `--no-redact` has no effect); no process memory scan
- [ ] Fake-`$HOME` / `t.Setenv` / temp-socket tests; *Agent-Context Preflight* docs guide (devcontainer + microVM examples)
