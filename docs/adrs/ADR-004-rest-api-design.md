# ADR-004: REST API Design

**Status**: APPROVED  
**Date**: 2026-04-07  
**Decision Makers**: @jbencardino  
**Depends on**: ADR-001 (Go + chi — APPROVED), ADR-002 (Storage — APPROVED), ADR-003 (CLI — APPROVED)

---

## Context

BentoTask needs a REST API to serve the web UI (M6), enable external integrations (M7), and support the future MCP server (M9). The API exposes the same operations as the CLI through the existing `App` layer — no new business logic, just HTTP transport.

Key questions:
1. **URL structure**: How are endpoints organized?
2. **Request/response format**: Payload shapes, error format, pagination
3. **Authentication**: Is auth needed for a local-first app?
4. **Server lifecycle**: How does the API server start, embed with CLI?
5. **Real-time**: WebSocket/SSE for live updates?
6. **CORS**: How to handle SvelteKit dev server → Go API?

---

## Decisions

### 1. URL Structure: Resource-Based REST

**Base path**: `/api/v1`

All endpoints follow REST conventions: plural nouns, HTTP verbs for actions, query parameters for filtering.

#### Task Endpoints

| Method | Path | Description | App Method |
|--------|------|-------------|------------|
| `POST` | `/api/v1/tasks` | Create a task | `AddTask` |
| `GET` | `/api/v1/tasks` | List tasks (with filters) | `ListTasks` |
| `GET` | `/api/v1/tasks/:id` | Get task details | `GetTask` |
| `PATCH` | `/api/v1/tasks/:id` | Update task fields | `UpdateTask` |
| `DELETE` | `/api/v1/tasks/:id` | Delete task | `DeleteTask` |
| `POST` | `/api/v1/tasks/:id/done` | Mark task complete | `CompleteTask` |
| `GET` | `/api/v1/tasks/search?q=` | Full-text search | `SearchTasks` |

#### Link Endpoints

| Method | Path | Description | App Method |
|--------|------|-------------|------------|
| `POST` | `/api/v1/tasks/:id/links` | Create a link | `LinkTasks` |
| `DELETE` | `/api/v1/tasks/:id/links/:targetId` | Remove a link | `UnlinkTasks` |
| `GET` | `/api/v1/tasks/:id/links` | Get task links | `GetTaskLinks` |

#### Habit Endpoints

| Method | Path | Description | App Method |
|--------|------|-------------|------------|
| `POST` | `/api/v1/habits` | Create a habit | `AddHabit` |
| `GET` | `/api/v1/habits` | List habits | `ListHabits` |
| `POST` | `/api/v1/habits/:id/log` | Log completion | `LogHabit` |
| `GET` | `/api/v1/habits/:id/stats` | Get habit stats | `HabitStats` |

#### Routine Endpoints

| Method | Path | Description | App Method |
|--------|------|-------------|------------|
| `POST` | `/api/v1/routines` | Create a routine | `AddRoutine` |
| `GET` | `/api/v1/routines` | List routines | `ListRoutines` |
| `GET` | `/api/v1/routines/:id` | Get routine details | `GetTask` (type=routine) |

#### Scheduling Endpoints

| Method | Path | Description | App Method |
|--------|------|-------------|------------|
| `GET` | `/api/v1/suggest?time=60&energy=medium&context=home&count=5` | Task suggestions | `Suggest` |
| `GET` | `/api/v1/plan/today?time=480&energy=medium&context=office` | Daily plan | `PlanDay` |

#### Admin Endpoints

| Method | Path | Description | App Method |
|--------|------|-------------|------------|
| `POST` | `/api/v1/index/rebuild` | Rebuild search index | `RebuildIndex` |
| `GET` | `/api/v1/meta/tags` | List all tags | `CompleteTags` |
| `GET` | `/api/v1/meta/boxes` | List all boxes | `CompleteBoxes` |
| `GET` | `/api/v1/meta/contexts` | List all contexts | `CompleteContexts` |

---

### 2. Request/Response Format

#### Content Type

All requests and responses use `application/json`. The API does not accept `multipart/form-data` or `application/x-www-form-urlencoded`.

#### Success Responses

Single resource:
```json
{
  "id": "01JQX...",
  "title": "Buy groceries",
  "type": "one-shot",
  "status": "pending",
  "priority": "medium",
  "energy": "low",
  "tags": ["errands"],
  "contexts": ["home"],
  "created_at": "2026-04-07T10:00:00Z",
  "updated_at": "2026-04-07T10:00:00Z"
}
```

Collection:
```json
{
  "items": [...],
  "count": 42
}
```

Collections always wrap in `{ "items": [], "count": N }` — never return a bare array. This is extensible (can add `cursor`, `has_more` later for pagination).

#### Create/Update Request Bodies

Create task:
```json
{
  "title": "Buy groceries",
  "priority": "medium",
  "energy": "low",
  "duration": 30,
  "due_date": "2026-04-10",
  "tags": ["errands"],
  "contexts": ["home"],
  "box": "projects/shopping"
}
```

Update task (PATCH — only changed fields):
```json
{
  "title": "Buy groceries and snacks",
  "priority": "high"
}
```

All fields except `title` (on create) are optional.

#### Error Responses

All errors use a consistent shape:

```json
{
  "error": {
    "code": "not_found",
    "message": "task not found: 01JQX..."
  }
}
```

Error codes (map to HTTP status):

| HTTP Status | Code | When |
|-------------|------|------|
| 400 | `bad_request` | Invalid JSON, missing required field, invalid enum value |
| 404 | `not_found` | Task/habit/routine not found |
| 409 | `conflict` | Duplicate link, task already done, dependency cycle |
| 422 | `validation_error` | Model validation failure (empty title, invalid link type) |
| 500 | `internal_error` | Unexpected server error |

---

### 3. Authentication: None (Local-First)

BentoTask is a **local-first, single-user** application. The API server binds to `localhost` only. No authentication is needed in Phase 4.

If remote access is needed in the future (M7 sync, multi-device), authentication can be added via:
- API key in `Authorization: Bearer <key>` header
- Generated on first run, stored in `~/.config/bentotask/api_key`
- Or delegated to a reverse proxy (Caddy, nginx)

**Decision**: No auth for now. Bind to `127.0.0.1` only. Add `--host` and `--port` flags for configuration.

---

### 4. Server Lifecycle

#### CLI Integration

The API server is started via:
```
bt serve                    # Start API server (default: localhost:7878)
bt serve --port 9090        # Custom port
bt serve --host 0.0.0.0     # Expose to network (future)
```

The `bt serve` command:
1. Opens the `App` (data dir, SQLite index)
2. Starts the file watcher (live-reloads external edits)
3. Creates the `chi` router with all API routes
4. Serves the embedded SvelteKit build on `/` (via `go:embed`)
5. Serves the API on `/api/v1/*`
6. Listens on `127.0.0.1:7878`
7. Handles `SIGINT`/`SIGTERM` for graceful shutdown

#### Concurrency

The `App` struct is currently not goroutine-safe. For the API server:
- Wrap `App` method calls with a `sync.RWMutex` in the API handler layer
- Reads (`GET`) take a read lock
- Writes (`POST`, `PATCH`, `DELETE`) take a write lock
- This is simple and correct for a single-user local app

---

### 5. Real-Time Updates: Deferred

Server-Sent Events (SSE) or WebSocket for live updates (e.g., file watcher detects external edit → push to UI) are deferred to a later phase. The web UI will poll or refresh on user action initially.

**Reason**: SSE/WebSocket adds complexity (connection management, reconnect logic) that isn't needed for a single-user local app where the user is the only one making changes.

**Future path**: When added, use SSE on `GET /api/v1/events` (simpler than WebSocket, works with HTTP/2, auto-reconnects).

---

### 6. CORS

During development, the SvelteKit dev server runs on `localhost:5173` and the Go API on `localhost:7878`. CORS headers are needed.

**Decision**: The API server adds CORS middleware that:
- In development: allows `localhost:*` origins
- In production: not needed (SvelteKit is embedded, same-origin)

Use `go-chi/cors` middleware with sensible defaults.

---

### 7. API Package Structure

```
internal/api/
├── server.go          # Server struct, NewServer(), ListenAndServe()
├── routes.go          # Route registration (chi router setup)
├── tasks.go           # Task handlers (CRUD, done, search)
├── habits.go          # Habit handlers (create, log, stats)
├── routines.go        # Routine handlers (create, list)
├── links.go           # Link handlers (create, delete, list)
├── schedule.go        # Scheduling handlers (suggest, plan)
├── middleware.go      # CORS, logging, recovery, content-type
├── errors.go          # Error response helpers
└── server_test.go     # API integration tests (httptest)
```

The `Server` struct holds:
```go
type Server struct {
    app    *app.App
    router chi.Router
    mu     sync.RWMutex  // protects app access
}
```

Handlers follow a consistent pattern:
```go
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request body
    // 2. Validate input
    // 3. Lock, call app method, unlock
    // 4. Write JSON response (or error)
}
```

---

### 8. JSON Types: Shared in `internal/api/types.go`

The CLI already has battle-tested JSON types (`TaskJSON`, `StepJSON`, `ScheduleJSON`, `SuggestionJSON`, `PlanJSON`) and converter functions (`taskToJSON`, `indexedToJSON`, `writeJSON`, `suggestionToJSON`, `planToJSON`) in `internal/cli/json.go` and `internal/cli/schedule.go`. These are the de facto JSON contract — every `--json` flag uses them.

**Decision**: Move these types and converters to `internal/api/` so both CLI and API share a single source of truth:

| Current Location | New Location | What Moves |
|---|---|---|
| `cli/json.go` | `api/types.go` | `TaskJSON`, `StepJSON`, `ScheduleJSON`, `taskToJSON`, `indexedToJSON`, `writeJSON` |
| `cli/schedule.go` | `api/types.go` | `SuggestionJSON`, `PlanJSON`, `suggestionToJSON`, `suggestionsToJSON`, `planToJSON` |
| `cli/links.go` | *(stays)* | `LinkDisplay`, `renderLinks` — CLI rendering only, not a JSON contract |

After the move:
- `internal/cli/` imports `internal/api/` for the types (no circular dependency — `api` depends on `model`/`store`/`engine`, same as `cli`)
- All types and converters become **exported** (e.g., `TaskToJSON`, `IndexedToJSON`, `WriteJSON`)
- JSON field names, `omitempty` tags, and null-safety rules (`tags`/`contexts` always `[]` never `null`) carry over unchanged
- Integration tests in `cli/` reference the types via `api.TaskJSON` etc.

This ensures any field added to the JSON contract is automatically available to both CLI `--json` output and API responses.

---

### 9. Default Port

**Port 7878**: "BT" in phone keypad (B=2→7, T=8→8... close enough). Easy to remember, unlikely to conflict.

Configurable via `--port` flag and `BT_PORT` environment variable.

---

## Consequences

- New package `internal/api/` with `chi` router
- New CLI command `bt serve` starts the HTTP server
- JSON types shared between CLI and API (single source of truth)
- No authentication initially (localhost only)
- No real-time updates initially (polling)
- `sync.RWMutex` for concurrent access in API handlers
- CORS middleware for development
- All responses wrapped in consistent envelope (`{ "items": [], "count": N }` or `{ "error": {} }`)

---

## References

- [chi router](https://github.com/go-chi/chi)
- [go-chi/cors](https://github.com/go-chi/cors)
- [JSON API specification (inspiration)](https://jsonapi.org/)
- [Huma (Go REST framework)](https://huma.rocks/) — considered but chi is simpler and already chosen in ADR-001
- [httptest package](https://pkg.go.dev/net/http/httptest) — for API testing
