---
id: "01a1d2c5-072f-7bcc-8113-323e53a43739"
title: "Wiki conventions"
type: "Guide"
tags:
  - "conventions"
  - "meta"
created: "2026-11-12T20:38:10-06:00"
timestamp: "2026-11-12T20:38:10-06:00"
description: "Layout, writing style and authoring rules for architecture notes, specs, ADRs, issues and daily notes in this wiki."
---

# Wiki conventions

This guide defines how the Whac-a-Leak OKF wiki (`.okfw/`) stores project knowledge.

- Create concepts with `okfw create` or the okfw MCP `create` tool. Do not write frontmatter by hand.
- Search before you write. Link to related concepts.
- Keep one idea per concept. Give each concept a one-sentence `description`.

## Layout
| Path | Type | Contents |
|---|---|---|
| `architecture/` | `Architecture` | System overview, package boundaries, load phase and scan phase, sources. |
| `specs/` | `Spec` | Contracts: rule schema, fingerprint format, CLI. The spec states intent. The code (for example `internal/rules/builtin/schema.cue`) implements it. |
| `decisions/` | `ADR` | Architecture decision records. |
| `issues/` | `Issue` (kind `issue`) | Tracked work: one issue per `PLAN.md` milestone, plus bugs and design debt. |
| `plans/` | `Plan` | Multi-issue plans beyond `PLAN.md`. Create this directory when you first need it. |
| `daily/` | daily note | `okfw daily`: what changed, what you learned, what is next. |

## PLAN.md and issues
`PLAN.md` is the milestone checklist for `CLAUDE.md` rule 1. Each milestone has a mirror issue (ISSUE-001 … ISSUE-006). Each `PLAN.md` heading has a `Tracking:` line that links to its issue. When a milestone is done, tick the `PLAN.md` checkboxes and the issue's acceptance criteria.

## ADRs (`decisions/`)
- Title: `ADR-NNNN: <decision>`. Number ADRs in sequence. Do not reuse a number.
- Command: `okfw create "ADR-NNNN: <title>" --path decisions --type ADR --tags adr,<topic> -d "<one sentence>" --body "<markdown>"`
- Body: **Status** (Proposed / Accepted / Superseded by ADR-XXXX), **Date**, **Source**, then **Context**, **Decision**, **Consequences**, **Alternatives considered**.
- To supersede an ADR, write the new ADR with a link to the old ADR. Update the Status line of the old ADR. Run `okfw set-lifecycle <id> deprecated`.
- Write an ADR before you add a dependency that is not on the pinned list in `CLAUDE.md`.

## Issues (`issues/`)
- Title: `ISSUE-NNN: <summary>`. Number issues in sequence.
- Command: `okfw create "ISSUE-NNN: <title>" --kind issue --path issues --type Issue --tags issue,open,phase-<n>,<area> -d "<summary>" --body "<markdown>"`
- Body: a **Status/Milestone** line, **Description**, **Acceptance Criteria** (checkboxes), **Notes**, **Related**.
- okfw has no status field for issues. Tags carry the state: `open` → `in-progress` → `closed`. Use `okfw remove-tags` and `okfw add-tags`. Keep the Status line the same as the tags. Add a **Resolution** section when you close an issue.
- Find open work with `okfw search --path issues --tag open`.

## Writing style
All documentation follows ASD-STE100 Simplified Technical English. This includes the wiki, the docs site, `README.md`, `PLAN.md` and `CLAUDE.md`.

- Use active voice and simple present tense: "The loader rejects the rule."
- Start each procedural step with an imperative verb: Run, Read, Inspect, Verify, Add, Set.
- Keep instructions to 20 words or fewer. Keep descriptive sentences to 25 words or fewer.
- Use a verb, not a noun form: "The engine verifies the samples", not "Verification of the samples occurs".
- Use one term for one concept. This wiki uses these terms:
  - *rule*, *builtin rule*, *sample* (a `mustMatch` or `mustIgnore` entry)
  - *layer* (a `.cue` file in the config dir), *config dir* (`.whac/`)
  - *secret*, *finding*, *fingerprint*, *allowlist*, *engine*, *milestone*
- Do not use the words `should`, `will`, `would`, `please`, `simply`, `just`, `very`, `really` or `easily`. Do not use marketing language.
- Run `make docs-style`. The check (`scripts/ste-check.py`) reports banned words and long sentences as errors and possible passive voice as warnings.

## Secrets hygiene
The repository includes the wiki. Do not paste a real or scanned secret into a concept. Identify a finding by its `whac1:` fingerprint. Use sample secrets that are clearly fake: `EXAMPLE` markers or invalid checksums.

## Housekeeping
- Use `okfw move` and `okfw delete`. Do not use `mv` or `rm`. The okfw commands repair links and update the `index.md` listings.
- `status` is the OKF lifecycle field (`draft` / `stable` / `deprecated`). It is not the task state. Mark superseded knowledge `deprecated`. Do not delete it.
- Do not edit `index.md` or `log.md`. okfw generates them.
- Run `make wiki` before you commit wiki changes.

See [Development workflow](/development-workflow--01a1d2c50728.md).
