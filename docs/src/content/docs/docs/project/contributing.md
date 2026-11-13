---
title: Contributing
description: How work on Whac-a-Leak is organized, and the rules every change follows.
---

## The loop

Every change goes through the same loop. Nothing is committed while a gate is red.

1. Pick the next unchecked milestone in `PLAN.md` and its mirror issue in
   `.okfw/issues/` (`okfw search --path issues --tag open`).
2. Implement only that milestone. Code lives in `internal/`, and `cmd/whac`
   stays a thin `main`.
3. Run `make lint vet-cue test` (plus `make build`). Fix the failures and
   rerun until everything is green. Never weaken or delete existing tests or
   rule samples to get there.
4. Update the wiki as you go: an ADR for each decision, an issue for each
   finding, and at the end the closed milestone issue with a **Resolution**.
   Then run `make wiki`.
5. Add or refresh the docs pages for what the milestone delivered, then run
   `make docs-build`.
6. Commit code, wiki and docs together.

## No real secrets, anywhere

Rule samples, test fixtures, wiki pages and these docs use values that
**cannot** be real credentials: `EXAMPLE` markers, impossible checksums,
fictional providers like *Acme*. The repository scans itself in CI, so a real
secret would be caught, but don't make it try.

## Writing rules

Every rule needs at least one `mustMatch` sample, and good rules also have
`mustIgnore` samples that match the regex but are rightly rejected
(placeholders, low entropy). Run `whac rules verify` (from Milestone 1) before
opening a pull request.

## Dependencies

The pinned library list is in `CLAUDE.md`. Adding a dependency outside it
needs an ADR first.
