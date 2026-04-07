# Current Handoff

## Active Task
- **Task ID**: M4
- **Milestone**: M4 — Routines & Links
- **Description**: Routine ordered steps, task linking, dependency validation
- **Status**: Planning
- **Assigned**: 2026-04-07

## Last Session Summary
- **Session 1–3 (2026-04-05–06)**: M0 + M1 + M2.1–M2.9
- **Session 4 (2026-04-06)**: M2.10–M2.12 — Tab completions, integration tests, --json
- **Session 5 (2026-04-07)**: M3 COMPLETE — Habits & Recurrence
- **Session 6 (2026-04-07)**: Bug fixes from M3 review (cmd.Println stderr, RebuildIndex habit completions)

## Current State
- **M0 COMPLETE**, **M1 COMPLETE**, **M2 COMPLETE**, **M3 COMPLETE**
- Module: `github.com/tesserabox/bentotask`
- `make test`: **162 tests** — 0 lint issues
- Packages:
  - `internal/model/` — Task struct, validation, helpers, ULID
  - `internal/store/` — Markdown I/O, SQLite index (FTS5 + habit_completions), file watcher
  - `internal/app/` — Application logic (CRUD, search, habits)
  - `internal/cli/` — Cobra commands, completions, JSON output, habit commands
  - `internal/style/` — Terminal colors/formatting via lipgloss
  - `internal/recurrence/` — RRULE parsing and next-occurrence calculation
  - `internal/habit/` — Streak calculation, completion parsing, body formatting

## M4 Plan — Grouped for Incremental Review

### Group A: Routines (M4.1 + M4.2 + M4.3)
Data model, CLI commands, and play mode for routines.

- **M4.1: Routine data model** — ordered step sequence
  - `RoutineStep` and `RoutineSchedule` already exist in `model/task.go`
  - Need: app-layer methods (AddRoutine, GetRoutine, etc.)
  - Need: routine storage in markdown (steps as frontmatter list)
- **M4.2: `bt routine create` / `bt routine play`** — CLI commands
  - `bt routine create "Morning routine" --step "Shower:5m" --step "Breakfast:15m"`
  - `bt routine list` — list all routines
- **M4.3: Routine play mode** — step-by-step terminal UX
  - Step-through with timers (bubbletea TUI or simple interactive)
  - Mark steps done/skip, show progress

### Group B: Task Linking (M4.4 + M4.5 + M4.6)
Dependency relationships between tasks.

- **M4.4: Task linking** — depends-on, blocks, related-to
  - `task_links` table already in schema
  - `LinkType` enum already in model
  - Need: app-layer methods (LinkTasks, UnlinkTasks, GetLinks)
- **M4.5: Dependency validation** — cycle detection
  - Graph traversal to detect circular dependencies
  - Reject links that would create cycles
- **M4.6: `bt link` / `bt unlink` commands** — CLI
  - `bt link <source-id> --depends-on <target-id>`
  - `bt unlink <source-id> <target-id>`
  - Show links in `bt show` output

### Group C: Tests (M4.7)
- **M4.7: Tests for routines and dependency graph**
  - Unit tests for routine model, app methods
  - Unit tests for cycle detection
  - Integration tests for CLI commands

## Next Steps
1. Start with **Group A** (M4.1 + M4.2 + M4.3) — routines
2. Commit and send for review
3. Then **Group B** (M4.4 + M4.5 + M4.6) — linking
4. Commit and send for review
5. Group C tests can be inline with each group

## Blockers
- None

## Context for Next Agent
- Routine model types already exist: `RoutineStep`, `RoutineSchedule` in `internal/model/task.go`
- Link types already defined: `depends-on`, `blocks`, `related-to` with `task_links` table already in schema
- `Validate()` already checks routine steps (type=routine requires non-empty Steps)
- Habit completions are stored in both SQLite and markdown body
- `rootCmd.SetOut(os.Stdout)` was recently set — all cmd.Println goes to stdout now
- `RebuildIndex` now repopulates habit_completions from markdown body
- Build binary with `make build` after code changes
