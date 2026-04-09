# BentoTask Makefile
# Run `make help` to see available targets.

# Variables
BINARY    := bt
MODULE    := github.com/tesserabox/bentotask
VERSION   ?= dev
LDFLAGS   := -ldflags "-X $(MODULE)/internal/cli.version=$(VERSION)"
GO        := go
GOTEST    := $(GO) test
LINT      := golangci-lint

.PHONY: build build-go web test lint clean fmt help

## build: Build web UI + Go binary (full release build)
build: web
	$(GO) build $(LDFLAGS) -o $(BINARY) ./cmd/bt/

## build-go: Build Go binary only (skip web, for fast iteration)
build-go:
	$(GO) build $(LDFLAGS) -o $(BINARY) ./cmd/bt/

## web: Build the SvelteKit web UI and copy to embed directory
web:
	cd web && npm run build
	rm -rf internal/api/static
	cp -r web/build internal/api/static

## test: Run all tests
test:
	$(GOTEST) ./... -v

## lint: Run golangci-lint
lint:
	$(LINT) run ./...

## fmt: Format all Go files
fmt:
	$(GO) fmt ./...
	@command -v goimports >/dev/null 2>&1 && goimports -w -local $(MODULE) . || echo "note: goimports not installed, skipping import sorting"

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf internal/api/static
	$(GO) clean

## help: Show this help message
help:
	@echo "BentoTask — available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
