# Current Handoff

## Active Task
- **Task ID**: M0.6
- **Milestone**: M0 — Project Bootstrap
- **Description**: Set up project scaffolding (go mod, folder structure, CI)
- **Status**: Not Started
- **Assigned**: 2026-04-05

## Last Session Summary
- **Session 1 (2026-04-05)**: Full planning phase completed
- Created initial SPEC.md (721 lines) covering all features
- Backed up original spec as SPEC.original.md (gitignored)
- Created project tracking infrastructure:
  - ROADMAP.md with 10 milestones
  - CLAUDE.md (agent orientation)
  - AGENTS.md (roles, handoff/supervisor protocols)
  - Handoff system (CURRENT.md + HISTORY.md)
  - Supervisor system (LOG.md)
- Wrote and approved 3 ADRs:
  - **ADR-001** (APPROVED): Go + SvelteKit
  - **ADR-002** (APPROVED): Markdown+YAML frontmatter, file-per-task (ULID filenames), SQLite index (pure Go, no CGO), hybrid mtime+hash sync, RFC 5545 RRULE for recurrence
  - **ADR-003** (APPROVED): Noun-verb commands (`bt task add`), Charm ecosystem (cobra+bubbletea+huh+lipgloss), --json/--quiet output, $EDITOR integration, dynamic shell completions

## Current State
- All 3 ADRs are APPROVED — architecture decisions are locked in
- No Go code exists yet
- M0 milestone: 5 of 8 tasks complete (all decisions done, scaffolding remains)
- 2 commits on `main` branch

## Next Steps (in order)
1. **M0.6: Project scaffolding**
   - `go mod init github.com/tesserabox/bentotask` (DONE)
   - Create folder structure per ADR-001: `cmd/bt/`, `internal/{model,store,engine,calendar,routine,graph,api}/`
   - Create `cmd/bt/main.go` with root Cobra command
   - `bt --version` and `bt --help` working
2. **M0.7: Coding standards**
   - Set up `golangci-lint` config
   - Create Makefile or Taskfile (build, test, lint targets)
   - Create CONTRIBUTING.md
3. **M0.8: First test**
   - Write and run at least one passing test
   - Set up GitHub Actions CI
4. **Then M1: Core Data Model** — implement task struct, markdown I/O, SQLite index

## Blockers
- None — all decisions approved, ready to code

## Context for Next Agent
- Start by reading this file, then `CLAUDE.md` for orientation
- All ADRs are in `docs/adrs/` — they define the tech stack, storage format, and CLI UX
- Key libraries to install: `cobra`, `bubbletea`, `huh`, `lipgloss`, `adrg/frontmatter`, `modernc.org/sqlite`, `oklog/ulid`, `teambition/rrule-go`, `fsnotify/fsnotify`
- Binary name is `bt` (short for BentoTask)
- The developer is learning Go — keep things idiomatic, explain patterns
- Data directory default: `~/.bentotask/data/`
- Check if Go is installed: `go version`
