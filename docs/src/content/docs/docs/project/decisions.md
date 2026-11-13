---
title: Design Decisions
description: The architecture decisions behind Whac-a-Leak, summarized from its ADRs.
---

The project OKF wiki (`.okfw/decisions/`) records each decision in full as an
ADR. This page gives a summary of each ADR.

## ADR-0001: An OKF wiki for project knowledge

An [okfw](https://github.com/infrashift/okf-wiki) wiki holds architecture
notes, specs, ADRs, issues and daily notes. The repository includes the wiki.
Agents read it through the `okfw mcp` server in `.mcp.json`. The wiki serves
contributors. This site serves users.

## ADR-0002: Pure Go, `CGO_ENABLED=0`, a Cobra CLI

A secret scanner runs on laptops, in pre-commit hooks, on CI runners and in
minimal containers. whac ships as one static binary. The CLI uses Cobra for
consistent help and shell completion. whac runs the `git` binary, not a Go git
library. A machine with a repository to scan has git.

The binary is `whac`, not `wal`. pywal, a popular Linux theming tool, already
installs a `wal` binary. The config dir (`.whac/`), the CUE package, the
fingerprint prefix (`whac1:`) and the inline marker (`whac:ignore`) use the
same name.

## ADR-0003: CUE at load time only

Rules are CUE. The scan is plain Go. At startup, the loader unifies the embedded
schema, the builtin rules and your `.whac/` layers. It validates the result and
decodes it into Go structs. Then it compiles regexes and an Aho-Corasick keyword
automaton. whac never evaluates CUE per file, line or match. That costs orders
of magnitude in throughput.

## ADR-0004: Rules keyed by ID, layered with defaults

CUE unification only *adds* information. It never removes information. A map
keyed by ID therefore holds the rules. A builtin rule sets a concrete
severity, so a layer cannot change it through `rules`. The `overrides` map
disables or re-grades a rule:

```cue
overrides: GENERIC_API_KEY: enabled: false     // disable a rule
overrides: AWS_ACCESS_KEY_ID: severity: "LOW"  // re-grade a rule
```

To loosen a regex, disable the rule and add your own rule with a new ID. If two
layers set conflicting values, the load fails. The error names both sources.
Allowlist entries and exclusions are maps, so two layers can each add entries.

## ADR-0005: Small extractors for structured files

Small Go parsers turn YAML, JSON, `.env`, properties/ini and HCL-style files
into positioned `(key path, value)` pairs. `assignment` rules match key names
and sensitive paths such as `**.credentials.password`. A placeholder filter
removes `${DB_PASSWORD}`, `changeme` and `<your-token>` from the results. A
parse of every config file through the CUE runtime is too slow. The project
rejected it.

## ADR-0006: Charset-aware entropy

The alphabet size limits Shannon entropy per character to log₂ of that size:
4.0 bits for hex, 6.0 for base64. A global threshold such as 4.3 therefore
never flags a hex key. Length also limits entropy: a 20-character string
reaches at most 4.32. Without an explicit minimum, whac uses
`0.75 × log2(min(length, alphabet size))`.

## ADR-0007: Fingerprints and allowlists

A finding's identity is `whac1:` + SHA-256 of `rule_id ‖ 0x00 ‖ secret`. The
fingerprint contains no path and no line number, so allowlist entries survive
rebases. SHA-256 needs no extra dependency. Every developer and CI job needs an
HMAC key, so an HMAC adds nothing.

**Allowlists hold false positives only.** An attacker can brute-force a
committed hash of a short real password. Rotate real secrets. Do not allowlist
them.

## ADR-0008: No live secret verification in v1

whac never calls a provider API to test a credential. It never transmits a
candidate secret. It behaves the same in air-gapped CI and on a laptop.

## ADR-0009: The engine runs the rule samples

`cue vet` checks each rule's `mustMatch` samples against its regex. Rule
authors get a fast check without the binary. The Go engine runs the
authoritative check. At every load, each sample goes through the same pipeline
as a real scan.

A `mustIgnore` check in CUE alone is wrong. Many ignore samples match the regex
on purpose. Entropy or placeholder filters then reject them.

## ADR-0010: Starlight docs on GitHub Pages

This site uses Astro Starlight. bun builds it.
`.github/workflows/docs-release.yml` deploys it to GitHub Pages. Pull requests
build the site. Pushes to `main` deploy it. Each milestone adds the docs for
its features. All documentation follows ASD-STE100 Simplified Technical
English.

## ADR-0011: Agent-context preflight

A coding agent reads everything its process can reach. `whac scan
--agent-context` audits that:

- **Environment variables**, as `env.<NAME>` key paths.
- **Well-known credential files** under `$HOME` and mounted roots: AWS SSO
  cache, gcloud ADC, Azure MSAL, `gh`, Docker, kube, `.npmrc`, `.netrc`,
  `.git-credentials` and shell history. CUE declares the list, so
  organizations can extend it.
- **Exposure checks**: a connectable `SSH_AUTH_SOCK` or an accessible
  `docker.sock`. The cloud metadata endpoint (IMDS) check is opt-in
  (`--probe-imds`). It is the only network call whac makes.

whac never prints values. Run the command as a preflight gate at devcontainer
or microVM startup. It exits 1 when a credential is reachable. Process memory
is out of scope. Access to it needs ptrace or root. Security tools cannot tell
it apart from credential-dumping malware.

## ADR-0012: One in-memory CUE build instance

The loader parses the builtin files and each layer, then builds them as one
CUE package in memory. It needs no `cue.mod` and no files on disk. Layers can
refer to schema definitions such as `#SecretRule`. The loader skips a schema
copy from `whac rules export-schema`, so a stale copy never conflicts.

## ADR-0013: Regexes run on keyword line windows

A rule's regex runs only on the lines that hold one of its keywords. The
first version ran each regex over the whole file and reached 24 MB/s. Line
windows raised that to about 245 MB/s. A rule with `multiline: true`, such as
`PRIVATE_KEY`, gets a window of up to 64 KiB. When two rules report the same
span, whac keeps one finding: the higher severity.

## ADR-0014: Git with pinned options

User git configuration changes diff output: prefixes, signatures, path
quoting and external diff tools. whac sets those options on every git call.
The diff parser counts the lines that each hunk announces, so an added line
that starts with `++` is never read as a file header.

## ADR-0015: Assembled samples and the self-scan

GitHub push protection blocks real-looking tokens. A rule file therefore
builds each sample for such a format from pieces, for example
`"sk" + "_live_" + "EXAMPLE…"`. CUE joins the pieces at load time. whac scans
its own repository in CI. The allowlist covers fixture directories by path,
and fake values in test files by fingerprint.


## ADR-0017: Report files with `--output`

`--format` sets the report on stdout. `--output FORMAT=PATH` writes one more
report to a file from the same scan. Repeat it for more files. The JSON
report stays one document. whac collects and sorts all findings before
output, so JSON Lines gives no streaming. The envelope (`apiVersion`,
`kind`, `summary`) is the versioned contract. Use `jq -c '.findings[]'` to
get one finding per line.
