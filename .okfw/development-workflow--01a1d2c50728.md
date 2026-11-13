---
id: "01a1d2c5-0728-76e3-94cc-270a0c61a835"
title: "Development workflow"
type: "Guide"
tags:
  - "conventions"
  - "workflow"
  - "meta"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "The one-milestone-at-a-time implement, lint, verify (repeat until green), update-wiki-and-docs, commit loop for Whac-a-Leak."
---

# Development workflow

Every change goes through this loop. Nothing is reported or committed while any gate is red.

1. **Pick the milestone.** Read `PLAN.md`, take the first unchecked milestone, and open its mirror issue (`okfw search --path issues --tag open`). Tag it in progress: `okfw remove-tags <id> open && okfw add-tags <id> in-progress`.
2. **Implement only that milestone.** Use the pinned libraries in `CLAUDE.md`. Don't add unrequested interfaces, factories or dependencies; a new dependency needs an ADR first. Code goes in `internal/`; `cmd/whac` stays a thin `main`.
3. **Lint and verify, then repeat until green.** Run `make lint && make vet-cue && make test`, and `make build` once the CLI exists. Fix what fails and run all of them again. Never weaken or delete existing tests or rule samples to get to green.
4. **Keep the wiki current as you go, not at the end.** Record a decision as an ADR when you make it. When you find a bug or design debt, file it as an issue. When the milestone is done, tick its `PLAN.md` checkboxes and the issue's acceptance criteria, add a **Resolution** section, and swap the tag `in-progress` → `closed`.
5. **Update the docs site.** Add or refresh the `docs/` pages for the features the milestone delivered, and update the roadmap page. Then run `make docs-build`.
6. **Validate the wiki.** `make wiki` (`okfw validate --all && okfw validate --okf`).
7. **Journal.** Add a line to today's `okfw daily` note: what changed, what was learned, what's next.
8. **Stop for a commit.** Code, wiki and docs go together, one commit per milestone.
