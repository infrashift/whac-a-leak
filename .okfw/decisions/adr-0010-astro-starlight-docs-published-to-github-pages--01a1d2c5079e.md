---
id: "01a1d2c5-079e-7af1-a434-7921f21a8c90"
title: "ADR-0010: Astro Starlight docs published to GitHub Pages"
type: "ADR"
tags:
  - "adr"
  - "docs"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "User documentation is an Astro Starlight site in docs/, built with bun and deployed to GitHub Pages by .github/workflows/docs-release.yml."
---

# ADR-0010: Astro Starlight docs published to GitHub Pages

**Status:** Accepted
**Date:** 2026-11-12
**Source:** Planning session; based on `vizzini-pr-reviewer/docs`

## Context
Users need a public guide: install, rule authoring, layering, pre-commit, CI. Sibling projects (vizzini) publish Starlight sites to GitHub Pages. Each site has a custom landing page with a demo that the tool itself generates.

## Decision
- `docs/` holds an Astro + `@astrojs/starlight` site. bun 1.3.9 builds it (`bun.lock` is in the repository). `site: https://infrashift.github.io`, `base: /whac-a-leak`.
- `src/pages/index.astro` is a custom landing page with the project logo (`src/assets/whac-a-leak.jpeg`). `astro:assets` optimizes the logo. From Milestone 5, `docs/src/demo/gen.go` generates the demo: it runs whac over a fixture of fake secrets.
- The theme uses an amber accent from the logo. It differs from vizzini (wine red), mrman (teal) and garmr (violet). Red marks findings only.
- `.github/workflows/docs-release.yml` builds the site for pull requests. It deploys on pushes to `main` that change `docs/**`. It uses `withastro/action` and `actions/deploy-pages`. The Pages source of the repository must be **GitHub Actions**.
- Each milestone adds the docs pages for the features it delivers. See [Development workflow](/development-workflow--01a1d2c50728.md).
- All documentation follows ASD-STE100 Simplified Technical English. `make docs-style` checks it.

## Consequences
- Two documentation surfaces exist: the site for users and the wiki for contributors. The site's *Design Decisions* page summarizes the ADRs and does not copy them.
- Docs examples use clearly fake secrets, because whac scans the site like all other files.
