# ADR-003: CLI Framework & UX Patterns

**Status**: APPROVED  
**Date**: 2026-04-05  
**Approved**: 2026-04-05  
**Decision Makers**: @jbencardino  
**Depends on**: ADR-001 (Go + SvelteKit — APPROVED), ADR-002 (Storage — APPROVED)

---

## Context

BentoTask's CLI (`bt`) is the primary interface for Phase 1. We need to decide:

1. **Command structure**: How are commands organized?
2. **Interactive vs plain output**: When to use TUI (BubbleTea) vs stdout?
3. **Output formatting**: How to support human and machine-readable output?
4. **Styling**: Colors, tables, error presentation
5. **Editor integration**: How to open `$EDITOR` for task bodies
6. **Shell completions**: Tab completion for commands and task IDs
7. **Error handling**: How to present errors with actionable suggestions

---

## Decisions

### 1. Command Structure: Noun-Verb with Top-Level Aliases

**Pattern**: `bt <noun> <verb> [args] [flags]`

This scales well because BentoTask has multiple resource types (tasks, habits, routines, boxes). Flat verbs would get ambiguous. This is the pattern `gh` (GitHub CLI) uses and Cobra is designed for.

#### Full Command Tree

```
bt
├── task                        (alias: t)
│   ├── add <title>             Create a task
│   ├── list                    List tasks (filtered)
│   ├── show <id>               Show task details
│   ├── edit <id>               Edit task (flags or $EDITOR)
│   ├── done <id>               Mark task complete
│   ├── delete <id>             Delete task
│   ├── link <id> <id>          Link two tasks
│   └── unlink <id> <id>        Remove link
│
├── habit                       (alias: h)
│   ├── add <title>             Create a habit
│   ├── log <id>                Log today's completion
│   ├── list                    List habits with streaks
│   ├── stats <id>              Show detailed habit stats
│   └── edit <id>               Edit habit
│
├── routine                     (alias: r)
│   ├── create <title>          Create a routine
│   ├── play <id>               Launch routine player (TUI)
│   ├── list                    List routines
│   ├── edit <id>               Edit routine steps
│   └── show <id>               Show routine details
│
├── box                         (alias: b)
│   ├── create <name>           Create a project/area
│   ├── list                    List boxes
│   ├── show <name>             Show box contents
│   └── archive <name>          Archive a box
│
├── now                         "What should I do?" (smart scheduling)
├── plan                        Generate today's plan
├── search <query>              Full-text search
├── index                       Index management
│   └── rebuild                 Rebuild SQLite index from files
├── doctor                      Check data integrity
├── server                      Start API server (for web UI)
│
├── completion                  Generate shell completions (built-in from Cobra)
└── version                     Print version info
```

#### Top-Level Aliases (Shortcuts)

For the most common operations, register top-level shortcuts:

| Shortcut | Expands to | Rationale |
|----------|-----------|-----------|
| `bt add <title>` | `bt task add <title>` | Most common operation |
| `bt list` | `bt task list` | Most common query |
| `bt done <id>` | `bt task done <id>` | Most common update |
| `bt now` | (top-level) | Core differentiator, deserves prominence |
| `bt plan` | (top-level) | Core differentiator |

#### Noun Aliases

| Full | Short | Plural |
|------|-------|--------|
| `task` | `t` | `tasks` |
| `habit` | `h` | `habits` |
| `routine` | `r` | `routines` |
| `box` | `b` | `boxes` |

All three forms work: `bt task list` = `bt t list` = `bt tasks list`

---

### 2. Interactive vs Plain Output

**Rule**: BubbleTea TUI is for **sessions** (multi-step, live-updating). Simple CRUD with args is plain stdout.

| Command | Mode | Why |
|---------|------|-----|
| `bt task add "buy milk"` | Plain stdout | Args provided, instant |
| `bt task add` (no args) | Interactive form (`huh`) | Needs input, guide the user |
| `bt task list` | Plain table | Quick glance |
| `bt routine play` | Full BubbleTea TUI | Multi-step session, timers |
| `bt now` | Plain stdout (default) | Quick suggestion |
| `bt now --interactive` | BubbleTea TUI | Browse/accept/skip suggestions |
| `bt task edit <id>` | Opens `$EDITOR` | Free-form editing |
| `bt task edit <id> --title "new"` | Plain stdout | Flag-based update, instant |

**Detection**: Auto-detect interactive terminal:

```go
func isInteractive() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}
```

- **Piped / scripted** (`bt list | grep urgent`): No color, no TUI, plain text
- **Interactive terminal**: Rich output, color, can prompt

**Libraries**:
- `charmbracelet/huh` — for interactive forms (task creation wizard)
- `charmbracelet/bubbletea` — for full TUI (routine player, interactive `now`)
- `charmbracelet/lipgloss` — for styling in all modes

---

### 3. Output Formatting

Support three output modes via a global flag:

```bash
bt task list                    # Default: human-readable table
bt task list --json             # Machine-readable JSON
bt task show 01JQX --json       # Single task as JSON
```

#### Formats

| Format | Flag | When | Library |
|--------|------|------|---------|
| **Text** (default) | (none) | Human reading in terminal | `lipgloss` |
| **JSON** | `--json` | Scripting, piping, integrations | `encoding/json` |
| **Quiet** | `--quiet` / `-q` | Only IDs, for piping | plain `fmt` |

#### Text Output Examples

**`bt task list`**:
```
 ID         TITLE                 PRIORITY  DUE         STATUS   ENERGY
 01JQX001   Buy groceries         high      2026-04-06  pending  low
 01JQX002   Paint bedroom         medium    Apr 7-13    pending  high
 01JQX003   Organize photos       low       —           pending  medium
```

**`bt task show 01JQX001`**:
```
┌─ Buy groceries ────────────────────────────┐
│ ID:       01JQX001ABCDEF123456             │
│ Status:   pending                           │
│ Priority: high                              │
│ Energy:   low                               │
│ Duration: ~45 min                           │
│ Due:      2026-04-06 (tomorrow)             │
│ Tags:     errands, home                     │
│ Box:      inbox                             │
│ Links:    none                              │
│ Created:  2026-04-05 10:30                  │
└─────────────────────────────────────────────┘

  Need to get items for the week.

  - [ ] Vegetables
  - [ ] Bread
  - [ ] Chicken
```

**`bt task list --json`**:
```json
[
  {
    "id": "01JQX001ABCDEF123456",
    "title": "Buy groceries",
    "status": "pending",
    "priority": "high",
    ...
  }
]
```

**`bt task list -q`**:
```
01JQX001
01JQX002
01JQX003
```

Quiet mode enables piping: `bt task list -q --tag errands | xargs -I{} bt task done {}`

---

### 4. Color and Styling

**Library**: `charmbracelet/lipgloss` (already in the Charm ecosystem with BubbleTea)

**Color behavior**:
- ON by default in interactive terminals
- OFF when piped (auto-detected by `lipgloss`/`termenv`)
- OFF when `NO_COLOR` env var is set (https://no-color.org — auto-respected)
- OFF with `--no-color` flag or `BT_NO_COLOR=1`

**Color scheme**:

| Element | Color | Example |
|---------|-------|---------|
| Priority urgent | Red | `urgent` |
| Priority high | Orange/Yellow | `high` |
| Priority medium | Blue | `medium` |
| Priority low | Gray | `low` |
| Status done | Green + strikethrough | `✓ done` |
| Status active | Cyan | `● active` |
| Status pending | Default | `○ pending` |
| Status blocked | Red | `✗ blocked` |
| Overdue | Red background | `Apr 3 (2d overdue)` |
| Due today | Yellow | `Apr 5 (today)` |
| Habit streak | Green gradient | `🔥 12 days` |
| Error | Red | `Error: no task found...` |
| Success | Green | `✓ Task marked done` |

**Adaptive colors**: `lipgloss` automatically adapts to light/dark terminal backgrounds via `termenv.HasDarkBackground()`.

---

### 5. Editor Integration

**Flow** (same as `git commit`):

```go
func openEditor(initialContent string) (string, error) {
    editor := firstNonEmpty(os.Getenv("EDITOR"), os.Getenv("VISUAL"), "vi")
    
    tmpFile, _ := os.CreateTemp("", "bt-edit-*.md")
    tmpFile.WriteString(initialContent)
    tmpFile.Close()
    
    cmd := exec.Command(editor, tmpFile.Name())
    cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
    err := cmd.Run()
    
    content, _ := os.ReadFile(tmpFile.Name())
    os.Remove(tmpFile.Name())
    return string(content), err
}
```

**When the editor opens**:

| Command | Editor opens? | Content |
|---------|--------------|---------|
| `bt task add` (no args) | Only if not interactive | Frontmatter template |
| `bt task edit <id>` | Yes (unless all changes via flags) | Full task file |
| `bt task add --edit` | Yes | Frontmatter template |
| `bt task add "title" -p high` | No | Created from flags |

**Template** shown in editor:

```markdown
---
title: 
type: one-shot
priority: none
energy: medium
estimated_duration: 
due_date: 
tags: []
context: [anywhere]
box: inbox
---

# Task Notes

<!-- Write any notes below. Lines starting with # in the YAML are comments. -->
<!-- Save and close to create the task. Empty title = abort. -->
```

---

### 6. Shell Completions

Cobra provides built-in completion generation:

```bash
# Install completions
bt completion bash > ~/.bash_completion.d/bt
bt completion zsh > ~/.zfunc/_bt
bt completion fish > ~/.config/fish/completions/bt.fish
```

**Dynamic completions** for task IDs (the killer feature):

```go
ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    tasks := loadPendingTasks()
    var comps []string
    for _, t := range tasks {
        // Tab shows: ID + title as description
        comps = append(comps, t.ID+"\t"+t.Title)
    }
    return comps, cobra.ShellCompDirectiveNoFileComp
},
```

Result when user presses Tab:
```
$ bt done <TAB>
01JQX001  Buy groceries
01JQX002  Paint bedroom
01JQX003  Organize photos
```

**Dynamic completions also for**:
- Tag names (`--tag <TAB>` shows existing tags)
- Box names (`--box <TAB>` shows existing boxes)
- Context values (`--context <TAB>` shows home, office, etc.)

---

### 7. Error Handling

**Pattern**: Error message + context + suggested fix (like `gh`).

```
Error: no task matching prefix "01Z"

  Did you mean one of these?
    01JQX001  Buy groceries
    01JQX002  Paint bedroom

  Run 'bt task list' to see all tasks.
```

```
Error: task 01JQX002 depends on 01JQX001 which is not complete

  Complete the dependency first:
    bt task done 01JQX001

  Or remove the dependency:
    bt task unlink 01JQX002 01JQX001
```

**Exit codes**:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (task not found, validation failed) |
| 2 | Usage error (invalid args, missing required flags) |

**Implementation**:
- All commands use `RunE` (not `Run`) — errors propagate to Cobra
- Errors go to stderr, results go to stdout
- Wrap errors with `fmt.Errorf("context: %w", err)` for chain

---

### 8. Global Flags

Available on every command:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--json` | | false | Output as JSON |
| `--quiet` | `-q` | false | Output only IDs |
| `--no-color` | | false | Disable color output |
| `--data-dir` | `-d` | `~/.bentotask/data` | Path to data directory |
| `--verbose` | `-v` | false | Verbose/debug output |

---

### 9. Common Flag Patterns for Task Operations

**`bt task add`**:

| Flag | Short | Example |
|------|-------|---------|
| `--priority` | `-p` | `-p high` |
| `--energy` | `-e` | `-e low` |
| `--duration` | | `--duration 45` |
| `--due` | | `--due 2026-04-10` or `--due tomorrow` |
| `--due-start` | | `--due-start 2026-04-07` |
| `--due-end` | | `--due-end 2026-04-13` |
| `--tag` | | `--tag errands --tag home` (repeatable) |
| `--context` | `-c` | `-c home` |
| `--box` | `-b` | `-b projects/home-renovation` |
| `--recurrence` | | `--recurrence "FREQ=WEEKLY;BYDAY=MO,WE,FR"` |
| `--edit` | | Open in $EDITOR after creating |

**`bt task list`**:

| Flag | Short | Example |
|------|-------|---------|
| `--status` | `-s` | `-s pending` |
| `--priority` | `-p` | `-p high,urgent` |
| `--tag` | | `--tag errands` |
| `--box` | `-b` | `-b inbox` |
| `--due` | | `--due today`, `--due this-week`, `--due overdue` |
| `--energy` | `-e` | `-e low` |
| `--sort` | | `--sort due`, `--sort priority`, `--sort created` |
| `--limit` | `-n` | `-n 10` |

**`bt now`**:

| Flag | Short | Example |
|------|-------|---------|
| `--time` | `-t` | `-t 45` (available minutes) |
| `--energy` | `-e` | `-e low` |
| `--context` | `-c` | `-c home` |
| `--interactive` | `-i` | Launch TUI to browse suggestions |

---

### 10. Natural Date Parsing (Future Enhancement)

For Phase 1, support ISO dates. Plan for future natural language:

```bash
# Phase 1 (strict)
bt add "Dentist" --due 2026-04-15

# Phase 2 (natural language)
bt add "Dentist" --due "next tuesday"
bt add "Laundry" --due "tomorrow"
bt add "Report" --due "in 3 days"
```

**Library for future**: `olebedev/when` or custom parser. For Phase 1, support a few keywords: `today`, `tomorrow`, `this-week`, `next-week` + ISO dates.

---

## Consequences

- Binary name: `bt`
- Commands: noun-verb pattern with top-level aliases
- Charm ecosystem: `cobra` + `bubbletea` + `huh` + `lipgloss`
- Output: text (default), `--json`, `--quiet` modes
- Color: on by default, `NO_COLOR` respected, `--no-color` flag
- Editor: `$EDITOR` → `$VISUAL` → `vi` fallback
- Completions: Cobra built-in with dynamic task ID/tag/box completion
- Errors: stderr, colored, with suggested fixes

---

## References

- [Cobra CLI framework](https://cobra.dev)
- [BubbleTea TUI framework](https://github.com/charmbracelet/bubbletea)
- [`huh` — interactive forms](https://github.com/charmbracelet/huh)
- [`lipgloss` — styling](https://github.com/charmbracelet/lipgloss)
- [`gh` CLI source](https://github.com/cli/cli) — reference implementation
- [NO_COLOR standard](https://no-color.org)
- [Cobra shell completions](https://github.com/spf13/cobra/blob/main/shell_completions.md)
