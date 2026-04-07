# Current Handoff

## Active Task
- **Task ID**: M1.6 / M2
- **Milestone**: M1 wrapping up → M2 Basic CLI
- **Description**: Final test coverage + start CLI commands
- **Status**: Not Started
- **Assigned**: 2026-04-06

## Last Session Summary
- **Session 1 (2026-04-05)**: Full planning phase — SPEC.md, 3 ADRs (all APPROVED)
- **Session 2 (2026-04-06)**: M0 COMPLETE + M1.1–M1.5 done
  - M1.1: Task data model — Task struct, enums, validation, helpers (18 tests)
  - M1.2: Markdown I/O — Parse, Marshal, WriteFile with atomic writes (11 tests)
  - M1.5: ULID generation — NewID, NewIDAt, IDTime, MatchesPrefix (6 tests)
  - M1.3: SQLite index — schema, upsert, query, filter, rebuild from files (13 tests)
  - M1.4: File watcher — fsnotify, create/modify/delete detection, subdirs (6 tests)

## Current State
- **M0 COMPLETE**, **M1.1–M1.5 COMPLETE** (M1.6 remaining tests are partial — already at 54)
- Module: `github.com/tesserabox/bentotask`
- Dependencies: cobra, adrg/frontmatter, yaml.v3, oklog/ulid/v2, modernc.org/sqlite, fsnotify
- `make test`: **54 tests** (3 CLI + 24 model + 30 store) — 0 lint issues
- Packages:
  - `internal/model/` — Task struct, validation, helpers, ULID generation
  - `internal/store/` — Markdown I/O, SQLite index, file watcher
  - `internal/cli/` — Root Cobra command with global flags

## Deferred ADR-002 items
| Section | What | Deferred to |
|---------|------|-------------|
| §5 Schema | `routine_steps` table | M4 (Routines & Links) |
| §5 Schema | `habit_completions` table | M3 (Habits & Recurrence) |
| §5 Schema | FTS5 full-text search | M2.9 (Search and filtering) |
| §6 Sync | mtime + xxhash incremental sync on startup | M2.1 (CLI scaffolding) |
| §7 | RRULE recurrence handling | M3 (Habits & Recurrence) |

## Next Steps (in order)
1. **M1.6: Unit tests** — fill any remaining coverage gaps (mostly done already)
2. **M2: Basic CLI** — `bt add`, `bt list`, `bt done`, `bt show`, `bt edit`, `bt delete`
3. Then M3: Habits & Recurrence

## Blockers
- None

## Context for Next Agent
- Start by reading this file, then `CLAUDE.md` for orientation
- The full data layer is built: model → markdown I/O → SQLite index → file watcher
- M2 connects the CLI commands to this data layer
- ADR-003 defines command structure, flags, output formats
- Use `make test` and `make lint` before committing
