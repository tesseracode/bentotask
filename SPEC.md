# BentoTask

> *Fit your life into your day — like a perfectly packed bento box.*

**BentoTask** is a local-first task, habit, and routine management system with intelligent scheduling. Like a bento box where every compartment holds something different yet everything fits together beautifully, BentoTask lets you nest tasks within tasks, group habits into routines, and intelligently pack your available time with what matters most.

---

## Table of Contents

1. [Vision](#vision)
2. [Core Concepts](#core-concepts)
3. [Features](#features)
4. [Architecture](#architecture)
5. [Data Model](#data-model)
6. [Scheduling Algorithm](#scheduling-algorithm)
7. [Interfaces](#interfaces)
8. [Integration Points](#integration-points)
9. [Future Extensibility](#future-extensibility)
10. [Prior Art & Differentiation](#prior-art--differentiation)
11. [Technical Stack](#technical-stack)
12. [Development Phases](#development-phases)

---

## 1. Vision

Most task managers are glorified checklists. Most habit trackers live in isolation. BentoTask combines:

- **Tasks** (one-shot, recurring, deadline-bound, or "whenever I'm free")
- **Habits** (streaks, frequency tracking, trend analysis)
- **Routines** (ordered sequences of tasks and habits)
- **Smart Scheduling** ("I have 45 minutes and medium energy — what should I do?")
- **Knowledge Context** (notes, documents, and linked concepts that give tasks meaning)

All stored locally in plain files you own, with optional sync and calendar integration.

---

## 2. Core Concepts

### 2.1 The Bento Model

```
+--[ Your Day ]----------------------------------+
|  +--[Morning Routine]---+  +--[Work Block]---+ |
|  | [] Meditate (habit)   |  | [] Deploy v2.1  | |
|  | [] Exercise (habit)   |  |   [] Run tests  | |
|  | [] Review inbox       |  |   [] Write docs | |
|  +----------------------+  |   [] Tag release | |
|                            +----------------+  |
|  +--[Chores]---+  +--[Evening]----------+     |
|  | [] Laundry   |  | [] Read 30min (habit)| |
|  | [] Groceries |  | [] Journal (habit)   | |
|  +-----------+  +---------------------+     |
+-------------------------------------------------+
```

- **Task**: A unit of work. Can be one-shot or recurring.
- **Habit**: A special task tracked for consistency/streaks over time.
- **Routine**: An ordered group of tasks/habits performed as a sequence.
- **Box** (Container): Any grouping — a project, an area of life, a context.
- **Link**: A relationship between items (dependency, related, blocks, blocked-by).
- **Slot**: A time window in your day where tasks can be packed.

### 2.2 Task Types

| Type | Description | Examples |
|------|-------------|---------|
| **One-shot** | Do once, then done | "Buy birthday gift for Mom" |
| **Dated** | Must happen on/by a specific date | "Submit tax return by April 15" |
| **Ranged** | Can happen within a date range | "Paint bedroom this week" |
| **Floating** | Do when free, no deadline | "Organize photo library" |
| **Recurring** | Repeats on a schedule | "Water plants every 3 days" |
| **Habit** | Tracked for streaks/consistency | "Read for 30 minutes daily" |

### 2.3 Recurrence Patterns

BentoTask supports rich recurrence:

- **Fixed interval**: Every N days/weeks/months
- **Day-of-week**: Every Monday, Wednesday, Friday
- **Day-of-month**: The 1st and 15th of every month
- **Relative**: Every 3rd Thursday of the month
- **After completion**: 2 weeks after last completion (e.g., haircut)
- **Custom cron**: For power users, cron-like expressions

---

## 3. Features

### 3.1 Core Task Management

- **FR-001**: Create, read, update, delete tasks with title, description, priority, energy level, estimated duration, and tags
- **FR-002**: Assign tasks to boxes (containers/projects/areas)
- **FR-003**: Nest tasks within tasks (subtasks, arbitrary depth)
- **FR-004**: Link tasks to each other with typed relationships:
  - `depends-on` / `blocks` (task B cannot start until task A completes)
  - `related-to` (informational link)
  - `parent-of` / `child-of` (nesting)
- **FR-005**: Set task status: `pending`, `active`, `paused`, `done`, `cancelled`, `waiting`
- **FR-006**: Support all task types from section 2.2
- **FR-007**: Support all recurrence patterns from section 2.3
- **FR-008**: Track time spent on tasks (optional timer/stopwatch)

### 3.2 Habit Tracking

- **FR-010**: Define habits with target frequency (daily, 3x/week, etc.)
- **FR-011**: Track completion with timestamps
- **FR-012**: Calculate and display streaks (current, longest, total completions)
- **FR-013**: Show completion rate over configurable periods (week, month, year)
- **FR-014**: Visual indicators (heatmaps, charts, trend lines)
- **FR-015**: Allow habits to be part of routines

### 3.3 Routines

- **FR-020**: Create routines as ordered sequences of tasks and habits
- **FR-021**: "Play mode" — step through a routine one item at a time (ideal for smart mirror/display)
- **FR-022**: Track routine completion time and individual step durations
- **FR-023**: Morning/evening/weekly routine templates
- **FR-024**: Routine can auto-schedule at specific times or be triggered manually

### 3.4 Smart Scheduling — "What Should I Do Now?"

- **FR-030**: User declares available time and current energy level (low/medium/high)
- **FR-031**: System suggests an optimal set of tasks using a scoring algorithm (see Section 6)
- **FR-032**: Factors considered:
  - Task priority and urgency (deadline proximity)
  - Estimated duration vs. available time (bin packing)
  - Energy required vs. energy available
  - Dependencies (only suggest tasks whose dependencies are met)
  - Context (location, tools available — home vs. office vs. commute)
  - Habit streaks at risk (prioritize habits about to break streak)
  - Task age (older floating tasks get boosted over time)
- **FR-033**: User can accept, skip, or snooze suggestions
- **FR-034**: Learn from user behavior over time (which suggestions are accepted/skipped)

### 3.5 Views

- **FR-040**: **Inbox** — unsorted/uncategorized tasks
- **FR-041**: **Today** — what's scheduled or suggested for today
- **FR-042**: **Calendar** — tasks on a timeline/calendar view
- **FR-043**: **Kanban** — tasks as cards in columns (by status or custom)
- **FR-044**: **Routine Player** — step-by-step view for executing routines
- **FR-045**: **Habits Dashboard** — streaks, heatmaps, statistics
- **FR-046**: **Focus Mode** — single task, timer, no distractions
- **FR-047**: **Smart Mirror View** — minimal, high-contrast, glanceable (current routine step, next task, weather/time)

### 3.6 Search & Filters

- **FR-050**: Full-text search across tasks, notes, and linked documents
- **FR-051**: Filter by: tags, priority, energy, duration, status, box, date range
- **FR-052**: Saved filters / smart lists
- **FR-053**: Natural language queries (future: AI-powered)

---

## 4. Architecture

### 4.1 Principles

- **Local-first**: All data stored locally. No account required. App works offline.
- **Plain files**: Data stored as Markdown + YAML frontmatter (human-readable, git-friendly)
- **Portable**: No vendor lock-in. Export/import easily.
- **Extensible**: Plugin/extension system for future integrations
- **Privacy-respecting**: No telemetry. Data stays on device unless user opts into sync.

### 4.2 High-Level Architecture

```
+---------------------------------------------------+
|                    Interfaces                      |
|  [CLI]  [Web UI]  [Native App]  [Mirror Display]  |
+---------------------------------------------------+
|                    API Layer                       |
|            (REST / GraphQL / IPC)                  |
+---------------------------------------------------+
|                   Core Engine                      |
|  [Task Manager] [Habit Tracker] [Scheduler]       |
|  [Routine Engine] [Search/Index] [Link Graph]     |
+---------------------------------------------------+
|                  Storage Layer                     |
|     [Markdown Files]  [SQLite Index/Cache]        |
|     [File Watcher]    [Sync Engine (future)]      |
+---------------------------------------------------+
|                  Integrations                      |
|  [CalDAV] [Google Cal] [Reminders] [Webhooks]     |
+---------------------------------------------------+
```

### 4.3 Storage Design

```
bentotask/
  data/
    inbox/
      buy-birthday-gift.md
    projects/
      home-renovation/
        _project.md          # Project metadata
        paint-bedroom.md
        fix-kitchen-sink.md
    routines/
      morning-routine.md     # Contains ordered list of steps
    habits/
      reading.md             # Habit definition + completion log
      exercise.md
    areas/
      health/
        _area.md
      work/
        _area.md
    notes/                   # Knowledge base (future)
      meeting-notes-2026-04-05.md
    .bentotask/
      config.yaml            # User configuration
      index.db               # SQLite index for fast search
      state.json             # App state (UI preferences, etc.)
```

### 4.4 File Format

Each task is a Markdown file with YAML frontmatter:

```markdown
---
id: 01JQXYZ123456
title: Paint bedroom
type: ranged
status: pending
priority: medium
energy: high
estimated_duration: 180  # minutes
due_start: 2026-04-07
due_end: 2026-04-13
tags: [home, renovation]
box: projects/home-renovation
links:
  - type: depends-on
    target: 01JQXYZ123400  # "Buy paint" task
  - type: related-to
    target: 01JQXYZ123499  # "Choose paint color" task
recurrence: null
created: 2026-04-05T10:30:00Z
updated: 2026-04-05T10:30:00Z
---

# Paint Bedroom

Need to repaint the bedroom walls. Going with the sage green that Maria picked.

## Notes
- Remove furniture first or cover with plastic
- Need painter's tape for the trim
- Two coats minimum

## Checklist
- [ ] Buy supplies
- [ ] Move furniture
- [ ] First coat
- [ ] Second coat
- [ ] Clean up
```

Habit file example:

```markdown
---
id: 01JQXYZ200000
title: Read for 30 minutes
type: habit
frequency:
  type: daily
  target: 1
energy: low
estimated_duration: 30
tags: [learning, personal-growth]
streak:
  current: 12
  longest: 45
  total_completions: 234
created: 2026-01-01T00:00:00Z
---

# Reading Habit

Daily reading goal: 30 minutes of non-fiction or technical books.

## Log
- 2026-04-05: 35min - "Designing Data-Intensive Applications" ch.7
- 2026-04-04: 30min - "Designing Data-Intensive Applications" ch.6
- 2026-04-03: 40min - "Designing Data-Intensive Applications" ch.6
```

---

## 5. Data Model

### 5.1 Entity Relationships

```
                    +-----------+
                    |   Box     |
                    | (project/ |
                    |  area)    |
                    +-----+-----+
                          |
                    contains (1:N)
                          |
+----------+        +-----v-----+        +----------+
|  Routine | -----> |   Task    | <----> |   Task   |
| (ordered |contains| (core     |  links | (another)|
|  sequence)|       |  entity)  |        |          |
+----------+        +-----+-----+        +----------+
                          |
                     extends
                          |
                    +-----v-----+
                    |   Habit   |
                    | (task +   |
                    |  tracking)|
                    +-----------+
```

### 5.2 Core Fields

**Task**:
- `id`: ULID (sortable, unique)
- `title`: string (required)
- `description`: markdown body
- `type`: enum (one-shot, dated, ranged, floating, recurring, habit)
- `status`: enum (pending, active, paused, done, cancelled, waiting)
- `priority`: enum (none, low, medium, high, urgent)
- `energy`: enum (low, medium, high)
- `estimated_duration`: integer (minutes)
- `due_date` / `due_start` / `due_end`: ISO date/datetime
- `recurrence`: recurrence rule object (nullable)
- `tags`: string[]
- `box`: path reference to container
- `links`: array of {type, target_id}
- `context`: enum[] (home, office, errands, commute, anywhere)
- `created` / `updated` / `completed`: ISO datetime

**Habit** (extends Task):
- `frequency`: {type, target, period}
- `streak`: {current, longest, total_completions}
- `completions`: array of {date, duration, notes}

**Routine**:
- `id`: ULID
- `title`: string
- `steps`: ordered array of task/habit references
- `schedule`: when to trigger (time-of-day, day-of-week, manual)
- `estimated_total_duration`: computed from steps

**Link**:
- `type`: enum (depends-on, blocks, related-to, parent-of, child-of)
- `source_id`: ULID
- `target_id`: ULID

---

## 6. Scheduling Algorithm

### 6.1 The Bento Packing Algorithm

The core "What Should I Do Now?" feature uses a variant of the **weighted knapsack problem**:

**Inputs**:
- Available time `T` (minutes)
- Current energy level `E` (1-3 scale)
- Current context `C` (home, office, etc.)
- Set of eligible tasks (status=pending, dependencies met, matches context)

**Scoring function** for each task `t`:

```
score(t) = w1 * urgency(t)
         + w2 * priority(t)
         + w3 * energy_match(t, E)
         + w4 * streak_risk(t)       # habits about to break streak
         + w5 * age_boost(t)         # older floating tasks
         + w6 * dependency_unlock(t) # tasks that unblock others
         + w7 * user_preference(t)   # learned from accept/skip history
```

**Packing strategy**:

1. Filter tasks: only those matching context `C` and energy `<= E`
2. Filter tasks: only those with `estimated_duration <= T`
3. Score all eligible tasks
4. Use **greedy knapsack** (sort by score/duration ratio, pack until full)
5. For remaining time gaps, try to fit smaller tasks (First Fit Decreasing)
6. Return ordered suggestion list

### 6.2 Default Weights

```yaml
weights:
  urgency: 0.25        # Deadline proximity
  priority: 0.20       # User-assigned priority
  energy_match: 0.15   # How well task energy matches current energy
  streak_risk: 0.15    # Habit streak about to break
  age_boost: 0.10      # Older tasks float up
  dependency_unlock: 0.10  # Completing this unblocks other tasks
  user_preference: 0.05    # Learned from history
```

### 6.3 Urgency Calculation

```
urgency(t) =
  if t.due_date is today:       1.0
  if t.due_date is tomorrow:    0.8
  if t.due_date is within 3d:   0.6
  if t.due_date is within 7d:   0.4
  if t.due_date is within 30d:  0.2
  if t.type is floating:        0.1 + age_factor
  else:                         0.0
```

---

## 7. Interfaces

### 7.1 CLI (Primary — Phase 1)

```bash
# Task operations
bento add "Buy groceries" --priority high --energy low --duration 45 --tags errands
bento add "Paint bedroom" --type ranged --start 2026-04-07 --end 2026-04-13 --energy high
bento done <task-id>
bento list                       # Show today's tasks
bento list --box work            # Filter by box
bento list --tag errands         # Filter by tag

# Habits
bento habit add "Read 30 minutes" --frequency daily
bento habit log reading           # Log today's completion
bento habit stats                 # Show streaks and stats

# Routines
bento routine create "Morning Routine" --steps meditate,exercise,review-inbox
bento routine play morning        # Step-by-step execution mode

# Smart scheduling
bento now                         # "What should I do now?"
bento now --time 45 --energy low  # "I have 45 min and low energy"
bento plan today                  # Generate today's schedule

# Search
bento search "paint"
bento filter --priority high --status pending
```

### 7.2 Web UI (Phase 2)

- Modern, responsive SPA
- Views: Inbox, Today, Calendar, Kanban, Habits, Routine Player
- Drag-and-drop task organization
- Dark mode / light mode
- PWA for mobile use

### 7.3 Smart Mirror / Display View (Phase 3)

- Minimal, high-contrast interface
- Shows: current routine step, next 3 tasks, habit streaks at risk
- Large text, glanceable from distance
- Auto-advance through routines
- Optional: weather, time, calendar events sidebar
- Accessible via any browser (Raspberry Pi + screen, tablet, etc.)

### 7.4 Native App (Phase 4 — Future)

- Consider: Tauri (Rust + Web), Flutter, or React Native
- System tray integration for reminders
- Native notifications
- Keyboard shortcuts

---

## 8. Integration Points

### 8.1 Calendar Integration

- **CalDAV** (Apple Calendar, Nextcloud, etc.):
  - Export tasks with due dates as calendar events
  - Import calendar events as time blocks (for scheduling around them)
  - Two-way sync for tasks with specific times

- **Google Calendar API**:
  - Same as CalDAV but via REST API
  - OAuth2 authentication

### 8.2 Reminders & Notifications

- **System notifications** (native or via ntfy.sh for self-hosted)
- **Email reminders** (optional, via SMTP)
- **Webhook triggers** (for automation with n8n, Home Assistant, etc.)
- **Apple Reminders** (via EventKit / Shortcuts — future)

### 8.3 Import/Export

- **Import from**: Todoist (JSON/CSV), Taskwarrior, Notion (MD export), Things 3 (JSON), OPML
- **Export to**: Markdown, JSON, CSV, iCal (.ics)
- **Sync**: Git-based sync (push/pull), Syncthing-compatible folder, future: CRDTs for conflict-free sync

---

## 9. Future Extensibility

### 9.1 Knowledge Base / Second Brain (Phase 5)

- Notes linked to tasks and projects
- Document attachments (PDFs, images)
- Concept mapping / knowledge graph
- Bi-directional links between notes
- Daily journal with task/habit summary auto-generated

### 9.2 AI Integration (Phase 6)

- **Natural language task creation**: "Remind me to call the dentist next Tuesday"
- **Smart categorization**: Auto-suggest tags, priority, energy level, duration
- **Task breakdown**: AI suggests subtasks for complex tasks
- **Schedule optimization**: AI-powered scheduling with more context
- **MCP tool integration**: Expose BentoTask as MCP tools for AI agents
- **AI skills**: Pluggable AI capabilities (summarize notes, suggest related tasks)

### 9.3 Plugin / Extension System

```yaml
# .bentotask/extensions/pomodoro.yaml
name: pomodoro
version: 1.0.0
hooks:
  on_task_start: start_pomodoro_timer
  on_task_complete: log_pomodoro_count
commands:
  - name: pomodoro
    description: Start a Pomodoro session for the current task
```

- **Event hooks**: on_task_create, on_task_complete, on_habit_log, on_routine_start, etc.
- **Custom commands**: Extensions can register CLI commands
- **Custom views**: Extensions can add web UI panels
- **MCP server**: Expose BentoTask data/actions as MCP tools

### 9.4 Collaboration (Phase 7 — Far Future)

- Shared boxes/projects
- Task assignment
- Comments/discussion on tasks
- CRDT-based real-time sync

---

## 10. Prior Art & Differentiation

### 10.1 Existing Tools Evaluated

| Tool | Strengths | Why BentoTask is Different |
|------|-----------|---------------------------|
| **Taskwarrior** | Powerful CLI, dependencies, urgency scoring | No habits, no routines, no smart scheduling, steep learning curve |
| **Vikunja** | CalDAV, web UI, recurring tasks | No habits, no "what to do now", no knowledge base |
| **Logseq** | Knowledge graph, local-first, markdown | Task management is basic, no habits, no scheduling |
| **Loop Habit Tracker** | Best FOSS habit tracking | Android only, no tasks, no routines |
| **Todoist** | Polished UX, natural language | Proprietary, no habits, no smart scheduling, no local-first |
| **Notion** | Flexible, databases, notes | Proprietary, slow, complex, not designed for habits/routines |
| **Obsidian** | Plugins, local markdown | Not truly open source, task management via plugins is fragile |

### 10.2 BentoTask's Unique Value

1. **Unified**: Tasks + Habits + Routines in one system (no one else does this well in FOSS)
2. **Smart Scheduling**: Knapsack-inspired "pack your day" algorithm
3. **Routine Player**: Step-by-step execution mode, perfect for smart mirrors
4. **Local-first + Plain Files**: Own your data, git-friendly, no lock-in
5. **Extensible**: Plugin system, MCP integration, AI-ready
6. **Bento Model**: Intuitive nested container metaphor

---

## 11. Technical Stack

### 11.1 Recommended Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Core engine** | Rust | Performance, reliability, single binary, cross-platform |
| **CLI** | Rust (clap) | Native CLI built into the same binary |
| **Storage** | Markdown + YAML frontmatter | Human-readable, git-friendly, portable |
| **Index/Cache** | SQLite (via rusqlite) | Fast search, computed views, no external dependency |
| **Web UI** | SvelteKit or Solid.js | Lightweight, fast, modern |
| **API** | REST (via axum or actix) | Simple, well-understood, easy to integrate |
| **Smart Mirror** | Minimal HTML/CSS/JS | Works on any browser, Raspberry Pi compatible |
| **Calendar sync** | CalDAV client library | Standard protocol, works with Apple/Google/Nextcloud |
| **IDs** | ULID | Sortable, unique, URL-safe |

### 11.2 Alternative Stack (if faster iteration preferred)

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Core engine** | TypeScript (Deno or Bun) | Faster development, shared code with web UI |
| **CLI** | TypeScript (commander.js) | Same language as core |
| **Storage** | Same (MD + YAML) | Same benefits |
| **Index** | SQLite (better-sqlite3) | Same benefits |
| **Web UI** | SvelteKit | Same |

---

## 12. Development Phases

### Phase 1: Foundation (MVP)
**Goal**: Functional CLI with core task management

- [ ] Project scaffolding and build system
- [ ] Data model implementation (task CRUD)
- [ ] Markdown file read/write with YAML frontmatter
- [ ] SQLite indexing for fast queries
- [ ] Basic CLI: add, list, done, edit, delete
- [ ] Task types: one-shot, dated, floating
- [ ] Tags and priority
- [ ] Basic search and filtering

### Phase 2: Habits & Routines
**Goal**: Habit tracking and routine execution

- [ ] Habit definition and logging
- [ ] Streak calculation and statistics
- [ ] Routine creation and step sequencing
- [ ] Routine "play mode" in CLI
- [ ] Recurring tasks (all patterns)
- [ ] Task linking (dependencies and relations)

### Phase 3: Smart Scheduling
**Goal**: The "What should I do now?" feature

- [ ] Bento Packing Algorithm implementation
- [ ] Energy level and context support
- [ ] `bento now` command
- [ ] `bento plan today` command
- [ ] Urgency scoring
- [ ] Dependency-aware suggestions

### Phase 4: Web UI
**Goal**: Visual interface for all features

- [ ] API server (REST)
- [ ] Inbox, Today, Calendar, Kanban views
- [ ] Habits dashboard with heatmaps
- [ ] Routine player (visual step-through)
- [ ] Smart mirror / display view
- [ ] Drag-and-drop organization
- [ ] PWA support

### Phase 5: Integrations
**Goal**: Connect with the outside world

- [ ] CalDAV sync (Apple Calendar, Nextcloud)
- [ ] Google Calendar API integration
- [ ] System notifications / reminders
- [ ] Import from Todoist, Taskwarrior, Notion
- [ ] Export to iCal, JSON, CSV
- [ ] Git-based sync between devices

### Phase 6: Knowledge Base
**Goal**: Second brain capabilities

- [ ] Notes system with bi-directional links
- [ ] Document and image attachments
- [ ] Knowledge graph visualization
- [ ] Link notes to tasks/projects
- [ ] Daily journal auto-generation

### Phase 7: AI & Extensions
**Goal**: Intelligence and extensibility

- [ ] Plugin/extension system
- [ ] MCP server (expose as AI tools)
- [ ] Natural language task creation
- [ ] AI-powered scheduling optimization
- [ ] Smart categorization and suggestions

---

## Non-Functional Requirements

### Performance
- **NFR-001**: CLI commands respond in < 100ms for datasets up to 10,000 tasks
- **NFR-002**: Web UI loads in < 2 seconds on localhost
- **NFR-003**: SQLite index rebuilds in < 5 seconds for 10,000 tasks

### Reliability
- **NFR-010**: No data loss — files are the source of truth
- **NFR-011**: Graceful handling of corrupted/malformed files
- **NFR-012**: Atomic file writes (write to temp, then rename)

### Usability
- **NFR-020**: CLI follows Unix conventions (pipes, exit codes, quiet mode)
- **NFR-021**: Helpful error messages with suggested fixes
- **NFR-022**: Comprehensive `--help` for all commands
- **NFR-023**: Tab completion for shells (bash, zsh, fish)

### Security & Privacy
- **NFR-030**: No network calls unless explicitly configured (sync, calendar)
- **NFR-031**: No telemetry or analytics
- **NFR-032**: Optional encryption for sync
- **NFR-033**: Sensitive data (API keys) stored in system keychain, not plain files

---

## Open Questions

1. **Rust vs TypeScript?** — Rust for performance and single-binary distribution, or TypeScript for faster iteration and shared web code?
2. **File-per-task vs single file?** — Individual .md files (better for git, but many files) vs. a single structured file per box?
3. **CRDT sync strategy** — Which CRDT library/approach for eventual multi-device sync?
4. **Mobile strategy** — PWA sufficient, or do we need a native mobile app?
5. **Smart mirror protocol** — Simple HTTP polling, WebSocket push, or MQTT for IoT integration?

---

*This spec is a living document. As BentoTask evolves, so will this specification.*
