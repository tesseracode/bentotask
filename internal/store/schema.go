package store

// schema defines all SQLite tables and indexes for the BentoTask index.
// This matches ADR-002 §5 exactly.
//
// The index is a derived cache — markdown files are the source of truth.
// All tables can be dropped and recreated from the .md files at any time.
const schema = `
-- Core task index
CREATE TABLE IF NOT EXISTS tasks (
    id                 TEXT PRIMARY KEY,
    title              TEXT NOT NULL,
    type               TEXT NOT NULL,
    status             TEXT NOT NULL,
    priority           TEXT,
    energy             TEXT,
    estimated_duration INTEGER,
    due_date           TEXT,
    due_start          TEXT,
    due_end            TEXT,
    box                TEXT,
    recurrence         TEXT,
    completed_at       TEXT,
    created_at         TEXT NOT NULL,
    updated_at         TEXT NOT NULL,
    file_path          TEXT NOT NULL
);

-- Tags (many-to-many)
CREATE TABLE IF NOT EXISTS task_tags (
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag     TEXT NOT NULL,
    PRIMARY KEY (task_id, tag)
);

-- Contexts (many-to-many)
CREATE TABLE IF NOT EXISTS task_contexts (
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    context TEXT NOT NULL,
    PRIMARY KEY (task_id, context)
);

-- Links between tasks
CREATE TABLE IF NOT EXISTS task_links (
    source_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    target_id TEXT NOT NULL,
    link_type TEXT NOT NULL,
    PRIMARY KEY (source_id, target_id, link_type)
);

-- Indexes for fast queries
CREATE INDEX IF NOT EXISTS idx_tasks_status   ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_type     ON tasks(type);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_due_end  ON tasks(due_end);
CREATE INDEX IF NOT EXISTS idx_tasks_box      ON tasks(box);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_task_tags_tag        ON task_tags(tag);
CREATE INDEX IF NOT EXISTS idx_task_contexts_ctx    ON task_contexts(context);
CREATE INDEX IF NOT EXISTS idx_task_links_target    ON task_links(target_id);
`
