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
DIST      := dist

.PHONY: build build-go web test lint clean fmt help dist dist-checksums \
        dist-darwin-amd64 dist-darwin-arm64 dist-linux-amd64 dist-linux-arm64 dist-windows-amd64

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

## dist: Build CLI binaries for all platforms
dist: dist-darwin-amd64 dist-darwin-arm64 dist-linux-amd64 dist-linux-arm64 dist-windows-amd64

dist-darwin-amd64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST)/bt-darwin-amd64 ./cmd/bt/

dist-darwin-arm64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST)/bt-darwin-arm64 ./cmd/bt/

dist-linux-amd64:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST)/bt-linux-amd64 ./cmd/bt/

dist-linux-arm64:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(DIST)/bt-linux-arm64 ./cmd/bt/

dist-windows-amd64:
	@mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(DIST)/bt-windows-amd64.exe ./cmd/bt/

## dist-checksums: Generate SHA256 checksums for all dist binaries
dist-checksums: dist
	cd $(DIST) && shasum -a 256 bt-* > checksums.txt

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf internal/api/static
	rm -rf $(DIST)
	$(GO) clean

## help: Show this help message
help:
	@echo "BentoTask — available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
