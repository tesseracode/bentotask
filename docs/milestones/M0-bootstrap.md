# Milestone 0: Project Bootstrap

**Status**: In Progress  
**Started**: 2026-04-05  
**Target**: 2026-04-12  

---

## Objective

Establish the project foundation: make key architecture decisions, set up the development environment, and create the initial project scaffolding.

---

## Tasks

### Decision Making
- [x] Write initial SPEC.md with full feature requirements
- [x] Back up original spec as SPEC.original.md
- [x] Create project tracking structure (roadmap, ADRs, milestones)
- [x] **ADR-001**: Choose programming language & tech stack — **APPROVED**
  - [x] Research and document options (Go, Rust, TS, Python, Elixir, FP langs)
  - [x] Create comparison matrix
  - [x] Write proposed decision
  - [x] Review and approve decision (2026-04-05)
- [x] Create CLAUDE.md (agent orientation)
- [x] Create AGENTS.md (roles, handoff protocol, supervisor protocol)
- [x] Set up handoff system (docs/handoff/)
- [x] Set up supervisor system (docs/supervisor/)
- [x] **ADR-002**: Storage format & indexing strategy — **APPROVED**
  - [x] Research Go libraries (frontmatter, SQLite, ULID, RRULE)
  - [x] Finalize markdown + YAML frontmatter format
  - [x] Define SQLite schema for index
  - [x] Decide on file-per-task vs bundled files (file-per-task, ULID filenames)
  - [x] Define index sync strategy (hybrid mtime + hash)
  - [x] Write proposed decision
  - [x] Review and approve decision (2026-04-05)
- [x] **ADR-003**: CLI framework & UX patterns — **APPROVED**
  - [x] Research CLI UX patterns (gh, kubectl, Taskwarrior, Charm ecosystem)
  - [x] Command structure and naming conventions (noun-verb + aliases)
  - [x] Output formatting (text, --json, --quiet)
  - [x] Interactive mode decisions (bubbletea for sessions, huh for forms, plain for CRUD)
  - [x] Color/styling, editor integration, shell completions, error handling
  - [x] Write proposed decision
  - [x] Review and approve decision (2026-04-05)

### Project Setup (after ADRs approved)
- [x] Initialize Go module (`go mod init github.com/tesserabox/bentotask`)
- [x] Set up folder structure per ADR-001
- [x] Configure linting (golangci-lint)
- [x] Set up test framework
- [x] Create Makefile / Taskfile
- [x] Write first test (proof of life)
- [x] Set up CI (GitHub Actions)
- [x] Create .gitignore
- [x] Create CONTRIBUTING.md

---

## Acceptance Criteria

- [x] All three ADRs are written and approved
- [x] `go build ./cmd/bt` compiles successfully
- [x] `bt --version` prints version info
- [x] `bt --help` shows command structure
- [x] At least one passing test
- [x] CI runs on push

---

## Notes

- Keep scope tight — this milestone is about decisions and scaffolding, not features
- Each ADR should be reviewed before moving to implementation
- The SPEC.md may be updated based on ADR outcomes
