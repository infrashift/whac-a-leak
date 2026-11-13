---
title: Make Targets
description: The GNU Make targets used to build, test, lint and document Whac-a-Leak.
---

Everything runs through GNU Make. `CGO_ENABLED=0` is exported for every target
except `test`, because Go's race detector needs cgo on Linux.

| Target | What it runs |
|---|---|
| `make build` | `go build -trimpath -o bin/whac ./cmd/whac` (static, pure Go) |
| `make test` | `go test -race ./...` |
| `make lint` | `golangci-lint run` with the strict `.golangci.yml`, plus `go mod tidy -diff` |
| `make vet-cue` | `cue vet ./internal/rules/...`: the embedded schema and default rules |
| `make bench` | Engine throughput and rule-load benchmarks |
| `make wiki` | `okfw validate --all && okfw validate --okf` on the project wiki |
| `make docs-dev` | This site with live reload (bun) |
| `make docs-build` | Build this site into `docs/dist`, the same as the Pages deploy |
| `make clean` | Remove `bin/` |

Before a commit, the gate is `make lint vet-cue test`. Once the CLI exists,
`make build` is added. Fix and rerun until all are green.
