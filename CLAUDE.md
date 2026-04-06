# CLAUDE.md — Agent Orientation for BentoTask

> This file is the entry point for any AI agent working on this project.
> Read this FIRST. It tells you where everything is and how to operate.

---

## What is BentoTask?

A local-first task, habit, and routine manager with smart scheduling, built in **Go** with a **SvelteKit** web UI. Think "bento box" — tasks nest inside tasks, habits group into routines, and the system intelligently packs your available time.

## Quick Orientation

| What you need | Where to find it |
|--------------|-----------------|
| **Full specification** | `SPEC.md` (living doc, may evolve) |
| **Original frozen spec** | `SPEC.original.md` (gitignored, may not be present) |
| **Roadmap & milestones** | `docs/ROADMAP.md` |
| **Architecture decisions** | `docs/adrs/ADR-*.md` |
| **Current milestone details** | `docs/milestones/M*-*.md` |
| **Agent handoff state** | `docs/handoff/CURRENT.md` |
| **Supervisor log** | `docs/supervisor/LOG.md` |
| **Agent instructions** | `AGENTS.md` (roles, protocols, rules) |

## Before You Start Working

1. **Read `docs/handoff/CURRENT.md`** — this tells you what task is active, what was last done, and what to do next
2. **Read the relevant milestone** in `docs/milestones/` — don't read the full SPEC unless needed
3. **Read the relevant ADR** if your task involves an architecture decision
4. **Check `docs/supervisor/LOG.md`** for any notes from the supervisor about your task

## Tech Stack (ADR-001 — APPROVED)

- **Language**: Go 1.22+
- **CLI**: `cobra` + `bubbletea` (Charm ecosystem)
- **Web UI**: SvelteKit (embedded via `go:embed`)
- **API**: `chi` router + `net/http`
- **Storage**: Markdown + YAML frontmatter (source of truth) + SQLite index (cache)
- **SQLite**: `modernc.org/sqlite` (pure Go, no CGO)
- **IDs**: ULID (`oklog/ulid`)
- **Testing**: `testing` + `testify`

## Project Structure (target)

```
bentotask/
├── cmd/bt/              # CLI entrypoint
├── internal/
│   ├── model/           # Data structures
│   ├── store/           # Markdown I/O + SQLite index
│   ├── engine/          # Scheduling algorithm
│   ├── calendar/        # CalDAV + Google Calendar
│   ├── routine/         # Routine engine
│   ├── graph/           # Task dependency graph
│   └── api/             # REST API server
├── web/                 # SvelteKit SPA
├── plugins/             # Future: Wasm plugins
├── docs/
│   ├── adrs/            # Architecture Decision Records
│   ├── milestones/      # Milestone details
│   ├── handoff/         # Agent handoff state
│   └── supervisor/      # Supervisor review log
├── SPEC.md              # Living specification
├── CLAUDE.md            # This file
├── AGENTS.md            # Agent roles & protocols
└── go.mod
```

## Rules for Agents

1. **Never modify CLAUDE.md or AGENTS.md** without explicit human approval
2. **Always update `docs/handoff/CURRENT.md`** when finishing a work session
3. **Never skip tests** — write tests for every feature
4. **Commit often** with clear messages
5. **If blocked**, document the blocker in the handoff file and stop
6. **Don't over-architect** — prefer simple, working code over clever abstractions
7. **Follow Go conventions** — `gofmt`, `golint`, idiomatic error handling
8. **Ask before making ADR-level decisions** — if something isn't covered by an existing ADR, flag it
