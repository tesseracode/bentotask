package app

import (
	"os"
	"testing"
)

func openTestApp(t *testing.T) *App {
	t.Helper()
	dataDir := t.TempDir()
	a, err := Open(dataDir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}

func TestOpenCreatesDirectories(t *testing.T) {
	dataDir := t.TempDir()
	a, err := Open(dataDir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer func() { _ = a.Close() }()

	// Check inbox exists
	if _, err := os.Stat(dataDir + "/inbox"); os.IsNotExist(err) {
		t.Error("inbox directory not created")
	}
	// Check .bentotask exists
	if _, err := os.Stat(dataDir + "/.bentotask"); os.IsNotExist(err) {
		t.Error(".bentotask directory not created")
	}
}

func TestAddTask(t *testing.T) {
	a := openTestApp(t)

	task, err := a.AddTask("Buy groceries", TaskOptions{
		Priority: "high",
		Energy:   "low",
		Tags:     []string{"errands"},
	})
	if err != nil {
		t.Fatalf("AddTask() error: %v", err)
	}

	if task.ID == "" {
		t.Error("task should have an ID")
	}
	if task.Title != "Buy groceries" {
		t.Errorf("Title = %q, want %q", task.Title, "Buy groceries")
	}
	if task.Status != "pending" {
		t.Errorf("Status = %q, want %q", task.Status, "pending")
	}
	if len(task.ID) != 26 {
		t.Errorf("ID length = %d, want 26 (ULID)", len(task.ID))
	}

	// Should be in the index
	count, _ := a.Index.TaskCount()
	if count != 1 {
		t.Errorf("TaskCount = %d, want 1", count)
	}

	// Should be on disk
	_, relPath, err := a.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if relPath == "" {
		t.Error("task should have a file path")
	}
}

func TestAddTaskWithDueDate(t *testing.T) {
	a := openTestApp(t)

	task, err := a.AddTask("Dentist", TaskOptions{DueDate: "2026-04-15"})
	if err != nil {
		t.Fatalf("AddTask() error: %v", err)
	}

	// Should auto-set type to "dated"
	if task.Type != "dated" {
		t.Errorf("Type = %q, want %q (auto-set from due_date)", task.Type, "dated")
	}
	if task.DueDate != "2026-04-15" {
		t.Errorf("DueDate = %q, want %q", task.DueDate, "2026-04-15")
	}
}

func TestAddTaskInBox(t *testing.T) {
	a := openTestApp(t)

	task, err := a.AddTask("Paint walls", TaskOptions{Box: "projects/home-reno"})
	if err != nil {
		t.Fatalf("AddTask() error: %v", err)
	}

	_, relPath, err := a.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if relPath != "projects/home-reno/"+task.ID+".md" {
		t.Errorf("file path = %q, want task in projects/home-reno/", relPath)
	}
}

func TestGetTaskByPrefix(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Test prefix", TaskOptions{})

	// Should work with prefix
	prefix := task.ID[:8]
	got, _, err := a.GetTask(prefix)
	if err != nil {
		t.Fatalf("GetTask(%q) error: %v", prefix, err)
	}
	if got.ID != task.ID {
		t.Errorf("GetTask prefix returned wrong task: %q", got.ID)
	}
}

func TestGetTaskNotFound(t *testing.T) {
	a := openTestApp(t)

	_, _, err := a.GetTask("NONEXISTENT")
	if err == nil {
		t.Error("GetTask should return error for nonexistent ID")
	}
}

func TestCompleteTask(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Complete me", TaskOptions{})

	completed, err := a.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("CompleteTask() error: %v", err)
	}
	if completed.Status != "done" {
		t.Errorf("Status = %q, want %q", completed.Status, "done")
	}
	if completed.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}

	// Verify persisted to disk
	reloaded, _, _ := a.GetTask(task.ID)
	if reloaded.Status != "done" {
		t.Errorf("persisted Status = %q, want %q", reloaded.Status, "done")
	}
}

func TestCompleteTaskAlreadyDone(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Already done", TaskOptions{})
	_, _ = a.CompleteTask(task.ID)

	_, err := a.CompleteTask(task.ID)
	if err == nil {
		t.Error("completing an already-done task should return error")
	}
}

func TestDeleteTask(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Delete me", TaskOptions{})

	deleted, err := a.DeleteTask(task.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error: %v", err)
	}
	if deleted.Title != "Delete me" {
		t.Errorf("Title = %q, want %q", deleted.Title, "Delete me")
	}

	// Should be gone
	count, _ := a.Index.TaskCount()
	if count != 0 {
		t.Errorf("TaskCount = %d, want 0 after delete", count)
	}

	_, _, err = a.GetTask(task.ID)
	if err == nil {
		t.Error("GetTask should fail after delete")
	}
}

func TestListTasks(t *testing.T) {
	a := openTestApp(t)

	_, _ = a.AddTask("Task A", TaskOptions{Tags: []string{"work"}})
	_, _ = a.AddTask("Task B", TaskOptions{Tags: []string{"home"}})
	_, _ = a.AddTask("Task C", TaskOptions{Tags: []string{"work"}})

	// All tasks
	all, err := a.ListTasks(nil)
	if err != nil {
		t.Fatalf("ListTasks(nil) error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListTasks(nil) = %d results, want 3", len(all))
	}
}

func TestRebuildIndex(t *testing.T) {
	a := openTestApp(t)

	_, _ = a.AddTask("Task 1", TaskOptions{})
	_, _ = a.AddTask("Task 2", TaskOptions{})

	count, err := a.RebuildIndex()
	if err != nil {
		t.Fatalf("RebuildIndex() error: %v", err)
	}
	if count != 2 {
		t.Errorf("RebuildIndex() indexed %d, want 2", count)
	}
}
