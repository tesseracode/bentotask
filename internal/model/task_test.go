package model

import (
	"testing"
	"time"
)

// newValidTask creates a minimal valid task for testing.
// Tests can then modify specific fields to test edge cases.
func newValidTask() Task {
	now := time.Now().UTC()
	return Task{
		ID:      "01JQX00001ABCDEF12345678",
		Title:   "Test task",
		Type:    TaskTypeOneShot,
		Status:  StatusPending,
		Created: now,
		Updated: now,
	}
}

func TestValidTask(t *testing.T) {
	task := newValidTask()
	errs := task.Validate()
	if len(errs) != 0 {
		t.Errorf("valid task returned errors: %v", errs)
	}
}

func TestValidateRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Task)
		wantErr string
	}{
		{"missing id", func(tk *Task) { tk.ID = "" }, "id is required"},
		{"missing title", func(tk *Task) { tk.Title = "" }, "title is required"},
		{"missing type", func(tk *Task) { tk.Type = "" }, "type is required"},
		{"missing status", func(tk *Task) { tk.Status = "" }, "status is required"},
		{"missing created", func(tk *Task) { tk.Created = time.Time{} }, "created is required"},
		{"missing updated", func(tk *Task) { tk.Updated = time.Time{} }, "updated is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newValidTask()
			tt.modify(&task)
			errs := task.Validate()
			if !containsError(errs, tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, errs)
			}
		})
	}
}

func TestValidateEnumValues(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Task)
		wantErr string
	}{
		{"invalid type", func(tk *Task) { tk.Type = "bogus" }, "invalid type"},
		{"invalid status", func(tk *Task) { tk.Status = "bogus" }, "invalid status"},
		{"invalid priority", func(tk *Task) { tk.Priority = "bogus" }, "invalid priority"},
		{"invalid energy", func(tk *Task) { tk.Energy = "bogus" }, "invalid energy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newValidTask()
			tt.modify(&task)
			errs := task.Validate()
			if !containsError(errs, tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, errs)
			}
		})
	}
}

func TestValidateTaskTypeSpecific(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Task)
		wantErr string
	}{
		{
			"dated without due_date",
			func(tk *Task) { tk.Type = TaskTypeDated },
			"dated tasks require due_date",
		},
		{
			"ranged without due_start",
			func(tk *Task) { tk.Type = TaskTypeRanged; tk.DueEnd = "2026-04-10" },
			"ranged tasks require due_start and due_end",
		},
		{
			"recurring without recurrence",
			func(tk *Task) { tk.Type = TaskTypeRecurring },
			"recurring tasks require recurrence rule",
		},
		{
			"habit without frequency",
			func(tk *Task) {
				tk.Type = TaskTypeHabit
				tk.Recurrence = "FREQ=DAILY"
			},
			"habits require frequency",
		},
		{
			"routine without steps",
			func(tk *Task) { tk.Type = TaskTypeRoutine },
			"routines require at least one step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := newValidTask()
			tt.modify(&task)
			errs := task.Validate()
			if !containsError(errs, tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, errs)
			}
		})
	}
}

func TestValidateDatedTaskValid(t *testing.T) {
	task := newValidTask()
	task.Type = TaskTypeDated
	task.DueDate = "2026-04-10"
	errs := task.Validate()
	if len(errs) != 0 {
		t.Errorf("valid dated task returned errors: %v", errs)
	}
}

func TestValidateRangedTaskValid(t *testing.T) {
	task := newValidTask()
	task.Type = TaskTypeRanged
	task.DueStart = "2026-04-07"
	task.DueEnd = "2026-04-13"
	errs := task.Validate()
	if len(errs) != 0 {
		t.Errorf("valid ranged task returned errors: %v", errs)
	}
}

func TestValidateHabitValid(t *testing.T) {
	task := newValidTask()
	task.Type = TaskTypeHabit
	task.Recurrence = "FREQ=DAILY"
	task.Frequency = &HabitFrequency{Type: "daily", Target: 1}
	errs := task.Validate()
	if len(errs) != 0 {
		t.Errorf("valid habit returned errors: %v", errs)
	}
}

func TestValidateRoutineValid(t *testing.T) {
	task := newValidTask()
	task.Type = TaskTypeRoutine
	task.Steps = []RoutineStep{{Ref: "01JQX00002ABCDEF12345678", Optional: false}}
	errs := task.Validate()
	if len(errs) != 0 {
		t.Errorf("valid routine returned errors: %v", errs)
	}
}

func TestValidateLinks(t *testing.T) {
	task := newValidTask()
	task.Links = []Link{
		{Type: LinkDependsOn, Target: "01JQX00002ABCDEF12345678"},
		{Type: "bogus", Target: "01JQX00003ABCDEF12345678"},
		{Type: LinkBlocks, Target: ""},
	}
	errs := task.Validate()
	if !containsError(errs, "link[1]: invalid type") {
		t.Errorf("expected link type error, got: %v", errs)
	}
	if !containsError(errs, "link[2]: target is required") {
		t.Errorf("expected link target error, got: %v", errs)
	}
}

func TestIsDone(t *testing.T) {
	task := newValidTask()

	task.Status = StatusPending
	if task.IsDone() {
		t.Error("pending task should not be done")
	}

	task.Status = StatusDone
	if !task.IsDone() {
		t.Error("done task should be done")
	}

	task.Status = StatusCancelled
	if !task.IsDone() {
		t.Error("cancelled task should be done")
	}
}

func TestShortID(t *testing.T) {
	task := newValidTask()

	got := task.ShortID(8)
	if got != "01JQX000" {
		t.Errorf("ShortID(8) = %q, want %q", got, "01JQX000")
	}

	// If n >= len(ID), return full ID
	got = task.ShortID(100)
	if got != task.ID {
		t.Errorf("ShortID(100) = %q, want full ID %q", got, task.ID)
	}
}

func TestHasTag(t *testing.T) {
	task := newValidTask()
	task.Tags = []string{"home", "Errands", "urgent"}

	if !task.HasTag("home") {
		t.Error("should find 'home' tag")
	}
	if !task.HasTag("ERRANDS") {
		t.Error("should find 'Errands' case-insensitively")
	}
	if task.HasTag("work") {
		t.Error("should not find 'work' tag")
	}
}

func TestHasContext(t *testing.T) {
	task := newValidTask()
	task.Context = []string{"home", "Office"}

	if !task.HasContext("home") {
		t.Error("should find 'home' context")
	}
	if !task.HasContext("OFFICE") {
		t.Error("should find 'Office' case-insensitively")
	}
	if task.HasContext("errands") {
		t.Error("should not find 'errands' context")
	}
}

func TestIsValid(t *testing.T) {
	task := newValidTask()
	if !task.IsValid() {
		t.Error("valid task should return IsValid() == true")
	}

	task.ID = ""
	if task.IsValid() {
		t.Error("invalid task should return IsValid() == false")
	}
}

// --- Helper ---

func containsError(errs []string, substr string) bool {
	for _, e := range errs {
		if contains(e, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
