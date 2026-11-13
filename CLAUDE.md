# CLAUDE.md — Whac-a-Leak secret scanner

## Purpose (WHY)
Whac-a-Leak (`whac`) is a pure-Go secret scanner. CUE defines its rules. Each rule has a type, organization layers and its own samples.

- CUE runs **only at load time**. The loader unifies, validates and decodes the rules, then compiles them into Go matchers.
- The scan is plain Go: Aho-Corasick prefilter → regex → entropy/charset → filters → allowlist.
- Sources: filesystem, staged changes (pre-commit), git history.
- Agent-context preflight sources: environment variables and well-known credential files (ADR-0011). whac never scans process memory.
- Outputs: text, JSON, SARIF.

## Tech Stack & Pinned Libraries (WHAT)
- **Language:** Go 1.25+ with `CGO_ENABLED=0`. Pure Go, no cgo.
- **Rules and config:** `cuelang.org/go` (CUE Go API, v0.17.x).
- **CLI:** `github.com/spf13/cobra`.
- **Keyword prefilter:** `github.com/petar-dambovaliev/aho-corasick`.
- **Path globs:** `github.com/bmatcuk/doublestar/v4`.
- **YAML positions:** the yaml.v3 module that `cuelang.org/go` already uses.
- **Git:** the `git` binary through `os/exec`. No Go git library.
- **Docs:** Astro Starlight in `docs/` (bun 1.3.9), deployed to GitHub Pages.

## Layout
- `cmd/whac/` holds only a thin `main` that calls `internal/cli`.
- `internal/` holds all logic: `cli`, `rules`, `engine`, `structural`, `source`, `allow`, `fingerprint`, `report`.
- Add no `pkg/` until a public Go API exists by design.
- Embedded CUE (`//go:embed`) sits next to the package that embeds it: `internal/rules/builtin/`.

## Verification & Workflow Commands (HOW)
Run these GNU Make targets after every code change:
- `make test`     → `go test -race ./...`
- `make vet-cue`  → `cue vet` on the embedded schema and builtin rules
- `make lint`     → `golangci-lint run` + `go mod tidy -diff`
- `make build`    → `CGO_ENABLED=0 go build -o bin/whac ./cmd/whac`
- `make bench`    → engine and rule-load benchmarks
- `make wiki`     → `okfw validate --all && okfw validate --okf`
- `make docs-style` → `scripts/ste-check.py`: ASD-STE100 check of the documentation
- `make docs-build` → build the Starlight site (the same build as the Pages deploy)

## Mandatory Execution Rules
1. **One milestone at a time.** Read `PLAN.md`. Implement only the first unchecked milestone. Run `make test && make vet-cue && make lint`. Tick the `PLAN.md` checkboxes, then commit.
2. **Test integrity.** Do not weaken or delete tests or rule samples (`tests.mustMatch`, `tests.mustIgnore`) to make a gate pass.
3. **No CUE in the hot path.** CUE runs only while rules load. Do not call the CUE runtime per file, line or match (ADR-0003).
4. **Minimal abstractions.** Keep the `cmd/` + `internal/` layout. Add no unrequested interfaces or factories. Write an ADR before you add a dependency outside the pinned list.
5. **No real secrets.** Samples, fixtures and docs demos use values that cannot be real credentials: `EXAMPLE` markers, invalid checksums.
6. **No scanned secrets in output.** Do not paste a scanned secret into output, logs, commits or the wiki. Identify a finding by its fingerprint.
7. **Wiki and docs stay in sync.** Each `PLAN.md` milestone has a mirror issue in `.okfw/issues/`. Its `Tracking:` line links to the issue. When a milestone is done:
   - Tick the issue's acceptance criteria. Add a **Resolution** section.
   - Swap the tag `in-progress` → `closed`.
   - Record new decisions as ADRs in `.okfw/decisions/`.
   - Add or update the `docs/` pages for the delivered features.
   - Run `make wiki`.

   Conventions: the *Wiki conventions* guide in `.okfw/`.
8. **Documentation follows ASD-STE100 (Simplified Technical English).** Use active voice, simple present tense and short sentences. Run `make docs-style`.

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
