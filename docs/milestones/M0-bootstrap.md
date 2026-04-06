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
- [ ] **ADR-002**: Storage format & indexing strategy
  - [ ] Finalize markdown + YAML frontmatter format
  - [ ] Define SQLite schema for index
  - [ ] Decide on file-per-task vs bundled files
- [ ] **ADR-003**: CLI framework & UX patterns
  - [ ] Command structure and naming conventions
  - [ ] Output formatting (table, JSON, plain)
  - [ ] Interactive mode decisions (bubbletea usage)

### Project Setup (after ADRs approved)
- [ ] Initialize Go module (`go mod init`)
- [ ] Set up folder structure per ADR-001
- [ ] Configure linting (golangci-lint)
- [ ] Set up test framework
- [ ] Create Makefile / Taskfile
- [ ] Write first test (proof of life)
- [ ] Set up CI (GitHub Actions)
- [x] Create .gitignore
- [ ] Create CONTRIBUTING.md

---

## Acceptance Criteria

- [ ] All three ADRs are written and approved
- [ ] `go build ./cmd/bt` compiles successfully
- [ ] `bt --version` prints version info
- [ ] `bt --help` shows command structure
- [ ] At least one passing test
- [ ] CI runs on push

---

## Notes

- Keep scope tight — this milestone is about decisions and scaffolding, not features
- Each ADR should be reviewed before moving to implementation
- The SPEC.md may be updated based on ADR outcomes
