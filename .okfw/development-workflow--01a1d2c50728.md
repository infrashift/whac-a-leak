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
description: "The loop for each Whac-a-Leak milestone: implement, lint, verify until green, update the wiki and docs, then commit."
---

# Development workflow

Every change follows this loop. Do not report or commit work while a gate fails.

1. **Pick the milestone.** Read `PLAN.md`. Take the first unchecked milestone. Open its mirror issue with `okfw search --path issues --tag open`.
2. **Tag the issue.** Run `okfw remove-tags <id> open && okfw add-tags <id> in-progress`.
3. **Implement only that milestone.** Use the pinned libraries in `CLAUDE.md`. Put code in `internal/`. Keep `cmd/whac` a thin `main`.
4. **Get approval for new dependencies.** Write an ADR before you add a dependency that is not on the pinned list. Do not add unrequested interfaces or factories.
5. **Lint and verify.** Run `make lint && make vet-cue && make test`. Run `make build` when the CLI exists.
6. **Repeat until green.** Fix each failure and run all gates again. Do not weaken or delete tests or rule samples.
7. **Update the wiki during the work.** Record each decision as an ADR when you make it. File each bug or design debt as an issue.
8. **Close the milestone.** Tick the `PLAN.md` checkboxes and the issue's acceptance criteria. Add a **Resolution** section. Swap the tag `in-progress` → `closed`.
9. **Update the docs site.** Add or refresh the `docs/` pages for the delivered features. Update the roadmap page. Run `make docs-build`.
10. **Check the writing style.** Run `make docs-style`. Fix every error. See [Wiki conventions](/wiki-conventions--01a1d2c5072f.md).
11. **Validate the wiki.** Run `make wiki`.
12. **Journal.** Add an entry to today's `okfw daily` note: what changed, what you learned, what is next.
13. **Commit.** Commit code, wiki and docs together. Make one commit per milestone.
