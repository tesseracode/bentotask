# Milestone 7: Integrations — Notion & Obsidian Detail

**New tasks M7.7 and M7.8** — added to the existing M7 milestone.

---

## M7.7: Notion Integration

### Feasibility: ✅ Fully Feasible

Notion has a [public REST API](https://developers.notion.com/) that supports reading/writing databases and pages.

### Approach

**Import (Notion → BentoTask)**:
- Authenticate via Notion integration token (user creates an "internal integration" in Notion settings)
- Read a Notion database via `POST /v1/databases/{id}/query`
- Map Notion properties to BentoTask fields:
  - `Title` property → task title
  - `Status` property → status (map Notion statuses to BentoTask statuses)
  - `Priority` select → priority
  - `Date` property → due_date
  - `Tags` multi-select → tags
  - Page body (blocks) → markdown body (basic block-to-markdown conversion)
- Create BentoTask tasks via `app.AddTask()` for each row
- Store Notion page ID in task metadata for future sync

**Export (BentoTask → Notion)**:
- Create pages in a target Notion database
- Map BentoTask fields back to Notion properties
- Update existing pages if a Notion page ID is stored

**CLI**: `bt import notion --token <token> --database <db-id>`
**CLI**: `bt export notion --token <token> --database <db-id>`

### Package Structure
```
internal/notion/
├── client.go     # Notion API client (REST)
├── import.go     # Notion → BentoTask task mapping
├── export.go     # BentoTask → Notion page mapping
└── types.go      # Notion API types (database, page, properties)
```

### Dependencies
- No external library needed — Notion API is simple REST + JSON
- Authentication: Bearer token in header

### Limitations
- Notion API rate limit: 3 requests/second
- Block content conversion is approximate (Notion blocks ≠ markdown)
- Two-way sync requires tracking Notion page IDs and change timestamps

---

## M7.8: Obsidian Vault Integration

### Feasibility: ✅ Natural Fit (Easiest Integration)

Obsidian uses **markdown files with YAML frontmatter** — the exact same format BentoTask uses. The file watcher already detects external edits. This integration is mostly about compatibility, not new code.

### Approach

**Shared Data Directory**:
- Point BentoTask's `--data-dir` at a folder inside an Obsidian vault
- Example: `bt serve --data-dir ~/Notes/BentoTask/`
- Obsidian sees the `.md` files as normal notes
- BentoTask sees them as tasks
- File watcher (M1.4) already picks up Obsidian edits → re-indexes automatically

**What needs to work**:

1. **Wikilinks in task bodies**: Obsidian uses `[[Page Name]]` links. BentoTask should:
   - Preserve them in the markdown body (already works — body is opaque markdown)
   - Optionally render them as links in the web UI (nice-to-have)

2. **Obsidian-compatible frontmatter**: Ensure BentoTask's YAML frontmatter doesn't conflict with Obsidian's expectations:
   - Obsidian reads `tags:` as its own tag system — BentoTask already uses `tags:` ✅
   - Obsidian uses `aliases:` — BentoTask doesn't use this, no conflict ✅
   - Obsidian uses `cssclass:` — no conflict ✅
   - BentoTask fields like `type:`, `status:`, `priority:` are non-standard for Obsidian but won't break anything ✅

3. **Inbox folder mapping**: BentoTask's `inbox/` folder should be configurable or mappable to an Obsidian folder structure

4. **`_box.md` metadata files**: These are BentoTask-specific. Add them to Obsidian's excluded files or prefix with `.` to hide them

5. **`.obsidian/` directory**: BentoTask's file watcher should skip `.obsidian/` (it already skips hidden directories starting with `.`) ✅

### CLI
- `bt obsidian init` — set up a BentoTask data dir inside an Obsidian vault:
  - Creates the folder structure
  - Adds a `.obsidian/` ignore pattern if needed
  - Creates an Obsidian plugin config snippet (optional)

### What's mostly free
- File watcher already handles external edits ✅
- YAML frontmatter is already the storage format ✅
- Tags are already shared ✅
- Hidden directory skipping already works ✅

### What needs new code
- Wikilink rendering in the web UI body display
- `bt obsidian init` command
- Documentation on how to set up the shared vault
- Optional: Obsidian community plugin that adds BentoTask-aware features (task status toggles, priority picker) — this would be a separate TypeScript project for the Obsidian plugin ecosystem

### Package Structure
```
internal/obsidian/
├── init.go       # bt obsidian init — vault setup helper
└── compat.go     # Obsidian compatibility checks and adjustments
```

---

## Implementation Priority

1. **M7.8 (Obsidian) first** — 90% already works, just needs init command + docs + wikilink support
2. **M7.7 (Notion) second** — requires API client + field mapping, more work but well-defined
