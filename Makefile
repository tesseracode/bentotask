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

.PHONY: build test lint clean fmt help

## build: Compile the bt binary
build:
	$(GO) build $(LDFLAGS) -o $(BINARY) ./cmd/bt/

## test: Run all tests
test:
	$(GOTEST) ./... -v

## lint: Run golangci-lint
lint:
	$(LINT) run ./...

## fmt: Format all Go files
fmt:
	$(GO) fmt ./...
	goimports -w -local $(MODULE) .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	$(GO) clean

## help: Show this help message
help:
	@echo "BentoTask — available targets:"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
