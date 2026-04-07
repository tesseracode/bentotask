package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// openTestIndex creates an in-memory SQLite index for testing.
func openTestIndex(t *testing.T) *Index {
	t.Helper()
	// Use a temp file so WAL mode works (in-memory doesn't support WAL well)
	path := filepath.Join(t.TempDir(), "test-index.db")
	idx, err := OpenIndex(path)
	if err != nil {
		t.Fatalf("OpenIndex() error: %v", err)
	}
	t.Cleanup(func() { _ = idx.Close() })
	return idx
}

func makeTestTask(id, title string) *model.Task {
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	return &model.Task{
		ID:      id,
		Title:   title,
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
	}
}

func TestOpenIndexCreatesSchema(t *testing.T) {
	idx := openTestIndex(t)

	// Should be able to query tasks table (it exists)
	count, err := idx.TaskCount()
	if err != nil {
		t.Fatalf("TaskCount() error: %v", err)
	}
	if count != 0 {
		t.Errorf("empty index should have 0 tasks, got %d", count)
	}
}

func TestUpsertAndGetTask(t *testing.T) {
	idx := openTestIndex(t)

	task := makeTestTask("01TEST001", "Buy groceries")
	task.Priority = model.PriorityHigh
	task.Energy = model.EnergyLow
	task.EstimatedDuration = 45
	task.DueDate = "2026-04-06"
	task.Tags = []string{"errands", "home"}
	task.Context = []string{"errands"}
	task.Box = "inbox"

	err := idx.UpsertTask(task, "inbox/01TEST001.md")
	if err != nil {
		t.Fatalf("UpsertTask() error: %v", err)
	}

	// Retrieve it
	got, err := idx.GetTask("01TEST001")
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}

	if got.ID != "01TEST001" {
		t.Errorf("ID = %q, want %q", got.ID, "01TEST001")
	}
	if got.Title != "Buy groceries" {
		t.Errorf("Title = %q, want %q", got.Title, "Buy groceries")
	}
	if got.FilePath != "inbox/01TEST001.md" {
		t.Errorf("FilePath = %q, want %q", got.FilePath, "inbox/01TEST001.md")
	}
	if len(got.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(got.Tags))
	}
	if len(got.Contexts) != 1 {
		t.Errorf("Contexts count = %d, want 1", len(got.Contexts))
	}
}

func TestUpsertUpdatesExistingTask(t *testing.T) {
	idx := openTestIndex(t)

	task := makeTestTask("01UPDATE01", "Original title")
	_ = idx.UpsertTask(task, "inbox/01UPDATE01.md")

	// Update it
	task.Title = "Updated title"
	task.Status = model.StatusDone
	task.Tags = []string{"new-tag"}
	_ = idx.UpsertTask(task, "inbox/01UPDATE01.md")

	got, err := idx.GetTask("01UPDATE01")
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if got.Title != "Updated title" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated title")
	}
	if got.Status != "done" {
		t.Errorf("Status = %q, want %q", got.Status, "done")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "new-tag" {
		t.Errorf("Tags = %v, want [new-tag]", got.Tags)
	}

	// Should still be only 1 task
	count, _ := idx.TaskCount()
	if count != 1 {
		t.Errorf("TaskCount = %d, want 1 after upsert", count)
	}
}

func TestDeleteTask(t *testing.T) {
	idx := openTestIndex(t)

	task := makeTestTask("01DELETE01", "To be deleted")
	task.Tags = []string{"tag1"}
	_ = idx.UpsertTask(task, "inbox/01DELETE01.md")

	err := idx.DeleteTask("01DELETE01")
	if err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}

	count, _ := idx.TaskCount()
	if count != 0 {
		t.Errorf("TaskCount = %d, want 0 after delete", count)
	}
}

func TestFindByPrefix(t *testing.T) {
	idx := openTestIndex(t)

	_ = idx.UpsertTask(makeTestTask("01AAAA0001", "Task A1"), "inbox/01AAAA0001.md")
	_ = idx.UpsertTask(makeTestTask("01AAAA0002", "Task A2"), "inbox/01AAAA0002.md")
	_ = idx.UpsertTask(makeTestTask("01BBBB0001", "Task B1"), "inbox/01BBBB0001.md")

	// Search by prefix
	results, err := idx.FindByPrefix("01AAAA")
	if err != nil {
		t.Fatalf("FindByPrefix() error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("FindByPrefix('01AAAA') returned %d results, want 2", len(results))
	}

	// Search for unique prefix
	results, err = idx.FindByPrefix("01BBBB")
	if err != nil {
		t.Fatalf("FindByPrefix() error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("FindByPrefix('01BBBB') returned %d results, want 1", len(results))
	}

	// No match
	results, err = idx.FindByPrefix("01ZZZZ")
	if err != nil {
		t.Fatalf("FindByPrefix() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("FindByPrefix('01ZZZZ') returned %d results, want 0", len(results))
	}
}

func TestListTasksNoFilter(t *testing.T) {
	idx := openTestIndex(t)

	_ = idx.UpsertTask(makeTestTask("01LIST0001", "Task 1"), "inbox/01LIST0001.md")
	_ = idx.UpsertTask(makeTestTask("01LIST0002", "Task 2"), "inbox/01LIST0002.md")
	_ = idx.UpsertTask(makeTestTask("01LIST0003", "Task 3"), "inbox/01LIST0003.md")

	results, err := idx.ListTasks(nil)
	if err != nil {
		t.Fatalf("ListTasks(nil) error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("ListTasks(nil) returned %d results, want 3", len(results))
	}
}

func TestListTasksFilterByStatus(t *testing.T) {
	idx := openTestIndex(t)

	pending := makeTestTask("01FILT0001", "Pending task")
	_ = idx.UpsertTask(pending, "inbox/01FILT0001.md")

	done := makeTestTask("01FILT0002", "Done task")
	done.Status = model.StatusDone
	_ = idx.UpsertTask(done, "inbox/01FILT0002.md")

	results, err := idx.ListTasks(&TaskFilter{Status: model.StatusPending})
	if err != nil {
		t.Fatalf("ListTasks(pending) error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "01FILT0001" {
		t.Errorf("ListTasks(pending) = %v, want [01FILT0001]", results)
	}
}

func TestListTasksFilterByTag(t *testing.T) {
	idx := openTestIndex(t)

	task1 := makeTestTask("01TAG00001", "Errands task")
	task1.Tags = []string{"errands", "home"}
	_ = idx.UpsertTask(task1, "inbox/01TAG00001.md")

	task2 := makeTestTask("01TAG00002", "Work task")
	task2.Tags = []string{"work"}
	_ = idx.UpsertTask(task2, "inbox/01TAG00002.md")

	results, err := idx.ListTasks(&TaskFilter{Tag: "errands"})
	if err != nil {
		t.Fatalf("ListTasks(tag=errands) error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "01TAG00001" {
		t.Errorf("ListTasks(tag=errands) returned wrong results")
	}
}

func TestListTasksFilterByContext(t *testing.T) {
	idx := openTestIndex(t)

	task1 := makeTestTask("01CTX00001", "Home task")
	task1.Context = []string{"home"}
	_ = idx.UpsertTask(task1, "inbox/01CTX00001.md")

	task2 := makeTestTask("01CTX00002", "Office task")
	task2.Context = []string{"office"}
	_ = idx.UpsertTask(task2, "inbox/01CTX00002.md")

	results, err := idx.ListTasks(&TaskFilter{Context: "home"})
	if err != nil {
		t.Fatalf("ListTasks(context=home) error: %v", err)
	}
	if len(results) != 1 || results[0].ID != "01CTX00001" {
		t.Errorf("ListTasks(context=home) returned wrong results")
	}
}

func TestListTasksWithLimit(t *testing.T) {
	idx := openTestIndex(t)

	for i := 0; i < 5; i++ {
		id := "01LIM" + string(rune('A'+i)) + "0001"
		_ = idx.UpsertTask(makeTestTask(id, "Task"), "inbox/"+id+".md")
	}

	results, err := idx.ListTasks(&TaskFilter{Limit: 2})
	if err != nil {
		t.Fatalf("ListTasks(limit=2) error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ListTasks(limit=2) returned %d results, want 2", len(results))
	}
}

func TestRebuildIndex(t *testing.T) {
	idx := openTestIndex(t)
	dataDir := t.TempDir()

	// Create some task files
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)

	task1 := &model.Task{
		ID: "01REBUILD001", Title: "Task one", Type: model.TaskTypeOneShot,
		Status: model.StatusPending, Created: now, Updated: now,
		Tags: []string{"rebuild-test"},
	}
	task2 := &model.Task{
		ID: "01REBUILD002", Title: "Task two", Type: model.TaskTypeDated,
		Status: model.StatusActive, DueDate: "2026-04-10",
		Created: now, Updated: now,
	}

	_ = WriteFile(filepath.Join(dataDir, "inbox", "01REBUILD001.md"), task1)
	_ = WriteFile(filepath.Join(dataDir, "inbox", "01REBUILD002.md"), task2)

	// Also create a malformed file to test graceful handling
	badDir := filepath.Join(dataDir, "inbox")
	_ = os.WriteFile(filepath.Join(badDir, "bad.md"), []byte("---\nnot: [[[valid\n---\n"), 0o644)

	// Rebuild
	count, err := idx.RebuildIndex(dataDir)
	if err != nil {
		t.Fatalf("RebuildIndex() error: %v", err)
	}
	if count != 2 {
		t.Errorf("RebuildIndex() indexed %d files, want 2", count)
	}

	// Verify tasks are in the index
	total, _ := idx.TaskCount()
	if total != 2 {
		t.Errorf("TaskCount after rebuild = %d, want 2", total)
	}

	// Verify specific task
	got, err := idx.GetTask("01REBUILD001")
	if err != nil {
		t.Fatalf("GetTask after rebuild error: %v", err)
	}
	if got.Title != "Task one" {
		t.Errorf("Title = %q, want %q", got.Title, "Task one")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "rebuild-test" {
		t.Errorf("Tags = %v, want [rebuild-test]", got.Tags)
	}
}

func TestRebuildIndexSkipsHiddenDirs(t *testing.T) {
	idx := openTestIndex(t)
	dataDir := t.TempDir()

	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID: "01VISIBLE01", Title: "Visible", Type: model.TaskTypeOneShot,
		Status: model.StatusPending, Created: now, Updated: now,
	}
	_ = WriteFile(filepath.Join(dataDir, "inbox", "01VISIBLE01.md"), task)

	// Create a file inside .bentotask/ (should be skipped)
	hiddenDir := filepath.Join(dataDir, ".bentotask")
	_ = os.MkdirAll(hiddenDir, 0o755)
	_ = os.WriteFile(filepath.Join(hiddenDir, "index.db"), []byte("fake"), 0o644)

	count, err := idx.RebuildIndex(dataDir)
	if err != nil {
		t.Fatalf("RebuildIndex() error: %v", err)
	}
	if count != 1 {
		t.Errorf("RebuildIndex() indexed %d files, want 1 (should skip .bentotask/)", count)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	idx := openTestIndex(t)

	_, err := idx.GetTask("NONEXISTENT")
	if err == nil {
		t.Error("GetTask(nonexistent) should return error")
	}
}

func TestSearchByTitle(t *testing.T) {
	idx := openTestIndex(t)

	task1 := makeTestTask("01SRCH0001", "Buy groceries for dinner")
	task1.Body = "Need milk, eggs, and bread"
	_ = idx.UpsertTask(task1, "inbox/01SRCH0001.md")

	task2 := makeTestTask("01SRCH0002", "Write project report")
	task2.Body = "Quarterly status update for management"
	_ = idx.UpsertTask(task2, "inbox/01SRCH0002.md")

	task3 := makeTestTask("01SRCH0003", "Plan grocery list")
	task3.Body = "Weekly shopping needs"
	_ = idx.UpsertTask(task3, "inbox/01SRCH0003.md")

	// Search by title word
	results, err := idx.Search("groceries")
	if err != nil {
		t.Fatalf("Search('groceries') error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search('groceries') returned %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].ID != "01SRCH0001" {
		t.Errorf("Search('groceries') returned %q, want 01SRCH0001", results[0].ID)
	}
}

func TestSearchByBody(t *testing.T) {
	idx := openTestIndex(t)

	task1 := makeTestTask("01BODY0001", "Daily standup")
	task1.Body = "Discuss blockers and progress with the engineering team"
	_ = idx.UpsertTask(task1, "inbox/01BODY0001.md")

	task2 := makeTestTask("01BODY0002", "Lunch meeting")
	task2.Body = "Meet with sales team at noon"
	_ = idx.UpsertTask(task2, "inbox/01BODY0002.md")

	// Search by body content
	results, err := idx.Search("blockers")
	if err != nil {
		t.Fatalf("Search('blockers') error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search('blockers') returned %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].ID != "01BODY0001" {
		t.Errorf("Search('blockers') returned %q, want 01BODY0001", results[0].ID)
	}
}

func TestSearchNoResults(t *testing.T) {
	idx := openTestIndex(t)

	task := makeTestTask("01NORES001", "Simple task")
	task.Body = "Nothing special here"
	_ = idx.UpsertTask(task, "inbox/01NORES001.md")

	results, err := idx.Search("nonexistentword")
	if err != nil {
		t.Fatalf("Search('nonexistentword') error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search('nonexistentword') returned %d results, want 0", len(results))
	}
}

func TestSearchAfterDelete(t *testing.T) {
	idx := openTestIndex(t)

	task := makeTestTask("01SRDEL001", "Searchable task to delete")
	task.Body = "This task has unique content xylophone"
	_ = idx.UpsertTask(task, "inbox/01SRDEL001.md")

	// Should find it
	results, _ := idx.Search("xylophone")
	if len(results) != 1 {
		t.Fatalf("Search before delete returned %d results, want 1", len(results))
	}

	// Delete and search again
	_ = idx.DeleteTask("01SRDEL001")
	results, err := idx.Search("xylophone")
	if err != nil {
		t.Fatalf("Search after delete error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search after delete returned %d results, want 0", len(results))
	}
}

func TestSearchAfterUpdate(t *testing.T) {
	idx := openTestIndex(t)

	task := makeTestTask("01SRUPD001", "Original title")
	task.Body = "Original body with keyword alpha"
	_ = idx.UpsertTask(task, "inbox/01SRUPD001.md")

	// Update the task — FTS should reflect new content
	task.Title = "Updated title"
	task.Body = "New body with keyword bravo"
	_ = idx.UpsertTask(task, "inbox/01SRUPD001.md")

	// Old keyword should not match
	results, _ := idx.Search("alpha")
	if len(results) != 0 {
		t.Errorf("Search('alpha') after update returned %d results, want 0", len(results))
	}

	// New keyword should match
	results, err := idx.Search("bravo")
	if err != nil {
		t.Fatalf("Search('bravo') error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search('bravo') after update returned %d results, want 1", len(results))
	}
}
