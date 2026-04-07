# Current Handoff

## Active Task
- **Task ID**: M1.4 / M1.6
- **Milestone**: M1 — Core Data Model
- **Description**: File watcher + remaining unit tests
- **Status**: Not Started
- **Assigned**: 2026-04-06

## Last Session Summary
- **Session 1 (2026-04-05)**: Full planning phase — SPEC.md, 3 ADRs (all APPROVED)
- **Session 2 (2026-04-06)**: M0 COMPLETE + M1.1, M1.2, M1.3, M1.5 done
  - M1.1: Task data model — Task struct, enums, validation, helpers (18 tests)
  - M1.2: Markdown I/O — Parse, Marshal, WriteFile with atomic writes (11 tests)
  - M1.5: ULID generation — NewID, NewIDAt, IDTime, MatchesPrefix (6 tests)
  - M1.3: SQLite index — schema, upsert, delete, query, filter, rebuild from files (13 tests)
  - Review feedback addressed: TODO for temp file naming, tracking docs

## Current State
- **M0 COMPLETE**, **M1.1-M1.3 + M1.5 COMPLETE**
- Module: `github.com/tesserabox/bentotask`
- Dependencies: cobra, adrg/frontmatter, yaml.v3, oklog/ulid/v2, modernc.org/sqlite
- `make test`: **48 tests** (3 CLI + 24 model + 24 store) — 0 lint issues
- Packages built:
  - `internal/model/` — Task struct, validation, helpers, ULID generation
  - `internal/store/` — Markdown I/O (atomic writes) + SQLite index (schema, CRUD, filtering, rebuild)
  - `internal/cli/` — Root Cobra command with global flags

## Next Steps (in order)
1. **M1.4: File watcher** — `fsnotify` for detecting external changes to .md files
2. **M1.6: Unit tests** — fill any remaining coverage gaps
3. **Then M2: Basic CLI** — `bt add`, `bt list`, `bt done`, etc.

## Blockers
- None

## Context for Next Agent
- Start by reading this file, then `CLAUDE.md` for orientation
- ADR-002 §6 defines the sync strategy (mtime + hash) for the file watcher
- Key library to add: `fsnotify/fsnotify` for filesystem events
- The index is a cache — RebuildIndex scans .md files and populates SQLite
- Use `make test` and `make lint` before committing
- **Remember to update tracking docs** when completing tasks
