<p align="center"><img src="docs/src/assets/whac-a-leak.jpeg" alt="Whac-a-Leak logo: a Go gopher at a whack-a-mole cabinet hits keys, tokens, passwords and .env files" width="720"></p>

# Whac-a-Leak (`whac`)

Whac-a-Leak is a pure-Go secret scanner. CUE defines its rules. Each rule has a type, organization layers and its own samples. If a sample fails, whac does not load the rules.

The scan is plain Go. An Aho-Corasick keyword pass runs first. Only the rules with a keyword hit run their regex and charset-aware entropy checks. whac reads YAML, JSON and `.env` files as key paths. Allowlists use `whac1:` fingerprints, which do not change on a rebase.

> **Pre-alpha.** The [ADRs](.okfw/decisions/) record the design. Each milestone adds part of the scanner. See [PLAN.md](PLAN.md).

- **Docs:** https://infrashift.github.io/whac-a-leak/
- **Build:** `make build` (Go 1.25+, `CGO_ENABLED=0`) → `bin/whac`
- **Contributing:** read [CLAUDE.md](CLAUDE.md) and the project wiki in [`.okfw/`](.okfw/). The wiki uses [OKF](https://github.com/infrashift/okf-wiki). Agents read it through `okfw mcp`.
