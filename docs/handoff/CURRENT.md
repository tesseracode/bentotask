# Current Handoff

## Active Task
- **Task ID**: M4
- **Milestone**: M4 — Routines & Links
- **Description**: Routine ordered steps, task linking, dependency validation
- **Status**: Not Started
- **Assigned**: 2026-04-07

## Last Session Summary
- **Session 1–3 (2026-04-05–06)**: M0 + M1 + M2.1–M2.9
- **Session 4 (2026-04-06)**: M2.10–M2.12 — Tab completions, integration tests, --json
- **Session 5 (2026-04-07)**: M3 COMPLETE — Habits & Recurrence
  - M3.1: Recurrence engine (internal/recurrence/) — RRULE parsing via teambition/rrule-go, NextAfter, NextAfterCompletion, Between, Frequency
  - M3.2: Next-occurrence calculation for fixed and completion-anchor modes
  - M3.3: Habit data model (internal/habit/) — Completion struct, body parsing, streak calc
  - M3.4: CLI commands — `bt habit add`, `bt habit log`, `bt habit stats`, `bt habit list`
  - M3.5: Streak engine — daily/weekly streaks, current/longest, completion rate
  - M3.6: Dual storage — habit_completions SQLite table + markdown body ## Completions section
  - M3.7: 46 new tests (13 recurrence, 17 habit, 7 app, 9 integration)
  - Bug fix: ListTasks/Search now load tags/contexts from junction tables

## Current State
- **M0 COMPLETE**, **M1 COMPLETE**, **M2 COMPLETE**, **M3 COMPLETE**
- Module: `github.com/tesserabox/bentotask`
- `make test`: **159 tests** — 0 lint issues
- Packages:
  - `internal/model/` — Task struct, validation, helpers, ULID
  - `internal/store/` — Markdown I/O, SQLite index (FTS5 + habit_completions), file watcher
  - `internal/app/` — Application logic (CRUD, search, habits)
  - `internal/cli/` — Cobra commands, completions, JSON output, habit commands
  - `internal/style/` — Terminal colors/formatting via lipgloss
  - `internal/recurrence/` — RRULE parsing and next-occurrence calculation
  - `internal/habit/` — Streak calculation, completion parsing, body formatting

## Next Steps
1. **M4.1: Routine data model** — ordered step sequence
2. **M4.2: `bt routine create` / `bt routine play`**
3. **M4.3: Routine play mode** — step-by-step terminal UX
4. **M4.4: Task linking** — depends-on, blocks, related-to
5. **M4.5: Dependency validation** — cycle detection
6. **M4.6: `bt link` / `bt unlink` commands**
7. **M4.7: Tests for routines and dependency graph**

## Blockers
- None

## Context for Next Agent
- Habit completions are stored in both SQLite (`habit_completions` table) and markdown body (`## Completions` section)
- Streaks are cached in frontmatter (`streak_current`, `streak_longest`) and recalculated on each log
- Recurrence engine supports fixed-anchor and completion-anchor modes
- `internal/cli/habits.go` contains all habit CLI commands
- Routines already have model types: `RoutineStep`, `RoutineSchedule`, and validation in `Validate()`
- Link types are defined: `depends-on`, `blocks`, `related-to` with `task_links` table already in schema
