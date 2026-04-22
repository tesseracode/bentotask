# Current Handoff

## Project Status: BentoTask
- **Date**: 2026-04-21
- **Commits**: 75
- **Tests**: 356 (0 lint, 0 vet issues)
- **Packages**: 12 Go packages + SvelteKit web UI
- **Binary**: Single 18MB binary via `go:embed` — `make build`

## Completed Milestones

| Milestone | Tasks | What was built |
|-----------|-------|---------------|
| M0 Bootstrap | 8/8 ✅ | Spec, 4 ADRs, CI, lint, scaffolding |
| M1 Data Model | 6/6 ✅ | Markdown+YAML storage, SQLite index, file watcher, ULIDs |
| M2 CLI | 12/12 ✅ | Full `bt` CLI — CRUD, search, filters, edit, --json, completions |
| M3 Habits | 7/7 ✅ | RRULE engine, streaks, completions, `bt habit` commands |
| M4 Routines & Links | 7/7 ✅ | Step sequences, play mode, depends-on/blocks with cycle detection |
| M5 Smart Scheduling | 9/9 ✅ | Bento Packing Algorithm, `bt now`, `bt plan today`, benchmarks |
| M6 REST API & Web UI | 12/12 ✅ | 22-endpoint API, 9 web views, `go:embed` single binary |
| M7 Integrations | 4/4 ✅ | Todoist/Taskwarrior import, JSON/CSV export, Notion, Obsidian |
| M8 Desktop (partial) | 3/3 ✅ | `bt serve --open`, cross-compilation, CI release pipeline |
| M10 AI (partial) | 4/4 ✅ | MCP server (20 tools, 5 resources, 4 prompts), NLP parser |

## Architecture Overview

```
cmd/bt/              CLI entrypoint
internal/
├── api/             REST API (chi router, 22 endpoints, go:embed static files)
├── app/             Business logic layer (29 methods — the core)
├── cli/             Cobra commands (task, habit, routine, link, export, import, obsidian, notion, serve, mcp)
├── engine/          Scoring engine + Bento Packing Algorithm
├── habit/           Streak calculation, completion tracking
├── mcp/             MCP server (JSON-RPC over stdio, 20 tools, 5 resources, 4 prompts)
├── model/           Data structures (Task, Link, HabitFrequency, RoutineStep)
├── nlp/             Natural language parser (dates, priority, tags, duration)
├── notion/          Notion API client + import
├── recurrence/      RRULE engine (daily, weekly, custom)
├── store/           Markdown I/O + SQLite index + file watcher
└── style/           Terminal styling (lipgloss)
web/                 SvelteKit SPA (9 routes, Bento Alt theme)
docs/
├── adrs/            4 ADRs (tech stack, storage, CLI, REST API)
├── milestones/      M0, M7, M8 detail docs
├── handoff/         This file + HISTORY.md
├── supervisor/      LOG.md (review log)
├── ROADMAP.md       Master milestone tracking
├── STYLE-GUIDE.md   Visual identity (Bento Alt dark + Clay Alt light archived)
└── DEPLOYMENT.md    4 deployment modes (single binary, split, SSR, headless)
```

## Web Views (9)

📥 Inbox · 📅 Today · 📆 Calendar · 📋 Kanban · 🔥 Habits · 🔄 Routines · 🪞 Mirror · 🎨 Design Explorer

## CLI Commands

```
bt add/list/show/edit/done/delete   Task CRUD
bt search                           Full-text search (FTS5)
bt link/unlink                      Task dependencies
bt habit add/log/stats/list         Habit tracking
bt routine create/list/play         Routine management
bt now                              Smart suggestions
bt plan today                       Day planning
bt export json/csv                  Export tasks
bt import todoist/taskwarrior       Import from other tools
bt notion import                    Import from Notion
bt obsidian init                    Set up Obsidian vault
bt serve [--open]                   Start API server + web UI
bt mcp                              Start MCP server for AI assistants
```

## Key Files for New Agents

| Need | File |
|------|------|
| Project rules & structure | `CLAUDE.md` |
| Agent roles & protocols | `AGENTS.md` |
| What to build next | `docs/ROADMAP.md` |
| How things connect | `docs/adrs/ADR-*.md` |
| Visual identity | `docs/STYLE-GUIDE.md` |
| Deployment options | `docs/DEPLOYMENT.md` |
| M7 integration details | `docs/milestones/M7-integrations-detail.md` |
| M8 desktop details | `docs/milestones/M8-desktop-distribution.md` |
| All App methods | `internal/app/app.go` (29 methods — the entire business logic) |
| API endpoints | `internal/api/routes.go` (22 routes) |
| MCP tools | `internal/mcp/tools.go` (20 tools) |
| Theme CSS | `web/src/lib/theme.css` |

## Remaining Milestones

### M8 Desktop — Remaining (Parked)
- M8.2 — Wails native desktop app
- M8.3–M8.5 — Platform installers (macOS .dmg, Windows .exe, Linux AppImage)
- See `docs/milestones/M8-desktop-distribution.md` for details
- **Note**: `bt serve --open` already provides a desktop-like experience

### M9 Knowledge Base — Not Started
- M9.1 — Notes system with bi-directional links
- M9.2 — Document & image attachments
- M9.3 — Knowledge graph visualization
- M9.4 — Link notes to tasks/projects
- M9.5 — Daily journal auto-generation

### M10 AI — Remaining
- M10.1 — Plugin architecture (Wasm via wazero)
- M10.4 — AI-powered scheduling optimization
- M10.5 — Smart categorization suggestions
- M10.6 — MCP resources enhancements
- M10.7 — MCP prompts enhancements

### M11 Calendar Sync & Notifications — Deferred
- M11.1 — CalDAV client
- M11.2 — Google Calendar (OAuth2)
- M11.3 — Notifications (system, webhooks, ntfy.sh)
- M11.4 — Git-based sync
- M11.5 — iCal export

## Known Design Decisions

- **No auth** on API (localhost only) — add API key via `Authorization: Bearer` when needed
- **`sync.RWMutex`** in API server for concurrent access (reads share, writes exclusive)
- **String-based error classification** in API (isNotFound/isConflict match on error messages) — works but fragile; sentinel errors would be cleaner
- **Habit `max_per_period=0`** means unlimited (backward compatible default)
- **Mirror view** uses hardcoded colors (intentional — own visual identity, not themed)
- **Dark/light mode toggle** not yet implemented — Clay Alt light theme CSS vars are archived in `theme.css`
- **Kanban** has no drag-and-drop (uses status dropdowns) — MVP decision

## For Supervisor Agents

Review checklist for any PR/commit:
1. `go build ./...` — compiles
2. `go vet ./...` — no issues
3. `golangci-lint run ./...` — 0 issues
4. `go test ./...` — 356+ tests pass
5. `cd web && npx svelte-check` — 0 errors (if web files changed)
6. No hardcoded hex colors in production `<style>` blocks (use `var(--*)`)
7. All new CLI commands support `--json` and `--quiet`
8. Error handling: `error = ''` at top of every async op in Svelte
9. MCP protocol: notifications (no ID) get no response

## Blockers
- None
