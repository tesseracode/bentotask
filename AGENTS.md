# AGENTS.md — Agent Roles, Protocols & Coordination

> Defines how AI agents work on BentoTask, how they hand off work,
> and how progress is supervised and validated.

---

## Agent Roles

### 1. Worker Agent

**Purpose**: Implements features, writes code, writes tests, fixes bugs.

**Capabilities**:
- Read/write code files
- Run commands (build, test, lint)
- Read project docs for context
- Create commits

**Workflow**:
1. Read `docs/handoff/CURRENT.md` to understand the active task
2. Read the relevant milestone doc in `docs/milestones/`
3. Read any relevant ADRs
4. Implement the task
5. Write/run tests
6. Update `docs/handoff/CURRENT.md` with progress
7. If task is complete, mark it done in handoff and milestone doc
8. If context is running out or connection may fail, write a detailed handoff

**Rules**:
- Work on ONE task at a time
- Commit at logical checkpoints (not just at the end)
- Never skip tests
- If a task is too large, break it down and document subtasks in the handoff
- If blocked, stop and document the blocker — don't guess

---

### 2. Supervisor Agent

**Purpose**: Reviews progress, validates completions, plans next steps, maintains quality.

**Capabilities**:
- Read all project files
- Review code changes (git diff)
- Run tests and check results
- Update roadmap and milestone docs
- Assign next tasks via handoff file

**Workflow**:
1. Read `docs/supervisor/LOG.md` for previous reviews
2. Read `docs/handoff/CURRENT.md` to see what the worker reported
3. Review the actual code changes (`git log`, `git diff`)
4. Run tests to verify claimed completions
5. Update milestone checkboxes in `docs/milestones/` and `docs/ROADMAP.md`
6. Log review findings in `docs/supervisor/LOG.md`
7. Determine next task and update `docs/handoff/CURRENT.md` with new assignment

**Review Checklist**:
- [ ] Does the code compile? (`go build ./...`)
- [ ] Do tests pass? (`go test ./...`)
- [ ] Is the code idiomatic Go?
- [ ] Are there edge cases not covered?
- [ ] Does the implementation match the spec?
- [ ] Is the handoff file accurate?
- [ ] What should be done next?

---

### 3. Solo Agent (Current Mode)

**Purpose**: Performs both Worker and Supervisor roles in a single session.

When only one agent is working (current setup), it operates as a Solo Agent:
- Implements features (Worker hat)
- Reviews its own work (Supervisor hat)
- Updates all tracking docs
- Plans next steps

The handoff and supervisor systems still exist so that:
- Work can resume after a disconnection
- A future multi-agent setup can take over seamlessly
- Progress is always tracked and never lost

---

## Handoff Protocol

### When to Write a Handoff

A handoff entry MUST be written when:
1. A task is **completed** (mark done, note what was accomplished)
2. A session is **ending** (context limit approaching, user signing off)
3. An agent is **blocked** (dependency, question, decision needed)
4. A task is **partially done** (save progress, describe remaining work)

### Handoff File Format

The handoff file lives at `docs/handoff/CURRENT.md`. Its format:

```markdown
# Current Handoff

## Active Task
- **Task ID**: M1.2
- **Milestone**: M1 — Core Data Model
- **Description**: Implement Markdown + YAML frontmatter reader/writer
- **Status**: In Progress | Complete | Blocked
- **Assigned**: 2026-04-06T10:00:00Z

## Last Session Summary
- What was done (bullet points)
- Files created/modified
- Tests written/passing

## Current State
- What works right now
- What's partially implemented
- Known issues or bugs

## Next Steps
1. Specific next action
2. Second action
3. etc.

## Blockers (if any)
- Description of blocker
- What's needed to unblock

## Context for Next Agent
- Any non-obvious decisions made and why
- Gotchas or tricky parts of the code
- Relevant file paths to start with
```

### Handoff History

When a handoff is superseded (task changes), the previous handoff is appended to `docs/handoff/HISTORY.md` with a timestamp. This creates an audit trail.

---

## Supervisor Protocol

### Supervisor Log Format

The supervisor log lives at `docs/supervisor/LOG.md`:

```markdown
# Supervisor Log

## [2026-04-06] Review: M1.2 — Markdown Reader/Writer

**Reviewed by**: Supervisor Agent
**Handoff claimed**: Task complete
**Verification**:
- [x] Code compiles
- [x] Tests pass (5/5)
- [ ] Edge case: empty frontmatter not handled
**Verdict**: NEEDS REVISION
**Action**: Reassigned to worker with note to handle empty frontmatter
**Next task**: M1.2 (revision), then M1.3

---
```

### Review Verdicts

| Verdict | Meaning | Action |
|---------|---------|--------|
| **APPROVED** | Task meets requirements, tests pass | Mark complete in roadmap, assign next task |
| **APPROVED WITH NOTES** | Acceptable but has minor issues | Mark complete, log improvement suggestions for later |
| **NEEDS REVISION** | Issues found that must be fixed | Reassign to worker with specific feedback |
| **BLOCKED** | Cannot validate due to external dependency | Log blocker, skip to next unblocked task |

---

## Task Lifecycle

```
                    ┌─────────┐
                    │ BACKLOG │  (defined in ROADMAP.md)
                    └────┬────┘
                         │ Supervisor assigns via handoff
                         v
                    ┌─────────┐
                    │ ASSIGNED│  (written in CURRENT.md)
                    └────┬────┘
                         │ Worker picks up
                         v
                    ┌───────────┐
                    │IN PROGRESS│  (worker is coding)
                    └────┬──────┘
                         │ Worker writes handoff
                    ┌────┴────┐
                    │         │
                    v         v
            ┌──────────┐ ┌─────────┐
            │ COMPLETE │ │ BLOCKED │
            │(claimed) │ │         │
            └────┬─────┘ └────┬────┘
                 │             │ Blocker resolved
                 │             └──> back to ASSIGNED
                 │ Supervisor reviews
            ┌────┴──────┐
            │           │
            v           v
      ┌──────────┐ ┌──────────┐
      │ APPROVED │ │ REVISION │
      │          │ │ NEEDED   │
      └────┬─────┘ └────┬─────┘
           │             └──> back to ASSIGNED
           │
           v
      ┌─────────┐
      │  DONE   │  (checked off in ROADMAP.md)
      └─────────┘
```

---

## File Ownership

| File | Who updates it |
|------|---------------|
| `SPEC.md` | Human (with agent suggestions) |
| `CLAUDE.md` | Human only |
| `AGENTS.md` | Human only |
| `docs/ROADMAP.md` | Supervisor (checkboxes, status) |
| `docs/adrs/ADR-*.md` | Worker proposes, Human approves |
| `docs/milestones/M*-*.md` | Supervisor (checkboxes, status) |
| `docs/handoff/CURRENT.md` | Worker (on every session end) |
| `docs/handoff/HISTORY.md` | Supervisor (archives old handoffs) |
| `docs/supervisor/LOG.md` | Supervisor (review entries) |
| Source code (`cmd/`, `internal/`) | Worker |
| Tests | Worker |
| `web/` | Worker |

---

## Naming Conventions

- **Milestones**: `M0`, `M1`, `M2`, ... (match ROADMAP.md)
- **Tasks**: `M1.2` = Milestone 1, task 2
- **ADRs**: `ADR-001`, `ADR-002`, ... (zero-padded, sequential)
- **Commits**: `M1.2: implement markdown frontmatter parser`
- **Branches** (future): `milestone/m1-core-data-model`, `feature/m1.2-markdown-parser`
