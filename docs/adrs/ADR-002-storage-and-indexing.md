# ADR-002: Storage Format & Indexing Strategy

**Status**: APPROVED  
**Date**: 2026-04-05  
**Approved**: 2026-04-05  
**Decision Makers**: @jbencardino  
**Depends on**: ADR-001 (Go + SvelteKit — APPROVED)

---

## Context

BentoTask stores tasks, habits, and routines as the user's data. We need to decide:

1. **File format**: How is each item serialized on disk?
2. **File organization**: One file per task, or bundled?
3. **Indexing**: How do we enable fast queries without re-parsing every file?
4. **Sync strategy**: How does the index stay in sync with files on disk?
5. **IDs**: How are items uniquely identified?
6. **Recurrence**: How are recurring schedules stored?

The guiding principles are: **local-first**, **human-readable**, **git-friendly**, **no vendor lock-in**.

---

## Decisions

### 1. File Format: Markdown + YAML Frontmatter

Each task/habit/routine is a single `.md` file with YAML frontmatter (delimited by `---`) followed by a Markdown body.

```markdown
---
id: 01JQXYZ123456
title: Paint bedroom
type: one-shot
status: pending
priority: medium
energy: high
estimated_duration: 180
due_start: 2026-04-07
due_end: 2026-04-13
tags: [home, renovation]
context: [home]
box: projects/home-renovation
links:
  - type: depends-on
    target: 01JQXYZ123400
recurrence: null
created: 2026-04-05T10:30:00Z
updated: 2026-04-05T10:30:00Z
---

# Paint Bedroom

Need to repaint the bedroom walls. Going with the sage green.

## Notes
- Remove furniture first
- Two coats minimum
```

**Library**: [`adrg/frontmatter`](https://github.com/adrg/frontmatter) — cleanest API, auto-detects delimiters, uses `go-yaml/yaml` under the hood.

```go
var matter TaskFrontmatter
body, err := frontmatter.Parse(reader, &matter)
```

**Rationale**:
- Human-readable — users can edit files in any text editor
- Git-friendly — clean diffs, meaningful merges
- Portable — standard format, no proprietary encoding
- Flexible — the body is free-form markdown for notes, checklists, links
- Hugo, Obsidian, Jekyll, and many other tools use this exact format

---

### 2. File Organization: One File Per Task, ULID as Filename

```
data/
├── inbox/
│   └── 01JQX00001.md          # Unsorted tasks
├── projects/
│   ├── home-renovation/
│   │   ├── _box.md             # Box/project metadata
│   │   ├── 01JQX00010.md      # "Buy paint"
│   │   └── 01JQX00011.md      # "Paint bedroom"
│   └── work-q2-launch/
│       └── _box.md
├── habits/
│   ├── 01JQX00100.md          # "Read 30 minutes"
│   └── 01JQX00101.md          # "Exercise"
├── routines/
│   ├── 01JQX00200.md          # "Morning routine"
│   └── 01JQX00201.md          # "Evening routine"
└── areas/
    ├── health/
    │   └── _box.md
    └── work/
        └── _box.md
```

**Why file-per-task**:

| Concern | File-per-task | Single file / DB |
|---------|--------------|-----------------|
| Git diffs | Clean — one file changed per task edit | Noisy — whole file changes |
| Merge conflicts | Per-task, easy to resolve | Whole-database conflicts |
| Filesystem at 1K tasks | ~2ms readdir, no issues | N/A |
| Filesystem at 10K tasks | Still fine (B-tree dirs on APFS/ext4) | N/A |
| External editing | Open specific file in $EDITOR | Need specialized tool |
| Sync (Syncthing/git) | File-level granularity | Full DB transfer |
| Deletion | `rm file.md` | Need tool |

**Why ULID as filename** (not slugified title):
- Titles change; filenames shouldn't
- ULIDs sort chronologically by default (`ls` shows creation order)
- No filename collision or sanitization issues
- The `title` field in frontmatter is the human-readable name
- 26 chars, URL-safe, no special characters

**Box metadata**: Each box (project/area) has a `_box.md` file with:
```yaml
---
id: 01JQX00009
title: Home Renovation
type: project  # or "area"
description: All tasks related to renovating the house
tags: [home]
created: 2026-04-05T10:00:00Z
---
```

---

### 3. Frontmatter Schema

#### Task Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string (ULID) | yes | Unique identifier |
| `title` | string | yes | Human-readable name |
| `type` | enum | yes | `one-shot`, `dated`, `ranged`, `floating`, `recurring`, `habit` |
| `status` | enum | yes | `pending`, `active`, `paused`, `done`, `cancelled`, `waiting` |
| `priority` | enum | no | `none` (default), `low`, `medium`, `high`, `urgent` |
| `energy` | enum | no | `low`, `medium`, `high` |
| `estimated_duration` | int | no | Minutes |
| `due_date` | ISO date | no | For `dated` tasks |
| `due_start` | ISO date | no | For `ranged` tasks — start of window |
| `due_end` | ISO date | no | For `ranged` tasks — end of window |
| `tags` | string[] | no | Freeform tags |
| `context` | string[] | no | `home`, `office`, `errands`, `commute`, `anywhere` |
| `box` | string | no | Path to parent box (e.g., `projects/home-renovation`) |
| `links` | object[] | no | `[{type: "depends-on", target: "ULID"}]` |
| `recurrence` | string | no | RFC 5545 RRULE (e.g., `FREQ=WEEKLY;BYDAY=MO,WE,FR`) |
| `completed_at` | ISO datetime | no | When task was marked done |
| `created` | ISO datetime | yes | Creation timestamp |
| `updated` | ISO datetime | yes | Last modification timestamp |

#### Habit-Specific Fields (in addition to task fields)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `frequency` | object | yes | `{type: "daily", target: 1}` or `{type: "weekly", target: 3}` |
| `streak_current` | int | no | Current streak count (computed, cached) |
| `streak_longest` | int | no | All-time longest streak (computed, cached) |

#### Habit Completion Log

Habit completions are stored in the markdown body as a structured section:

```markdown
## Completions
- 2026-04-05T08:30:00Z | 35min | "DDIA ch.7"
- 2026-04-04T09:00:00Z | 30min | "DDIA ch.6"
```

**Why in the body, not frontmatter?** Frontmatter stays small and structured. The completion log can grow large and is append-only — keeping it in the body means the frontmatter stays fast to parse for indexing.

#### Routine Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | ULID | yes | Unique identifier |
| `title` | string | yes | Name |
| `type` | literal | yes | Always `routine` |
| `steps` | object[] | yes | `[{ref: "ULID", optional: false}]` |
| `schedule` | object | no | `{time: "07:00", days: ["mon","tue","wed","thu","fri"]}` |

---

### 4. ID Strategy: ULID

**Library**: [`oklog/ulid`](https://github.com/oklog/ulid)

```go
id := ulid.Make() // "01H5N3K9QM8RJPW2XVZGY7T4E"
```

| Property | Value |
|----------|-------|
| Length | 26 characters |
| Sortable by time | Yes (millisecond precision) |
| URL-safe | Yes (Crockford Base32) |
| Collision-resistant | 80 bits of randomness per millisecond |
| Timestamp extractable | Yes — `ulid.Time(id.Time())` |
| Fits UUID column | Yes — 128 bits |

**Filename convention**: `{ULID}.md` (e.g., `01JQX00010.md`)

**Short IDs for CLI**: For user-facing commands, support prefix matching:
```bash
bt done 01JQX    # Matches if unique prefix
bt show 01JQX00010  # Full ID always works
```

---

### 5. SQLite Index / Cache

The SQLite database is a **derived cache**, never the source of truth. It can be deleted and rebuilt from the markdown files at any time.

**Location**: `data/.bentotask/index.db`

**Library**: [`modernc.org/sqlite`](https://modernc.org/sqlite) (pure Go, no CGO)

**Why pure Go SQLite?**
- No C toolchain required
- Trivial cross-compilation (`GOOS=linux GOARCH=arm go build`)
- Performance is negligible for our scale (~10% slower than CGO for <10K rows)
- `go build` just works everywhere

**Connection string**:
```go
"file:data/.bentotask/index.db?_pragma=journal_mode(wal)&_pragma=foreign_keys(1)"
```

#### Schema

```sql
-- Core task index
CREATE TABLE tasks (
    id              TEXT PRIMARY KEY,   -- ULID
    title           TEXT NOT NULL,
    type            TEXT NOT NULL,       -- 'one-shot','dated','ranged','floating','recurring','habit'
    status          TEXT NOT NULL,       -- 'pending','active','paused','done','cancelled','waiting'
    priority        TEXT DEFAULT 'none', -- 'none','low','medium','high','urgent'
    energy          TEXT,                -- 'low','medium','high'
    estimated_duration INTEGER,          -- minutes
    due_date        TEXT,                -- ISO date
    due_start       TEXT,                -- ISO date
    due_end         TEXT,                -- ISO date
    box             TEXT,                -- path to parent box
    recurrence      TEXT,                -- RRULE string
    completed_at    TEXT,                -- ISO datetime
    created_at      TEXT NOT NULL,       -- ISO datetime
    updated_at      TEXT NOT NULL,       -- ISO datetime
    file_path       TEXT NOT NULL,       -- relative path from data/
    file_mtime      INTEGER NOT NULL,    -- unix timestamp (seconds)
    file_hash       TEXT NOT NULL        -- xxhash of file content
);

-- Tags (many-to-many)
CREATE TABLE task_tags (
    task_id     TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag         TEXT NOT NULL,
    PRIMARY KEY (task_id, tag)
);

-- Context (many-to-many)
CREATE TABLE task_contexts (
    task_id     TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    context     TEXT NOT NULL,
    PRIMARY KEY (task_id, context)
);

-- Links between tasks
CREATE TABLE task_links (
    source_id   TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    target_id   TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    link_type   TEXT NOT NULL,  -- 'depends-on','blocks','related-to'
    PRIMARY KEY (source_id, target_id, link_type)
);

-- Routine steps (ordered)
CREATE TABLE routine_steps (
    routine_id  TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    step_order  INTEGER NOT NULL,
    task_id     TEXT NOT NULL REFERENCES tasks(id),
    optional    INTEGER DEFAULT 0,
    PRIMARY KEY (routine_id, step_order)
);

-- Habit completions (for fast streak queries)
CREATE TABLE habit_completions (
    habit_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    completed_at TEXT NOT NULL,  -- ISO datetime
    duration    INTEGER,         -- minutes
    note        TEXT,
    PRIMARY KEY (habit_id, completed_at)
);

-- Indexes
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_due_end ON tasks(due_end);
CREATE INDEX idx_tasks_box ON tasks(box);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_task_tags_tag ON task_tags(tag);
CREATE INDEX idx_task_contexts_context ON task_contexts(context);
CREATE INDEX idx_task_links_target ON task_links(target_id);
CREATE INDEX idx_habit_completions_date ON habit_completions(completed_at);
```

#### Full-Text Search

```sql
-- FTS5 virtual table for full-text search across titles and body content
CREATE VIRTUAL TABLE tasks_fts USING fts5(
    id UNINDEXED,
    title,
    body,
    content=tasks,
    content_rowid=rowid
);
```

---

### 6. Index Sync Strategy: Hybrid mtime + Hash

The index must stay in sync with the markdown files. Strategy:

```
On CLI startup (every command):
  1. Quick check: compare data/ directory mtime with last sync timestamp
     - If unchanged → skip sync, use cached index (< 1ms)
  2. If changed → incremental sync:
     a. Walk data/ directory
     b. For each .md file:
        - Compare file mtime with stored mtime in SQLite
        - If mtime unchanged → skip (trust cache)
        - If mtime changed → hash file content (xxhash)
          - If hash unchanged → update mtime only (editor touched file)
          - If hash changed → re-parse frontmatter, update index
     c. Detect deleted files (in index but not on disk) → remove from index
     d. Detect new files (on disk but not in index) → parse and add
     e. Update last sync timestamp

While API server is running:
  - Use fsnotify to watch data/ directory
  - On CREATE/WRITE → parse and upsert
  - On REMOVE → delete from index
  - On RENAME → delete old + insert new
  - Ignore CHMOD events (editors trigger these)

Full rebuild (manual):
  - `bt index rebuild` — drops and recreates index from all files
  - Used as fallback if index gets corrupted
```

**Performance targets**:
- No-change startup sync: < 1ms
- Incremental sync (10 files changed out of 1000): < 50ms
- Full rebuild (1000 files): < 500ms

---

### 7. Recurrence: RFC 5545 RRULE

**Library**: [`teambition/rrule-go`](https://github.com/teambition/rrule-go)

Store recurrence as a standard RRULE string:

```yaml
recurrence: "FREQ=WEEKLY;BYDAY=MO,WE,FR"
```

| Pattern | RRULE |
|---------|-------|
| Every day | `FREQ=DAILY` |
| Every 3 days | `FREQ=DAILY;INTERVAL=3` |
| Mon, Wed, Fri | `FREQ=WEEKLY;BYDAY=MO,WE,FR` |
| 1st and 15th of month | `FREQ=MONTHLY;BYMONTHDAY=1,15` |
| 3rd Thursday of month | `FREQ=MONTHLY;BYDAY=3TH` |
| Every 2 weeks after completion | `FREQ=WEEKLY;INTERVAL=2` + `after_completion: true` |

**Why RRULE?**
- Industry standard (Google Calendar, Apple Calendar, Thunderbird)
- Import/export without conversion to iCal
- `rrule-go` handles all the edge cases (DST, leap years, end-of-month)

**Extension**: For "after completion" recurrence (not part of RFC 5545), add a custom field:
```yaml
recurrence: "FREQ=WEEKLY;INTERVAL=2"
recurrence_anchor: completion  # "fixed" (default) or "completion"
```

---

## Data Integrity

### Atomic Writes

To prevent corruption from crashes mid-write:

```go
// Write to temp file, then atomic rename
tmpFile := filepath.Join(dir, ".tmp-"+ulid.Make().String())
os.WriteFile(tmpFile, data, 0644)
os.Rename(tmpFile, targetPath)  // atomic on POSIX
```

### Validation

On parse, validate frontmatter against the schema:
- Required fields present
- Enum values are valid
- Dates parse correctly
- Referenced ULIDs exist (warn, don't block)
- Gracefully handle unknown fields (forward compatibility)

### Recovery

If a file is malformed:
- Log a warning
- Skip the file (don't index it)
- `bt doctor` command to list and attempt repair of malformed files

---

## Consequences

- Every task is a `.md` file — users can edit with any text editor
- The SQLite index is disposable — `bt index rebuild` recreates it
- File operations use atomic writes — no corruption on crash
- ULIDs as filenames — sortable, collision-free, URL-safe
- RRULE for recurrence — interoperable with calendar standards
- Pure Go SQLite — no CGO, cross-compiles to ARM
- Tags and contexts use junction tables — fast filtering
- FTS5 for full-text search across task titles and notes

---

## Open Questions (for ADR-003 or later)

1. Should `bt add` open `$EDITOR` for the markdown body, or just set frontmatter via flags?
2. Should we store the body text in SQLite too (for FTS), or only index frontmatter?
3. Maximum file size / body length — should we warn on very large task files?

---

## References

- [`adrg/frontmatter`](https://github.com/adrg/frontmatter) — Go frontmatter parser
- [`modernc.org/sqlite`](https://modernc.org/sqlite) — Pure Go SQLite
- [`oklog/ulid`](https://github.com/oklog/ulid) — Go ULID implementation
- [`teambition/rrule-go`](https://github.com/teambition/rrule-go) — RFC 5545 RRULE for Go
- [`fsnotify/fsnotify`](https://github.com/fsnotify/fsnotify) — Cross-platform file watcher
- [RFC 5545 — iCalendar RRULE](https://tools.ietf.org/html/rfc5545#section-3.3.10)
- [Hugo frontmatter docs](https://gohugo.io/content-management/front-matter/)
