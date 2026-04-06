# Contributing to BentoTask

Thanks for your interest in contributing to BentoTask!

## Development Setup

### Prerequisites

- **Go 1.22+** — [install](https://go.dev/doc/install)
- **golangci-lint v2** — [install](https://golangci-lint.run/welcome/install/)

### Getting Started

```bash
git clone https://github.com/tesserabox/bentotask.git
cd bentotask
make build
./bt --version
```

### Common Commands

```bash
make build    # Compile the bt binary
make test     # Run all tests
make lint     # Run golangci-lint
make fmt      # Format all Go files
make clean    # Remove build artifacts
```

## Code Style

- Run `make lint` before committing — CI will reject PRs that don't pass
- Follow standard Go conventions (`gofmt`, idiomatic error handling)
- Import order: stdlib, then external, then internal (`github.com/tesserabox/bentotask/...`)
- Use `cmd.Printf` (not `fmt.Printf`) in Cobra commands so output is testable

## Project Structure

```
cmd/bt/           CLI entry point (thin — just calls internal/cli)
internal/cli/     Cobra commands and CLI logic
internal/model/   Data structures (task, habit, routine)
internal/store/   Markdown I/O + SQLite index
internal/engine/  Scheduling algorithm
internal/...      Other domain packages
```

## Testing

- Write tests for every feature — files named `*_test.go` next to the code
- Use the standard `testing` package
- Run `make test` to execute all tests

## Architecture Decisions

Key decisions are documented in `docs/adrs/`. Read the relevant ADR before working on a feature:

- **ADR-001**: Tech stack (Go + SvelteKit)
- **ADR-002**: Storage format (Markdown + YAML frontmatter, SQLite index)
- **ADR-003**: CLI patterns (noun-verb commands, Charm ecosystem)

## Commit Messages

Use the format: `M<milestone>.<task>: <imperative description>`

Examples:
```
M0.6: Add project scaffolding — Go module, folder structure, root CLI command
M1.2: Implement markdown frontmatter reader/writer
```

## Reporting Issues

Open an issue on GitHub with:
1. What you expected to happen
2. What actually happened
3. Steps to reproduce
4. Your Go version (`go version`) and OS
