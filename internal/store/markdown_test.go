package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tesserabox/bentotask/internal/model"
)

// sampleTaskMarkdown is a realistic task file matching ADR-002 format.
const sampleTaskMarkdown = `---
id: 01JQX00010ABCDEF12345678
title: Paint bedroom
type: one-shot
status: pending
priority: medium
energy: high
estimated_duration: 180
due_start: "2026-04-07"
due_end: "2026-04-13"
tags: [home, renovation]
context: [home]
box: projects/home-renovation
links:
  - type: depends-on
    target: 01JQX00009ABCDEF12345678
created: 2026-04-05T10:30:00Z
updated: 2026-04-05T10:30:00Z
---

# Paint Bedroom

Need to repaint the bedroom walls. Going with the sage green.

## Notes
- Remove furniture first
- Two coats minimum`

func TestParseBasicTask(t *testing.T) {
	r := strings.NewReader(sampleTaskMarkdown)
	task, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Required fields
	assertEqual(t, "ID", task.ID, "01JQX00010ABCDEF12345678")
	assertEqual(t, "Title", task.Title, "Paint bedroom")
	assertEqual(t, "Type", string(task.Type), "one-shot")
	assertEqual(t, "Status", string(task.Status), "pending")

	// Optional fields
	assertEqual(t, "Priority", string(task.Priority), "medium")
	assertEqual(t, "Energy", string(task.Energy), "high")
	if task.EstimatedDuration != 180 {
		t.Errorf("EstimatedDuration = %d, want 180", task.EstimatedDuration)
	}
	assertEqual(t, "DueStart", task.DueStart, "2026-04-07")
	assertEqual(t, "DueEnd", task.DueEnd, "2026-04-13")
	assertEqual(t, "Box", task.Box, "projects/home-renovation")

	// Tags
	if len(task.Tags) != 2 || task.Tags[0] != "home" || task.Tags[1] != "renovation" {
		t.Errorf("Tags = %v, want [home, renovation]", task.Tags)
	}

	// Context
	if len(task.Context) != 1 || task.Context[0] != "home" {
		t.Errorf("Context = %v, want [home]", task.Context)
	}

	// Links
	if len(task.Links) != 1 {
		t.Fatalf("Links count = %d, want 1", len(task.Links))
	}
	assertEqual(t, "Link.Type", string(task.Links[0].Type), "depends-on")
	assertEqual(t, "Link.Target", task.Links[0].Target, "01JQX00009ABCDEF12345678")

	// Timestamps
	if task.Created.IsZero() {
		t.Error("Created should not be zero")
	}

	// Body
	if !strings.Contains(task.Body, "Paint Bedroom") {
		t.Errorf("Body should contain 'Paint Bedroom', got: %q", task.Body[:50])
	}
	if !strings.Contains(task.Body, "Two coats minimum") {
		t.Errorf("Body should contain 'Two coats minimum'")
	}
}

func TestParseMinimalTask(t *testing.T) {
	md := `---
id: 01ABC
title: Simple task
type: floating
status: pending
created: 2026-04-05T10:00:00Z
updated: 2026-04-05T10:00:00Z
---
`
	task, err := Parse(strings.NewReader(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	assertEqual(t, "ID", task.ID, "01ABC")
	assertEqual(t, "Title", task.Title, "Simple task")
	assertEqual(t, "Type", string(task.Type), "floating")
	assertEqual(t, "Body", task.Body, "") // No body content
}

func TestParseHabit(t *testing.T) {
	md := `---
id: 01HABIT001
title: Read 30 minutes
type: habit
status: active
recurrence: "FREQ=DAILY"
frequency:
  type: daily
  target: 1
streak_current: 12
streak_longest: 45
created: 2026-04-01T08:00:00Z
updated: 2026-04-05T08:00:00Z
---

## Completions
- 2026-04-05T08:30:00Z | 35min | "DDIA ch.7"
- 2026-04-04T09:00:00Z | 30min | "DDIA ch.6"`

	task, err := Parse(strings.NewReader(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	assertEqual(t, "Type", string(task.Type), "habit")
	if task.Frequency == nil {
		t.Fatal("Frequency should not be nil for habits")
	}
	assertEqual(t, "Frequency.Type", task.Frequency.Type, "daily")
	if task.Frequency.Target != 1 {
		t.Errorf("Frequency.Target = %d, want 1", task.Frequency.Target)
	}
	if task.StreakCurrent != 12 {
		t.Errorf("StreakCurrent = %d, want 12", task.StreakCurrent)
	}
	if task.StreakLongest != 45 {
		t.Errorf("StreakLongest = %d, want 45", task.StreakLongest)
	}
	if !strings.Contains(task.Body, "Completions") {
		t.Error("Body should contain completions section")
	}
}

func TestParseRoutine(t *testing.T) {
	md := `---
id: 01ROUTINE001
title: Morning routine
type: routine
status: active
steps:
  - ref: 01HABIT001
    optional: false
  - ref: 01HABIT002
    optional: true
schedule:
  time: "07:00"
  days: [mon, tue, wed, thu, fri]
created: 2026-04-01T08:00:00Z
updated: 2026-04-05T08:00:00Z
---
`
	task, err := Parse(strings.NewReader(md))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	assertEqual(t, "Type", string(task.Type), "routine")
	if len(task.Steps) != 2 {
		t.Fatalf("Steps count = %d, want 2", len(task.Steps))
	}
	assertEqual(t, "Step[0].Ref", task.Steps[0].Ref, "01HABIT001")
	if task.Steps[0].Optional {
		t.Error("Step[0] should not be optional")
	}
	if !task.Steps[1].Optional {
		t.Error("Step[1] should be optional")
	}
	if task.Schedule == nil {
		t.Fatal("Schedule should not be nil")
	}
	assertEqual(t, "Schedule.Time", task.Schedule.Time, "07:00")
	if len(task.Schedule.Days) != 5 {
		t.Errorf("Schedule.Days count = %d, want 5", len(task.Schedule.Days))
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 5, 10, 30, 0, 0, time.UTC)
	original := &model.Task{
		ID:                "01JQX00010ABCDEF12345678",
		Title:             "Buy groceries",
		Type:              model.TaskTypeDated,
		Status:            model.StatusPending,
		Priority:          model.PriorityHigh,
		Energy:            model.EnergyLow,
		EstimatedDuration: 45,
		DueDate:           "2026-04-06",
		Tags:              []string{"errands", "home"},
		Context:           []string{"errands"},
		Box:               "inbox",
		Created:           now,
		Updated:           now,
		Body:              "Need to get items for the week.\n\n- [ ] Vegetables\n- [ ] Bread\n- [ ] Chicken",
	}

	// Marshal to bytes
	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// Verify it looks like proper markdown with frontmatter
	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Error("marshaled content should start with ---")
	}
	if !strings.Contains(content, "title: Buy groceries") {
		t.Error("marshaled content should contain title")
	}
	if !strings.Contains(content, "Vegetables") {
		t.Error("marshaled content should contain body")
	}

	// Parse it back
	parsed, err := Parse(strings.NewReader(content))
	if err != nil {
		t.Fatalf("Parse(Marshal()) error: %v", err)
	}

	// Verify round-trip fidelity
	assertEqual(t, "ID", parsed.ID, original.ID)
	assertEqual(t, "Title", parsed.Title, original.Title)
	assertEqual(t, "Type", string(parsed.Type), string(original.Type))
	assertEqual(t, "Status", string(parsed.Status), string(original.Status))
	assertEqual(t, "Priority", string(parsed.Priority), string(original.Priority))
	assertEqual(t, "Energy", string(parsed.Energy), string(original.Energy))
	if parsed.EstimatedDuration != original.EstimatedDuration {
		t.Errorf("EstimatedDuration = %d, want %d", parsed.EstimatedDuration, original.EstimatedDuration)
	}
	assertEqual(t, "DueDate", parsed.DueDate, original.DueDate)
	assertEqual(t, "Box", parsed.Box, original.Box)

	if len(parsed.Tags) != len(original.Tags) {
		t.Errorf("Tags count = %d, want %d", len(parsed.Tags), len(original.Tags))
	}

	if !strings.Contains(parsed.Body, "Vegetables") {
		t.Error("round-tripped body should contain 'Vegetables'")
	}
}

func TestWriteFileAndParseFile(t *testing.T) {
	tmpDir := t.TempDir()

	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID:      "01TESTFILE001",
		Title:   "File I/O test",
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
		Body:    "This task tests file writing and reading.",
	}

	path := filepath.Join(tmpDir, "inbox", "01TESTFILE001.md")

	// Write
	err := WriteFile(path, task)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("WriteFile() did not create the file")
	}

	// Read back
	parsed, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile() error: %v", err)
	}

	assertEqual(t, "ID", parsed.ID, task.ID)
	assertEqual(t, "Title", parsed.Title, task.Title)
	assertEqual(t, "Body", parsed.Body, task.Body)
}

func TestWriteFileAtomicity(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "01ATOMIC.md")

	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID:      "01ATOMIC",
		Title:   "Atomic write test",
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
	}

	// Write the file
	err := WriteFile(path, task)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Verify no temp files remain
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file not cleaned up: %s", e.Name())
		}
	}
}

func TestWriteFileCreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Deeply nested path that doesn't exist yet
	path := filepath.Join(tmpDir, "projects", "home-renovation", "01NESTED.md")

	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID:      "01NESTED",
		Title:   "Nested directory test",
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
	}

	err := WriteFile(path, task)
	if err != nil {
		t.Fatalf("WriteFile() should create parent dirs, got error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("file was not created in nested directory")
	}
}

func TestParseFileNotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/path/to/file.md")
	if err == nil {
		t.Error("ParseFile() should return error for nonexistent file")
	}
}

func TestParseMalformedFrontmatter(t *testing.T) {
	md := `---
this is not valid yaml: [[[
---

Some body.`

	_, err := Parse(strings.NewReader(md))
	if err == nil {
		t.Error("Parse() should return error for malformed YAML")
	}
}

func TestMarshalEmptyBody(t *testing.T) {
	now := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID:      "01NOBODY",
		Title:   "No body task",
		Type:    model.TaskTypeOneShot,
		Status:  model.StatusPending,
		Created: now,
		Updated: now,
		Body:    "",
	}

	data, err := Marshal(task)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	content := string(data)
	// Should end with the closing --- and no trailing blank body section
	if strings.Contains(content, "\n\n\n") {
		t.Error("empty body should not produce extra blank lines")
	}
}

// assertEqual is a simple test helper to reduce boilerplate.
func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}
