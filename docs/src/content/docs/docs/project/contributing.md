---
title: Contributing
description: How the Whac-a-Leak project organizes work, and the rules for every change.
---

## The loop

Every change follows the same loop. Do not commit while a gate fails.

1. Pick the next unchecked milestone in `PLAN.md`. Open its mirror issue in
   `.okfw/issues/` with `okfw search --path issues --tag open`.
2. Implement only that milestone. Put code in `internal/`. Keep `cmd/whac` a
   thin `main`.
3. Run `make lint vet-cue test` and `make build`. Fix each failure. Run the
   gates again until all pass.
4. Do not weaken or delete tests or rule samples to make a gate pass.
5. Update the wiki during the work. Write an ADR for each decision. File an
   issue for each new problem.
6. Close the milestone issue with a **Resolution** section. Run `make wiki`.
7. Add or update the docs pages for the milestone. Run `make docs-style` and
   `make docs-build`.
8. Commit code, wiki and docs together.

## Writing style

All documentation follows ASD-STE100 Simplified Technical English:

- Use active voice and simple present tense.
- Start each step with an imperative verb.
- Keep instructions to 20 words or fewer. Keep descriptions to 25 words or
  fewer.
- Use one term for one concept.

Run `make docs-style` to check the docs, the README and the wiki.

## No real secrets

Rule samples, test fixtures, wiki pages and these docs use values that
**cannot** be real credentials: `EXAMPLE` markers, invalid checksums, fictional
providers such as *Acme*. CI scans the repository with whac. Do not commit a
real secret.

## Rule samples

Each rule needs one or more `mustMatch` samples. A good rule also has
`mustIgnore` samples. These samples match the regex, and a filter rejects
them: placeholders, low entropy. Run `whac rules verify` (from Milestone 1)
before you open a pull request.

## Dependencies

`CLAUDE.md` lists the pinned libraries. Write an ADR before you add a
dependency that is not on the list.

## License

Whac-a-Leak uses the Apache License 2.0. A contribution is under the same
license. See the `LICENSE` file in the repository.

## Releases

1. Make sure CI passes on `main`.
2. Create an annotated tag: `git tag -a v0.2.0 -m "whac v0.2.0"`.
3. Push the tag: `git push origin v0.2.0`.

* **Expected result:** the Release workflow builds four archives, runs each
  binary on its platform, and publishes the release with `SHA256SUMS` and a
  build provenance attestation.
