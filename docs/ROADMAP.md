# BentoTask Roadmap

> Living document tracking milestones, progress, and decisions.

## Status Legend

| Symbol | Meaning |
|--------|---------|
| :white_large_square: | Not started |
| :construction: | In progress |
| :white_check_mark: | Complete |
| :no_entry: | Blocked |
| :recycle: | Needs revision |

---

## Milestone 0: Project Bootstrap
**Goal**: Set up the project foundation, make key architecture decisions.

- [x] M0.1 — Write initial SPEC.md
- [x] M0.2 — Initialize git repository
- [x] M0.3 — ADR-001: Choose programming language & tech stack (APPROVED: Go + SvelteKit)
- [x] M0.4 — ADR-002: Choose storage format & indexing strategy (APPROVED: MD+YAML, file-per-task, SQLite index, ULID, RRULE)
- [x] M0.5 — ADR-003: Choose CLI framework & UX patterns (APPROVED: noun-verb, Charm stack)
- [x] M0.6 — Set up project scaffolding (go mod, folder structure, CLI root command)
- [x] M0.7 — Define coding standards & contribution guidelines
- [x] M0.8 — Set up testing framework & first test

---

## Milestone 1: Core Data Model (Phase 1a)
**Goal**: Tasks can be created, stored as markdown, indexed, and queried.

- [x] M1.1 — Implement task data model (struct + serialization)
- [x] M1.2 — Markdown + YAML frontmatter reader/writer
- [x] M1.3 — SQLite index: schema, create, rebuild from files
- [x] M1.4 — File watcher for external changes
- [x] M1.5 — ULID generation for task IDs
- [x] M1.6 — Unit tests for data model & storage layer

---

## Milestone 2: Basic CLI (Phase 1b)
**Goal**: Functional CLI for task CRUD operations.

- [x] M2.1 — CLI scaffolding (command structure, help, version)
- [x] M2.2 — `bt add` — create tasks with flags
- [x] M2.3 — `bt list` — list tasks with filters
- [x] M2.4 — `bt done` — mark task complete
- [x] M2.5 — `bt edit` — modify existing task
- [x] M2.6 — `bt delete` — remove task
- [x] M2.7 — `bt show <id>` — display task details
- [x] M2.8 — Tags, priority, energy level support
- [x] M2.9 — Search and filtering (by tag, priority, status, box)
- [x] M2.10 — Tab completion (bash, zsh, fish)
- [x] M2.11 — Integration tests for CLI commands
- [x] M2.12 — `--json` output mode for all commands

---

## Milestone 3: Habits & Recurrence (Phase 2a)
**Goal**: Habits can be tracked with streaks, tasks can recur.

- [x] M3.1 — Recurrence rule model (all patterns from spec)
- [x] M3.2 — Recurring task instance generation
- [x] M3.3 — Habit data model (extends task)
- [x] M3.4 — `bt habit add` / `bt habit log` / `bt habit stats`
- [x] M3.5 — Streak calculation engine
- [x] M3.6 — Habit completion history storage
- [x] M3.7 — Tests for recurrence and streak logic

---

## Milestone 4: Routines & Links (Phase 2b)
**Goal**: Tasks can be grouped into routines and linked to each other.

- [x] M4.1 — Routine data model (ordered step sequence)
- [x] M4.2 — `bt routine create` / `bt routine play`
- [x] M4.3 — Routine play mode in terminal (step-by-step, timers)
- [x] M4.4 — Task linking (depends-on, blocks, related-to)
- [x] M4.5 — Dependency validation (cycle detection)
- [x] M4.6 — `bt link` / `bt unlink` commands
- [ ] M4.7 — Tests for routines and dependency graph

---

## Milestone 5: Smart Scheduling (Phase 3)
**Goal**: The "What should I do now?" feature works.

- [ ] M5.1 — Bento Packing Algorithm implementation
- [ ] M5.2 — Urgency scoring function
- [ ] M5.3 — Energy matching logic
- [ ] M5.4 — Streak risk detection
- [ ] M5.5 — `bt now` command (suggest next task)
- [ ] M5.6 — `bt plan today` command (generate daily plan)
- [ ] M5.7 — Context support (home, office, errands)
- [ ] M5.8 — Algorithm benchmarks and tuning
- [ ] M5.9 — Tests for scheduling algorithm

---

## Milestone 6: REST API & Web UI (Phase 4)
**Goal**: Visual interface accessible via browser.

- [ ] M6.1 — REST API design & OpenAPI spec
- [ ] M6.2 — API server implementation
- [ ] M6.3 — Web UI scaffolding (SvelteKit)
- [ ] M6.4 — Inbox view
- [ ] M6.5 — Today view
- [ ] M6.6 — Calendar view
- [ ] M6.7 — Kanban view
- [ ] M6.8 — Habits dashboard (heatmaps, charts)
- [ ] M6.9 — Routine player (visual step-through)
- [ ] M6.10 — Smart mirror / display view
- [ ] M6.11 — Embed web UI in Go binary (go:embed)
- [ ] M6.12 — API & UI tests

---

## Milestone 7: Integrations (Phase 5)
**Goal**: Calendar sync, reminders, import/export.

- [ ] M7.1 — CalDAV client (Apple Calendar, Nextcloud)
- [ ] M7.2 — Google Calendar API integration
- [ ] M7.3 — Notification system (system notifications, webhooks)
- [ ] M7.4 — Import: Todoist, Taskwarrior, Notion markdown
- [ ] M7.5 — Export: iCal, JSON, CSV
- [ ] M7.6 — Git-based sync between devices

---

## Milestone 8: Knowledge Base (Phase 6)
**Goal**: Second brain — notes, documents, concept mapping.

- [ ] M8.1 — Notes system with bi-directional links
- [ ] M8.2 — Document & image attachments
- [ ] M8.3 — Knowledge graph visualization
- [ ] M8.4 — Link notes to tasks/projects
- [ ] M8.5 — Daily journal auto-generation

---

## Milestone 9: AI & Extensions (Phase 7)
**Goal**: Plugin system, MCP, AI features.

- [ ] M9.1 — Plugin/extension architecture (Wasm via wazero)
- [ ] M9.2 — MCP server implementation
- [ ] M9.3 — Natural language task creation
- [ ] M9.4 — AI-powered scheduling optimization
- [ ] M9.5 — Smart categorization suggestions

---

## Decision Log

| ID | Date | Decision | Status |
|----|------|----------|--------|
| ADR-001 | 2026-04-05 | Tech stack: Go + SvelteKit | :white_check_mark: Approved |
| ADR-002 | 2026-04-05 | Storage: MD+YAML, file-per-task, SQLite index, ULID, RRULE | :white_check_mark: Approved |
| ADR-003 | 2026-04-05 | CLI: noun-verb, Charm stack, --json, $EDITOR, completions | :white_check_mark: Approved |
