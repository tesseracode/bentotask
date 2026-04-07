// Package app provides the core application logic for BentoTask.
//
// It sits between the CLI (or API) and the store layer, coordinating
// operations that touch both markdown files and the SQLite index.
// Commands should call App methods rather than using store directly.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
)

// App is the core application. It manages the data directory, index,
// and provides high-level CRUD operations for tasks.
type App struct {
	DataDir  string
	Index    *store.Index
	indexDir string
}

// Open initializes the application with the given data directory.
// It creates the directory structure and opens the SQLite index.
func Open(dataDir string) (*App, error) {
	// Expand ~ if needed
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".bentotask", "data")
	}

	// Create the standard directory structure
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "inbox"),
		filepath.Join(dataDir, ".bentotask"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}

	// Open the index
	indexPath := filepath.Join(dataDir, ".bentotask", "index.db")
	idx, err := store.OpenIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("open index: %w", err)
	}

	return &App{
		DataDir:  dataDir,
		Index:    idx,
		indexDir: filepath.Join(dataDir, ".bentotask"),
	}, nil
}

// Close shuts down the application cleanly.
func (a *App) Close() error {
	return a.Index.Close()
}

// AddTask creates a new task, writes it to disk, and indexes it.
func (a *App) AddTask(title string, opts TaskOptions) (*model.Task, error) {
	now := time.Now().UTC()
	task := &model.Task{
		ID:      model.NewID(),
		Title:   title,
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
	}

	// Apply options
	if opts.Type != "" {
		task.Type = opts.Type
	}
	if opts.Priority != "" {
		task.Priority = opts.Priority
	}
	if opts.Energy != "" {
		task.Energy = opts.Energy
	}
	if opts.Duration > 0 {
		task.EstimatedDuration = opts.Duration
	}
	if opts.DueDate != "" {
		task.DueDate = opts.DueDate
		if task.Type == model.TaskTypeOneShot {
			task.Type = model.TaskTypeDated
		}
	}
	if opts.DueStart != "" {
		task.DueStart = opts.DueStart
	}
	if opts.DueEnd != "" {
		task.DueEnd = opts.DueEnd
		if task.DueStart != "" && task.Type == model.TaskTypeOneShot {
			task.Type = model.TaskTypeRanged
		}
	}
	if len(opts.Tags) > 0 {
		task.Tags = opts.Tags
	}
	if len(opts.Context) > 0 {
		task.Context = opts.Context
	}
	if opts.Box != "" {
		task.Box = opts.Box
	}

	// Validate
	if errs := task.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("validation failed: %s", errs[0])
	}

	// Determine file path based on box
	relPath := taskFilePath(task)
	absPath := filepath.Join(a.DataDir, relPath)

	// Write to disk
	if err := store.WriteFile(absPath, task); err != nil {
		return nil, fmt.Errorf("write task: %w", err)
	}

	// Index
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return nil, fmt.Errorf("index task: %w", err)
	}

	return task, nil
}

// ListTasks returns tasks matching the given filter.
func (a *App) ListTasks(f *store.TaskFilter) ([]*store.IndexedTask, error) {
	return a.Index.ListTasks(f)
}

// GetTask retrieves a single task by exact ID or prefix.
// If a prefix matches exactly one task, it returns that task.
func (a *App) GetTask(idOrPrefix string) (*model.Task, string, error) {
	// Try exact match first
	indexed, err := a.Index.GetTask(idOrPrefix)
	if err == nil {
		return a.loadTask(indexed)
	}

	// Try prefix match
	matches, err := a.Index.FindByPrefix(idOrPrefix)
	if err != nil {
		return nil, "", fmt.Errorf("search: %w", err)
	}

	switch len(matches) {
	case 0:
		return nil, "", fmt.Errorf("no task matching %q", idOrPrefix)
	case 1:
		return a.loadTask(matches[0])
	default:
		return nil, "", fmt.Errorf("ambiguous prefix %q matches %d tasks", idOrPrefix, len(matches))
	}
}

// CompleteTask marks a task as done.
func (a *App) CompleteTask(idOrPrefix string) (*model.Task, error) {
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}

	if task.IsDone() {
		return nil, fmt.Errorf("task %q is already %s", task.Title, task.Status)
	}

	now := time.Now().UTC()
	task.Status = model.StatusDone
	task.CompletedAt = &now
	task.Updated = now

	return task, a.saveTask(task, relPath)
}

// DeleteTask removes a task from disk and the index.
func (a *App) DeleteTask(idOrPrefix string) (*model.Task, error) {
	task, relPath, err := a.GetTask(idOrPrefix)
	if err != nil {
		return nil, err
	}

	absPath := filepath.Join(a.DataDir, relPath)
	if err := os.Remove(absPath); err != nil {
		return nil, fmt.Errorf("delete file: %w", err)
	}

	if err := a.Index.DeleteTask(task.ID); err != nil {
		return nil, fmt.Errorf("remove from index: %w", err)
	}

	return task, nil
}

// RebuildIndex drops the index and rebuilds it from all .md files.
func (a *App) RebuildIndex() (int, error) {
	return a.Index.RebuildIndex(a.DataDir)
}

// --- Helpers ---

// loadTask reads the full task from disk given an indexed reference.
func (a *App) loadTask(indexed *store.IndexedTask) (*model.Task, string, error) {
	absPath := filepath.Join(a.DataDir, indexed.FilePath)
	task, err := store.ParseFile(absPath)
	if err != nil {
		return nil, "", fmt.Errorf("read %s: %w", indexed.FilePath, err)
	}
	return task, indexed.FilePath, nil
}

// saveTask writes a task to disk and updates the index.
func (a *App) saveTask(task *model.Task, relPath string) error {
	absPath := filepath.Join(a.DataDir, relPath)
	if err := store.WriteFile(absPath, task); err != nil {
		return fmt.Errorf("write task: %w", err)
	}
	if err := a.Index.UpsertTask(task, relPath); err != nil {
		return fmt.Errorf("index task: %w", err)
	}
	return nil
}

// taskFilePath determines the relative file path for a task based on its box.
func taskFilePath(task *model.Task) string {
	dir := "inbox"
	if task.Box != "" {
		dir = task.Box
	}
	return filepath.Join(dir, task.ID+".md")
}

// TaskOptions holds optional fields for creating a task.
type TaskOptions struct {
	Type     model.TaskType
	Priority model.Priority
	Energy   model.Energy
	Duration int
	DueDate  string
	DueStart string
	DueEnd   string
	Tags     []string
	Context  []string
	Box      string
}
