# decisions

<!-- okfw:generated — maintained by okfw; edits here are overwritten -->

## Concepts

* [ADR-0001: Adopt an OKF wiki (okfw) for project knowledge](adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1d2c5074c.md) - Project knowledge (architecture, specs, ADRs, issues, daily notes) lives in an okfw-managed OKF wiki at .okfw/.
* [ADR-0002: Pure Go with CGO_ENABLED=0 and a Cobra CLI](adr-0002-pure-go-with-cgo-enabled0-and-a-cobra-cli--01a1d2c50754.md) - whac is a single static pure-Go binary (CGO_ENABLED=0) whose command surface is built with spf13/cobra.
* [ADR-0003: CUE at load time only; the scan engine is plain Go](adr-0003-cue-at-load-time-only-the-scan-engine-is-plain-go--01a1d2c5075d.md) - CUE unifies, validates and decodes rules at startup; scanning never touches the CUE runtime and runs as an Aho-Corasick prefilter followed by Go regexp and entropy checks.
* [ADR-0004: Rules are keyed by ID and layered with CUE default markers](adr-0004-rules-are-keyed-by-id-and-layered-with-cue-default-markers--01a1d2c50766.md) - Rules live in a map keyed by ID with defaulted enabled/severity fields, so org layers can add, tighten, disable or re-grade rules through unification without forking.
* [ADR-0005: Structured files are parsed by lightweight Go extractors, not the CUE runtime](adr-0005-structured-files-are-parsed-by-lightweight-go-extractors-not-the-cue-runtime--01a1d2c50770.md) - YAML, JSON, dotenv, properties/ini and HCL-style files are turned into positioned (keyPath, value) assignments by small Go extractors; CUE only declares which key paths are sensitive.
* [ADR-0006: Entropy thresholds are charset-aware](adr-0006-entropy-thresholds-are-charset-aware--01a1d2c50779.md) - Shannon entropy is compared against per-charset thresholds and minimum lengths, because the maximum entropy of a string is bounded by its alphabet.
* [ADR-0007: Domain-separated SHA-256 fingerprints; allowlists are for false positives only](adr-0007-domain-separated-sha-256-fingerprints-allowlists-are-for-false-positives-only--01a1d2c50782.md) - Findings are identified by whac1:SHA-256(rule_id || 0x00 || secret), which is stable across rebases; committed fingerprints are acceptable because allowlists only hold false positives.
* [ADR-0008: No live secret verification in v1](adr-0008-no-live-secret-verification-in-v1--01a1d2c5078c.md) - whac never makes network calls to check whether a found credential is live; detection is fully offline.
* [ADR-0009: Rule tests run through the real engine pipeline at load time](adr-0009-rule-tests-run-through-the-real-engine-pipeline-at-load-time--01a1d2c50795.md) - Each rule's mustMatch and mustIgnore samples are executed by the Go engine whenever rules load, and a failure aborts loading; CUE checks mustMatch against the regex for rule authors.
* [ADR-0010: Astro Starlight docs published to GitHub Pages](adr-0010-astro-starlight-docs-published-to-github-pages--01a1d2c5079e.md) - User documentation is an Astro Starlight site in docs/, built with bun and deployed to GitHub Pages by .github/workflows/docs-release.yml.
* [ADR-0011: Agent-context preflight scans the environment and credential files, never process memory](adr-0011-agent-context-preflight-scans-the-environment-and-credential-files-never-process-memory--01a1d2cc210e.md) - whac scan --agent-context audits what an agent's process can reach (environment variables, well-known credential files, credential sockets, optionally IMDS) and reports it as a preflight gate; process memory is out of scope.

