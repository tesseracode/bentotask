# Current Handoff

## Active Task
- **Task ID**: M6 Group A (API Foundation)
- **Milestone**: M6 — REST API & Web UI
- **Description**: Move JSON types to api/, build REST API server with chi, implement all endpoints
- **Status**: COMPLETE
- **Assigned**: 2026-04-07

## Last Session Summary
- **Session 1–3 (2026-04-05–06)**: M0 + M1 + M2.1–M2.9
- **Session 4 (2026-04-06)**: M2.10–M2.12 — Tab completions, integration tests, --json
- **Session 5 (2026-04-07)**: M3 COMPLETE — Habits & Recurrence
- **Session 6 (2026-04-07)**: Bug fixes from M3 review + M4 Group A (Routines)
- **Session 7 (2026-04-07)**: M4 Group B (Linking) + fixes from reviews
- **Session 8 (2026-04-07)**: M4 closure — fixed flaky recurrence tests, planned M5
- **Session 9 (2026-04-07)**: M5 COMPLETE — Smart Scheduling
- **Session 10 (2026-04-07)**: M5 fixes (lint, blocks bug, age_boost snap) + ADR-004 written and approved
- **Session 11 (2026-04-07)**: M6 Group A COMPLETE — REST API Foundation
  - Step 1: Moved JSON types (TaskJSON, SuggestionJSON, PlanJSON, etc.) from cli/ to api/types.go
  - Step 2: Built full REST API server — 22 endpoints across 9 files, chi router, CORS, sync.RWMutex
  - Step 3: Added bt serve CLI command (--port, --host, graceful shutdown)
  - Step 4: 16 httptest-based API integration tests covering all endpoints + error cases

## Current State
- **M0–M5 COMPLETE**, **M6.1 + M6.2 + M6.12 COMPLETE**
- Module: `github.com/tesserabox/bentotask`
- `make test`: **315 tests** — 0 lint issues, 0 vet issues
- Binary: `make build` — up to date
- ADR-004 (REST API Design) — APPROVED and IMPLEMENTED

## M6 Group A Implementation Summary

### Files Created
- `internal/api/types.go` — Shared JSON types + converters (moved from cli/)
- `internal/api/server.go` — Server struct, NewServer, ListenAndServe, Shutdown
- `internal/api/routes.go` — chi router with 22 endpoints under /api/v1
- `internal/api/tasks.go` — Task CRUD + done + search (7 handlers)
- `internal/api/habits.go` — Habit create + list + log + stats (4 handlers)
- `internal/api/routines.go` — Routine create + list + get (3 handlers)
- `internal/api/links.go` — Link create + delete + get (3 handlers)
- `internal/api/schedule.go` — Suggest + plan + rebuild + meta (5 handlers)
- `internal/api/middleware.go` — CORS, recovery, request logging, JSON content-type
- `internal/api/errors.go` — respondJSON/respondError helpers, ADR-004 error envelope
- `internal/api/server_test.go` — 16 integration tests
- `internal/cli/serve.go` — bt serve command

### Files Modified
- `internal/cli/json.go` — Now contains type aliases + thin wrappers to api/
- `internal/cli/schedule.go` — Removed moved type/converter defs
- `internal/cli/integration_test.go` — Added serveCmd to resetFlags()

## Next Steps
- M6.3: Web UI scaffolding (SvelteKit)
- M6.4–M6.10: Web UI views

## Blockers
- None

## Context for Next Agent
- API is fully functional at /api/v1 — start bt serve and hit endpoints
- JSON types are in api/types.go — shared between CLI --json and REST API
- CORS is configured for localhost:5173 (SvelteKit dev server)
- Dependencies: chi/v5 and go-chi/cors are in go.mod
- All 22 endpoints match ADR-004 specification exactly
- Collection responses use { "items": [], "count": N } envelope
- Error responses use { "error": { "code": "...", "message": "..." } }
- sync.RWMutex wraps all App calls — reads=RLock, writes=Lock
