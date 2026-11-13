.PHONY: all test vet-cue lint build dist release bench wiki docs-style docs-demo docs-dev docs-build clean

# golangci-lint runs through `go run` at a pinned version, so the current Go
# toolchain builds it. A binary built with an older Go cannot type-check a newer
# standard library.
GOLANGCI ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
CUE      ?= cue

PKG     := github.com/infrashift/whac-a-leak/internal/cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X $(PKG).version=$(VERSION) -X $(PKG).commit=$(COMMIT) -X $(PKG).date=$(DATE)

# Pure Go everywhere (ADR-0002). The one exception is `test`: Go's race
# detector runtime needs cgo on Linux, so it turns cgo back on for that target only.
export CGO_ENABLED := 0

all: lint test build

## test: unit tests with the race detector
test:
	CGO_ENABLED=1 go test -race ./...

## vet-cue: validate the embedded schema and builtin rules (concrete, incl. mustMatch samples)
vet-cue:
	$(CUE) vet -c ./internal/rules/builtin/

## lint: golangci-lint (strict config in .golangci.yml) plus a go mod tidy check
lint:
	$(GOLANGCI) run
	go mod tidy -diff

## build: static, stripped whac binary with version information
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/whac ./cmd/whac

# Release targets: Linux and macOS on amd64 and arm64 (arm64 = AArch64,
# including Apple silicon). Pure Go, so every target cross-compiles.
PLATFORMS   ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
RELVERSION  := $(VERSION:v%=%)

## dist: build one release archive in dist/ for GOOS/GOARCH (default: the host)
dist:
	@goos=$${GOOS:-$$(go env GOOS)}; goarch=$${GOARCH:-$$(go env GOARCH)}; \
	name=whac_$(RELVERSION)_$${goos}_$${goarch}; \
	mkdir -p dist/stage/$$name && \
	GOOS=$$goos GOARCH=$$goarch go build -trimpath -ldflags "$(LDFLAGS)" -o dist/stage/$$name/whac ./cmd/whac && \
	cp README.md LICENSE dist/stage/$$name/ && \
	tar -C dist/stage --owner=0 --group=0 --numeric-owner -czf dist/$$name.tar.gz $$name && \
	echo "dist/$$name.tar.gz"

## release: build every platform archive and dist/SHA256SUMS for VERSION
release:
	@for p in $(PLATFORMS); do $(MAKE) --no-print-directory dist GOOS=$${p%/*} GOARCH=$${p#*/} || exit 1; done
	cd dist && sha256sum whac_$(RELVERSION)_*.tar.gz > SHA256SUMS && cat SHA256SUMS

## bench: engine throughput and rule-load benchmarks
bench:
	go test -run '^$$' -bench . -benchmem ./internal/...

## wiki: validate the OKF wiki
wiki:
	okfw validate --all
	okfw validate --okf

## docs-style: check the docs, README, PLAN.md, CLAUDE.md and wiki against ASD-STE100 rules
docs-style:
	python3 -I scripts/ste-check.py

## docs-demo: regenerate the landing-page scan output with whac itself (exit 1 = findings, expected)
docs-demo: build
	cd docs/src/demo && { ../../../bin/whac scan --config-dir . app.env > scan-output.txt; test $$? -eq 1; }

## docs-dev: run the Astro Starlight site locally (bun)
docs-dev:
	cd docs && bun install && bun --bun run dev

## docs-build: build the site into docs/dist (what CI and the Pages deploy run)
docs-build:
	cd docs && bun install --frozen-lockfile && bun --bun run build

clean:
	rm -rf bin/ dist/
