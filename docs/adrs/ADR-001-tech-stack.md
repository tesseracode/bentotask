# ADR-001: Programming Language & Tech Stack

**Status**: APPROVED  
**Date**: 2026-04-05  
**Approved**: 2026-04-05  
**Decision Makers**: @jbencardino  

---

## Context

BentoTask is a local-first task/habit/routine manager with smart scheduling. We need to choose a primary language for the core engine + CLI, and a strategy for the web UI. The system must support:

- CLI with <100ms startup time
- Markdown + YAML file I/O
- SQLite indexing
- REST API server
- Embeddable web UI
- CalDAV/Google Calendar integration
- Future: plugin system, MCP server, CRDT sync

The developer knows JavaScript/Python and wants to learn either Go or a functional programming language (FP).

---

## Options Evaluated

### Option A: Go (core + CLI + API) + SvelteKit (web UI)

**Strengths**:
| Criteria | Assessment |
|----------|-----------|
| CLI startup | ~5-10ms — best-in-class for a GC'd language |
| Single binary | Excellent — `go build` produces one static binary |
| SQLite | `modernc.org/sqlite` (pure Go, no CGO) — trivial cross-compilation |
| Markdown/YAML | `goldmark` + `go-yaml` — battle-tested |
| Web server | `chi`, `echo`, or `gin` — mature, fast, well-documented |
| CalDAV | `emersion/go-webdav` — niche but functional |
| Learning curve | Low-moderate coming from JS/Python (1-2 weeks to productive) |
| Cross-compilation | `GOOS=linux GOARCH=arm go build` — trivial, great for Raspberry Pi |
| Web UI embedding | `go:embed` bundles SvelteKit build into binary |
| Plugin system | `hashicorp/go-plugin` (gRPC) or Wasm via `wazero` |
| CLI framework | `cobra` (commands) + `bubbletea` (TUI) — gold standard |
| Community | Massive — extensive libraries, tutorials, hiring pool |
| Error handling | Explicit but verbose (`if err != nil`) — some love it, some don't |

**Weaknesses**:
- No generics until Go 1.18, and still limited compared to Rust/TS
- Error handling is repetitive
- No pattern matching or algebraic data types
- Less expressive than functional languages for algorithm code

**Real-world comps**: Taskwarrior-like tools (Ultralist), Charm CLI suite (glow, soft-serve), Hugo (markdown+YAML engine), Caddy, Docker CLI.

---

### Option B: Rust (core + CLI + API) + SvelteKit (web UI)

**Strengths**:
| Criteria | Assessment |
|----------|-----------|
| CLI startup | ~3-5ms — absolute fastest |
| Single binary | Excellent — static linking, small binaries |
| SQLite | `rusqlite` — excellent, well-maintained |
| Web server | `axum` — modern, async, powerful |
| Performance | Best possible — zero-cost abstractions |
| Type system | Most powerful — enums, pattern matching, traits |
| Plugin system | Wasm via `wasmtime` — first-class support |

**Weaknesses**:
| Concern | Reality |
|---------|---------|
| Learning curve | **High** — 3-6 months to feel productive |
| Borrow checker | Will slow feature development significantly, especially early |
| Async complexity | `async` Rust has sharp edges (pinning, lifetimes in futures) |
| Compile times | 30s-2min incremental builds (Go: 1-5s) |
| Overkill? | A task manager doesn't need zero-cost abstractions |

**Real-world comps**: ripgrep, fd, bat, zoxide (all CLI tools). Fewer full-stack apps.

**Verdict**: Amazing language, wrong time. The bottleneck is *feature velocity*, not performance. Revisit if Go proves too slow (unlikely).

---

### Option C: TypeScript/Bun (full stack)

**Strengths**:
| Criteria | Assessment |
|----------|-----------|
| Shared code | Types and logic shared between CLI, API, and web UI |
| SQLite | `bun:sqlite` / `better-sqlite3` — excellent, synchronous |
| Web UI | Same language everywhere — no context switching |
| Ecosystem | Largest — library for everything |
| Rapid prototyping | Fastest iteration speed |

**Weaknesses**:
| Concern | Reality |
|---------|---------|
| CLI startup | Bun ~30ms, Node ~70-150ms — borderline |
| Binary size | `bun build --compile` produces 50-100MB binaries |
| Single binary | Works but bloated, not production-elegant |
| Learning | Developer already knows JS — not learning anything new |
| Runtime dependency | Bun/Node needs to be installed (or huge compiled binary) |

**Verdict**: The code-sharing story is compelling but doesn't offset the binary size, startup concerns, and lack of learning opportunity.

---

### Option D: Python (Click/Typer + FastAPI)

**Strengths**:
| Criteria | Assessment |
|----------|-----------|
| Libraries | `typer` + `rich` = beautiful CLI; `FastAPI` = great API |
| Rapid prototyping | Very fast to build features |
| AI/ML | Best ecosystem for future AI features |

**Weaknesses**:
| Concern | Reality |
|---------|---------|
| CLI startup | **150-300ms** — fails the <100ms requirement |
| Single binary | PyInstaller/Nuitka — fragile, 50MB+ |
| Distribution | Dependency management is painful |

**Verdict**: **Eliminated.** Startup time alone disqualifies it for a CLI-first tool.

---

### Option E: Elixir/Phoenix LiveView

**Strengths**:
| Criteria | Assessment |
|----------|-----------|
| Real-time web | LiveView is *incredible* — no separate SPA needed |
| Smart mirror | Server-rendered real-time updates over WebSocket — perfect |
| Concurrency | BEAM VM — best concurrency model in any language |
| Functional programming | Pure FP with pattern matching, immutability |
| Fault tolerance | OTP supervisors — self-healing processes |

**Weaknesses**:
| Concern | Reality |
|---------|---------|
| CLI startup | ~100-200ms for BEAM boot — borderline/fail |
| Single binary | No — requires BEAM VM (Burrito/Bakeware exist but are hacky) |
| Local-first | BEAM is designed for servers, not local desktop apps |
| SQLite | Ecto works with SQLite3, but it's not the primary target |
| Ecosystem size | Smaller — fewer libraries for CalDAV, markdown processing |

**Verdict**: Fascinating for the web layer. Terrible for the CLI layer. Could be a future option *just for the web UI* if LiveView's real-time capabilities are needed, while Go handles the core.

---

### Option F: OCaml / Haskell / F#

| Language | CLI Speed | Binary | Ecosystem | Practical? |
|----------|-----------|--------|-----------|-----------|
| OCaml | ~5ms | Single binary | Tiny | Impractical — no CalDAV, limited web |
| Haskell | ~15ms | Large binary | Medium | Possible but painful dependency management |
| F# | ~100ms | .NET dependency | Medium | Requires .NET runtime |

**Verdict**: Rewarding to learn, impractical for this project. The ecosystem gaps would constantly slow you down.

---

## Comparison Matrix

| Criteria (weighted) | Go | Rust | TypeScript | Python | Elixir |
|---------------------|-----|------|-----------|--------|--------|
| CLI startup (<100ms) ⭐⭐⭐ | 5-10ms ✅ | 3-5ms ✅ | 30-150ms ⚠️ | 200ms+ ❌ | 100-200ms ⚠️ |
| Single binary ⭐⭐ | ✅ | ✅ | ⚠️ 50MB+ | ❌ | ❌ |
| Feature velocity ⭐⭐⭐ | Fast | Slow | Fast | Fastest | Medium |
| Learning opportunity ⭐⭐ | ✅ New | ✅ New | ❌ Known | ❌ Known | ✅ New |
| SQLite ecosystem ⭐⭐ | ✅ | ✅ | ✅ | ✅ | ⚠️ |
| Web framework ⭐⭐ | ✅ | ✅ | ✅ | ✅ | ✅✅ |
| Cross-compile (Pi) ⭐ | Trivial | Hard | N/A | N/A | N/A |
| Plugin system ⭐ | Wasm/gRPC | Wasm | Dynamic | Dynamic | OTP |
| CalDAV libraries ⭐ | ⚠️ Niche | ⚠️ Niche | ⚠️ Niche | ⚠️ Niche | ⚠️ Niche |

---

## Proposed Decision

### **Go (core + CLI + API) + SvelteKit (web UI)**

```
bentotask/
├── cmd/
│   └── bt/              # CLI entrypoint (Cobra)
├── internal/
│   ├── model/           # Task, Habit, Routine structs
│   ├── store/           # Markdown I/O + SQLite index
│   ├── engine/          # Scheduling algorithm (knapsack)
│   ├── calendar/        # CalDAV + Google Calendar
│   ├── routine/         # Routine engine
│   ├── graph/           # Task dependency graph
│   └── api/             # REST API (Chi)
├── web/                 # SvelteKit SPA
├── plugins/             # Future: Wasm plugins (wazero)
├── docs/
│   └── adrs/
├── SPEC.md
└── go.mod
```

### Rationale

1. **Meets all hard requirements**: <10ms CLI, single binary, cross-compilation
2. **Best feature velocity** among compiled languages — you'll ship fast
3. **Learning opportunity** — new language without the 6-month Rust tax
4. **`go:embed`** bundles the SvelteKit build into the Go binary — one artifact serves CLI + API + web UI
5. **Pure Go SQLite** (`modernc.org/sqlite`) means no CGO, trivial cross-compilation to ARM (Raspberry Pi)
6. **`cobra` + `bubbletea`** — the best CLI framework ecosystem in any language
7. **`wazero`** — Wasm runtime in pure Go for future plugin system
8. **Massive community** — you'll never be stuck without a library or answer

### Tradeoffs Accepted

- Go's type system is less expressive than Rust or Haskell (no sum types, limited generics)
- Error handling is verbose (`if err != nil` everywhere)
- Not a functional language (but you can write functional-*style* Go)
- CalDAV libraries are niche in *every* language — may need to contribute upstream

### FP Learning Path (Bonus)

If the FP itch remains, consider:
- Writing the scheduling algorithm in a functional style in Go (pure functions, no mutation)
- Building a small Elixir/Phoenix LiveView prototype of just the smart mirror view later
- Exploring OCaml or Gleam (Erlang VM + ML-like syntax) for side projects

---

## Consequences

- Project will be written in Go 1.22+
- CLI framework: `cobra` with `bubbletea` for interactive modes
- Web UI: SvelteKit, embedded via `go:embed`
- API: `chi` router with standard `net/http`
- SQLite: `modernc.org/sqlite` (pure Go)
- Markdown: `goldmark` + `go-yaml/yaml`
- IDs: `oklog/ulid`
- Testing: standard `testing` package + `testify`

---

## References

- [Charm tools (Go CLI gold standard)](https://charm.sh)
- [modernc.org/sqlite (pure Go SQLite)](https://modernc.org/sqlite)
- [wazero (Go Wasm runtime)](https://wazero.io)
- [Cobra CLI framework](https://cobra.dev)
- [BubbleTea TUI framework](https://github.com/charmbracelet/bubbletea)
