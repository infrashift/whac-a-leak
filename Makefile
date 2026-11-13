.PHONY: all test vet-cue lint build bench wiki docs-dev docs-build clean

GOBIN    := $(shell go env GOPATH)/bin
GOLANGCI ?= $(GOBIN)/golangci-lint
CUE      ?= cue

# Pure Go everywhere (ADR-0002). The one exception is `test`: Go's race
# detector runtime needs cgo on Linux, so it turns cgo back on for that target only.
export CGO_ENABLED := 0

# True once the module contains at least one Go package (Milestone 1 onward).
HAS_GO := $(shell test -n "$$(go list ./... 2>/dev/null)" && echo yes)

all: lint test build

## test: unit tests with the race detector
test:
ifeq ($(HAS_GO),yes)
	CGO_ENABLED=1 go test -race ./...
else
	@echo "test: no Go packages yet, skipping"
endif

## vet-cue: validate the embedded CUE schema and default rules
vet-cue:
	@if [ -d internal/rules ]; then $(CUE) vet ./internal/rules/...; else echo "vet-cue: no internal/rules/ yet, skipping"; fi

## lint: golangci-lint (strict config in .golangci.yml) plus a go mod tidy check
lint:
ifeq ($(HAS_GO),yes)
	$(GOLANGCI) run
	go mod tidy -diff
else
	@echo "lint: no Go packages yet, skipping"
endif

## build: static whac binary
build:
	go build -trimpath -o bin/whac ./cmd/whac

## bench: engine throughput and rule-load benchmarks
bench:
	go test -run '^$$' -bench . -benchmem ./internal/...

## wiki: validate the OKF wiki
wiki:
	okfw validate --all
	okfw validate --okf

## docs-dev: run the Astro Starlight site locally (bun)
docs-dev:
	cd docs && bun install && bun --bun run dev

## docs-build: build the site into docs/dist (what CI and the Pages deploy run)
docs-build:
	cd docs && bun install --frozen-lockfile && bun --bun run build

clean:
	rm -rf bin/
