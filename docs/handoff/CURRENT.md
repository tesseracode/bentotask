# Current Handoff

## Active Task
- **Task ID**: M6 Group D (Ship It)
- **Milestone**: M6 — REST API & Web UI
- **Description**: go:embed single binary, smart mirror, routine player
- **Status**: COMPLETE
- **Assigned**: 2026-04-08

## Last Session Summary
- **Sessions 1–13**: M0–M5 + M6 Groups A/B/C complete
- **Session 14**: M6 Group D COMPLETE — Ship It
  - Part 1 (M6.11): go:embed — SvelteKit static adapter + embedded in Go binary
    - `make build` runs full pipeline: web build → copy to static/ → Go compile
    - `./bt serve` serves web UI + API from single binary at localhost:7878
    - SPA fallback: all client-side routes work on refresh
    - Placeholder index.html committed so `go build` works without npm
  - Part 2 (M6.10): Smart mirror — full-screen, high-contrast display
    - Own layout (no sidebar), pure black bg, large clock, top 3 tasks, habit streaks
    - Auto-refreshes every 60 seconds, clock ticks every second
  - Part 3 (M6.9): Routine player — step-through with timers
    - Routine list at /routines, player at /routines/[id]
    - Play mode: step highlighting, elapsed timer, progress bars, Next/Skip/Stop
    - Summary on completion with total time

## Current State
- **M0–M5 COMPLETE**, **M6: 10 of 12 items complete**
- Remaining: M6.6 (Calendar view), M6.7 (Kanban view)
- Go backend: 315 tests, 0 lint issues
- Web UI: 0 svelte-check errors, builds with adapter-static
- Single binary: `make build` → `./bt serve` serves everything
- Theme: Bento Alt (dark) applied via CSS custom properties
- Clay Alt (light) archived in theme.css for future toggle

## Next Steps
- M6.6: Calendar view — tasks on timeline/calendar
- M6.7: Kanban view — tasks as cards in status columns
- M7: Sync & external integrations
- Future: Dark/light mode toggle, design polish

## Blockers
- None
