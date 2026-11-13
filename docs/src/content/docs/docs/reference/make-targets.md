---
title: Make Targets
description: The GNU Make targets that build, test, lint and document Whac-a-Leak.
---

All tasks run through GNU Make. The Makefile exports `CGO_ENABLED=0` for every
target except `test`. Go's race detector needs cgo on Linux.

| Target | Command |
|---|---|
| `make build` | `go build -trimpath` with version ldflags, stripped (`-s -w`) → `bin/whac` (static, pure Go) |
| `make test` | `go test -race ./...` |
| `make lint` | golangci-lint v2.15.0 through `go run` with the strict `.golangci.yml`, then `go mod tidy -diff` |
| `make vet-cue` | `cue vet -c ./internal/rules/builtin/`: the schema, the builtin rules and their `mustMatch` samples |
| `make dist` | One release archive for `GOOS`/`GOARCH` (default: the host) in `dist/` |
| `make release` | Archives for Linux and macOS on amd64 and arm64, plus `dist/SHA256SUMS` |
| `make bench` | Engine throughput and rule-load benchmarks |
| `make wiki` | `okfw validate --all && okfw validate --okf` on the project wiki |
| `make docs-style` | `scripts/ste-check.py`: checks the documentation against ASD-STE100 rules |
| `make docs-dev` | Serve this site with live reload (bun) |
| `make docs-build` | Build this site into `docs/dist`, the same build as the Pages deploy |
| `make clean` | Remove `bin/` |

Before a commit, run `make lint vet-cue test docs-style`. When the CLI exists,
also run `make build`. Fix each failure and run the gates again until all pass.
