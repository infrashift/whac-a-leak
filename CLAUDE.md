# CLAUDE.md — Whac-a-Leak secret scanner

## Purpose (WHY)
Whac-a-Leak (`whac`) is a lightweight, pure-Go secret scanner. Its detection rules are written in CUE: typed, layered by organization policy, and self-testing, since every rule ships its own positive and negative samples. CUE is used only **at load time**, to unify, validate and decode rules, which then compile into Go matchers. Scanning (Aho-Corasick prefilter → regex → entropy/charset → filters → allowlist) is plain Go. Sources: filesystem, staged changes (pre-commit), git history and, for agent-context preflight, environment variables and well-known credential files (ADR-0011; never process memory). Outputs: text, JSON, SARIF.

## Tech Stack & Pinned Libraries (WHAT)
- **Language:** Go 1.25+ with `CGO_ENABLED=0` (strictly pure Go, no cgo).
- **Rules & config:** `cuelang.org/go` (CUE Go API, v0.17.x).
- **CLI:** `github.com/spf13/cobra`.
- **Keyword prefilter:** `github.com/petar-dambovaliev/aho-corasick`.
- **Path globs:** `github.com/bmatcuk/doublestar/v4`.
- **YAML positions:** the yaml.v3 module that `cuelang.org/go` already depends on.
- Git is driven through the `git` binary (`os/exec`), not a Go git library.
- Docs: Astro Starlight in `docs/` (bun 1.3.9), deployed to GitHub Pages.

## Layout
`cmd/whac/` holds only a thin `main` that calls `internal/cli`. All logic lives in `internal/` (`cli`, `rules`, `engine`, `structural`, `source`, `allow`, `fingerprint`, `report`). There is no `pkg/` until a public Go API is deliberately designed. Embedded CUE (`//go:embed`) lives next to the package that embeds it (`internal/rules/schema`, `internal/rules/defaults`).

## Verification & Workflow Commands (HOW)
Run these through GNU Make after every code change:
- `make test`     → `go test -race ./...`
- `make vet-cue`  → `cue vet ./internal/rules/...`
- `make lint`     → `golangci-lint run` + `go mod tidy -diff`
- `make build`    → `CGO_ENABLED=0 go build -o bin/whac ./cmd/whac`
- `make bench`    → engine and rule-load benchmarks
- `make wiki`     → `okfw validate --all && okfw validate --okf`
- `make docs-build` → build the Starlight site (what the Pages deploy runs)

## Mandatory Execution Rules
1. **Work One Phase at a Time:** Read `PLAN.md`, implement only the current unchecked milestone, run `make test && make vet-cue && make lint`, tick the milestone's checkboxes in `PLAN.md`, and stop for a commit.
2. **Test Integrity:** Never weaken or delete existing tests (or a rule's `tests.mustMatch` / `tests.mustIgnore` samples) to make a failing implementation pass.
3. **CUE Stays Out of the Hot Path:** CUE is evaluated only while rules load. Never call the CUE runtime per file, per line or per match (ADR-0003).
4. **Minimal Abstractions:** Keep the `cmd/` + `internal/` layout. Don't add unrequested interfaces, factory patterns, or third-party dependencies outside the pinned list; a new dependency needs an ADR first.
5. **No Real Secrets, Anywhere:** Test samples, fixtures and docs demos use values that cannot be real credentials (`EXAMPLE` markers, invalid checksums). Never paste a scanned secret into output, logs, commits or the wiki; refer to findings by fingerprint.
6. **Keep the Wiki and Docs in Sync:** Each `PLAN.md` milestone has a mirror issue in `.okfw/issues/` (linked from its `Tracking:` line). When a milestone is done:
   - Tick the issue's acceptance criteria and add a **Resolution** section.
   - Swap the tag `in-progress` → `closed`.
   - Record new decisions as ADRs in `.okfw/decisions/`.
   - Add or refresh the `docs/` pages for the features it delivered.
   - Run `make wiki`.

   Conventions: `.okfw/` *Wiki conventions* guide.

<!-- BEGIN okfw -->
## Project knowledge base (okfw)

whac-a-leak keeps a local-first OKF wiki in `.okfw/`: markdown concepts with
YAML frontmatter, linked to each other. The files are the whole database —
search and the link graph are computed from them on demand, so there is no
index to build and nothing to sync. Each directory's `index.md` is generated
and gives a cheap overview before you open anything (OKF section 8).

- **Search before you derive.** `okfw search "<query>"` / `okfw_search`. The
  wiki holds what the code does not say — why things are built the way they
  are, what was considered and rejected, dependency and security findings,
  operational context, open work. Check it before explaining a design,
  assessing risk, or running an analysis someone may already have recorded and
  adjudicated.
- **Browse top-down.** Each directory's `index.md` lists its concepts with
  their descriptions (`okfw_read {"ref":"<dir>/"}` over MCP). Open only what fits.
- **Search before you write**, too, so a new concept links to what already
  exists instead of duplicating it.
- **Link actively** — `[Title](/dir/file.md)` from the wiki root, or
  `./file.md`. Links are what makes the wiki more than a folder.
- **One idea per concept**, with a one-sentence `description`.
- **Use `okfw move` / `okfw delete`, never `mv` / `rm`.** `move` repairs every
  link to the concept and its own relative links, and both keep the `index.md`
  listings current; the shell commands break the links and leave the listings
  stale.
- **`status` is the OKF lifecycle field** (`draft`/`stable`/`deprecated`), not
  task state. Mark superseded knowledge rather than deleting it.
- **Custom metadata goes under `x:`** — the schema is closed and rejects any
  other unrecognized key.
- Don't edit `index.md` or `log.md`; okfw generates them.

Results are capped and say so: `truncated`, or `… N more`, means you did not
see everything — raise the limit or page with `--offset`/`"offset"`.

`okfw mcp` serves all of this as MCP stdio tools named `okfw_*`.
<!-- END okfw -->
