# Supervisor Log

> Review entries for completed tasks. Newest first.

---

## [2026-04-07] Review: Bug Fix — ListTasks/Search Not Loading Tags/Contexts

**Reviewed by**: Supervisor Agent
**Commit**: `a143c78` — Fix ListTasks/Search not loading tags and contexts from junction tables
**Origin**: Bug found during M2.10-12 review (see below)

**Verification**:
- [x] `go build ./...` — compiles cleanly
- [x] `go vet ./...` — 0 issues
- [x] `golangci-lint run ./...` — 0 issues
- [x] `make test` — **116 tests PASS** (3 new: 2 store-level regression + 1 integration)
- [x] Smoke: `bt add --tag work --tag urgent -c office --json` then `bt list --json` — tags and contexts now appear correctly in list output

**Fix approach**: Promoted `collectTasks()` from a free function to an `Index` method so it can call `loadTags()`/`loadContexts()` per result. This applies to `ListTasks`, `FindByPrefix`, and `Search` — all three paths now return complete tag/context data.

**Also fixed**: Removed inline `file_path` computation in `runAdd` JSON output, replaced with `a.GetTask(task.ID)` to match other commands (eliminates the minor divergence noted in previous review).

**Tests added**:
- Store: `TestListTasksFilterByTag` — now asserts `Tags` field populated on result
- Store: `TestListTasksFilterByContext` — now asserts `Contexts` field populated on result
- Integration: `TestIntegrationJSONListShowsTags` — end-to-end add-with-tags then list-json verification

**Verdict**: **APPROVED** ✅ — clean, minimal fix with proper regression tests at both layers.

---

## [2026-04-07] Review: M2.10 + M2.11 + M2.12 — Tab Completions, Integration Tests, JSON Output

**Reviewed by**: Supervisor Agent
**Commit**: `b9da759` — M2.10+M2.11+M2.12: Add tab completions, integration tests, and --json output
**Handoff claimed**: M2 COMPLETE — all 12 tasks done

**Verification**:
- [x] `go build ./...` — compiles cleanly
- [x] `go vet ./...` — 0 issues
- [x] `golangci-lint run ./...` — 0 issues
- [x] `make test` — **113 tests PASS** across 5 packages (25 new integration tests)
- [x] Smoke: `bt add --json` — valid JSON, correct fields, tags/contexts never null
- [x] Smoke: `bt list --json` — valid JSON array
- [x] Smoke: `bt done --json` — status=done, completed_at set
- [x] Smoke: `bt search --json` — valid JSON array
- [x] Smoke: `bt index rebuild --json` — `{"indexed": N}`
- [x] Smoke: `bt completion --help` — shell completion subcommand available
- [x] Tracking docs updated: ROADMAP (M2.10-12 checked), handoff (M3 next), session 4 summary

### M2.10: Tab Completions — `internal/cli/completions.go` (165 lines, new)

- [x] `registerCompletions()` called in `init()` — wires up all commands
- [x] **Task ID completions**: `completeTaskIDs` → `App.CompleteTasks()` → `Index.ListTasks(nil)` — returns `ID\tTitle` format, filters out done/cancelled tasks
- [x] **Dynamic flag completions**: `--tag` → `App.CompleteTags()` → `Index.DistinctTags()`, `--box` → `App.CompleteBoxes()` → `Index.DistinctBoxes()`
- [x] **Static enum completions**: `--status` (6 values with descriptions), `--priority` (5 values), `--energy` (3 values), `--context` (4 fixed values)
- [x] All completion functions return `cobra.ShellCompDirectiveNoFileComp` (no file fallback)
- [x] Completions registered for paired commands: `taskAddCmd`+`addCmd`, `taskListCmd`+`listCmd`
- [x] Edit command gets own `registerEditCompletions` with all enum flags

**New index methods** (`index.go`, 34 new lines):
- [x] `DistinctTags()`, `DistinctBoxes()`, `DistinctContexts()` — all use shared `distinctStrings()` helper
- [x] `DistinctBoxes` correctly filters `NULL` and empty strings

**New app methods** (`app.go`, 31 new lines):
- [x] `CompleteTasks()`, `CompleteTags()`, `CompleteBoxes()`, `CompleteContexts()`

### M2.11: Integration Tests — `internal/cli/integration_test.go` (601 lines, new)

- [x] **25 end-to-end tests** covering full CLI flow through real Cobra execution
- [x] `executeCmdInDir()` test helper — sets `--data-dir`, captures stdout, resets flags
- [x] `resetFlags()` + `resetFlag()` — prevents Cobra global state leaks between tests (handles `StringSlice` via `pflag.SliceValue.Replace`)
- [x] **CRUD lifecycle tests**: AddAndList, AddAndShow, AddAndDone, AddAndDelete, EditWithFlags
- [x] **Search tests**: Search (finds match), SearchNoResults
- [x] **Filter tests**: ListFilters (--tag, --priority — positive and negative assertions)
- [x] **Output mode tests**: QuietMode (ULID length check), JSONAdd, JSONList, JSONShow, JSONSearch, JSONDone, JSONEmptyList, JSONNullSafety
- [x] **Edge cases**: PrefixMatch (8-char prefix), NotFound (error returned), DoneAlreadyComplete (double-done error)
- [x] **ADR-003 compliance**: NounVerb (`bt task add`), TaskAlias (`bt t add`), AddWithDueDate (auto-promotes to `dated` type)
- [x] **JSON integrity**: Parses output with `json.Unmarshal`, verifies fields, checks `tags` is `[]` not `null`

### M2.12: JSON Output — `internal/cli/json.go` (130 lines, new)

- [x] `TaskJSON` struct — proper `json:"field_name"` tags, `omitempty` on optional fields
- [x] `Tags`/`Contexts` fields: `[]string` without `omitempty` — enforces `[]` never `null`
- [x] `taskToJSON()` — converts `model.Task` + `relPath` → `TaskJSON`, nil-safe slice init
- [x] `indexedToJSON()` — converts `store.IndexedTask` → `TaskJSON`, handles `*string` pointers
- [x] `writeJSON()` — `json.NewEncoder` with `SetIndent("", "  ")` for readable output
- [x] `isJSON(cmd)` helper in `commands.go` — reads global `--json` flag

**JSON integrated into all commands** (`commands.go`, 60 new lines):
- [x] `runAdd` — returns single `TaskJSON`
- [x] `runList` — returns `[]TaskJSON` array
- [x] `taskShowCmd` — returns single `TaskJSON` with body
- [x] `runDone` — returns single `TaskJSON` with status=done + completed_at
- [x] `editWithFlags` / `editWithEditor` — returns single `TaskJSON`
- [x] `taskDeleteCmd` — returns single `TaskJSON` (empty file_path)
- [x] `searchCmd` — returns `[]TaskJSON` array
- [x] `indexRebuildCmd` — returns `{"indexed": N}`

### Convention & ADR Compliance

- [x] Commit message format: `M2.10+M2.11+M2.12:` prefix — matches AGENTS.md convention
- [x] Error wrapping with `%w` — consistent throughout new methods
- [x] `defer func() { _ = rows.Close() }()` — resource cleanup pattern maintained
- [x] `defer func() { _ = a.Close() }()` — app cleanup in all commands
- [x] ADR-003 §3 output modes: text (default), JSON (`--json`), quiet (`--quiet`) — all implemented
- [x] ADR-003 §6 completions: dynamic task IDs with `ID\tTitle` format, dynamic tags/boxes, static enums
- [x] `cmd.Printf` / `cmd.Println` — uses Cobra output writers (not `fmt.Println`), testable
- [x] Test naming: `TestIntegration*` prefix — clear integration test identification

### Issues Found

**🐛 Bug: `ListTasks`/`Search`/`FindByPrefix` don't load tags or contexts from junction tables**

The `collectTasks()` scanner only reads the 16 columns from the `tasks` table. Tags and contexts live in `task_tags`/`task_contexts` junction tables and are only loaded in `GetTask()` (which calls `loadTags`/`loadContexts`).

This means:
- `bt list --json` returns `"tags": []` for all tasks (even those with tags)
- `bt search --json` same issue
- `bt list` (styled) also can't show tags properly — though it tries (line 265-271 in commands.go)

**This is a pre-existing bug from M1.3**, not introduced in this commit. The impact was hidden until JSON output made the data visible. The styled `list` command happened to hide it because the empty tag slice just meant no tags were shown.

**Severity**: Medium — affects data completeness in list/search views but not in show/edit/done.
**Fix**: Either (a) add per-task tag/context loading in `collectTasks`, or (b) use a LEFT JOIN to `task_tags`/`task_contexts` in the ListTasks query, or (c) batch-load tags/contexts for all returned tasks in one query.

**⚠️ Minor: `bt add --json` computes `file_path` differently than stored**

In `runAdd` (line 146-149), the JSON path is computed inline:
```go
relPath := "inbox/" + task.ID + ".md"
if opts.Box != "" { relPath = opts.Box + "/" + task.ID + ".md" }
```
But the actual `taskFilePath()` helper in `app.go` uses `filepath.Join()` which is platform-aware. These will produce the same result on Unix but could diverge on Windows. Minor since BentoTask is Unix-focused, but worth noting.

**Verdict**: **APPROVED WITH NOTES** ✅

M2 is complete. All 12 milestones tasks pass verification. The tag/context loading bug is pre-existing and doesn't block M2 closure, but should be fixed early in M3 or as a standalone fix before M3 work begins.

**Recommended actions**:
1. Fix the tags/contexts loading bug in `ListTasks`/`Search`/`FindByPrefix` (ideally before M3 starts)
2. Begin M3: Habits & Recurrence (M3.1: RRULE model)

---

## [2026-04-06] Review: M2.8 + M2.9 — Styled Output & Full-Text Search (UNCOMMITTED)

**Reviewed by**: Supervisor Agent  
**Status**: Work done but NOT committed. Agent lost context mid-task. Supervisor fixed a bug and verified.

**Verification**:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` no issues
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `make test` — **78/78 PASS** (96 including subtests)
- [x] Functional smoke tests: styled list, search by title, search by body, no-results case

### M2.8: Styled Output — `internal/style/` (new package)

**`internal/style/style.go` (150 lines)**:
- [x] Priority colors: urgent (red), high (yellow), medium (blue), low (gray)
- [x] Status icons + colors: ✓ done (green), ● active (cyan), ○ pending (default), ✗ blocked (red), ◌ paused (dim), ⊘ cancelled (dim)
- [x] Energy indicators: ⚡ high, ~ medium, · low
- [x] Tags: `#cyan` styling
- [x] General helpers: `Success()`, `ErrorMsg()`, `Dim()`, `Bold()`, `Header()`
- [x] Uses lipgloss with 256-color ANSI codes (adaptive to terminal)
- [x] Package comment claims NO_COLOR/--no-color/piped output are auto-handled

**CLI integration**:
- [x] `bt list` — bold headers, styled status/priority/tags, dimmed IDs and metadata
- [x] `bt show` — bold title, styled status/priority/energy, dimmed timestamps/file path, styled tags
- [x] All success messages (Created/Updated/Completed/Deleted/Rebuilt) use `style.Success()`
- [x] Search results use styled output

### M2.9: Full-Text Search

**Schema** (`schema.go`):
- [x] `tasks_fts` — FTS5 virtual table with `id UNINDEXED`, `title`, `body`
- [x] Standalone (not content-synced) — populated manually

**Index** (`index.go`):
- [x] `UpsertTask` — now deletes+inserts FTS entry within the same transaction
- [x] `DeleteTask` — clears FTS entry before main row
- [x] `Search(query)` — FTS5 MATCH with JOIN to tasks table, ordered by rank
- [x] **Bug found & fixed**: `RebuildIndex` was NOT clearing `tasks_fts` before rebuild — orphan FTS rows would survive for deleted tasks. Fixed by adding `"tasks_fts"` to the clear loop.

**App** (`app.go`):
- [x] `SearchTasks(query)` — validates non-empty query, delegates to index

**CLI** (`commands.go`):
- [x] `bt search <query>` — top-level command, multi-word queries supported
- [x] Quiet mode outputs IDs only
- [x] Styled results: dimmed ID, styled status, title, priority, tags

**Test Coverage — 7 new tests**:
- [x] Store: `TestSearchByTitle`, `TestSearchByBody`, `TestSearchNoResults`, `TestSearchAfterDelete`, `TestSearchAfterUpdate`
- [x] App: `TestSearchTasks`, `TestSearchTasksEmptyQuery`

### Convention & Style Audit (context-loss check)

| Convention | Status | Notes |
|------------|--------|-------|
| Commit message format (`M<n>.<m>: ...`) | ⚠️ NOT COMMITTED YET | Need to commit with proper message |
| Package comments | ✅ | `style.go` has proper doc comment |
| Error wrapping with `%w` | ✅ | All new errors use `fmt.Errorf("...: %w", err)` |
| `cmd.Printf` (not `fmt.Printf`) | ✅ | All CLI output uses `cmd.Printf`/`cmd.Println` |
| `defer rows.Close()` / `defer a.Close()` | ✅ | All resource cleanup in place |
| Test helpers (`openTestIndex`, `makeTestTask`) | ✅ | Reused consistently |
| Tracking docs updated | ✅ | Roadmap (M2.8+M2.9 ✅), handoff (M2.10-12 assigned) |

### Issues Found

| Severity | Issue | Status |
|----------|-------|--------|
| 🐛 **BUG** | `RebuildIndex` didn't clear `tasks_fts` — orphan FTS rows | **Fixed by supervisor** |
| ⚠️ **Gap** | `--no-color` flag exists but isn't wired to style package. Lipgloss respects `NO_COLOR` env var automatically, but the CLI flag is disconnected. | Noted for later |
| ⚠️ **Gap** | `internal/style/` has no tests | Acceptable for a pure-presentation package, but could add basic tests later |
| ⚠️ **Gap** | Changes are **uncommitted** — all M2.8+M2.9 work + supervisor log entries sitting in working tree | Needs commit |
| ℹ️ **Minor** | `bt search` is top-level only (no `bt task search`) — inconsistent with other commands which have both forms. Fine for UX but worth noting. |

**Verdict**: **APPROVED** ✅ (after FTS bug fix)

**Next steps**: Commit this work, then M2.10–M2.12 or jump to M3.

---

## [2026-04-06] Review: M2.5 — bt edit (Commit 07cb8df)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M2.5 complete

**Verification**:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` no issues
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `make test` — **69/69 PASS** (4 new app tests + previous 65)
- [x] Functional smoke test: flag-based edit works (title, priority, status, tags)
- [x] Roadmap updated: M2.5 ✅, M2.12 (`--json`) added as new task

**App Layer — 3 new methods (58 lines)**:
- [x] `UpdateTask(id, apply func(*Task))` — apply-function pattern, auto-updates timestamp, validates before saving. Rolls back if validation fails.
- [x] `EditTaskFile(id)` → returns absolute path for `$EDITOR`
- [x] `ReloadTask(id)` — re-reads from disk after external edit, re-indexes

**CLI — `bt task edit` (162 lines)**:
- [x] **Flag mode**: `bt task edit <id> --title "new" -p high` — applies directly via `UpdateTask`
- [x] **Editor mode**: `bt task edit <id>` (no flags) — opens `$EDITOR`, reloads on close
- [x] Editor chain: `$EDITOR → $VISUAL → vi` per ADR-003 §5
- [x] Handles editor commands with args (e.g., `code --wait`) via `strings.Fields`
- [x] `cmd.Flags().Changed()` used correctly to detect which flags were explicitly set
- [x] All fields editable: title, priority, energy, duration, due, due-start, due-end, tag, context, box, status
- [x] Quiet mode supported

**Bug fix included**:
- [x] `bt show` now displays `CompletedAt` as "Done: ..." timestamp

**Test Coverage — 4 new tests**:
- [x] `TestUpdateTask` — modifies title/priority/tags, verifies persistence
- [x] `TestUpdateTaskValidation` — empty title rejected, original unchanged
- [x] `TestEditTaskFile` — returns absolute .md path
- [x] `TestReloadTask` — simulates external edit via `store.WriteFile`, verifies reload picks up changes

**ADR-003 §5 Compliance**:
- [x] Flag-based quick edits ✅
- [x] `$EDITOR` integration ✅
- [x] Editor fallback chain (`$EDITOR → $VISUAL → vi`) ✅

**Notes**:
- The `UpdateTask` apply-function pattern is elegant — lets the CLI pass a closure that modifies specific fields, while the App layer handles timestamp updates, validation, and persistence. Keeps CLI thin.
- Validation-before-save in `UpdateTask` is good defensive design — rejects invalid edits without corrupting the file.
- `_ = task` in `ReloadTask` is a minor code smell (fetches original task just to get relPath, discards the task). Could refactor `GetTask` to return relPath separately, but it's fine for now.
- M2.12 (`--json` output) was added to the roadmap — good proactive tracking of a known gap.

**Verdict**: **APPROVED** ✅

---

## [2026-04-06] Review: M2.1–M2.4+M2.6+M2.7 — Core CLI Commands (Commit e4af3b1)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M2.1, M2.2, M2.3, M2.4, M2.6, M2.7 complete (6 tasks in 1 commit)

**Verification**:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` no issues
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `make test` — **65/65 PASS** (11 new app tests + previous 54)
- [x] Tracking docs updated: roadmap (6 tasks ✅) + handoff (M2.5/M2.8-11 assigned)

**Architecture Review — `internal/app/` (new package, 259 lines)**:
- [x] Clean separation: `App` struct wraps `DataDir` + `Index`, provides high-level CRUD
- [x] `Open(dataDir)` — creates dirs (data, inbox, .bentotask), opens SQLite index
- [x] `AddTask(title, opts)` — creates task with ULID, applies options, auto-detects type (one-shot→dated if due_date, one-shot→ranged if due_start+due_end), validates, writes file + indexes
- [x] `GetTask(idOrPrefix)` — exact match first, then prefix match; errors on 0 or ambiguous matches
- [x] `CompleteTask(idOrPrefix)` — sets status=done, timestamp, re-saves to disk + index; errors if already done
- [x] `DeleteTask(idOrPrefix)` — removes file from disk + index entry
- [x] `ListTasks(*TaskFilter)` — delegates to index
- [x] `RebuildIndex()` — delegates to index
- [x] `taskFilePath()` — routes to `inbox/` or `box/` directory based on task.Box
- [x] `TaskOptions` struct — clean option pattern for AddTask

**CLI Review — `internal/cli/commands.go` (401 lines)**:
- [x] **Noun-verb structure** per ADR-003: `bt task {add,list,done,show,delete}`
- [x] **Top-level aliases** per ADR-003: `bt add`, `bt list`, `bt done`
- [x] **Noun aliases**: `bt t`, `bt tasks`
- [x] **Quiet mode** (`-q`/`--quiet`): outputs only IDs for piping
- [x] **Flag support**: `-p` priority, `-e` energy, `--due`, `--tag` (repeatable), `-c` context, `-b` box, `--duration`, `--due-start`, `--due-end`, `-s` status, `-n` limit
- [x] **Table output** for `bt list`: ID (8 chars), TITLE (truncated at 28), STATUS, PRIORITY, DUE
- [x] **Detail view** for `bt show`: all fields including file path, timestamps, body
- [x] **Confirmation output**: `✓ Created task <shortID>`, `✓ Completed: <title>`, `✓ Deleted: <title>`
- [x] `bt index rebuild` — admin command for re-indexing
- [x] `openApp(cmd)` helper — reads `--data-dir` flag, resolves to abs path, opens App
- [x] All commands use `cmd.Printf` (not `fmt.Printf`) — testable output
- [x] All commands properly `defer a.Close()` — no resource leaks

**Functional Smoke Tests**:
- [x] `bt add "Buy groceries" -p high --tag errands --tag home` → creates task ✅
- [x] `bt add "Write report" --due 2026-04-10` → auto-detects `dated` type ✅
- [x] `bt add "Paint bedroom" -b projects/reno -c home` → writes to box dir ✅
- [x] `bt list` → table output with all 3 tasks ✅
- [x] `bt done <id>` → marks task done, persists to disk ✅
- [x] `bt task show <id>` → full detail view with status=done ✅
- [x] `bt task delete <id>` → removes file + index, list shows "No tasks found" ✅
- [x] `bt task --help` → shows subcommands, `bt t --help` alias works ✅
- [x] `bt task add --help` → shows all flags ✅
- [x] `bt index rebuild --help` → shows help ✅

**App Test Coverage — 11 tests**:
- [x] Open creates directories (inbox, .bentotask)
- [x] AddTask: basic, with due date (auto-type), in box (custom path)
- [x] GetTask: by prefix, not found
- [x] CompleteTask: success, already done error
- [x] DeleteTask: removes file + index
- [x] ListTasks: returns all
- [x] RebuildIndex: re-indexes from files

**ADR-003 Compliance**:
- [x] Noun-verb: `bt task add` ✅
- [x] Top-level shortcuts: `bt add`, `bt list`, `bt done` ✅
- [x] `--quiet` outputs only IDs ✅
- [x] Flag short forms: `-p`, `-e`, `-c`, `-b`, `-s`, `-n`, `-q` ✅
- [x] `--tag` is repeatable (StringSlice) ✅
- ⚠️ `--json` flag exists but not yet implemented (outputs plain text) — OK for now
- ⚠️ No `$EDITOR` integration yet — M2.5 (`bt edit`)
- ⚠️ No color/lipgloss styling yet — M2.8

**Notes**:
- This is the biggest and most impactful commit so far (920 lines). The architecture is clean — the app layer properly separates orchestration from CLI concerns. Commands can't bypass validation or forget to index.
- Smart type auto-detection: setting `--due` automatically promotes one-shot→dated, setting `--due-start`+`--due-end` promotes one-shot→ranged. Reduces user friction.
- The flag duplication pattern (registering same flags on both `taskAddCmd` and `addCmd`) is a pragmatic workaround for Cobra aliases not sharing flag sets. Slightly verbose but correct.
- `bt show` displays `CompletedAt` missing — it shows timestamps for created/updated but not completed_at. Minor display gap.
- `.gitkeep` files removed from `internal/{api,calendar,engine,graph,routine}/` — good cleanup now that real code exists nearby.

**Verdict**: **APPROVED** ✅

**Next tasks**: M2.5 (`bt edit`), M2.8 (enhanced display), M2.9 (search/FTS), M2.10 (completions), M2.11 (integration tests)

---

## [2026-04-06] Review: M1.4 — File Watcher (Commit a77e2a1)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M1.4 complete

**Verification**:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` no issues
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `make test` — **54/54 PASS** (6 new watcher tests + previous 48)
- [x] Tracking docs updated: roadmap (M1.4 ✅) + handoff (M1.6/M2 assigned)

**Code Review — `internal/store/watcher.go` (193 lines)**:
- [x] `NewWatcher(dataDir, *Index)` — creates fsnotify watcher, recursively adds all non-hidden subdirs, starts background goroutine
- [x] `Close()` — signals `done` channel, closes fsnotify, waits for goroutine via `sync.WaitGroup`
- [x] Event loop: `select` on `done`, `Events`, and `Errors` channels — clean shutdown
- [x] `handleEvent`:
  - CREATE (directory) → adds to watcher (skips hidden dirs)
  - CREATE/WRITE (.md file) → parses and upserts into index
  - REMOVE/RENAME → deletes from index by ULID extracted from filename
  - Filters: skips non-.md, `_box.md`, `.tmp-*` temp files, hidden directory paths
- [x] `OnError` / `OnIndex` callback hooks — testable, defaults to stderr
- [x] `addRecursive` — walks dir tree, skips hidden dirs, adds each to fsnotify
- [x] Proper error wrapping with `%w` throughout

**Concurrency Design**:
- [x] Background goroutine with `sync.WaitGroup` for clean shutdown ✅
- [x] `done` channel for graceful stop — no goroutine leaks ✅
- [x] `atomic.Int32` used in tests for safe cross-goroutine counting ✅

**Test Coverage — 6 tests**:
- [x] `TestWatcherDetectsNewFile` — creates .md file, waits for OnIndex callback, verifies in index
- [x] `TestWatcherDetectsModifiedFile` — modifies existing file, verifies title updated in index
- [x] `TestWatcherDetectsDeletedFile` — removes file, verifies TaskCount drops to 0
- [x] `TestWatcherIgnoresNonMarkdownFiles` — creates .txt file, verifies nothing indexed
- [x] `TestWatcherDetectsNewSubdirectory` — creates nested dirs one level at a time, verifies file indexed
- [x] `TestWatcherClose` — verifies close doesn't hang or panic

**ADR-002 §6 Compliance**:
- [x] Watches data directory while API server is running ✅
- [x] Detects create, modify, delete events for .md files ✅
- [x] Skips hidden directories (`.bentotask/`) ✅
- ⚠️ No mtime/hash incremental sync on startup — deferred to M2.1 (correct per handoff)

**Notes**:
- Clean concurrent design. The `done` channel + WaitGroup pattern is idiomatic Go.
- `removeFile` extracts task ID from filename (strips `.md` suffix) — assumes ULID filenames per ADR-002. This is correct but fragile if filenames ever differ from IDs. Acceptable for now.
- The `TestWatcherDetectsNewSubdirectory` test creates dirs one level at a time with `time.Sleep(200ms)` between — smart workaround for fsnotify not catching deeply nested `MkdirAll` calls. Slightly slow (0.43s) but reliable.
- `waitFor` polling helper with 20ms intervals and 2s timeout is reasonable for filesystem event tests.
- Temp file filtering (`.tmp-*`) correctly prevents indexing partial writes from the atomic write path in `WriteFile`. Nice integration between the two subsystems.
- The "deferred ADR-002 items" table in the handoff is an excellent addition — makes it clear what was intentionally left out vs missed.

**Verdict**: **APPROVED** ✅

**M1 Status**: 5 of 6 tasks complete (M1.1–M1.5). M1.6 (remaining unit tests) is partially satisfied — 54 tests already exist across the data layer. The handoff correctly notes this and suggests moving to M2.

**Next**: M1.6 (fill remaining test gaps if any) → M2: Basic CLI

---

## [2026-04-06] Review: M1.5 — ULID Generation (Commit 6915a05)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M1.5 complete (done out of order to unblock M1.3)

**Verification**:
- [x] `go build ./...` compiles
- [x] `go vet` + `golangci-lint` — 0 issues
- [x] 6 new tests, all pass

**Code Review — `internal/model/id.go` (47 lines)**:
- [x] `NewID()` — generates ULID with `crypto/rand` for entropy (secure)
- [x] `NewIDAt(time.Time)` — useful for tests and imports
- [x] `IDTime(string)` — extracts timestamp, returns zero time on invalid input
- [x] `MatchesPrefix(id, prefix)` — case-insensitive via `strings.EqualFold`, handles edge cases (empty prefix, prefix longer than ID)
- [x] Package comment duplicated from `task.go` — minor, Go allows it

**Test Quality**:
- [x] Uniqueness: two `NewID()` calls produce different values
- [x] Length: exactly 26 characters
- [x] Round-trip: `NewIDAt(ts)` → `IDTime()` matches within 1ms
- [x] Sortability: earlier timestamps produce lexicographically smaller IDs
- [x] Prefix matching: full match, partial, case-insensitive, no-match, empty, too-long

**Also in this commit**:
- [x] Added TODO comment on `markdown.go` temp file naming (per previous review feedback)
- [x] Dependencies: `oklog/ulid/v2`, `modernc.org/sqlite` + transitive deps

**Verdict**: **APPROVED** ✅

---

## [2026-04-06] Review: M1.3 — SQLite Index (Commit 13260d4)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M1.3 complete

**Verification**:
- [x] `go build ./...` compiles
- [x] `go vet` + `golangci-lint` — **0 issues**
- [x] `make test` — **48/48 PASS** (13 new index tests + 11 store + 24 model + 3 CLI)

**Code Review — `internal/store/schema.go` (61 lines)**:
- [x] 4 tables: `tasks`, `task_tags`, `task_contexts`, `task_links`
- [x] All `IF NOT EXISTS` — safe for re-runs
- [x] Foreign keys with `ON DELETE CASCADE` on junction tables
- [x] 9 indexes for fast queries (status, type, due_date, due_end, box, priority, tag, context, link target)

**Code Review — `internal/store/index.go` (442 lines)**:
- [x] `OpenIndex` — creates dirs, opens with WAL mode + foreign keys (matches ADR-002)
- [x] `UpsertTask` — transactional, handles 16 columns + junction tables (delete + re-insert for tags/contexts/links)
- [x] `DeleteTask` — cascade deletes junction rows via FK
- [x] `GetTask` — loads tags and contexts via separate queries
- [x] `FindByPrefix` — `LIKE ? || '%'` with `strings.ToUpper` (ULID is Crockford Base32)
- [x] `ListTasks` — composable filters (status, type, priority, energy, box, tag, context, limit), JOINs added dynamically
- [x] `RebuildIndex` — walks directory, parses .md files, skips hidden dirs and `_box.md`, graceful error handling (warnings to stderr)
- [x] `TaskCount` — simple aggregate
- [x] Proper `defer` on `rows.Close()`, transaction rollback
- [x] Null-handling helpers: `nullIfEmpty`, `nullIfZero`, `timePtr`

**Test Coverage — 13 tests**:
- [x] Schema creation, upsert+get, upsert updates, delete, find by prefix (match, multi, no-match)
- [x] List: no filter, filter by status, filter by tag, filter by context, limit
- [x] Rebuild: parses files + skips malformed, skips hidden dirs
- [x] Error case: GetTask for nonexistent ID

**ADR-002 §5 Cross-Reference Audit**:
- ✅ Core `tasks` table matches — all query-relevant columns present
- ✅ Junction tables (`task_tags`, `task_contexts`, `task_links`) match
- ✅ WAL mode and foreign keys enabled via DSN pragma
- ⚠️ **Missing from ADR-002** (intentionally deferred):
  - `file_mtime` + `file_hash` columns (for sync strategy §6) — needed for M1.4 file watcher
  - `routine_steps` table — needed for M4 routines
  - `habit_completions` table — needed for M3 habits
  - `tasks_fts` FTS5 virtual table — needed for search
- ⚠️ **Minor gaps**:
  - `priority` DEFAULT 'none' in ADR vs nullable in schema — code uses `nullIfEmpty` which works but differs from spec
  - `task_links.target_id` — ADR has FK reference, schema omits it (target task may not be indexed yet — this is defensible)
  - Index name `idx_task_contexts_ctx` vs ADR's `idx_task_contexts_context` — cosmetic

**Notes**:
- The missing tables/columns are **correct engineering decisions** — they belong to later milestones (M3, M4) and adding them now would be premature. The schema uses `IF NOT EXISTS` so they can be added incrementally.
- `RebuildIndex` uses `fmt.Fprintf(os.Stderr, ...)` for warnings — this works but isn't testable. Consider a logger or callback in the future. Non-blocking.
- Tracking docs were updated in this commit (roadmap + handoff). Good improvement from M1.2 feedback.
- `FindByPrefix` uses `strings.ToUpper` before LIKE — correct since ULIDs use uppercase Crockford Base32 and SQLite LIKE is case-insensitive for ASCII.

**Verdict**: **APPROVED** ✅

**Next tasks**: M1.4 (file watcher) + M1.6 (remaining unit tests)

---

## [2026-04-06] Review: M1.2 — Markdown + YAML Frontmatter Reader/Writer (Commit 8fac6e0)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M1.2 complete (but tracking docs not updated in commit — fixed by supervisor)

**Verification**:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` no issues
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `make test` — **29/29 PASS** (11 new store tests + 18 model + 3 CLI)

**Code Review — `internal/store/markdown.go` (116 lines)**:
- [x] `Parse(io.Reader)` — parses YAML frontmatter via `adrg/frontmatter`, captures body, trims whitespace
- [x] `ParseFile(path)` — convenience wrapper, opens file and delegates to Parse
- [x] `Marshal(*model.Task)` — serializes Task to `---\nyaml\n---\n\nbody\n` format
- [x] `WriteFile(path, *model.Task)` — atomic writes (temp file + os.Rename), auto-creates parent dirs
- [x] Proper error wrapping with `%w` throughout
- [x] `defer func() { _ = f.Close() }()` — explicit discard of close error (passes linter)
- [x] `.gitkeep` removed from `internal/store/` (replaced by real code)

**Test Coverage — `internal/store/markdown_test.go` (412 lines, 11 tests)**:
- [x] `TestParseBasicTask` — full-featured task with all optional fields, links, body
- [x] `TestParseMinimalTask` — minimal valid task, empty body
- [x] `TestParseHabit` — habit with frequency, streaks, recurrence
- [x] `TestParseRoutine` — routine with steps (optional/required), schedule
- [x] `TestMarshalRoundTrip` — write → read → compare fields
- [x] `TestWriteFileAndParseFile` — end-to-end file I/O with `t.TempDir()`
- [x] `TestWriteFileAtomicity` — verifies no `.tmp-*` files left behind
- [x] `TestWriteFileCreatesDirectories` — nested `projects/home-renovation/` path
- [x] `TestParseFileNotFound` — error on nonexistent file
- [x] `TestParseMalformedFrontmatter` — error on invalid YAML
- [x] `TestMarshalEmptyBody` — no extra blank lines when body is empty

**ADR-002 Compliance**:
- [x] File format: YAML frontmatter between `---` delimiters + Markdown body ✅
- [x] Atomic writes: temp file + rename (ADR-002 §8) ✅
- [x] Recovery: temp file cleanup on rename failure ✅
- [x] Auto-creates parent directories for box-based file organization ✅

**Dependencies Added**:
- `github.com/adrg/frontmatter v0.2.0` (direct) — YAML/TOML/JSON frontmatter parser
- `gopkg.in/yaml.v3 v3.0.1` (direct) — YAML marshaling
- `github.com/BurntSushi/toml v0.3.1` (indirect, via adrg/frontmatter)
- `gopkg.in/yaml.v2 v2.3.0` (indirect, via adrg/frontmatter)

**Notes**:
- Very clean implementation. 116 lines of code with thorough test coverage (412 lines of tests — ~3.5x test-to-code ratio).
- Good use of `io.Reader` abstraction — `Parse()` is testable without touching the filesystem.
- Atomic write pattern is correct: write to `.tmp-write-<filename>` in same directory, then `os.Rename`. Same-filesystem rename is atomic on POSIX.
- Temp file naming uses a fixed prefix (`.tmp-write-`). If two processes write the same file simultaneously, they'd race on the same temp path. Fine for single-user local-first app, but worth noting for future multi-device sync scenarios.
- The round-trip test verifies field-level fidelity but doesn't do byte-level comparison (Marshal→Parse→Marshal→compare bytes). This is acceptable since YAML formatting can vary, but could be tightened later.
- **Tracking docs were not updated** in this commit. The handoff still shows M1.1 as active. Fixed by supervisor in this review.

**Verdict**: **APPROVED** ✅

**Next task**: M1.3 — SQLite index (schema, create, rebuild from files)

---

## [2026-04-06] Review: 5dc5e2c — M0 Review Feedback Fix

**Reviewed by**: Supervisor Agent  
**Commit**: `5dc5e2c` — "Address M0.7+M0.8 review feedback — cleaner tests, graceful fmt target"

**Changes**:
- `TestExecute` now captures output via `rootCmd.SetOut(buf)` — no more help text in test logs
- Makefile `fmt` target gracefully skips `goimports` if not installed
- Supervisor log updated with M0.6 and M0.7+M0.8 review entries

**Verdict**: APPROVED — Both feedback items addressed cleanly.

---

## [2026-04-06] Review: M1.1 — Task Data Model (Commit 6a18bcb)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M1.1 complete

**Verification**:
- [x] `go build ./...` compiles cleanly
- [x] `go vet ./...` no issues
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `make test` — **18/18 PASS** (15 new model tests + 3 existing CLI tests)
- [x] `task.go` — Task struct with all ADR-002 fields: required (id, title, type, status, created, updated) + optional common + recurrence + habit + routine fields
- [x] Type-safe enums: TaskType (7 values), Status (6), Priority (5), Energy (3), LinkType (3), RecurrenceAnchor (2)
- [x] `validate.go` — Validation for required fields, enum values, type-specific rules, link validation
- [x] Helper methods: IsDone, ShortID, HasTag, HasContext, IsValid
- [x] `task_test.go` — Table-driven tests covering all validation paths + helpers
- [x] `Body` field uses `yaml:"-"` tag (excluded from frontmatter, stored as markdown body) ✅

**ADR-002 Cross-Reference Audit**:
- [x] All 18 task fields from ADR-002 §3 present in struct
- [x] All 3 habit fields present (frequency, streak_current, streak_longest)
- [x] All 4 routine fields present (steps, schedule + step sub-fields)
- [x] All enum values match ADR-002
- **Minor spec gaps (NOT code bugs)**:
  - `routine` is missing from the `type` enum list in ADR-002 §3 (but documented in Routine Fields section) — code is correct
  - `recurrence_anchor` not in §3 field table (but documented in §7 Recurrence) — code is correct
  - `time.Time` used for datetime fields vs ADR's "ISO datetime" string — functionally compatible via go-yaml marshaling

**Code Quality Notes**:
- Good separation: types in `task.go`, logic in `validate.go`
- Table-driven tests with `newValidTask()` helper — idiomatic Go testing pattern
- `containsError` / `searchString` helpers in tests could use `strings.Contains` from stdlib instead of hand-rolled string search. Non-blocking, minor.
- `DueDate`, `DueStart`, `DueEnd` are `string` (not `time.Time`) — this is intentional per ADR-002 which uses ISO date strings (YYYY-MM-DD), distinct from datetime fields. Makes sense.

**Verdict**: **APPROVED** ✅

**Next task**: M1.2 — Markdown + YAML frontmatter reader/writer

---

## [2026-04-06] Review: M0.7+M0.8 — Coding Standards, Tests & CI (M0 Completion)

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: M0.7 and M0.8 complete; M0 milestone fully done

**Verification — M0.7 (Coding Standards)**:
- [x] `.golangci.yml` — v2 config with revive, gocritic, misspell, errcheck, staticcheck, unused, ineffassign, govet
- [x] Formatters configured: gofmt + goimports with local prefix
- [x] `golangci-lint run ./...` — **0 issues**
- [x] `Makefile` — targets: build, test, lint, fmt, clean, help
- [x] Build target uses `-ldflags` to inject version correctly
- [x] `CONTRIBUTING.md` — covers dev setup, code style, project structure, testing, ADRs, commit conventions

**Verification — M0.8 (First Test + CI)**:
- [x] `internal/cli/root_test.go` — 3 tests: TestExecute, TestVersionCommand, TestRootHasGlobalFlags
- [x] `make test` — **3/3 PASS**
- [x] `.github/workflows/ci.yml` — test + lint jobs, triggers on push/PR to main
- [x] CI uses `go-version-file: go.mod` (auto Go version) + `golangci-lint-action@v7`
- [x] Code fix: `versionCmd` now uses `cmd.Printf` instead of `fmt.Printf` for testability

**Verification — M0 Acceptance Criteria (ALL MET)**:
- [x] All three ADRs written and approved (ADR-001, 002, 003)
- [x] `go build ./cmd/bt` compiles successfully
- [x] `bt --version` prints version info (`bt version dev`)
- [x] `bt --help` shows command structure with usage examples
- [x] At least one passing test (3 tests pass)
- [x] CI runs on push (GitHub Actions configured)

**Verdict**: **APPROVED** ✅

**Notes**:
- Clean, thorough work. All M0 acceptance criteria are met.
- Good practice: used `cmd.Printf` over `fmt.Printf` in Cobra commands for testability — this shows awareness of idiomatic Cobra patterns.
- The golangci-lint config is thoughtful — enables useful linters (revive, gocritic) without being overly strict for early development.
- Makefile `fmt` target depends on `goimports` being installed — not a blocker since CI uses golangci-lint formatters, but the CONTRIBUTING.md could mention installing it.
- Minor: `TestExecute` prints the full help text to stdout during test runs. Consider capturing output to keep test output clean. Non-blocking.
- Worker agent updated all tracking docs correctly this time (learned from M0.6 feedback). Good improvement.

**M0 Milestone: COMPLETE** 🎉

**Next milestone**: M1 — Core Data Model
**First task**: M1.1 — Implement task data model (struct + serialization)

---

## [2026-04-06] Review: M0.6 — Project Scaffolding

**Reviewed by**: Supervisor Agent  
**Handoff claimed**: Task complete (code present, but handoff file was stale — still said "Not Started")

**Verification**:
- [x] `go build ./...` compiles cleanly, no errors
- [x] `go vet ./...` passes with no issues
- [x] `bt --version` prints `bt version 0.1.0`
- [x] `bt --help` shows proper help text with usage examples
- [x] Global flags match ADR-003: `--json`, `--quiet`, `--no-color`, `--data-dir`, `--verbose`
- [x] Folder structure matches ADR-001: `cmd/bt/`, `internal/{model,store,engine,calendar,routine,graph,api,cli}/`, `plugins/`, `web/`
- [x] `cmd/bt/main.go` is minimal — delegates to `cli.Execute()`
- [x] `internal/cli/root.go` uses Cobra idiomatically with `version` subcommand
- [x] `/bt` binary is gitignored
- [ ] No test files exist yet (expected — M0.8 covers this)
- [ ] Tracking docs were stale (handoff, milestone, roadmap not updated after scaffolding)

**Verdict**: APPROVED WITH NOTES

**Notes**:
- Scaffolding is solid and well-structured. Code is idiomatic Go.
- Version is injectable via `-ldflags` (`var version = "dev"` default) — good pattern, but no Makefile yet to standardize the build. M0.7 will address this.
- `SPEC.original.md` exists locally despite being gitignored — low risk but noted.
- Tracking docs (handoff, milestone, roadmap) were not updated after commit. Fixed in this review session.

**Next assignments**:
1. M0.7: Coding standards (golangci-lint, Makefile, CONTRIBUTING.md)
2. M0.8: First test + CI (GitHub Actions)
3. Then M1: Core Data Model

---

## [2026-04-05] Review: M0.4–M0.5 — ADR-002 & ADR-003

**Reviewed by**: Solo Agent (supervisor hat)  
**Tasks reviewed**:
- M0.4: ADR-002 — Storage format & indexing strategy
- M0.5: ADR-003 — CLI framework & UX patterns

**Verification**:
- [x] ADR-002 covers: file format, file organization, frontmatter schema, ID strategy, SQLite schema, sync strategy, recurrence, atomic writes, recovery
- [x] ADR-002 approved by human
- [x] ADR-003 covers: command structure, interactive vs plain, output formats, color/styling, editor integration, shell completions, error handling, flag patterns
- [x] ADR-003 approved by human
- [x] All tracking docs updated (ROADMAP.md, M0-bootstrap.md, handoff, decision log)

**Verdict**: APPROVED

**Notes**:
- All three ADRs are now approved. Architecture phase is complete.
- The handoff file is ready for a new agent to pick up scaffolding work.
- Key decision: all 3 ADRs consistently chose the Charm ecosystem + pure Go libraries (no CGO).

**Next assignments**:
1. M0.6: Project scaffolding (go mod, folder structure, main.go)
2. M0.7: Coding standards (golangci-lint, Makefile)
3. M0.8: First test + CI

---

## [2026-04-05] Review: M0.1–M0.3 — Initial Spec, Repo Setup, ADR-001

**Reviewed by**: Solo Agent (supervisor hat)  
**Tasks reviewed**:
- M0.1: Write initial SPEC.md
- M0.2: Initialize git repository
- M0.3: ADR-001 — Tech stack selection

**Verification**:
- [x] SPEC.md exists with comprehensive feature requirements (721 lines)
- [x] Git repo initialized on `main` branch
- [x] SPEC.original.md backed up and gitignored
- [x] ADR-001 written with 6 options evaluated, comparison matrix, clear recommendation
- [x] ADR-001 approved by human: **Go + SvelteKit**
- [x] ROADMAP.md created with 10 milestones
- [x] M0-bootstrap.md milestone doc created
- [x] CLAUDE.md agent orientation file created
- [x] AGENTS.md roles/protocols file created
- [x] Handoff system initialized (CURRENT.md + HISTORY.md)
- [x] .gitignore configured

**Verdict**: APPROVED  

**Notes**:
- Excellent foundation. All planning artifacts are in place.
- ADR-001 was thorough — the Go + SvelteKit choice is well-justified for the requirements.
- The handoff/supervisor system is set up and ready for multi-agent use when needed.

**Next assignments**:
1. ADR-002: Storage format & indexing strategy
2. ADR-003: CLI framework & UX patterns
3. Project scaffolding (go mod, folder structure, Makefile)

---
