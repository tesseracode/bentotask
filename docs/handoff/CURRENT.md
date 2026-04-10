# Current Handoff

## Active Task
- **Task ID**: M6 Complete
- **Milestone**: M6 — REST API & Web UI
- **Description**: All 12 M6 items complete
- **Status**: COMPLETE
- **Assigned**: 2026-04-10

## Last Session Summary
- **Sessions 1–14**: M0–M5 + M6 Groups A–D (API, views, themes, go:embed, mirror, routines)
- **Session 15**: Final M6 items — routine editing + Calendar + Kanban
  - Part 1: Routine CRUD — edit mode (steps, title, priority), delete with confirm,
    create routine from web UI with inline step builder
  - Part 2: Routine player UX — tooltips on all buttons, play mode hint text,
    keyboard shortcuts (Enter/→ = Next, S = Skip, Esc = Stop)
  - Part 3: Fix mirror — parallel habit stats loading via Promise.all
  - Part 4: Calendar view (M6.6) — monthly grid, task pills by due date,
    month navigation, today highlight
  - Part 4: Kanban view (M6.7) — columns by status, card-per-task,
    status dropdown to move between columns, filters out routines/habits

## Current State
- **M0–M6 ALL COMPLETE** (12/12 M6 items)
- Go backend: 315 tests, 0 lint issues
- Web UI: 0 svelte-check errors, builds with adapter-static
- Single binary: `make build` → `./bt serve` serves everything
- 8 web views: Inbox, Today, Calendar, Kanban, Habits, Routines, Mirror, Design
- Theme: Bento Alt (dark) via CSS custom properties
- API: 22 endpoints, steps/schedule updatable via PATCH

## M6 Summary — All 12 Items

| Item | Description | Status |
|------|-------------|--------|
| M6.1 | REST API design (ADR-004) | ✅ |
| M6.2 | API server (chi, 22 endpoints) | ✅ |
| M6.3 | SvelteKit scaffolding | ✅ |
| M6.4 | Inbox (filters, expand, edit) | ✅ |
| M6.5 | Today (suggest, plan, controls) | ✅ |
| M6.6 | Calendar (monthly grid) | ✅ |
| M6.7 | Kanban (status columns) | ✅ |
| M6.8 | Habits dashboard (streaks, stats) | ✅ |
| M6.9 | Routine player (timers, keyboard) | ✅ |
| M6.10 | Smart mirror (full-screen display) | ✅ |
| M6.11 | go:embed (single binary) | ✅ |
| M6.12 | API integration tests (16) | ✅ |

## Next Steps
- M7: Sync & external integrations (CalDAV, Google Calendar)
- Future: Dark/light mode toggle, design polish, drag-and-drop Kanban
- Future: PWA support, notifications

## Blockers
- None
