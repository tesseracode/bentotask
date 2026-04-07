# Current Handoff

## Active Task
- **Task ID**: M2.5 / M2.8–M2.11
- **Milestone**: M2 — Basic CLI
- **Description**: Edit command, tags/priority/energy, search, completions, integration tests
- **Status**: Not Started
- **Assigned**: 2026-04-06

## Last Session Summary
- **Session 1 (2026-04-05)**: Full planning phase — SPEC.md, 3 ADRs (all APPROVED)
- **Session 2 (2026-04-06)**: M0 COMPLETE + M1 COMPLETE + M2 core commands
  - M1.1–M1.5: Data model, Markdown I/O, ULID, SQLite index, file watcher
  - M2 app layer: `internal/app/` — AddTask, ListTasks, GetTask, CompleteTask, DeleteTask, RebuildIndex
  - M2 CLI commands: `bt add`, `bt list`, `bt done`, `bt show`, `bt delete`, `bt index rebuild`
  - Top-level aliases: `bt add`, `bt list`, `bt done` (per ADR-003)
  - Noun aliases: `bt t`, `bt tasks` (per ADR-003)
  - Quiet mode (-q) for piping, all flags per ADR-003

## Current State
- **M0 COMPLETE**, **M1 COMPLETE**, M2 core CRUD done
- Module: `github.com/tesserabox/bentotask`
- `make test`: **65+ tests** — 0 lint issues
- Working CLI: `bt add`, `bt list`, `bt done`, `bt show`, `bt task delete`, `bt index rebuild`
- Packages:
  - `internal/model/` — Task struct, validation, helpers, ULID
  - `internal/store/` — Markdown I/O, SQLite index, file watcher
  - `internal/app/` — Application logic (CRUD orchestration)
  - `internal/cli/` — Cobra commands with flags and aliases

## Next Steps
1. **M2.5: `bt edit`** — modify task via flags or $EDITOR
2. **M2.8: Tags, priority, energy** — enhance list output with color (lipgloss)
3. **M2.9: Search and filtering** — full-text search (FTS5)
4. **M2.10: Tab completion** — dynamic Cobra completions for task IDs
5. **M2.11: Integration tests** — end-to-end CLI tests

## Blockers
- None

## Context for Next Agent
- `internal/app/app.go` is the service layer — CLI commands call App methods
- ADR-003 defines all flags, output formats, and UX patterns
- `bt edit` should support both `--title "new"` flag updates and `$EDITOR` opening
- Color output (lipgloss) is deferred — current output is plain text tables
