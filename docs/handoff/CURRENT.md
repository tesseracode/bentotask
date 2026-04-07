# Current Handoff

## Active Task
- **Task ID**: M5 (Smart Scheduling)
- **Milestone**: M5 — Smart Scheduling
- **Description**: Bento Packing Algorithm, urgency scoring, energy matching, bt now, bt plan today
- **Status**: Planning Complete, Implementation Next
- **Assigned**: 2026-04-07

## Last Session Summary
- **Session 1–3 (2026-04-05–06)**: M0 + M1 + M2.1–M2.9
- **Session 4 (2026-04-06)**: M2.10–M2.12 — Tab completions, integration tests, --json
- **Session 5 (2026-04-07)**: M3 COMPLETE — Habits & Recurrence
- **Session 6 (2026-04-07)**: Bug fixes from M3 review + M4 Group A (Routines)
- **Session 7 (2026-04-07)**: M4 Group B (Linking) + fixes from reviews
- **Session 8 (2026-04-07)**: M4 closure — fixed flaky TestNextAfterDaily, marked M4.7 complete, planned M5
  - Fixed flaky recurrence tests by adding SetDTStart method and pinning DTSTART in tests
  - Also fixed TestNextAfterWeeklyMWF, TestBetween, TestBetweenWeekly (same root cause)
  - Marked M4 COMPLETE in ROADMAP (all 7 tasks checked)
  - Planned M5 in grouped subtasks below

## Current State
- **M0 COMPLETE**, **M1 COMPLETE**, **M2 COMPLETE**, **M3 COMPLETE**, **M4 COMPLETE**
- Module: `github.com/tesserabox/bentotask`
- `make test`: **215 tests** — 0 lint issues, 0 vet issues
- Binary: `make build` — up to date

## M5 Plan — Smart Scheduling

### Group A: Scoring Engine (M5.1 + M5.2 + M5.3 + M5.4)
Core scoring functions — no CLI yet, just the `internal/engine/` package.

- **M5.1: Urgency scoring** — `urgency(t) → float64` per spec §6.3
  - Due today → 1.0, tomorrow → 0.8, 3d → 0.6, 7d → 0.4, 30d → 0.2
  - Floating tasks get 0.1 + age_factor
  - No due date → 0.0
- **M5.2: Priority scoring** — `priority(t) → float64`
  - urgent → 1.0, high → 0.75, medium → 0.5, low → 0.25, none → 0.0
- **M5.3: Energy matching** — `energyMatch(t, E) → float64`
  - Exact match → 1.0, one level below → 0.5, two levels below → 0.2
  - Filter: tasks with energy > E are excluded (done in packing, not scoring)
- **M5.4: Streak risk detection** — `streakRisk(t) → float64`
  - Daily habit not completed today → 1.0
  - Daily habit completed today → 0.0
  - Weekly habit: (target - completions this week) / target, boosted if deadline approaching
  - Non-habits → 0.0

### Group B: Algorithm + Filters (M5.1 continued + M5.5 + M5.6)
The packing algorithm that combines scores and produces suggestions.

- **M5.5: Bento Packing Algorithm** — per spec §6.1
  - Filter: context match, energy ≤ E, duration ≤ T, dependencies met
  - Score each task using weighted sum (spec §6.2 default weights)
  - Greedy knapsack: sort by score/duration ratio, pack until full
  - For remaining time gaps, fit smaller tasks (First Fit Decreasing)
  - Returns ordered suggestion list
- **M5.6: Age boost + dependency unlock sub-functions**
  - `ageBoost(t)`: logarithmic growth based on days since creation, cap at 1.0
  - `dependencyUnlock(t)`: count of tasks blocked by t, normalized

### Group C: CLI Commands (M5.7 + M5.8)
Wire the engine to CLI.

- **M5.7: `bt now` command** — "What should I do now?"
  - `bt now` — default 60 min, medium energy, any context
  - `bt now --time 45 --energy low --context home`
  - Shows top N suggestions with score breakdown
  - `--json` output mode
- **M5.8: `bt plan today` command** — Generate today's schedule
  - Takes total available time (default configurable)
  - Packs tasks into a day plan ordered by time
  - Shows a time-blocked plan view

### Group D: Tests + Tuning (M5.9)
- **M5.9: Tests and benchmarks**
  - Unit tests for each scoring function
  - Integration tests for bt now / bt plan today
  - Algorithm benchmarks for tuning weights

### Implementation Notes
- New package: `internal/engine/` — scoring functions + packing algorithm
- Engine depends on: `model` (Task, Priority, Energy, LinkType), `store` (queries)
- Engine does NOT depend on: `cli`, `habit`, `recurrence`
- The `user_preference` factor (spec §6.1, w7) is deferred — requires accept/skip history storage
- Context is already supported on tasks (model.Task.Context field, task_contexts table)
- Dependencies already exist via task_links (depends-on/blocks) from M4

## Blockers
- None

## Context for Next Agent
- SPEC.md §3.4, §6.1–§6.3 define the algorithm in detail
- Default weights in spec §6.2: urgency=0.25, priority=0.20, energy=0.15, streak_risk=0.15, age=0.10, dep_unlock=0.10, user_pref=0.05
- `user_preference` (w7=0.05) is deferred — no accept/skip history table yet
- Energy is a 3-level enum: low=1, medium=2, high=3
- Habit completions are in `habit.ParseCompletionsFromBody()` (markdown SOT)
- Streak calculation is in `habit.CalculateStreak()`
- Task links are in `store.DependencyGraph()` and `app.GetTaskLinks()`
- Start with `internal/engine/score.go` for the scoring functions
- Key files: SPEC.md §6, internal/model/task.go, internal/app/app.go, internal/store/index.go
