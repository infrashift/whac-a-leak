<p align="center"><img src="docs/src/assets/whac-a-leak.jpeg" alt="Whac-a-Leak: a Go gopher at a whack-a-mole arcade cabinet, whacking keys, tokens, passwords and .env files as they pop up" width="720"></p>

# Whac-a-Leak (`whac`)

A lightweight, pure-Go secret scanner. Its rules are **CUE**: typed, layered by organization policy, and **self-testing**, so a rule that fails its own samples never loads. Scanning is plain Go: an Aho-Corasick keyword pass first, then regex and charset-aware entropy only for the rules whose keywords appear. It reads YAML, JSON and `.env` files as key paths, and allowlists use rebase-proof `whac1:` fingerprints.

> **Pre-alpha.** The design is settled ([ADRs](.okfw/decisions/)) and the scanner is being built one milestone at a time. See [PLAN.md](PLAN.md).

- **Docs:** https://infrashift.github.io/whac-a-leak/
- **Build:** `make build` (Go 1.25+, `CGO_ENABLED=0`) → `bin/whac`
- **Contributing:** see [CLAUDE.md](CLAUDE.md) and the project wiki in [`.okfw/`](.okfw/) (an [OKF](https://github.com/infrashift/okf-wiki) wiki, also served to agents over `okfw mcp`)
