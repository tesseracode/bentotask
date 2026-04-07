# Current Handoff

## Active Task
- **Task ID**: M4 (Group B)
- **Milestone**: M4 — Routines & Links
- **Description**: Task linking, dependency validation, link CLI commands
- **Status**: Group A Complete, Group B Next
- **Assigned**: 2026-04-07

## Last Session Summary
- **Session 1–3 (2026-04-05–06)**: M0 + M1 + M2.1–M2.9
- **Session 4 (2026-04-06)**: M2.10–M2.12 — Tab completions, integration tests, --json
- **Session 5 (2026-04-07)**: M3 COMPLETE — Habits & Recurrence
- **Session 6 (2026-04-07)**: Bug fixes from M3 review + M4 Group A (Routines)
  - M4.1: Expanded RoutineStep model (Title, Duration, Optional fields alongside Ref)
  - M4.2: CLI commands — `bt routine create`, `bt routine list`, `bt routine show`
  - M4.3: Play mode — interactive step-by-step with timer, skip optional steps
  - Step parsing: "Title:Duration?" format with optional marker
  - App layer: AddRoutine, ListRoutines, auto-computed EstimatedDuration
  - 22 new tests (2 model, 5 app, 5 unit, 10 integration)

## Current State
- **M0 COMPLETE**, **M1 COMPLETE**, **M2 COMPLETE**, **M3 COMPLETE**
- **M4 Group A COMPLETE** (M4.1 + M4.2 + M4.3)
- Module: `github.com/tesserabox/bentotask`
- `make test`: **184 tests** — 0 lint issues
- Packages:
  - `internal/model/` — Task struct, validation, helpers, ULID
  - `internal/store/` — Markdown I/O, SQLite index (FTS5 + habit_completions), file watcher
  - `internal/app/` — Application logic (CRUD, search, habits, routines)
  - `internal/cli/` — Cobra commands, completions, JSON output, habit + routine commands
  - `internal/style/` — Terminal colors/formatting via lipgloss
  - `internal/recurrence/` — RRULE parsing and next-occurrence calculation
  - `internal/habit/` — Streak calculation, completion parsing, body formatting

## Next Steps (Group B)
1. **M4.4: Task linking** — depends-on, blocks, related-to
2. **M4.5: Dependency validation** — cycle detection
3. **M4.6: `bt link` / `bt unlink` commands**
4. **M4.7: Tests** — inline with Group B

## Blockers
- None

## Context for Next Agent
- RoutineStep now has: Title, Duration, Ref, Optional (Title+Duration for inline steps, Ref for linked tasks)
- Routines stored as type=routine in inbox/ like other tasks, steps in YAML frontmatter
- Play mode is interactive (stdin) — reads Enter/s for each step, tracks elapsed time
- Play mode JSON outputs step listing without interactive execution
- `task_links` table already exists in schema, LinkType enum already defined
- `parseStepFlags` handles "Title:Duration?" format for CLI step creation
- Build binary with `make build` after code changes
