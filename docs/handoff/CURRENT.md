# Current Handoff

## Active Task
- **Task ID**: M7 Integrations (partial)
- **Milestone**: M7 — Integrations
- **Description**: Obsidian, Export, Import, Notion
- **Status**: 4/8 COMPLETE
- **Assigned**: 2026-04-18

## Last Session Summary
- **Sessions 1–15**: M0–M6 ALL COMPLETE
- **Session 16**: M7 partial — Obsidian, Export, Import, Notion
  - Part 1 (M7.8): Obsidian vault integration — bt obsidian init + wikilink rendering in web UI
  - Part 2 (M7.5): Export — bt export json, bt export csv with filters and file output
  - Part 3 (M7.4): Import — bt import todoist (CSV), bt import taskwarrior (JSON)
  - Part 4 (M7.7): Notion integration — bt notion import with API client + property mapping + mock tests

## Current State
- **M0–M6 ALL COMPLETE**
- **M7: 4/8 complete** (M7.4, M7.5, M7.7, M7.8)
- **Deferred M7 items**: M7.1 (CalDAV), M7.2 (Google Calendar), M7.3 (Notifications), M7.6 (Git sync)
- Go backend: 325+ tests, 0 lint issues
- Web UI: 0 svelte-check errors
- New packages: internal/notion/ (Notion API client + import)
- New CLI commands: bt obsidian init, bt export json/csv, bt import todoist/taskwarrior, bt notion import

## M7 Status

| Item | Description | Status |
|------|-------------|--------|
| M7.1 | CalDAV (Apple/Nextcloud) | ⏳ Deferred |
| M7.2 | Google Calendar | ⏳ Deferred |
| M7.3 | Notifications | ⏳ Deferred |
| M7.4 | Import: Todoist + Taskwarrior | ✅ |
| M7.5 | Export: JSON + CSV | ✅ |
| M7.6 | Git-based sync | ⏳ Deferred |
| M7.7 | Notion integration | ✅ |
| M7.8 | Obsidian vault integration | ✅ |

## Next Steps
- M7.1–M7.3, M7.6: CalDAV, Google Calendar, notifications, git sync (more infrastructure needed)
- M8: Desktop app & distribution (Wails, cross-compilation, installers)
- Future: dark/light mode toggle, PWA, drag-and-drop Kanban

## Blockers
- None
