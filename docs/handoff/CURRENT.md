# Current Handoff

## Active Task
- **Task ID**: M0.4
- **Milestone**: M0 — Project Bootstrap
- **Description**: ADR-002 — Storage format & indexing strategy
- **Status**: Not Started
- **Assigned**: 2026-04-05

## Last Session Summary
- Created initial SPEC.md with full feature requirements
- Backed up original spec as SPEC.original.md (gitignored)
- Created project tracking structure: ROADMAP.md, ADR directory, milestone docs
- Researched and wrote ADR-001 (tech stack) — evaluated Go, Rust, TypeScript, Python, Elixir, FP languages
- **ADR-001 APPROVED**: Go core + SvelteKit web UI
- Created CLAUDE.md (agent orientation) and AGENTS.md (roles, handoff protocol, supervisor protocol)
- Created .gitignore
- Set up handoff and supervisor systems

## Current State
- Project structure is fully set up for tracking
- ADR-001 is approved — Go + SvelteKit is the tech stack
- No code exists yet — still in planning/architecture phase
- M0 is partially complete (decisions + scaffolding docs done, code scaffolding not started)

## Completed Tasks
- [x] M0.1 — Write initial SPEC.md
- [x] M0.2 — Initialize git repository
- [x] M0.3 — ADR-001: Tech stack selection (APPROVED)

## Next Steps
1. **Write ADR-002**: Storage format & indexing strategy
   - Finalize markdown + YAML frontmatter format (field names, types, validation)
   - Design SQLite schema for the index/cache
   - Decide: file-per-task vs bundled files per box
   - Define how the index stays in sync with files
2. **Write ADR-003**: CLI framework & UX patterns
   - Command naming and structure (`bt` as the binary name)
   - Output formatting (table, JSON, plain text modes)
   - Interactive mode decisions (when to use bubbletea TUI vs simple output)
3. **Project scaffolding**: `go mod init`, folder structure, Makefile, CI
4. **First code**: `bt --version` and `bt --help` working

## Blockers
- None

## Context for Next Agent
- All project docs are in `/Users/jbencardino/Documents/Proyectos/bentotask/`
- Start by reading this file, then `CLAUDE.md` for orientation
- The SPEC.md has the full feature spec — reference it for ADR-002 (Section 4.3, 4.4 cover storage)
- ADR-001 is at `docs/adrs/ADR-001-tech-stack.md` — the tech stack is decided, don't re-evaluate
- The developer wants to learn Go — keep things idiomatic
