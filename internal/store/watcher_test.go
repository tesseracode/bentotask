package store

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// waitFor polls a condition up to timeout. Returns true if the condition was met.
func waitFor(t *testing.T, timeout time.Duration, condition func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestWatcherDetectsNewFile(t *testing.T) {
	dataDir := t.TempDir()
	inboxDir := filepath.Join(dataDir, "inbox")
	_ = os.MkdirAll(inboxDir, 0o755)

	idx := openTestIndex(t)
	var indexed atomic.Int32

	w, err := NewWatcher(dataDir, idx)
	if err != nil {
		t.Fatalf("NewWatcher() error: %v", err)
	}
	w.OnError = func(err error) { t.Logf("watcher error: %v", err) }
	w.OnIndex = func(_ string) { indexed.Add(1) }
	defer func() { _ = w.Close() }()

	// Create a task file
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID: "01WATCH0001", Title: "Watched task", Type: model.TaskTypeOneShot,
		Status: model.StatusPending, Created: now, Updated: now,
	}
	_ = WriteFile(filepath.Join(inboxDir, "01WATCH0001.md"), task)

	// Wait for the watcher to pick it up
	if !waitFor(t, 2*time.Second, func() bool { return indexed.Load() >= 1 }) {
		t.Fatal("watcher did not index the new file within timeout")
	}

	// Verify it's in the index
	got, err := idx.GetTask("01WATCH0001")
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if got.Title != "Watched task" {
		t.Errorf("Title = %q, want %q", got.Title, "Watched task")
	}
}

func TestWatcherDetectsModifiedFile(t *testing.T) {
	dataDir := t.TempDir()
	inboxDir := filepath.Join(dataDir, "inbox")
	_ = os.MkdirAll(inboxDir, 0o755)

	idx := openTestIndex(t)
	var indexed atomic.Int32

	// Pre-create a task file before starting the watcher
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID: "01WATCH0002", Title: "Original title", Type: model.TaskTypeOneShot,
		Status: model.StatusPending, Created: now, Updated: now,
	}
	taskPath := filepath.Join(inboxDir, "01WATCH0002.md")
	_ = WriteFile(taskPath, task)

	// Index it first
	_ = idx.UpsertTask(task, "inbox/01WATCH0002.md")

	w, err := NewWatcher(dataDir, idx)
	if err != nil {
		t.Fatalf("NewWatcher() error: %v", err)
	}
	w.OnError = func(err error) { t.Logf("watcher error: %v", err) }
	w.OnIndex = func(_ string) { indexed.Add(1) }
	defer func() { _ = w.Close() }()

	// Modify the file
	task.Title = "Updated title"
	_ = WriteFile(taskPath, task)

	// Wait for the watcher to process the update
	if !waitFor(t, 2*time.Second, func() bool { return indexed.Load() >= 1 }) {
		t.Fatal("watcher did not detect the file modification within timeout")
	}

	// Verify the index was updated
	got, err := idx.GetTask("01WATCH0002")
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if got.Title != "Updated title" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated title")
	}
}

func TestWatcherDetectsDeletedFile(t *testing.T) {
	dataDir := t.TempDir()
	inboxDir := filepath.Join(dataDir, "inbox")
	_ = os.MkdirAll(inboxDir, 0o755)

	idx := openTestIndex(t)

	// Create and index a task
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID: "01WATCH0003", Title: "To be deleted", Type: model.TaskTypeOneShot,
		Status: model.StatusPending, Created: now, Updated: now,
	}
	taskPath := filepath.Join(inboxDir, "01WATCH0003.md")
	_ = WriteFile(taskPath, task)
	_ = idx.UpsertTask(task, "inbox/01WATCH0003.md")

	w, err := NewWatcher(dataDir, idx)
	if err != nil {
		t.Fatalf("NewWatcher() error: %v", err)
	}
	w.OnError = func(err error) { t.Logf("watcher error: %v", err) }
	defer func() { _ = w.Close() }()

	// Delete the file
	_ = os.Remove(taskPath)

	// Wait for the watcher to detect deletion
	if !waitFor(t, 2*time.Second, func() bool {
		count, _ := idx.TaskCount()
		return count == 0
	}) {
		t.Fatal("watcher did not detect file deletion within timeout")
	}
}

func TestWatcherIgnoresNonMarkdownFiles(t *testing.T) {
	dataDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dataDir, "inbox"), 0o755)

	idx := openTestIndex(t)
	var indexed atomic.Int32

	w, err := NewWatcher(dataDir, idx)
	if err != nil {
		t.Fatalf("NewWatcher() error: %v", err)
	}
	w.OnError = func(err error) { t.Logf("watcher error: %v", err) }
	w.OnIndex = func(_ string) { indexed.Add(1) }
	defer func() { _ = w.Close() }()

	// Create a non-markdown file
	_ = os.WriteFile(filepath.Join(dataDir, "inbox", "notes.txt"), []byte("not a task"), 0o644)

	// Give the watcher time to (not) process it
	time.Sleep(200 * time.Millisecond)

	if indexed.Load() != 0 {
		t.Error("watcher should not index .txt files")
	}
	count, _ := idx.TaskCount()
	if count != 0 {
		t.Errorf("TaskCount = %d, want 0", count)
	}
}

func TestWatcherDetectsNewSubdirectory(t *testing.T) {
	dataDir := t.TempDir()

	idx := openTestIndex(t)
	var indexed atomic.Int32

	w, err := NewWatcher(dataDir, idx)
	if err != nil {
		t.Fatalf("NewWatcher() error: %v", err)
	}
	w.OnError = func(err error) { t.Logf("watcher error: %v", err) }
	w.OnIndex = func(_ string) { indexed.Add(1) }
	defer func() { _ = w.Close() }()

	// Create directories one level at a time so fsnotify registers each.
	// MkdirAll creates nested dirs atomically, which fsnotify may miss.
	projectsDir := filepath.Join(dataDir, "projects")
	_ = os.Mkdir(projectsDir, 0o755)
	time.Sleep(200 * time.Millisecond)

	newDir := filepath.Join(projectsDir, "new-project")
	_ = os.Mkdir(newDir, 0o755)
	time.Sleep(200 * time.Millisecond)

	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID: "01SUBDIR001", Title: "Subdir task", Type: model.TaskTypeOneShot,
		Status: model.StatusPending, Created: now, Updated: now,
	}
	_ = WriteFile(filepath.Join(newDir, "01SUBDIR001.md"), task)

	if !waitFor(t, 2*time.Second, func() bool { return indexed.Load() >= 1 }) {
		t.Fatal("watcher did not index file in new subdirectory within timeout")
	}
}

func TestWatcherClose(t *testing.T) {
	dataDir := t.TempDir()
	idx := openTestIndex(t)

	w, err := NewWatcher(dataDir, idx)
	if err != nil {
		t.Fatalf("NewWatcher() error: %v", err)
	}

	// Close should not hang or panic
	err = w.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}
}
