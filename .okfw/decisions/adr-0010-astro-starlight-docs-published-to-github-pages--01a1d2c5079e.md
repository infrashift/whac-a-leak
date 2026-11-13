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
**Source:** Planning session; modelled on `vizzini-pr-reviewer/docs`

## Context
Users need a public, versionless guide: install, writing rules, layering, pre-commit, CI. Sibling projects (vizzini) already publish Starlight sites to GitHub Pages, with a custom landing page whose demo is produced by the tool itself.

## Decision
- `docs/` holds an Astro + `@astrojs/starlight` site, built with bun 1.3.9 (`bun.lock` committed), `site: https://infrashift.github.io`, `base: /whac-a-leak`.
- `src/pages/index.astro` is a custom landing page with the project logo (`src/assets/whac-a-leak.jpeg`, optimized through `astro:assets`). From Milestone 5 its demo comes from `docs/src/demo/gen.go` running whac over a fake-secret fixture.
- The theme uses an arcade amber/gold accent taken from the logo, distinct from vizzini's wine-red, mrman's teal and garmr's violet. Signal red is kept for findings.
- `.github/workflows/docs-release.yml` builds on pull requests and deploys on pushes to `main` (paths `docs/**`) with `withastro/action` + `actions/deploy-pages`. The repository's Pages source must be **GitHub Actions**.
- Each milestone ships the docs pages for the features it delivers (see [Development workflow](/development-workflow--01a1d2c50728.md)).

## Consequences
- Two doc surfaces: the site for users and the wiki for builders. The site's *Design Decisions* page summarizes and links ADRs rather than duplicating them.
- Docs examples must use obviously fake secrets, because the site is scanned like everything else.
