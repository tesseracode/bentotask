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
