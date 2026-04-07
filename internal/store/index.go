// Package store handles reading and writing BentoTask data to disk.
//
// It has two layers:
//   - Markdown I/O (markdown.go): reads/writes task files
//   - SQLite index (index.go): derived cache for fast queries
//
// The markdown files are the source of truth. The SQLite index can be
// deleted and rebuilt at any time with RebuildIndex.
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver

	"github.com/tesserabox/bentotask/internal/habit"
	"github.com/tesserabox/bentotask/internal/model"
)

// Index is a SQLite-backed query cache for BentoTask data.
// It mirrors the frontmatter from markdown files into relational tables
// for fast filtering, sorting, and full-text search.
//
// The index is disposable — it can always be rebuilt from the markdown files.
type Index struct {
	db *sql.DB
}

// OpenIndex opens (or creates) the SQLite index at the given path.
// It creates all tables and indexes if they don't exist.
func OpenIndex(path string) (*Index, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create index directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(wal)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open index: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping index: %w", err)
	}

	idx := &Index{db: db}

	if err := idx.createSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return idx, nil
}

// Close closes the SQLite database connection.
func (idx *Index) Close() error {
	return idx.db.Close()
}

// createSchema creates all tables and indexes if they don't exist.
// Schema matches ADR-002 §5.
func (idx *Index) createSchema() error {
	_, err := idx.db.Exec(schema)
	return err
}

// UpsertTask inserts or updates a task in the index.
// It handles the main task row plus junction tables (tags, contexts, links).
func (idx *Index) UpsertTask(task *model.Task, filePath string) error {
	tx, err := idx.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Upsert main task row
	_, err = tx.Exec(`
		INSERT INTO tasks (id, title, type, status, priority, energy,
			estimated_duration, due_date, due_start, due_end, box,
			recurrence, completed_at, created_at, updated_at, file_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title=excluded.title, type=excluded.type, status=excluded.status,
			priority=excluded.priority, energy=excluded.energy,
			estimated_duration=excluded.estimated_duration,
			due_date=excluded.due_date, due_start=excluded.due_start,
			due_end=excluded.due_end, box=excluded.box,
			recurrence=excluded.recurrence, completed_at=excluded.completed_at,
			updated_at=excluded.updated_at, file_path=excluded.file_path`,
		task.ID, task.Title, string(task.Type), string(task.Status),
		nullIfEmpty(string(task.Priority)), nullIfEmpty(string(task.Energy)),
		nullIfZero(task.EstimatedDuration),
		nullIfEmpty(task.DueDate), nullIfEmpty(task.DueStart), nullIfEmpty(task.DueEnd),
		nullIfEmpty(task.Box), nullIfEmpty(task.Recurrence),
		timePtr(task.CompletedAt),
		task.Created.UTC().Format("2006-01-02T15:04:05Z"),
		task.Updated.UTC().Format("2006-01-02T15:04:05Z"),
		filePath,
	)
	if err != nil {
		return fmt.Errorf("upsert task: %w", err)
	}

	// Replace tags (delete + re-insert is simplest for updates)
	if _, err := tx.Exec("DELETE FROM task_tags WHERE task_id = ?", task.ID); err != nil {
		return fmt.Errorf("clear tags: %w", err)
	}
	for _, tag := range task.Tags {
		if _, err := tx.Exec("INSERT INTO task_tags (task_id, tag) VALUES (?, ?)", task.ID, tag); err != nil {
			return fmt.Errorf("insert tag %q: %w", tag, err)
		}
	}

	// Replace contexts
	if _, err := tx.Exec("DELETE FROM task_contexts WHERE task_id = ?", task.ID); err != nil {
		return fmt.Errorf("clear contexts: %w", err)
	}
	for _, ctx := range task.Context {
		if _, err := tx.Exec("INSERT INTO task_contexts (task_id, context) VALUES (?, ?)", task.ID, ctx); err != nil {
			return fmt.Errorf("insert context %q: %w", ctx, err)
		}
	}

	// Replace links
	if _, err := tx.Exec("DELETE FROM task_links WHERE source_id = ?", task.ID); err != nil {
		return fmt.Errorf("clear links: %w", err)
	}
	for _, link := range task.Links {
		if _, err := tx.Exec("INSERT INTO task_links (source_id, target_id, link_type) VALUES (?, ?, ?)",
			task.ID, link.Target, string(link.Type)); err != nil {
			return fmt.Errorf("insert link: %w", err)
		}
	}

	// Upsert FTS entry
	if _, err := tx.Exec("DELETE FROM tasks_fts WHERE id = ?", task.ID); err != nil {
		return fmt.Errorf("clear fts: %w", err)
	}
	if _, err := tx.Exec("INSERT INTO tasks_fts (id, title, body) VALUES (?, ?, ?)",
		task.ID, task.Title, task.Body); err != nil {
		return fmt.Errorf("insert fts: %w", err)
	}

	return tx.Commit()
}

// DeleteTask removes a task and its related data from the index.
func (idx *Index) DeleteTask(id string) error {
	_, _ = idx.db.Exec("DELETE FROM tasks_fts WHERE id = ?", id)
	_, err := idx.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// GetTask retrieves a single task from the index by exact ID.
func (idx *Index) GetTask(id string) (*IndexedTask, error) {
	row := idx.db.QueryRow(`
		SELECT id, title, type, status, priority, energy,
			estimated_duration, due_date, due_start, due_end, box,
			recurrence, completed_at, created_at, updated_at, file_path
		FROM tasks WHERE id = ?`, id)

	task, err := scanTask(row)
	if err != nil {
		return nil, err
	}

	// Load tags
	tags, err := idx.loadTags(id)
	if err != nil {
		return nil, err
	}
	task.Tags = tags

	// Load contexts
	contexts, err := idx.loadContexts(id)
	if err != nil {
		return nil, err
	}
	task.Contexts = contexts

	return task, nil
}

// FindByPrefix finds tasks whose ID starts with the given prefix.
// Used for CLI short-ID matching (ADR-002 §4).
func (idx *Index) FindByPrefix(prefix string) ([]*IndexedTask, error) {
	rows, err := idx.db.Query(`
		SELECT id, title, type, status, priority, energy,
			estimated_duration, due_date, due_start, due_end, box,
			recurrence, completed_at, created_at, updated_at, file_path
		FROM tasks WHERE id LIKE ? || '%'`, strings.ToUpper(prefix))
	if err != nil {
		return nil, fmt.Errorf("find by prefix: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return idx.collectTasks(rows)
}

// ListTasks returns all tasks matching the given filter.
// Pass nil for no filtering (returns all tasks).
func (idx *Index) ListTasks(f *TaskFilter) ([]*IndexedTask, error) {
	query := `
		SELECT DISTINCT t.id, t.title, t.type, t.status, t.priority, t.energy,
			t.estimated_duration, t.due_date, t.due_start, t.due_end, t.box,
			t.recurrence, t.completed_at, t.created_at, t.updated_at, t.file_path
		FROM tasks t`

	var conditions []string
	var args []any

	if f != nil {
		if f.Status != "" {
			conditions = append(conditions, "t.status = ?")
			args = append(args, string(f.Status))
		}
		if f.Type != "" {
			conditions = append(conditions, "t.type = ?")
			args = append(args, string(f.Type))
		}
		if f.Priority != "" {
			conditions = append(conditions, "t.priority = ?")
			args = append(args, string(f.Priority))
		}
		if f.Energy != "" {
			conditions = append(conditions, "t.energy = ?")
			args = append(args, string(f.Energy))
		}
		if f.Box != "" {
			conditions = append(conditions, "t.box = ?")
			args = append(args, f.Box)
		}
		if f.Tag != "" {
			query += " JOIN task_tags tt ON t.id = tt.task_id"
			conditions = append(conditions, "tt.tag = ?")
			args = append(args, f.Tag)
		}
		if f.Context != "" {
			query += " JOIN task_contexts tc ON t.id = tc.task_id"
			conditions = append(conditions, "tc.context = ?")
			args = append(args, f.Context)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY t.created_at DESC"

	if f != nil && f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", f.Limit)
	}

	rows, err := idx.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return idx.collectTasks(rows)
}

// TaskCount returns the total number of tasks in the index.
func (idx *Index) TaskCount() (int, error) {
	var count int
	err := idx.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	return count, err
}

// DistinctTags returns all unique tags across all tasks, sorted alphabetically.
func (idx *Index) DistinctTags() ([]string, error) {
	return idx.distinctStrings("SELECT DISTINCT tag FROM task_tags ORDER BY tag")
}

// DistinctBoxes returns all unique box values across all tasks, sorted alphabetically.
func (idx *Index) DistinctBoxes() ([]string, error) {
	return idx.distinctStrings("SELECT DISTINCT box FROM tasks WHERE box IS NOT NULL AND box != '' ORDER BY box")
}

// DistinctContexts returns all unique context values across all tasks, sorted alphabetically.
func (idx *Index) DistinctContexts() ([]string, error) {
	return idx.distinctStrings("SELECT DISTINCT context FROM task_contexts ORDER BY context")
}

// distinctStrings runs a single-column query and returns the results as a string slice.
func (idx *Index) distinctStrings(query string) ([]string, error) {
	rows, err := idx.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// Search performs a full-text search across task titles and bodies.
// Uses SQLite FTS5 for fast matching. Returns tasks ranked by relevance.
func (idx *Index) Search(query string) ([]*IndexedTask, error) {
	rows, err := idx.db.Query(`
		SELECT t.id, t.title, t.type, t.status, t.priority, t.energy,
			t.estimated_duration, t.due_date, t.due_start, t.due_end, t.box,
			t.recurrence, t.completed_at, t.created_at, t.updated_at, t.file_path
		FROM tasks_fts fts
		JOIN tasks t ON fts.id = t.id
		WHERE tasks_fts MATCH ?
		ORDER BY rank`, query)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return idx.collectTasks(rows)
}

// --- Task Links ---

// IndexedLink represents a link relationship as stored in the SQLite index.
type IndexedLink struct {
	SourceID string
	TargetID string
	LinkType string
}

// LoadLinks returns all outgoing links from a task (source → target).
func (idx *Index) LoadLinks(taskID string) ([]IndexedLink, error) {
	rows, err := idx.db.Query(
		"SELECT source_id, target_id, link_type FROM task_links WHERE source_id = ? ORDER BY link_type, target_id",
		taskID)
	if err != nil {
		return nil, fmt.Errorf("load links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []IndexedLink
	for rows.Next() {
		var l IndexedLink
		if err := rows.Scan(&l.SourceID, &l.TargetID, &l.LinkType); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// LoadBacklinks returns all incoming links to a task (source → this task).
func (idx *Index) LoadBacklinks(taskID string) ([]IndexedLink, error) {
	rows, err := idx.db.Query(
		"SELECT source_id, target_id, link_type FROM task_links WHERE target_id = ? ORDER BY link_type, source_id",
		taskID)
	if err != nil {
		return nil, fmt.Errorf("load backlinks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []IndexedLink
	for rows.Next() {
		var l IndexedLink
		if err := rows.Scan(&l.SourceID, &l.TargetID, &l.LinkType); err != nil {
			return nil, fmt.Errorf("scan backlink: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// DependencyGraph loads all depends-on and blocks edges from the index.
// Returns a map of sourceID → []targetID for cycle detection.
func (idx *Index) DependencyGraph() (map[string][]string, error) {
	rows, err := idx.db.Query(
		"SELECT source_id, target_id FROM task_links WHERE link_type IN ('depends-on', 'blocks')")
	if err != nil {
		return nil, fmt.Errorf("load dependency graph: %w", err)
	}
	defer func() { _ = rows.Close() }()

	graph := make(map[string][]string)
	for rows.Next() {
		var src, tgt string
		if err := rows.Scan(&src, &tgt); err != nil {
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		graph[src] = append(graph[src], tgt)
	}
	return graph, rows.Err()
}

// --- Habit Completions ---

// LogHabitCompletion inserts a completion record for a habit.
func (idx *Index) LogHabitCompletion(habitID string, completedAt time.Time, duration int, note string) error {
	_, err := idx.db.Exec(`
		INSERT OR REPLACE INTO habit_completions (habit_id, completed_at, duration, note)
		VALUES (?, ?, ?, ?)`,
		habitID,
		completedAt.UTC().Format("2006-01-02T15:04:05Z"),
		nullIfZero(duration),
		nullIfEmpty(note),
	)
	if err != nil {
		return fmt.Errorf("log habit completion: %w", err)
	}
	return nil
}

// HabitCompletions returns all completions for a habit, ordered by date descending.
func (idx *Index) HabitCompletions(habitID string) ([]HabitCompletion, error) {
	rows, err := idx.db.Query(`
		SELECT completed_at, duration, note
		FROM habit_completions
		WHERE habit_id = ?
		ORDER BY completed_at DESC`, habitID)
	if err != nil {
		return nil, fmt.Errorf("query habit completions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var completions []HabitCompletion
	for rows.Next() {
		var c HabitCompletion
		if err := rows.Scan(&c.CompletedAt, &c.Duration, &c.Note); err != nil {
			return nil, fmt.Errorf("scan habit completion: %w", err)
		}
		completions = append(completions, c)
	}
	return completions, rows.Err()
}

// HabitCompletionCount returns the number of completions for a habit.
func (idx *Index) HabitCompletionCount(habitID string) (int, error) {
	var count int
	err := idx.db.QueryRow("SELECT COUNT(*) FROM habit_completions WHERE habit_id = ?", habitID).Scan(&count)
	return count, err
}

// HabitCompletion represents a single habit completion record from the index.
type HabitCompletion struct {
	CompletedAt string
	Duration    *int
	Note        *string
}

// RebuildIndex drops all data and re-indexes every .md file under dataDir.
func (idx *Index) RebuildIndex(dataDir string) (int, error) {
	// Clear existing data
	for _, table := range []string{"habit_completions", "task_links", "task_contexts", "task_tags", "tasks", "tasks_fts"} {
		if _, err := idx.db.Exec("DELETE FROM " + table); err != nil {
			return 0, fmt.Errorf("clear %s: %w", table, err)
		}
	}

	count := 0
	err := filepath.WalkDir(dataDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories (like .bentotask/)
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		// Only process .md files, skip _box.md metadata files for now
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") || d.Name() == "_box.md" {
			return nil
		}

		task, err := ParseFile(path)
		if err != nil {
			// Log warning but continue — don't let one bad file break the rebuild
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			return nil
		}

		// Store path relative to dataDir
		relPath, _ := filepath.Rel(dataDir, path)

		if err := idx.UpsertTask(task, relPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: indexing %s: %v\n", path, err)
			return nil
		}

		// Repopulate habit_completions from the markdown body (source of truth)
		if task.Type == model.TaskTypeHabit && task.Body != "" {
			completions := habit.ParseCompletionsFromBody(task.Body)
			for _, c := range completions {
				if err := idx.LogHabitCompletion(task.ID, c.CompletedAt, c.Duration, c.Note); err != nil {
					fmt.Fprintf(os.Stderr, "warning: indexing completion for %s: %v\n", task.ID, err)
				}
			}
		}

		count++
		return nil
	})

	return count, err
}

// --- Types ---

// IndexedTask represents a task as stored in the SQLite index.
// It's a flat view of the data optimized for queries (no nested structs).
type IndexedTask struct {
	ID                string
	Title             string
	Type              string
	Status            string
	Priority          *string
	Energy            *string
	EstimatedDuration *int
	DueDate           *string
	DueStart          *string
	DueEnd            *string
	Box               *string
	Recurrence        *string
	CompletedAt       *string
	CreatedAt         string
	UpdatedAt         string
	FilePath          string
	Tags              []string
	Contexts          []string
}

// TaskFilter controls which tasks are returned by ListTasks.
type TaskFilter struct {
	Status   model.Status
	Type     model.TaskType
	Priority model.Priority
	Energy   model.Energy
	Box      string
	Tag      string
	Context  string
	Limit    int
}

// --- Helpers ---

func scanTask(row *sql.Row) (*IndexedTask, error) {
	var t IndexedTask
	err := row.Scan(
		&t.ID, &t.Title, &t.Type, &t.Status, &t.Priority, &t.Energy,
		&t.EstimatedDuration, &t.DueDate, &t.DueStart, &t.DueEnd, &t.Box,
		&t.Recurrence, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt, &t.FilePath,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}
	return &t, nil
}

func (idx *Index) collectTasks(rows *sql.Rows) ([]*IndexedTask, error) {
	var tasks []*IndexedTask
	for rows.Next() {
		var t IndexedTask
		err := rows.Scan(
			&t.ID, &t.Title, &t.Type, &t.Status, &t.Priority, &t.Energy,
			&t.EstimatedDuration, &t.DueDate, &t.DueStart, &t.DueEnd, &t.Box,
			&t.Recurrence, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt, &t.FilePath,
		)
		if err != nil {
			return nil, fmt.Errorf("scan task row: %w", err)
		}
		tasks = append(tasks, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load tags and contexts for each task from junction tables
	for _, t := range tasks {
		tags, err := idx.loadTags(t.ID)
		if err != nil {
			return nil, fmt.Errorf("load tags for %s: %w", t.ID, err)
		}
		t.Tags = tags

		contexts, err := idx.loadContexts(t.ID)
		if err != nil {
			return nil, fmt.Errorf("load contexts for %s: %w", t.ID, err)
		}
		t.Contexts = contexts
	}

	return tasks, nil
}

func (idx *Index) loadTags(taskID string) ([]string, error) {
	rows, err := idx.db.Query("SELECT tag FROM task_tags WHERE task_id = ? ORDER BY tag", taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (idx *Index) loadContexts(taskID string) ([]string, error) {
	rows, err := idx.db.Query("SELECT context FROM task_contexts WHERE task_id = ? ORDER BY context", taskID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var contexts []string
	for rows.Next() {
		var ctx string
		if err := rows.Scan(&ctx); err != nil {
			return nil, err
		}
		contexts = append(contexts, ctx)
	}
	return contexts, rows.Err()
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullIfZero(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

func timePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format("2006-01-02T15:04:05Z")
	return &s
}
