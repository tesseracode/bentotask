# Current Handoff

## Active Task
- **Task ID**: M1.1
- **Milestone**: M1 — Core Data Model
- **Description**: Implement task data model (struct + serialization)
- **Status**: Not Started
- **Assigned**: 2026-04-06

## Last Session Summary
- **Session 1 (2026-04-05)**: Full planning phase — SPEC.md, 3 ADRs (all APPROVED), tracking infrastructure
- **Session 2 (2026-04-06)**: M0 Bootstrap COMPLETE
  - M0.6: Go module (`github.com/tesserabox/bentotask`), folder structure, root Cobra command
  - M0.7: golangci-lint v2 config, Makefile (build/test/lint/fmt/clean/help), CONTRIBUTING.md
  - M0.8: 3 passing tests (Execute, VersionCommand, RootHasGlobalFlags), GitHub Actions CI
  - Fixed: version command uses `cmd.Printf` for testability, unused `args` param, lint v2 config

## Current State
- **M0 is COMPLETE** — all 8 tasks done, all acceptance criteria met
- Module: `github.com/tesserabox/bentotask`
- Dependencies: cobra v1.10.2
- 5 commits on `main` branch
- `make build`, `make test`, `make lint` all pass (0 lint issues, 3 tests)
- CI configured: `.github/workflows/ci.yml` (test + lint on push/PR)

## Next Steps (in order)
1. **M1.1: Task data model** — struct definition, field types per ADR-002 frontmatter schema
2. **M1.2: Markdown + YAML frontmatter reader/writer** — using `adrg/frontmatter`
3. **M1.3: SQLite index** — schema, create, rebuild from files
4. **M1.4: File watcher** — `fsnotify` for external changes
5. **M1.5: ULID generation** — `oklog/ulid` for task IDs
6. **M1.6: Unit tests** — full coverage for data model & storage layer

## Blockers
- None

## Context for Next Agent
- Start by reading this file, then `CLAUDE.md` for orientation
- ADR-002 defines the frontmatter schema and SQLite table structure — read it carefully for M1
- Key libraries to add: `adrg/frontmatter`, `modernc.org/sqlite`, `oklog/ulid`, `teambition/rrule-go`, `fsnotify/fsnotify`
- Charm TUI libraries (`bubbletea`, `huh`, `lipgloss`) are for M2, not needed yet
- Use `make test` and `make lint` before committing
- The developer is learning Go — keep things idiomatic, explain patterns
