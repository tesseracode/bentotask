package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// executeCmdInDir runs a bt command using a specific data directory.
// It resets Cobra's flag state between calls to prevent test pollution.
func executeCmdInDir(t *testing.T, dataDir string, args ...string) (string, error) {
	t.Helper()

	fullArgs := append([]string{"--data-dir", dataDir}, args...)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(fullArgs)

	err := rootCmd.Execute()

	// Reset all persistent and local flag state to prevent leaking between calls.
	// Cobra caches parsed flag values on global command instances, so without this
	// flags like --json, --tag accumulate across test invocations.
	resetFlags()

	return buf.String(), err
}

// resetFlags resets all flag values on every command to their defaults.
// This prevents state from one Execute() call leaking into the next.
func resetFlags() {
	// Walk all commands and reset their flags
	allCmds := []*cobra.Command{
		rootCmd, taskCmd,
		taskAddCmd, addCmd,
		taskListCmd, listCmd,
		taskDoneCmd, doneCmd,
		taskShowCmd,
		taskDeleteCmd,
		taskEditCmd,
		searchCmd,
		indexCmd, indexRebuildCmd,
		habitCmd, habitAddCmd, habitLogCmd, habitStatsCmd, habitListCmd,
		routineCmd, routineCreateCmd, routineListCmd, routineShowCmd, routinePlayCmd,
		linkCmd, unlinkCmd,
		nowCmd, planCmd, planTodayCmd,
		serveCmd,
	}
	for _, cmd := range allCmds {
		cmd.Flags().VisitAll(resetFlag)
		cmd.InheritedFlags().VisitAll(resetFlag)
	}
}

// resetFlag resets a single flag to its default value.
// StringSlice flags need special handling because Set("[]") would
// create a slice with a literal "[]" element.
func resetFlag(f *pflag.Flag) {
	if f.Value.Type() == "stringSlice" {
		// For StringSlice, the only safe reset is to set to empty string
		// which Cobra interprets as an empty slice.
		_ = f.Value.Set("[]")
		// But the above actually sets ["[]"], so use the SliceValue interface
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			_ = sv.Replace(nil)
		}
	} else {
		_ = f.Value.Set(f.DefValue)
	}
	f.Changed = false
}

// --- Integration tests ---

func TestIntegrationAddAndList(t *testing.T) {
	dataDir := t.TempDir()

	// Add a task
	out, err := executeCmdInDir(t, dataDir, "add", "Buy groceries", "-p", "high", "--tag", "errands")
	if err != nil {
		t.Fatalf("add error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Created task") {
		t.Errorf("add output should contain 'Created task', got: %s", out)
	}

	// List should show it
	out, err = executeCmdInDir(t, dataDir, "list")
	if err != nil {
		t.Fatalf("list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Buy groceries") {
		t.Errorf("list output should contain task title, got: %s", out)
	}
}

func TestIntegrationAddAndShow(t *testing.T) {
	dataDir := t.TempDir()

	// Add task in quiet mode to get the ID
	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Test show command")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Show it
	out, err = executeCmdInDir(t, dataDir, "task", "show", id)
	if err != nil {
		t.Fatalf("show error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Test show command") {
		t.Errorf("show should contain title, got: %s", out)
	}
	if !strings.Contains(out, id) {
		t.Errorf("show should contain ID, got: %s", out)
	}
}

func TestIntegrationAddAndDone(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Complete this task")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Complete it
	out, err = executeCmdInDir(t, dataDir, "done", id)
	if err != nil {
		t.Fatalf("done error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Completed:") {
		t.Errorf("done output should contain 'Completed:', got: %s", out)
	}

	// Show should reflect done status
	out, err = executeCmdInDir(t, dataDir, "task", "show", id)
	if err != nil {
		t.Fatalf("show error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "done") {
		t.Errorf("show after done should contain 'done', got: %s", out)
	}
}

func TestIntegrationAddAndDelete(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Delete me")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Delete it
	out, err = executeCmdInDir(t, dataDir, "task", "delete", id)
	if err != nil {
		t.Fatalf("delete error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Deleted:") {
		t.Errorf("delete output should contain 'Deleted:', got: %s", out)
	}

	// List should be empty
	out, err = executeCmdInDir(t, dataDir, "list")
	if err != nil {
		t.Fatalf("list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "No tasks found") {
		t.Errorf("list after delete should show 'No tasks found', got: %s", out)
	}
}

func TestIntegrationEditWithFlags(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Original title")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Edit title and priority
	out, err = executeCmdInDir(t, dataDir, "task", "edit", id, "--title", "New title", "-p", "urgent")
	if err != nil {
		t.Fatalf("edit error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Updated:") {
		t.Errorf("edit output should contain 'Updated:', got: %s", out)
	}

	// Show should reflect changes
	out, err = executeCmdInDir(t, dataDir, "task", "show", id)
	if err != nil {
		t.Fatalf("show error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "New title") {
		t.Errorf("show should contain updated title, got: %s", out)
	}
	if !strings.Contains(out, "urgent") {
		t.Errorf("show should contain updated priority, got: %s", out)
	}
}

func TestIntegrationSearch(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Buy groceries at the store")
	_, _ = executeCmdInDir(t, dataDir, "add", "Write quarterly report")
	_, _ = executeCmdInDir(t, dataDir, "add", "Clean kitchen")

	// Search should find matching task
	out, err := executeCmdInDir(t, dataDir, "search", "groceries")
	if err != nil {
		t.Fatalf("search error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Buy groceries") {
		t.Errorf("search should find 'Buy groceries', got: %s", out)
	}
	if !strings.Contains(out, "1 results") {
		t.Errorf("search should show 1 result, got: %s", out)
	}
}

func TestIntegrationSearchNoResults(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Some task")

	out, err := executeCmdInDir(t, dataDir, "search", "nonexistent")
	if err != nil {
		t.Fatalf("search error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "No results for") {
		t.Errorf("search should show 'No results for', got: %s", out)
	}
}

func TestIntegrationListFilters(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Work task", "--tag", "work", "-p", "high")
	_, _ = executeCmdInDir(t, dataDir, "add", "Home task", "--tag", "home", "-p", "low")

	// Filter by tag
	out, err := executeCmdInDir(t, dataDir, "list", "--tag", "work")
	if err != nil {
		t.Fatalf("list --tag error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Work task") {
		t.Errorf("list --tag=work should contain 'Work task', got: %s", out)
	}
	if strings.Contains(out, "Home task") {
		t.Errorf("list --tag=work should NOT contain 'Home task', got: %s", out)
	}

	// Filter by priority
	out, err = executeCmdInDir(t, dataDir, "list", "-p", "low")
	if err != nil {
		t.Fatalf("list -p low error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Home task") {
		t.Errorf("list -p low should contain 'Home task', got: %s", out)
	}
	if strings.Contains(out, "Work task") {
		t.Errorf("list -p low should NOT contain 'Work task', got: %s", out)
	}
}

func TestIntegrationQuietMode(t *testing.T) {
	dataDir := t.TempDir()

	// Quiet mode should output only the ID
	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Quiet task")
	if err != nil {
		t.Fatalf("add -q error: %v", err)
	}
	id := strings.TrimSpace(out)
	if len(id) != 26 {
		t.Errorf("quiet add should output 26-char ULID, got %d chars: %q", len(id), id)
	}

	// Quiet list
	out, err = executeCmdInDir(t, dataDir, "list", "-q")
	if err != nil {
		t.Fatalf("list -q error: %v", err)
	}
	if strings.TrimSpace(out) != id {
		t.Errorf("quiet list should output just the ID, got: %q", out)
	}
}

func TestIntegrationJSONAdd(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "--json", "JSON task", "-p", "high", "--tag", "test")
	if err != nil {
		t.Fatalf("add --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Title != "JSON task" {
		t.Errorf("JSON title = %q, want %q", result.Title, "JSON task")
	}
	if result.Priority != "high" {
		t.Errorf("JSON priority = %q, want %q", result.Priority, "high")
	}
	if len(result.Tags) != 1 || result.Tags[0] != "test" {
		t.Errorf("JSON tags = %v, want [test]", result.Tags)
	}
	if result.Status != "pending" {
		t.Errorf("JSON status = %q, want %q", result.Status, "pending")
	}
}

func TestIntegrationJSONList(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Task A")
	_, _ = executeCmdInDir(t, dataDir, "add", "Task B")

	out, err := executeCmdInDir(t, dataDir, "list", "--json")
	if err != nil {
		t.Fatalf("list --json error: %v\noutput: %s", err, out)
	}

	var results []TaskJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if len(results) != 2 {
		t.Errorf("JSON list length = %d, want 2", len(results))
	}
}

func TestIntegrationJSONShow(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Show JSON task", "-e", "high")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	out, err = executeCmdInDir(t, dataDir, "task", "show", id, "--json")
	if err != nil {
		t.Fatalf("show --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Title != "Show JSON task" {
		t.Errorf("JSON title = %q, want %q", result.Title, "Show JSON task")
	}
	if result.Energy != "high" {
		t.Errorf("JSON energy = %q, want %q", result.Energy, "high")
	}
	if result.ID != id {
		t.Errorf("JSON id = %q, want %q", result.ID, id)
	}
}

func TestIntegrationJSONSearch(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Unique searchable item")

	out, err := executeCmdInDir(t, dataDir, "search", "--json", "searchable")
	if err != nil {
		t.Fatalf("search --json error: %v\noutput: %s", err, out)
	}

	var results []TaskJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if len(results) != 1 {
		t.Errorf("JSON search results = %d, want 1", len(results))
	}
	if len(results) > 0 && results[0].Title != "Unique searchable item" {
		t.Errorf("JSON search title = %q, want %q", results[0].Title, "Unique searchable item")
	}
}

func TestIntegrationJSONDone(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Complete with JSON")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	out, err = executeCmdInDir(t, dataDir, "done", "--json", id)
	if err != nil {
		t.Fatalf("done --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Status != "done" {
		t.Errorf("JSON status = %q, want %q", result.Status, "done")
	}
	if result.CompletedAt == "" {
		t.Error("JSON completed_at should be set")
	}
}

func TestIntegrationIndexRebuild(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Task for rebuild")

	out, err := executeCmdInDir(t, dataDir, "index", "rebuild")
	if err != nil {
		t.Fatalf("index rebuild error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Rebuilt index:") {
		t.Errorf("rebuild output should contain 'Rebuilt index:', got: %s", out)
	}
	if !strings.Contains(out, "1 tasks indexed") {
		t.Errorf("rebuild output should contain '1 tasks indexed', got: %s", out)
	}
}

func TestIntegrationIndexRebuildJSON(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Task A")
	_, _ = executeCmdInDir(t, dataDir, "add", "Task B")

	out, err := executeCmdInDir(t, dataDir, "index", "rebuild", "--json")
	if err != nil {
		t.Fatalf("index rebuild --json error: %v\noutput: %s", err, out)
	}

	var result map[string]int
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result["indexed"] != 2 {
		t.Errorf("JSON indexed = %d, want 2", result["indexed"])
	}
}

func TestIntegrationPrefixMatch(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Prefix test task")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)
	prefix := id[:8]

	// Show by prefix
	out, err = executeCmdInDir(t, dataDir, "task", "show", prefix)
	if err != nil {
		t.Fatalf("show by prefix error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Prefix test task") {
		t.Errorf("show by prefix should find task, got: %s", out)
	}
}

func TestIntegrationNotFound(t *testing.T) {
	dataDir := t.TempDir()

	_, err := executeCmdInDir(t, dataDir, "task", "show", "NONEXISTENT")
	if err == nil {
		t.Error("show nonexistent task should return error")
	}
}

func TestIntegrationDoneAlreadyComplete(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Already done test")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	_, _ = executeCmdInDir(t, dataDir, "done", id)

	// Second done should error
	_, err = executeCmdInDir(t, dataDir, "done", id)
	if err == nil {
		t.Error("completing already-done task should return error")
	}
}

func TestIntegrationJSONEmptyList(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "list", "--json")
	if err != nil {
		t.Fatalf("list --json error: %v\noutput: %s", err, out)
	}

	var results []TaskJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if len(results) != 0 {
		t.Errorf("empty list JSON should be [], got %d items", len(results))
	}
}

func TestIntegrationNounVerb(t *testing.T) {
	dataDir := t.TempDir()

	// bt task add (noun-verb form)
	out, err := executeCmdInDir(t, dataDir, "task", "add", "-q", "Noun verb test")
	if err != nil {
		t.Fatalf("task add error: %v", err)
	}
	id := strings.TrimSpace(out)
	if len(id) != 26 {
		t.Errorf("expected 26-char ULID, got %d: %q", len(id), id)
	}

	// bt task list (noun-verb form)
	out, err = executeCmdInDir(t, dataDir, "task", "list", "-q")
	if err != nil {
		t.Fatalf("task list error: %v", err)
	}
	if strings.TrimSpace(out) != id {
		t.Errorf("task list should return same ID, got: %q", out)
	}
}

func TestIntegrationTaskAlias(t *testing.T) {
	dataDir := t.TempDir()

	// bt t add (alias form)
	out, err := executeCmdInDir(t, dataDir, "t", "add", "-q", "Alias test")
	if err != nil {
		t.Fatalf("t add error: %v", err)
	}
	id := strings.TrimSpace(out)
	if len(id) != 26 {
		t.Errorf("expected 26-char ULID, got %d: %q", len(id), id)
	}
}

func TestIntegrationAddWithDueDate(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "--json", "Dated task", "--due", "2026-12-25")
	if err != nil {
		t.Fatalf("add with due error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.DueDate != "2026-12-25" {
		t.Errorf("JSON due_date = %q, want %q", result.DueDate, "2026-12-25")
	}
	if result.Type != "dated" {
		t.Errorf("JSON type = %q, want %q (auto-promoted from due date)", result.Type, "dated")
	}
}

func TestIntegrationJSONNullSafety(t *testing.T) {
	dataDir := t.TempDir()

	// Add a minimal task — tags and contexts should be [] not null
	out, err := executeCmdInDir(t, dataDir, "add", "--json", "Minimal task")
	if err != nil {
		t.Fatalf("add --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Tags == nil {
		t.Error("JSON tags should be [], not null")
	}
	if result.Contexts == nil {
		t.Error("JSON contexts should be [], not null")
	}

	// Also verify raw JSON doesn't have "null"
	if strings.Contains(out, `"tags": null`) {
		t.Error("JSON output should use [] for empty tags, not null")
	}
}

func TestIntegrationJSONListShowsTags(t *testing.T) {
	dataDir := t.TempDir()

	// Add tasks with tags and contexts
	_, _ = executeCmdInDir(t, dataDir, "add", "Tagged task", "--tag", "work", "--tag", "urgent", "-c", "office")
	_, _ = executeCmdInDir(t, dataDir, "add", "Plain task")

	// List in JSON should include tags from junction tables
	out, err := executeCmdInDir(t, dataDir, "list", "--json")
	if err != nil {
		t.Fatalf("list --json error: %v\noutput: %s", err, out)
	}

	var results []TaskJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}

	// Find the tagged task (listed in reverse chronological order)
	var tagged *TaskJSON
	for i := range results {
		if results[i].Title == "Tagged task" {
			tagged = &results[i]
			break
		}
	}
	if tagged == nil {
		t.Fatal("Tagged task not found in list results")
	}
	if len(tagged.Tags) != 2 {
		t.Errorf("list --json tags = %v, want 2 tags [urgent work]", tagged.Tags)
	}
	if len(tagged.Contexts) != 1 || tagged.Contexts[0] != "office" {
		t.Errorf("list --json contexts = %v, want [office]", tagged.Contexts)
	}
}

// --- Habit integration tests ---

func TestIntegrationHabitAddAndList(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "add", "Read 30 minutes", "--freq", "daily")
	if err != nil {
		t.Fatalf("habit add error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Created habit") {
		t.Errorf("habit add output should contain 'Created habit', got: %s", out)
	}

	// List habits
	out, err = executeCmdInDir(t, dataDir, "habit", "list")
	if err != nil {
		t.Fatalf("habit list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Read 30 minutes") {
		t.Errorf("habit list should contain habit title, got: %s", out)
	}
}

func TestIntegrationHabitAddJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "add", "--json", "Meditate", "--freq", "daily", "--tag", "wellness")
	if err != nil {
		t.Fatalf("habit add --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Title != "Meditate" {
		t.Errorf("JSON title = %q, want %q", result.Title, "Meditate")
	}
	if result.Type != "habit" {
		t.Errorf("JSON type = %q, want %q", result.Type, "habit")
	}
	if result.Status != "active" {
		t.Errorf("JSON status = %q, want %q", result.Status, "active")
	}
}

func TestIntegrationHabitLog(t *testing.T) {
	dataDir := t.TempDir()

	// Create a habit
	out, err := executeCmdInDir(t, dataDir, "habit", "add", "-q", "Exercise")
	if err != nil {
		t.Fatalf("habit add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Log a completion
	out, err = executeCmdInDir(t, dataDir, "habit", "log", id, "--duration", "30", "-n", "Morning run")
	if err != nil {
		t.Fatalf("habit log error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Logged:") {
		t.Errorf("habit log output should contain 'Logged:', got: %s", out)
	}
}

func TestIntegrationHabitLogJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "add", "-q", "Read")
	if err != nil {
		t.Fatalf("habit add error: %v", err)
	}
	id := strings.TrimSpace(out)

	out, err = executeCmdInDir(t, dataDir, "habit", "log", "--json", id, "--duration", "25")
	if err != nil {
		t.Fatalf("habit log --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Title != "Read" {
		t.Errorf("JSON title = %q, want %q", result.Title, "Read")
	}
}

func TestIntegrationHabitStats(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "add", "-q", "Meditate")
	if err != nil {
		t.Fatalf("habit add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Log a completion
	_, _ = executeCmdInDir(t, dataDir, "habit", "log", id)

	// Get stats
	out, err = executeCmdInDir(t, dataDir, "habit", "stats", id)
	if err != nil {
		t.Fatalf("habit stats error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Meditate") {
		t.Errorf("habit stats should contain title, got: %s", out)
	}
	if !strings.Contains(out, "Current streak") {
		t.Errorf("habit stats should contain 'Current streak', got: %s", out)
	}
	if !strings.Contains(out, "Total completions") {
		t.Errorf("habit stats should contain 'Total completions', got: %s", out)
	}
}

func TestIntegrationHabitStatsJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "add", "-q", "Exercise")
	if err != nil {
		t.Fatalf("habit add error: %v", err)
	}
	id := strings.TrimSpace(out)

	_, _ = executeCmdInDir(t, dataDir, "habit", "log", id)
	_, _ = executeCmdInDir(t, dataDir, "habit", "log", id)

	out, err = executeCmdInDir(t, dataDir, "habit", "stats", "--json", id)
	if err != nil {
		t.Fatalf("habit stats --json error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result["title"] != "Exercise" {
		t.Errorf("JSON title = %v, want Exercise", result["title"])
	}
	if result["total_completions"].(float64) != 2 {
		t.Errorf("JSON total_completions = %v, want 2", result["total_completions"])
	}
}

func TestIntegrationHabitLogNonHabit(t *testing.T) {
	dataDir := t.TempDir()

	// Create a regular task
	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Regular task")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Try to log — should fail
	_, err = executeCmdInDir(t, dataDir, "habit", "log", id)
	if err == nil {
		t.Error("habit log on non-habit task should return error")
	}
}

func TestIntegrationHabitListEmpty(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "list")
	if err != nil {
		t.Fatalf("habit list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "No habits found") {
		t.Errorf("empty habit list should show hint, got: %s", out)
	}
}

func TestIntegrationHabitWeekly(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "habit", "add", "--json", "Exercise", "--freq", "weekly", "--target", "3")
	if err != nil {
		t.Fatalf("habit add weekly error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result.Type != "habit" {
		t.Errorf("type = %q, want habit", result.Type)
	}
}

func TestIntegrationRebuildPreservesHabitCompletions(t *testing.T) {
	dataDir := t.TempDir()

	// Create a habit
	out, err := executeCmdInDir(t, dataDir, "habit", "add", "-q", "Rebuild habit")
	if err != nil {
		t.Fatalf("habit add error: %v", err)
	}
	id := strings.TrimSpace(out)

	// Log two completions
	_, err = executeCmdInDir(t, dataDir, "habit", "log", id, "--duration", "30", "-n", "Morning")
	if err != nil {
		t.Fatalf("habit log 1 error: %v", err)
	}
	_, err = executeCmdInDir(t, dataDir, "habit", "log", id, "--duration", "20")
	if err != nil {
		t.Fatalf("habit log 2 error: %v", err)
	}

	// Rebuild the index
	out, err = executeCmdInDir(t, dataDir, "index", "rebuild")
	if err != nil {
		t.Fatalf("index rebuild error: %v\noutput: %s", err, out)
	}

	// Stats should still show 2 completions (from body, which is SOT)
	out, err = executeCmdInDir(t, dataDir, "habit", "stats", "--json", id)
	if err != nil {
		t.Fatalf("habit stats error: %v\noutput: %s", err, out)
	}

	var stats map[string]any
	if err := json.Unmarshal([]byte(out), &stats); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if stats["total_completions"].(float64) != 2 {
		t.Errorf("total_completions after rebuild = %v, want 2", stats["total_completions"])
	}
}

// --- Routine integration tests ---

func TestIntegrationRoutineCreateAndList(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "Morning Routine",
		"--step", "Shower:5", "--step", "Breakfast:15", "--step", "Review inbox:10")
	if err != nil {
		t.Fatalf("routine create error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Created routine") {
		t.Errorf("routine create output should contain 'Created routine', got: %s", out)
	}
	if !strings.Contains(out, "3 steps") {
		t.Errorf("routine create output should contain '3 steps', got: %s", out)
	}

	// List routines
	out, err = executeCmdInDir(t, dataDir, "routine", "list")
	if err != nil {
		t.Fatalf("routine list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Morning Routine") {
		t.Errorf("routine list should contain title, got: %s", out)
	}
}

func TestIntegrationRoutineCreateJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "--json", "Evening",
		"--step", "Journal:10", "--step", "Read:30", "--tag", "wellness")
	if err != nil {
		t.Fatalf("routine create --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Title != "Evening" {
		t.Errorf("JSON title = %q, want %q", result.Title, "Evening")
	}
	if result.Type != "routine" {
		t.Errorf("JSON type = %q, want %q", result.Type, "routine")
	}
	if result.Status != "active" {
		t.Errorf("JSON status = %q, want %q", result.Status, "active")
	}
	if result.EstimatedDuration != 40 {
		t.Errorf("JSON estimated_duration = %d, want 40", result.EstimatedDuration)
	}
	// Verify steps are included in JSON output
	if len(result.Steps) != 2 {
		t.Fatalf("JSON steps count = %d, want 2", len(result.Steps))
	}
	if result.Steps[0].Title != "Journal" || result.Steps[0].Duration != 10 {
		t.Errorf("JSON steps[0] = %+v, want Journal:10", result.Steps[0])
	}
	if result.Steps[1].Title != "Read" || result.Steps[1].Duration != 30 {
		t.Errorf("JSON steps[1] = %+v, want Read:30", result.Steps[1])
	}
}

func TestIntegrationRoutineCreateQuiet(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "-q", "Quick",
		"--step", "Step one:5")
	if err != nil {
		t.Fatalf("routine create -q error: %v", err)
	}
	id := strings.TrimSpace(out)
	if len(id) != 26 {
		t.Errorf("quiet create should output 26-char ULID, got %d chars: %q", len(id), id)
	}
}

func TestIntegrationRoutineShow(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "-q", "Show me",
		"--step", "A:5", "--step", "B:10")
	if err != nil {
		t.Fatalf("routine create error: %v", err)
	}
	id := strings.TrimSpace(out)

	out, err = executeCmdInDir(t, dataDir, "routine", "show", id)
	if err != nil {
		t.Fatalf("routine show error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Show me") {
		t.Errorf("routine show should contain title, got: %s", out)
	}
	if !strings.Contains(out, "1. A") {
		t.Errorf("routine show should list step 1, got: %s", out)
	}
	if !strings.Contains(out, "2. B") {
		t.Errorf("routine show should list step 2, got: %s", out)
	}
}

func TestIntegrationRoutineShowJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "-q", "JSON Show",
		"--step", "Step:5")
	if err != nil {
		t.Fatalf("routine create error: %v", err)
	}
	id := strings.TrimSpace(out)

	out, err = executeCmdInDir(t, dataDir, "routine", "show", "--json", id)
	if err != nil {
		t.Fatalf("routine show --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Title != "JSON Show" {
		t.Errorf("JSON title = %q, want %q", result.Title, "JSON Show")
	}
	// Verify steps are included in show --json output
	if len(result.Steps) != 1 {
		t.Fatalf("JSON steps count = %d, want 1", len(result.Steps))
	}
	if result.Steps[0].Title != "Step" || result.Steps[0].Duration != 5 {
		t.Errorf("JSON steps[0] = %+v, want Step:5", result.Steps[0])
	}
}

func TestIntegrationRoutinePlayJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "-q", "Play me",
		"--step", "A:5", "--step", "B:10")
	if err != nil {
		t.Fatalf("routine create error: %v", err)
	}
	id := strings.TrimSpace(out)

	out, err = executeCmdInDir(t, dataDir, "routine", "play", "--json", id)
	if err != nil {
		t.Fatalf("routine play --json error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result["title"] != "Play me" {
		t.Errorf("JSON title = %v, want 'Play me'", result["title"])
	}
	steps, ok := result["steps"].([]any)
	if !ok || len(steps) != 2 {
		t.Errorf("JSON steps should have 2 items, got: %v", result["steps"])
	}
}

func TestIntegrationRoutineListEmpty(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "list")
	if err != nil {
		t.Fatalf("routine list error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "No routines found") {
		t.Errorf("empty routine list should show hint, got: %s", out)
	}
}

func TestIntegrationRoutineCreateNoSteps(t *testing.T) {
	dataDir := t.TempDir()

	_, err := executeCmdInDir(t, dataDir, "routine", "create", "Empty routine")
	if err == nil {
		t.Error("routine create with no steps should return error")
	}
}

func TestIntegrationRoutineShowNonRoutine(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "add", "-q", "Regular task")
	if err != nil {
		t.Fatalf("add error: %v", err)
	}
	id := strings.TrimSpace(out)

	_, err = executeCmdInDir(t, dataDir, "routine", "show", id)
	if err == nil {
		t.Error("routine show on non-routine should return error")
	}
}

func TestIntegrationRoutineCreateWithSchedule(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "routine", "create", "--json", "Scheduled",
		"--step", "Meditate:10", "--schedule-time", "07:00", "--schedule-days", "mon,wed,fri")
	if err != nil {
		t.Fatalf("routine create with schedule error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result.Type != "routine" {
		t.Errorf("type = %q, want routine", result.Type)
	}
	// Verify schedule is included in JSON output
	if result.Schedule == nil {
		t.Fatal("JSON schedule should not be nil")
	}
	if result.Schedule.Time != "07:00" {
		t.Errorf("JSON schedule.time = %q, want 07:00", result.Schedule.Time)
	}
	if len(result.Schedule.Days) != 3 {
		t.Errorf("JSON schedule.days = %v, want [mon wed fri]", result.Schedule.Days)
	}
}

func TestIntegrationRoutineAlias(t *testing.T) {
	dataDir := t.TempDir()

	// bt r create (alias form)
	out, err := executeCmdInDir(t, dataDir, "r", "create", "-q", "Alias test",
		"--step", "Step:5")
	if err != nil {
		t.Fatalf("r create error: %v", err)
	}
	id := strings.TrimSpace(out)
	if len(id) != 26 {
		t.Errorf("expected 26-char ULID, got %d: %q", len(id), id)
	}
}

// --- Link integration tests ---

func TestIntegrationLinkAndShow(t *testing.T) {
	dataDir := t.TempDir()

	// Create two tasks
	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	// Link them
	out, err := executeCmdInDir(t, dataDir, "link", idA, idB, "-t", "depends-on")
	if err != nil {
		t.Fatalf("link error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Linked:") {
		t.Errorf("link output should contain 'Linked:', got: %s", out)
	}

	// Show should display the link
	out, err = executeCmdInDir(t, dataDir, "task", "show", idA)
	if err != nil {
		t.Fatalf("show error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Links:") {
		t.Errorf("show should contain 'Links:' section, got: %s", out)
	}
	if !strings.Contains(out, "depends-on") {
		t.Errorf("show should contain link type, got: %s", out)
	}
}

func TestIntegrationLinkJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Source")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Target")
	idB := strings.TrimSpace(out)

	out, err := executeCmdInDir(t, dataDir, "link", "--json", idA, idB, "-t", "blocks")
	if err != nil {
		t.Fatalf("link --json error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result["link_type"] != "blocks" {
		t.Errorf("JSON link_type = %v, want 'blocks'", result["link_type"])
	}
	if result["source_id"] != idA {
		t.Errorf("JSON source_id = %v, want %q", result["source_id"], idA)
	}
	if result["target_id"] != idB {
		t.Errorf("JSON target_id = %v, want %q", result["target_id"], idB)
	}
}

func TestIntegrationLinkDefaultType(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	// Default type should be related-to
	out, err := executeCmdInDir(t, dataDir, "link", "--json", idA, idB)
	if err != nil {
		t.Fatalf("link error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if result["link_type"] != "related-to" {
		t.Errorf("default link_type = %v, want 'related-to'", result["link_type"])
	}
}

func TestIntegrationUnlink(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	// Link then unlink
	_, _ = executeCmdInDir(t, dataDir, "link", idA, idB)

	out, err := executeCmdInDir(t, dataDir, "unlink", idA, idB)
	if err != nil {
		t.Fatalf("unlink error: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "Unlinked:") {
		t.Errorf("unlink output should contain 'Unlinked:', got: %s", out)
	}

	// Show should no longer have links
	out, err = executeCmdInDir(t, dataDir, "task", "show", idA)
	if err != nil {
		t.Fatalf("show error: %v\noutput: %s", err, out)
	}
	if strings.Contains(out, "Links:") {
		t.Errorf("show after unlink should NOT contain 'Links:', got: %s", out)
	}
}

func TestIntegrationUnlinkJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	_, _ = executeCmdInDir(t, dataDir, "link", idA, idB, "-t", "depends-on")

	out, err := executeCmdInDir(t, dataDir, "unlink", "--json", idA, idB, "-t", "depends-on")
	if err != nil {
		t.Fatalf("unlink --json error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if result["removed"] != true {
		t.Errorf("JSON removed = %v, want true", result["removed"])
	}
}

func TestIntegrationLinkSelfError(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task")
	id := strings.TrimSpace(out)

	_, err := executeCmdInDir(t, dataDir, "link", id, id)
	if err == nil {
		t.Error("self-link should return error")
	}
}

func TestIntegrationLinkCycleError(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task C")
	idC := strings.TrimSpace(out)

	// A depends-on B, B depends-on C
	_, _ = executeCmdInDir(t, dataDir, "link", idA, idB, "-t", "depends-on")
	_, _ = executeCmdInDir(t, dataDir, "link", idB, idC, "-t", "depends-on")

	// C depends-on A would create cycle
	_, err := executeCmdInDir(t, dataDir, "link", idC, idA, "-t", "depends-on")
	if err == nil {
		t.Error("cycle should be detected and rejected")
	}
}

func TestIntegrationLinkDuplicateError(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	_, _ = executeCmdInDir(t, dataDir, "link", idA, idB)
	_, err := executeCmdInDir(t, dataDir, "link", idA, idB)
	if err == nil {
		t.Error("duplicate link should return error")
	}
}

func TestIntegrationUnlinkNotFoundError(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	_, err := executeCmdInDir(t, dataDir, "unlink", idA, idB)
	if err == nil {
		t.Error("unlinking non-existent link should return error")
	}
}

func TestIntegrationShowLinksJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	_, _ = executeCmdInDir(t, dataDir, "link", idA, idB, "-t", "depends-on")

	out, err := executeCmdInDir(t, dataDir, "task", "show", "--json", idA)
	if err != nil {
		t.Fatalf("show --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if len(result.Links) != 1 {
		t.Fatalf("JSON links = %d, want 1", len(result.Links))
	}
	if result.Links[0]["type"] != "depends-on" {
		t.Errorf("JSON link type = %q, want 'depends-on'", result.Links[0]["type"])
	}
	if result.Links[0]["direction"] != "outgoing" {
		t.Errorf("JSON link direction = %q, want 'outgoing'", result.Links[0]["direction"])
	}
	if result.Links[0]["task_title"] != "Task B" {
		t.Errorf("JSON link task_title = %q, want 'Task B'", result.Links[0]["task_title"])
	}
}

func TestIntegrationShowBacklinksJSON(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	// A depends-on B. Showing B should show incoming link from A.
	_, _ = executeCmdInDir(t, dataDir, "link", idA, idB, "-t", "depends-on")

	out, err := executeCmdInDir(t, dataDir, "task", "show", "--json", idB)
	if err != nil {
		t.Fatalf("show --json error: %v\noutput: %s", err, out)
	}

	var result TaskJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, out)
	}
	if len(result.Links) != 1 {
		t.Fatalf("JSON links = %d, want 1 (incoming)", len(result.Links))
	}
	if result.Links[0]["direction"] != "incoming" {
		t.Errorf("JSON link direction = %q, want 'incoming'", result.Links[0]["direction"])
	}
	if result.Links[0]["task_title"] != "Task A" {
		t.Errorf("JSON link task_title = %q, want 'Task A'", result.Links[0]["task_title"])
	}
}

func TestIntegrationLinkQuiet(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Task A")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task B")
	idB := strings.TrimSpace(out)

	out, err := executeCmdInDir(t, dataDir, "link", "-q", idA, idB)
	if err != nil {
		t.Fatalf("link -q error: %v", err)
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		t.Errorf("quiet link should output 2 IDs, got: %q", out)
	}
}

// --- bt now integration tests ---

func TestIntegrationNowBasic(t *testing.T) {
	dataDir := t.TempDir()

	// Create some tasks
	_, _ = executeCmdInDir(t, dataDir, "add", "Buy groceries", "-p", "low", "-e", "low", "--duration", "30")
	_, _ = executeCmdInDir(t, dataDir, "add", "Write report", "-p", "high", "-e", "medium", "--duration", "60", "--due", "2026-04-09")
	_, _ = executeCmdInDir(t, dataDir, "add", "Review code", "-p", "medium", "-e", "medium", "--duration", "20")

	out, err := executeCmdInDir(t, dataDir, "now")
	if err != nil {
		t.Fatalf("now error: %v\noutput: %s", err, out)
	}

	// Should show a header and suggestions
	if !strings.Contains(out, "What to do now") {
		t.Errorf("now output should contain header, got: %s", out)
	}
	// Should show at least one task
	if !strings.Contains(out, "Write report") && !strings.Contains(out, "Review code") && !strings.Contains(out, "Buy groceries") {
		t.Errorf("now output should suggest tasks, got: %s", out)
	}
}

func TestIntegrationNowEmpty(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "now")
	if err != nil {
		t.Fatalf("now error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, "No tasks match") {
		t.Errorf("now with no tasks should show 'No tasks match', got: %s", out)
	}
}

func TestIntegrationNowWithFlags(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Low task", "-e", "low", "--duration", "15")
	_, _ = executeCmdInDir(t, dataDir, "add", "High task", "-e", "high", "--duration", "15")

	// With low energy — should only show low task
	out, err := executeCmdInDir(t, dataDir, "now", "--energy", "low")
	if err != nil {
		t.Fatalf("now error: %v\noutput: %s", err, out)
	}
	if strings.Contains(out, "High task") {
		t.Errorf("now with low energy should not show high energy task, got: %s", out)
	}
	if !strings.Contains(out, "Low task") {
		t.Errorf("now with low energy should show low energy task, got: %s", out)
	}
}

func TestIntegrationNowContextFilter(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Home task", "-c", "home", "--duration", "15")
	_, _ = executeCmdInDir(t, dataDir, "add", "Office task", "-c", "office", "--duration", "15")

	out, err := executeCmdInDir(t, dataDir, "now", "--context", "home")
	if err != nil {
		t.Fatalf("now error: %v\noutput: %s", err, out)
	}
	if strings.Contains(out, "Office task") {
		t.Errorf("now with home context should not show office task, got: %s", out)
	}
	if !strings.Contains(out, "Home task") {
		t.Errorf("now with home context should show home task, got: %s", out)
	}
}

func TestIntegrationNowJSON(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "JSON test task", "-p", "high", "-e", "medium", "--duration", "30")

	out, err := executeCmdInDir(t, dataDir, "now", "--json", "-n", "1")
	if err != nil {
		t.Fatalf("now --json error: %v\noutput: %s", err, out)
	}

	var results []SuggestionJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("now --json parse error: %v\noutput: %s", err, out)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 JSON result, got %d", len(results))
	}
	if results[0].Title != "JSON test task" {
		t.Errorf("expected title 'JSON test task', got %q", results[0].Title)
	}
	if results[0].Score.Priority != 0.75 {
		t.Errorf("expected priority score 0.75, got %v", results[0].Score.Priority)
	}
	if results[0].Duration != 30 {
		t.Errorf("expected duration 30, got %d", results[0].Duration)
	}
}

func TestIntegrationNowCountLimit(t *testing.T) {
	dataDir := t.TempDir()

	for i := 0; i < 10; i++ {
		_, _ = executeCmdInDir(t, dataDir, "add", "-q", "Task number "+string(rune('A'+i)), "--duration", "5")
	}

	out, err := executeCmdInDir(t, dataDir, "now", "--json", "-n", "3")
	if err != nil {
		t.Fatalf("now --json error: %v\noutput: %s", err, out)
	}

	var results []SuggestionJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results with -n 3, got %d", len(results))
	}
}

func TestIntegrationNowExcludesDoneTasks(t *testing.T) {
	dataDir := t.TempDir()

	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Done task", "--duration", "10")
	id := strings.TrimSpace(out)
	_, _ = executeCmdInDir(t, dataDir, "done", id)

	_, _ = executeCmdInDir(t, dataDir, "add", "Pending task", "--duration", "10")

	out, err := executeCmdInDir(t, dataDir, "now", "--json")
	if err != nil {
		t.Fatalf("now error: %v", err)
	}

	if strings.Contains(out, "Done task") {
		t.Errorf("now should not suggest done tasks, got: %s", out)
	}
}

// --- bt plan today integration tests ---

func TestIntegrationPlanTodayBasic(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Task A", "-p", "high", "--duration", "30")
	_, _ = executeCmdInDir(t, dataDir, "add", "Task B", "-p", "medium", "--duration", "20")
	_, _ = executeCmdInDir(t, dataDir, "add", "Task C", "-p", "low", "--duration", "15")

	out, err := executeCmdInDir(t, dataDir, "plan", "today", "--time", "60")
	if err != nil {
		t.Fatalf("plan today error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, "Today's Plan") {
		t.Errorf("plan today should contain header, got: %s", out)
	}
	if !strings.Contains(out, "Total:") {
		t.Errorf("plan today should contain total line, got: %s", out)
	}
}

func TestIntegrationPlanTodayEmpty(t *testing.T) {
	dataDir := t.TempDir()

	out, err := executeCmdInDir(t, dataDir, "plan", "today")
	if err != nil {
		t.Fatalf("plan today error: %v\noutput: %s", err, out)
	}

	if !strings.Contains(out, "No tasks to plan") {
		t.Errorf("plan today with no tasks should show empty message, got: %s", out)
	}
}

func TestIntegrationPlanTodayJSON(t *testing.T) {
	dataDir := t.TempDir()

	_, _ = executeCmdInDir(t, dataDir, "add", "Plan task A", "-p", "high", "--duration", "30")
	_, _ = executeCmdInDir(t, dataDir, "add", "Plan task B", "-p", "medium", "--duration", "20")

	out, err := executeCmdInDir(t, dataDir, "plan", "today", "--json", "--time", "60")
	if err != nil {
		t.Fatalf("plan today --json error: %v\noutput: %s", err, out)
	}

	var result PlanJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("plan today JSON parse error: %v\noutput: %s", err, out)
	}

	if result.AvailableTime != 60 {
		t.Errorf("expected available_time 60, got %d", result.AvailableTime)
	}
	if result.TotalDuration+result.TimeRemaining != 60 {
		t.Errorf("total + remaining should equal available: %d + %d != 60",
			result.TotalDuration, result.TimeRemaining)
	}
	if len(result.Suggestions) == 0 {
		t.Error("expected at least 1 suggestion in plan")
	}
}

func TestIntegrationPlanTodayRespectsTimeLimit(t *testing.T) {
	dataDir := t.TempDir()

	// Add 3 tasks totaling 60 min
	_, _ = executeCmdInDir(t, dataDir, "add", "T1", "--duration", "20")
	_, _ = executeCmdInDir(t, dataDir, "add", "T2", "--duration", "20")
	_, _ = executeCmdInDir(t, dataDir, "add", "T3", "--duration", "20")

	out, err := executeCmdInDir(t, dataDir, "plan", "today", "--json", "--time", "30")
	if err != nil {
		t.Fatalf("plan today error: %v", err)
	}

	var result PlanJSON
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if result.TotalDuration > 30 {
		t.Errorf("total duration %d exceeds available time 30", result.TotalDuration)
	}
}

func TestIntegrationNowWithDependencies(t *testing.T) {
	dataDir := t.TempDir()

	// Create two tasks where B depends on A
	out, _ := executeCmdInDir(t, dataDir, "add", "-q", "Blocker task", "--duration", "15")
	idA := strings.TrimSpace(out)
	out, _ = executeCmdInDir(t, dataDir, "add", "-q", "Blocked task", "--duration", "15")
	idB := strings.TrimSpace(out)

	// B depends on A
	_, _ = executeCmdInDir(t, dataDir, "link", idB, idA, "--type", "depends-on")

	// bt now should not suggest B (unmet dependency)
	out, err := executeCmdInDir(t, dataDir, "now", "--json")
	if err != nil {
		t.Fatalf("now error: %v", err)
	}

	var results []SuggestionJSON
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("parse error: %v\noutput: %s", err, out)
	}

	for _, r := range results {
		if r.TaskID == idB {
			t.Error("now should not suggest a task with unmet dependencies")
		}
	}

	// A should be suggested (it's the blocker)
	foundA := false
	for _, r := range results {
		if r.TaskID == idA {
			foundA = true
			break
		}
	}
	if !foundA {
		t.Error("now should suggest the blocker task")
	}
}
