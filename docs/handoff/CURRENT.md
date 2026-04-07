# Current Handoff

## Active Task
- **Task ID**: M4 (Group B Complete, Group C Next)
- **Milestone**: M4 — Routines & Links
- **Description**: Task linking, dependency validation, link CLI commands
- **Status**: Group A+B Complete, Group C (tests review) Next
- **Assigned**: 2026-04-07

## Last Session Summary
- **Session 1–3 (2026-04-05–06)**: M0 + M1 + M2.1–M2.9
- **Session 4 (2026-04-06)**: M2.10–M2.12 — Tab completions, integration tests, --json
- **Session 5 (2026-04-07)**: M3 COMPLETE — Habits & Recurrence
- **Session 6 (2026-04-07)**: Bug fixes from M3 review + M4 Group A (Routines)
- **Session 7 (2026-04-07)**: Bug fix from M4-A review (TaskJSON steps/schedule) + M4 Group B (Linking)
  - M4.4: Task linking — depends-on, blocks, related-to
  - M4.5: Dependency validation — DFS cycle detection for depends-on/blocks
  - M4.6: `bt link` / `bt unlink` CLI commands with --json, --quiet, completions
  - Links displayed in `bt task show` (both outgoing and incoming/backlinks)
  - Store: LoadLinks, LoadBacklinks, DependencyGraph methods
  - App: LinkTasks, UnlinkTasks, GetTaskLinks, hasCycle
  - 28 new tests (17 app, 5 cycle-detection unit, 15 integration)

## Current State
- **M0 COMPLETE**, **M1 COMPLETE**, **M2 COMPLETE**, **M3 COMPLETE**
- **M4 Group A COMPLETE** (M4.1 + M4.2 + M4.3)
- **M4 Group B COMPLETE** (M4.4 + M4.5 + M4.6)
- Module: `github.com/tesserabox/bentotask`
- `make test`: **234 tests** — 0 lint issues
- Packages:
  - `internal/model/` — Task struct, validation, helpers, ULID, Link/LinkType
  - `internal/store/` — Markdown I/O, SQLite index (FTS5 + habit_completions + task_links), file watcher
  - `internal/app/` — Application logic (CRUD, search, habits, routines, linking, cycle detection)
  - `internal/cli/` — Cobra commands, completions, JSON output, habit + routine + link commands
  - `internal/style/` — Terminal colors/formatting via lipgloss
  - `internal/recurrence/` — RRULE parsing and next-occurrence calculation
  - `internal/habit/` — Streak calculation, completion parsing, body formatting

## Next Steps (M4.7 / Group C)
1. **M4.7**: Review test coverage for routines and dependency graph
   - Tests are already inline with Groups A and B (22 routine + 28 linking = 50 new tests)
   - M4.7 may be considered complete already — supervisor to verify

## Blockers
- None

## Context for Next Agent
- Link types: depends-on (scheduling), blocks (scheduling), related-to (informational)
- Cycle detection uses DFS with 3-color marking (white/gray/black) — only for depends-on and blocks
- related-to links are bidirectional and skip cycle detection
- Links stored in task YAML frontmatter (`links:` field) and indexed in `task_links` table
- `bt task show` displays both outgoing (→) and incoming (←) links with titles
- `bt show --json` includes `links` array with type, direction, task_id, task_title
- `bt link` defaults to `related-to`, override with `-t depends-on` or `-t blocks`
- Self-links and duplicate links are rejected
- `task_links` table has composite PK on (source_id, target_id, link_type)
