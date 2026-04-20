package app

import (
	"os"
	"strings"
	"testing"

	"github.com/tesserabox/bentotask/internal/model"
	"github.com/tesserabox/bentotask/internal/store"
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

func TestUpdateTask(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Original", TaskOptions{Priority: "low"})

	updated, err := a.UpdateTask(task.ID, func(tk *model.Task) {
		tk.Title = "Modified"
		tk.Priority = "high"
		tk.Tags = []string{"updated"}
	})
	if err != nil {
		t.Fatalf("UpdateTask() error: %v", err)
	}
	if updated.Title != "Modified" {
		t.Errorf("Title = %q, want %q", updated.Title, "Modified")
	}
	if updated.Priority != "high" {
		t.Errorf("Priority = %q, want %q", updated.Priority, "high")
	}

	// Verify persisted
	reloaded, _, _ := a.GetTask(task.ID)
	if reloaded.Title != "Modified" {
		t.Errorf("persisted Title = %q, want %q", reloaded.Title, "Modified")
	}
	if len(reloaded.Tags) != 1 || reloaded.Tags[0] != "updated" {
		t.Errorf("persisted Tags = %v, want [updated]", reloaded.Tags)
	}
}

func TestUpdateTaskValidation(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Valid task", TaskOptions{})

	_, err := a.UpdateTask(task.ID, func(tk *model.Task) {
		tk.Title = "" // Invalid — title is required
	})
	if err == nil {
		t.Error("UpdateTask with empty title should return validation error")
	}

	// Original should be unchanged
	reloaded, _, _ := a.GetTask(task.ID)
	if reloaded.Title != "Valid task" {
		t.Errorf("Title should be unchanged after failed update, got %q", reloaded.Title)
	}
}

func TestEditTaskFile(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Edit me", TaskOptions{})

	path, err := a.EditTaskFile(task.ID)
	if err != nil {
		t.Fatalf("EditTaskFile() error: %v", err)
	}
	if path == "" {
		t.Error("path should not be empty")
	}
	// Should be an absolute path ending in .md
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("path should end with .md, got %q", path)
	}
}

func TestReloadTask(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Before edit", TaskOptions{})

	// Simulate an external edit by modifying the file directly
	filePath, _ := a.EditTaskFile(task.ID)
	updated := *task
	updated.Title = "After edit"
	_ = store.WriteFile(filePath, &updated)

	reloaded, err := a.ReloadTask(task.ID)
	if err != nil {
		t.Fatalf("ReloadTask() error: %v", err)
	}
	if reloaded.Title != "After edit" {
		t.Errorf("Title = %q, want %q", reloaded.Title, "After edit")
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

func TestSearchTasks(t *testing.T) {
	a := openTestApp(t)

	_, _ = a.AddTask("Buy groceries at the store", TaskOptions{Tags: []string{"errands"}})
	_, _ = a.AddTask("Write quarterly report", TaskOptions{Tags: []string{"work"}})
	_, _ = a.AddTask("Clean the kitchen", TaskOptions{Tags: []string{"home"}})

	results, err := a.SearchTasks("groceries")
	if err != nil {
		t.Fatalf("SearchTasks() error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("SearchTasks('groceries') = %d results, want 1", len(results))
	}
	if len(results) > 0 && results[0].Title != "Buy groceries at the store" {
		t.Errorf("SearchTasks title = %q, want %q", results[0].Title, "Buy groceries at the store")
	}
}

func TestSearchTasksEmptyQuery(t *testing.T) {
	a := openTestApp(t)

	_, err := a.SearchTasks("")
	if err == nil {
		t.Error("SearchTasks('') should return error")
	}
}

// --- Habit tests ---

func TestAddHabit(t *testing.T) {
	a := openTestApp(t)

	h, err := a.AddHabit("Read 30 minutes", HabitOptions{
		FreqType:   "daily",
		FreqTarget: 1,
		Recurrence: "FREQ=DAILY",
	})
	if err != nil {
		t.Fatalf("AddHabit() error: %v", err)
	}
	if h.Type != "habit" {
		t.Errorf("Type = %q, want 'habit'", h.Type)
	}
	if h.Status != "active" {
		t.Errorf("Status = %q, want 'active'", h.Status)
	}
	if h.Frequency == nil || h.Frequency.Type != "daily" {
		t.Errorf("Frequency = %v, want daily", h.Frequency)
	}
	if h.Recurrence != "FREQ=DAILY" {
		t.Errorf("Recurrence = %q, want 'FREQ=DAILY'", h.Recurrence)
	}
}

func TestAddHabitInvalidRRULE(t *testing.T) {
	a := openTestApp(t)

	_, err := a.AddHabit("Bad habit", HabitOptions{
		FreqType:   "daily",
		FreqTarget: 1,
		Recurrence: "INVALID",
	})
	if err == nil {
		t.Error("AddHabit with invalid RRULE should return error")
	}
}

func TestLogHabit(t *testing.T) {
	a := openTestApp(t)

	h, _ := a.AddHabit("Exercise", HabitOptions{
		FreqType:   "daily",
		FreqTarget: 1,
		Recurrence: "FREQ=DAILY",
	})

	updated, err := a.LogHabit(h.ID, 30, "Morning run")
	if err != nil {
		t.Fatalf("LogHabit() error: %v", err)
	}
	if updated.StreakCurrent != 1 {
		t.Errorf("StreakCurrent = %d, want 1", updated.StreakCurrent)
	}
	if !strings.Contains(updated.Body, "## Completions") {
		t.Error("Body should contain ## Completions section")
	}
	if !strings.Contains(updated.Body, "30min") {
		t.Error("Body should contain duration")
	}
	if !strings.Contains(updated.Body, "Morning run") {
		t.Error("Body should contain note")
	}
}

func TestLogHabitNonHabit(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Regular task", TaskOptions{})
	_, err := a.LogHabit(task.ID, 0, "")
	if err == nil {
		t.Error("LogHabit on non-habit should return error")
	}
}

func TestHabitStats(t *testing.T) {
	a := openTestApp(t)

	h, _ := a.AddHabit("Meditate", HabitOptions{
		FreqType:   "daily",
		FreqTarget: 1,
		Recurrence: "FREQ=DAILY",
	})

	// Log a completion
	_, _ = a.LogHabit(h.ID, 20, "")

	task, stats, err := a.HabitStats(h.ID)
	if err != nil {
		t.Fatalf("HabitStats() error: %v", err)
	}
	if task.Title != "Meditate" {
		t.Errorf("Title = %q, want 'Meditate'", task.Title)
	}
	if stats.TotalCompletions != 1 {
		t.Errorf("TotalCompletions = %d, want 1", stats.TotalCompletions)
	}
	if stats.CurrentStreak != 1 {
		t.Errorf("CurrentStreak = %d, want 1", stats.CurrentStreak)
	}
}

func TestListHabits(t *testing.T) {
	a := openTestApp(t)

	// Add a habit and a regular task
	_, _ = a.AddHabit("Read", HabitOptions{
		FreqType: "daily", FreqTarget: 1, Recurrence: "FREQ=DAILY",
	})
	_, _ = a.AddTask("Regular task", TaskOptions{})

	habits, err := a.ListHabits()
	if err != nil {
		t.Fatalf("ListHabits() error: %v", err)
	}
	if len(habits) != 1 {
		t.Errorf("ListHabits() = %d, want 1 (should only return habits)", len(habits))
	}
}

// --- Routine tests ---

func TestAddRoutine(t *testing.T) {
	a := openTestApp(t)

	r, err := a.AddRoutine("Morning Routine", RoutineOptions{
		Steps: []model.RoutineStep{
			{Title: "Shower", Duration: 5},
			{Title: "Breakfast", Duration: 15},
			{Title: "Review inbox", Duration: 10},
		},
	})
	if err != nil {
		t.Fatalf("AddRoutine() error: %v", err)
	}
	if r.Type != "routine" {
		t.Errorf("Type = %q, want 'routine'", r.Type)
	}
	if r.Status != "active" {
		t.Errorf("Status = %q, want 'active'", r.Status)
	}
	if len(r.Steps) != 3 {
		t.Errorf("Steps = %d, want 3", len(r.Steps))
	}
	if r.EstimatedDuration != 30 {
		t.Errorf("EstimatedDuration = %d, want 30 (sum of step durations)", r.EstimatedDuration)
	}
}

func TestAddRoutineWithSchedule(t *testing.T) {
	a := openTestApp(t)

	r, err := a.AddRoutine("Evening Wind-down", RoutineOptions{
		Steps: []model.RoutineStep{
			{Title: "Journal", Duration: 10},
			{Title: "Read", Duration: 30},
		},
		Schedule: &model.RoutineSchedule{
			Time: "21:00",
			Days: []string{"mon", "tue", "wed", "thu", "fri"},
		},
	})
	if err != nil {
		t.Fatalf("AddRoutine() error: %v", err)
	}
	if r.Schedule == nil {
		t.Fatal("Schedule should not be nil")
	}
	if r.Schedule.Time != "21:00" {
		t.Errorf("Schedule.Time = %q, want '21:00'", r.Schedule.Time)
	}
	if len(r.Schedule.Days) != 5 {
		t.Errorf("Schedule.Days = %d, want 5", len(r.Schedule.Days))
	}
}

func TestAddRoutineNoSteps(t *testing.T) {
	a := openTestApp(t)

	_, err := a.AddRoutine("Empty", RoutineOptions{
		Steps: []model.RoutineStep{},
	})
	if err == nil {
		t.Error("AddRoutine with no steps should return validation error")
	}
}

func TestAddRoutineOptionalSteps(t *testing.T) {
	a := openTestApp(t)

	r, err := a.AddRoutine("Flexible morning", RoutineOptions{
		Steps: []model.RoutineStep{
			{Title: "Coffee", Duration: 5},
			{Title: "Stretch", Duration: 10, Optional: true},
		},
	})
	if err != nil {
		t.Fatalf("AddRoutine() error: %v", err)
	}
	if !r.Steps[1].Optional {
		t.Error("Step[1] should be optional")
	}
}

func TestListRoutines(t *testing.T) {
	a := openTestApp(t)

	// Add a routine and a regular task
	_, _ = a.AddRoutine("Morning", RoutineOptions{
		Steps: []model.RoutineStep{{Title: "Wake up"}},
	})
	_, _ = a.AddTask("Regular task", TaskOptions{})

	routines, err := a.ListRoutines()
	if err != nil {
		t.Fatalf("ListRoutines() error: %v", err)
	}
	if len(routines) != 1 {
		t.Errorf("ListRoutines() = %d, want 1 (should only return routines)", len(routines))
	}
}

// --- Linking tests ---

func TestLinkTasks(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	source, target, err := a.LinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	if err != nil {
		t.Fatalf("LinkTasks() error: %v", err)
	}
	if source.ID != taskA.ID {
		t.Errorf("source.ID = %q, want %q", source.ID, taskA.ID)
	}
	if target.ID != taskB.ID {
		t.Errorf("target.ID = %q, want %q", target.ID, taskB.ID)
	}
	if len(source.Links) != 1 {
		t.Fatalf("source.Links = %d, want 1", len(source.Links))
	}
	if source.Links[0].Type != model.LinkRelatedTo {
		t.Errorf("link type = %q, want %q", source.Links[0].Type, model.LinkRelatedTo)
	}
	if source.Links[0].Target != taskB.ID {
		t.Errorf("link target = %q, want %q", source.Links[0].Target, taskB.ID)
	}
}

func TestLinkTasksDependsOn(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	_, _, err := a.LinkTasks(taskA.ID, taskB.ID, model.LinkDependsOn)
	if err != nil {
		t.Fatalf("LinkTasks(depends-on) error: %v", err)
	}

	// Verify persisted to disk
	reloaded, _, _ := a.GetTask(taskA.ID)
	if len(reloaded.Links) != 1 {
		t.Fatalf("persisted links = %d, want 1", len(reloaded.Links))
	}
	if reloaded.Links[0].Type != model.LinkDependsOn {
		t.Errorf("persisted link type = %q, want depends-on", reloaded.Links[0].Type)
	}
}

func TestLinkTasksInvalidType(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	_, _, err := a.LinkTasks(taskA.ID, taskB.ID, model.LinkType("garbage"))
	if err == nil {
		t.Error("invalid link type should return error")
	}
}

func TestLinkTasksSelfLink(t *testing.T) {
	a := openTestApp(t)

	task, _ := a.AddTask("Task", TaskOptions{})

	_, _, err := a.LinkTasks(task.ID, task.ID, model.LinkRelatedTo)
	if err == nil {
		t.Error("self-link should return error")
	}
}

func TestLinkTasksDuplicate(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	_, _, _ = a.LinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	_, _, err := a.LinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	if err == nil {
		t.Error("duplicate link should return error")
	}
}

func TestLinkTasksDifferentTypes(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	// Same pair can have different link types
	_, _, err := a.LinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	if err != nil {
		t.Fatalf("first link error: %v", err)
	}
	_, _, err = a.LinkTasks(taskA.ID, taskB.ID, model.LinkDependsOn)
	if err != nil {
		t.Fatalf("second link (different type) error: %v", err)
	}

	reloaded, _, _ := a.GetTask(taskA.ID)
	if len(reloaded.Links) != 2 {
		t.Errorf("links = %d, want 2 (different types allowed)", len(reloaded.Links))
	}
}

func TestLinkTasksCycleDetection(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})
	taskC, _ := a.AddTask("Task C", TaskOptions{})

	// A depends-on B
	_, _, _ = a.LinkTasks(taskA.ID, taskB.ID, model.LinkDependsOn)
	// B depends-on C
	_, _, _ = a.LinkTasks(taskB.ID, taskC.ID, model.LinkDependsOn)
	// C depends-on A would create a cycle
	_, _, err := a.LinkTasks(taskC.ID, taskA.ID, model.LinkDependsOn)
	if err == nil {
		t.Error("cycle should be detected and rejected")
	}
}

func TestLinkTasksCycleDetectionBlocks(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	// A blocks B
	_, _, _ = a.LinkTasks(taskA.ID, taskB.ID, model.LinkBlocks)
	// B blocks A would create a cycle
	_, _, err := a.LinkTasks(taskB.ID, taskA.ID, model.LinkBlocks)
	if err == nil {
		t.Error("blocks cycle should be detected")
	}
}

func TestLinkTasksRelatedToNoCycleCheck(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	// related-to links are bidirectional/informational — no cycle check
	_, _, _ = a.LinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	_, _, err := a.LinkTasks(taskB.ID, taskA.ID, model.LinkRelatedTo)
	if err != nil {
		t.Errorf("related-to should allow bidirectional links, got: %v", err)
	}
}

func TestUnlinkTasks(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	_, _, _ = a.LinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)

	source, target, err := a.UnlinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	if err != nil {
		t.Fatalf("UnlinkTasks() error: %v", err)
	}
	if source.ID != taskA.ID || target.ID != taskB.ID {
		t.Errorf("unexpected source/target IDs")
	}

	// Verify link removed from disk
	reloaded, _, _ := a.GetTask(taskA.ID)
	if len(reloaded.Links) != 0 {
		t.Errorf("links after unlink = %d, want 0", len(reloaded.Links))
	}
}

func TestUnlinkTasksNotFound(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})

	_, _, err := a.UnlinkTasks(taskA.ID, taskB.ID, model.LinkRelatedTo)
	if err == nil {
		t.Error("unlinking non-existent link should return error")
	}
}

func TestGetTaskLinks(t *testing.T) {
	a := openTestApp(t)

	taskA, _ := a.AddTask("Task A", TaskOptions{})
	taskB, _ := a.AddTask("Task B", TaskOptions{})
	taskC, _ := a.AddTask("Task C", TaskOptions{})

	// A depends-on B, C related-to A
	_, _, _ = a.LinkTasks(taskA.ID, taskB.ID, model.LinkDependsOn)
	_, _, _ = a.LinkTasks(taskC.ID, taskA.ID, model.LinkRelatedTo)

	links, err := a.GetTaskLinks(taskA.ID)
	if err != nil {
		t.Fatalf("GetTaskLinks() error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("links = %d, want 2 (1 outgoing + 1 incoming)", len(links))
	}

	// Check outgoing
	var outgoing, incoming int
	for _, l := range links {
		if l.Direction == "outgoing" {
			outgoing++
			if l.TaskTitle != "Task B" {
				t.Errorf("outgoing link title = %q, want 'Task B'", l.TaskTitle)
			}
		} else {
			incoming++
			if l.TaskTitle != "Task C" {
				t.Errorf("incoming link title = %q, want 'Task C'", l.TaskTitle)
			}
		}
	}
	if outgoing != 1 || incoming != 1 {
		t.Errorf("outgoing=%d incoming=%d, want 1 and 1", outgoing, incoming)
	}
}

// --- Cycle detection unit tests ---

func TestHasCycleSimple(t *testing.T) {
	// A → B → C → A (cycle)
	graph := map[string][]string{
		"A": {"B"},
		"B": {"C"},
		"C": {"A"},
	}
	if !hasCycle(graph, "A") {
		t.Error("should detect cycle A→B→C→A")
	}
}

func TestHasCycleNoCycle(t *testing.T) {
	// A → B → C (no cycle)
	graph := map[string][]string{
		"A": {"B"},
		"B": {"C"},
	}
	if hasCycle(graph, "A") {
		t.Error("should not detect cycle in linear chain")
	}
}

func TestHasCycleSelfLoop(t *testing.T) {
	graph := map[string][]string{
		"A": {"A"},
	}
	if !hasCycle(graph, "A") {
		t.Error("should detect self-loop")
	}
}

func TestHasCycleDiamond(t *testing.T) {
	// Diamond: A → B, A → C, B → D, C → D (no cycle)
	graph := map[string][]string{
		"A": {"B", "C"},
		"B": {"D"},
		"C": {"D"},
	}
	if hasCycle(graph, "A") {
		t.Error("diamond graph should not have a cycle")
	}
}

func TestHasCycleDisconnected(t *testing.T) {
	// A → B, C → D (disconnected, no cycle from A)
	graph := map[string][]string{
		"A": {"B"},
		"C": {"D"},
	}
	if hasCycle(graph, "A") {
		t.Error("disconnected graph should not have cycle from A")
	}
}

func TestLogHabitMaxPerPeriod(t *testing.T) {
	a := openTestApp(t)

	h, err := a.AddHabit("Meditate", HabitOptions{
		FreqType: "daily", FreqTarget: 1, MaxPerPeriod: 1,
		Recurrence: "FREQ=DAILY",
	})
	if err != nil {
		t.Fatalf("add habit: %v", err)
	}

	// First log should succeed
	_, err = a.LogHabit(h.ID, 10, "")
	if err != nil {
		t.Fatalf("first log error: %v", err)
	}

	// Second log should fail
	_, err = a.LogHabit(h.ID, 10, "")
	if err == nil {
		t.Error("second log should fail with max_per_period=1")
	}
}

func TestLogHabitUnlimited(t *testing.T) {
	a := openTestApp(t)

	h, err := a.AddHabit("Drink water", HabitOptions{
		FreqType: "daily", FreqTarget: 8, MaxPerPeriod: 0,
		Recurrence: "FREQ=DAILY",
	})
	if err != nil {
		t.Fatalf("add habit: %v", err)
	}

	// Multiple logs should all succeed
	for i := 0; i < 3; i++ {
		_, err := a.LogHabit(h.ID, 0, "")
		if err != nil {
			t.Fatalf("log %d error: %v", i+1, err)
		}
	}
}
