package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tesserabox/bentotask/internal/app"
)

// setupTestServer creates a test server with a temp data directory.
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dataDir := t.TempDir()
	a, err := app.Open(dataDir)
	if err != nil {
		t.Fatalf("open app: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	srv := NewServer(a)
	return httptest.NewServer(srv)
}

// doRequest is a test helper for making API requests.
func doRequest(t *testing.T, ts *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, ts.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// decodeJSON decodes a response body into the given value.
func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

// assertStatus checks that the response has the expected status code.
func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}

// --- Task CRUD lifecycle ---

func TestAPITaskCRUDLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title":    "Buy groceries",
		"priority": "high",
		"energy":   "low",
		"duration": 30,
		"tags":     []string{"errands"},
		"contexts": []string{"home"},
	})
	assertStatus(t, resp, http.StatusCreated)
	var created TaskJSON
	decodeJSON(t, resp, &created)
	if created.Title != "Buy groceries" {
		t.Errorf("title = %q, want %q", created.Title, "Buy groceries")
	}
	if created.Priority != "high" {
		t.Errorf("priority = %q, want %q", created.Priority, "high")
	}
	if created.ID == "" {
		t.Fatal("created task has no ID")
	}

	taskID := created.ID

	// Get
	resp = doRequest(t, ts, "GET", "/api/v1/tasks/"+taskID, nil)
	assertStatus(t, resp, http.StatusOK)
	var fetched TaskJSON
	decodeJSON(t, resp, &fetched)
	if fetched.ID != taskID {
		t.Errorf("GET returned wrong ID: %s", fetched.ID)
	}
	if fetched.Title != "Buy groceries" {
		t.Errorf("GET title = %q", fetched.Title)
	}

	// List
	resp = doRequest(t, ts, "GET", "/api/v1/tasks", nil)
	assertStatus(t, resp, http.StatusOK)
	var listResp struct {
		Items []TaskJSON `json:"items"`
		Count int        `json:"count"`
	}
	decodeJSON(t, resp, &listResp)
	if listResp.Count != 1 {
		t.Errorf("list count = %d, want 1", listResp.Count)
	}
	if len(listResp.Items) != 1 {
		t.Fatalf("list items len = %d, want 1", len(listResp.Items))
	}
	if listResp.Items[0].ID != taskID {
		t.Errorf("list item ID = %q", listResp.Items[0].ID)
	}

	// Update
	resp = doRequest(t, ts, "PATCH", "/api/v1/tasks/"+taskID, map[string]any{
		"title":    "Buy groceries and snacks",
		"priority": "urgent",
	})
	assertStatus(t, resp, http.StatusOK)
	var updated TaskJSON
	decodeJSON(t, resp, &updated)
	if updated.Title != "Buy groceries and snacks" {
		t.Errorf("updated title = %q", updated.Title)
	}
	if updated.Priority != "urgent" {
		t.Errorf("updated priority = %q", updated.Priority)
	}

	// Complete
	resp = doRequest(t, ts, "POST", "/api/v1/tasks/"+taskID+"/done", nil)
	assertStatus(t, resp, http.StatusOK)
	var completed TaskJSON
	decodeJSON(t, resp, &completed)
	if completed.Status != "done" {
		t.Errorf("completed status = %q, want done", completed.Status)
	}

	// Delete
	resp = doRequest(t, ts, "DELETE", "/api/v1/tasks/"+taskID, nil)
	assertStatus(t, resp, http.StatusOK)

	// Verify deleted
	resp = doRequest(t, ts, "GET", "/api/v1/tasks/"+taskID, nil)
	assertStatus(t, resp, http.StatusNotFound)
	_ = resp.Body.Close()
}

// --- Habit lifecycle ---

func TestAPIHabitLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create habit
	resp := doRequest(t, ts, "POST", "/api/v1/habits", map[string]any{
		"title":       "Meditate",
		"freq_type":   "daily",
		"freq_target": 1,
		"priority":    "medium",
		"energy":      "low",
	})
	assertStatus(t, resp, http.StatusCreated)
	var habit TaskJSON
	decodeJSON(t, resp, &habit)
	if habit.Title != "Meditate" {
		t.Errorf("title = %q", habit.Title)
	}
	if habit.Type != "habit" {
		t.Errorf("type = %q, want habit", habit.Type)
	}

	habitID := habit.ID

	// Log completion
	resp = doRequest(t, ts, "POST", "/api/v1/habits/"+habitID+"/log", map[string]any{
		"duration": 15,
		"note":     "Morning session",
	})
	assertStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()

	// Get stats
	resp = doRequest(t, ts, "GET", "/api/v1/habits/"+habitID+"/stats", nil)
	assertStatus(t, resp, http.StatusOK)
	var statsResp map[string]any
	decodeJSON(t, resp, &statsResp)
	if _, ok := statsResp["task"]; !ok {
		t.Error("stats response missing 'task' field")
	}
	if _, ok := statsResp["stats"]; !ok {
		t.Error("stats response missing 'stats' field")
	}

	// List habits
	resp = doRequest(t, ts, "GET", "/api/v1/habits", nil)
	assertStatus(t, resp, http.StatusOK)
	var habitList struct {
		Items []TaskJSON `json:"items"`
		Count int        `json:"count"`
	}
	decodeJSON(t, resp, &habitList)
	if habitList.Count != 1 {
		t.Errorf("habit list count = %d, want 1", habitList.Count)
	}
}

// --- Routine lifecycle ---

func TestAPIRoutineLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create routine
	resp := doRequest(t, ts, "POST", "/api/v1/routines", map[string]any{
		"title": "Morning Routine",
		"steps": []map[string]any{
			{"title": "Wake up", "duration": 5},
			{"title": "Stretch", "duration": 10},
			{"title": "Shower", "duration": 15},
		},
		"priority": "high",
	})
	assertStatus(t, resp, http.StatusCreated)
	var routine TaskJSON
	decodeJSON(t, resp, &routine)
	if routine.Title != "Morning Routine" {
		t.Errorf("title = %q", routine.Title)
	}
	if routine.Type != "routine" {
		t.Errorf("type = %q, want routine", routine.Type)
	}
	if len(routine.Steps) != 3 {
		t.Errorf("steps len = %d, want 3", len(routine.Steps))
	}

	routineID := routine.ID

	// List routines
	resp = doRequest(t, ts, "GET", "/api/v1/routines", nil)
	assertStatus(t, resp, http.StatusOK)
	var routineList struct {
		Items []TaskJSON `json:"items"`
		Count int        `json:"count"`
	}
	decodeJSON(t, resp, &routineList)
	if routineList.Count != 1 {
		t.Errorf("routine list count = %d, want 1", routineList.Count)
	}

	// Get routine
	resp = doRequest(t, ts, "GET", "/api/v1/routines/"+routineID, nil)
	assertStatus(t, resp, http.StatusOK)
	var fetched TaskJSON
	decodeJSON(t, resp, &fetched)
	if fetched.ID != routineID {
		t.Errorf("GET routine ID = %q", fetched.ID)
	}
}

// --- Link lifecycle ---

func TestAPILinkLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create two tasks
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "Task A"})
	assertStatus(t, resp, http.StatusCreated)
	var taskA TaskJSON
	decodeJSON(t, resp, &taskA)

	resp = doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "Task B"})
	assertStatus(t, resp, http.StatusCreated)
	var taskB TaskJSON
	decodeJSON(t, resp, &taskB)

	// Create link
	resp = doRequest(t, ts, "POST", "/api/v1/tasks/"+taskA.ID+"/links", map[string]any{
		"target_id": taskB.ID,
		"type":      "depends-on",
	})
	assertStatus(t, resp, http.StatusCreated)
	_ = resp.Body.Close()

	// Get links
	resp = doRequest(t, ts, "GET", "/api/v1/tasks/"+taskA.ID+"/links", nil)
	assertStatus(t, resp, http.StatusOK)
	var linkList struct {
		Items []map[string]string `json:"items"`
		Count int                 `json:"count"`
	}
	decodeJSON(t, resp, &linkList)
	if linkList.Count != 1 {
		t.Errorf("link count = %d, want 1", linkList.Count)
	}
	if linkList.Items[0]["type"] != "depends-on" {
		t.Errorf("link type = %q", linkList.Items[0]["type"])
	}

	// Delete link
	resp = doRequest(t, ts, "DELETE", "/api/v1/tasks/"+taskA.ID+"/links/"+taskB.ID+"?type=depends-on", nil)
	assertStatus(t, resp, http.StatusOK)
	_ = resp.Body.Close()

	// Verify deleted
	resp = doRequest(t, ts, "GET", "/api/v1/tasks/"+taskA.ID+"/links", nil)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &linkList)
	if linkList.Count != 0 {
		t.Errorf("link count after delete = %d, want 0", linkList.Count)
	}
}

// --- Scheduling ---

func TestAPISuggest(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create tasks
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title": "Write report", "priority": "high", "energy": "medium", "duration": 30,
	})
	_ = resp.Body.Close()
	resp = doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title": "Read emails", "priority": "low", "energy": "low", "duration": 15,
	})
	_ = resp.Body.Close()

	// Get suggestions
	resp = doRequest(t, ts, "GET", "/api/v1/suggest?time=60&energy=medium&count=5", nil)
	assertStatus(t, resp, http.StatusOK)
	var suggestResp struct {
		Items []SuggestionJSON `json:"items"`
		Count int              `json:"count"`
	}
	decodeJSON(t, resp, &suggestResp)
	if suggestResp.Count == 0 {
		t.Error("expected at least 1 suggestion")
	}
}

func TestAPIPlanToday(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create tasks
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title": "Deep work", "priority": "high", "energy": "medium", "duration": 60,
	})
	_ = resp.Body.Close()

	// Plan
	resp = doRequest(t, ts, "GET", "/api/v1/plan/today?time=120&energy=medium", nil)
	assertStatus(t, resp, http.StatusOK)
	var planResp PlanJSON
	decodeJSON(t, resp, &planResp)
	if planResp.AvailableTime != 120 {
		t.Errorf("available_time = %d, want 120", planResp.AvailableTime)
	}
	if planResp.TotalDuration+planResp.TimeRemaining != 120 {
		t.Errorf("total + remaining = %d, want 120", planResp.TotalDuration+planResp.TimeRemaining)
	}
}

// --- Error cases ---

func TestAPINotFound(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp := doRequest(t, ts, "GET", "/api/v1/tasks/nonexistent-id", nil)
	assertStatus(t, resp, http.StatusNotFound)

	var errResp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "not_found" {
		t.Errorf("error code = %q, want not_found", errResp.Error.Code)
	}
}

func TestAPIConflictDuplicateLink(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create two tasks
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "A"})
	var taskA TaskJSON
	decodeJSON(t, resp, &taskA)

	resp = doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "B"})
	var taskB TaskJSON
	decodeJSON(t, resp, &taskB)

	// Create link
	resp = doRequest(t, ts, "POST", "/api/v1/tasks/"+taskA.ID+"/links", map[string]any{
		"target_id": taskB.ID,
		"type":      "related-to",
	})
	assertStatus(t, resp, http.StatusCreated)
	_ = resp.Body.Close()

	// Duplicate link → conflict
	resp = doRequest(t, ts, "POST", "/api/v1/tasks/"+taskA.ID+"/links", map[string]any{
		"target_id": taskB.ID,
		"type":      "related-to",
	})
	assertStatus(t, resp, http.StatusConflict)
	_ = resp.Body.Close()
}

func TestAPIValidationEmptyTitle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title": "",
	})
	assertStatus(t, resp, http.StatusUnprocessableEntity)

	var errResp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeJSON(t, resp, &errResp)
	if errResp.Error.Code != "validation_error" {
		t.Errorf("error code = %q, want validation_error", errResp.Error.Code)
	}
}

func TestAPIValidationInvalidLinkType(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "A"})
	var task TaskJSON
	decodeJSON(t, resp, &task)

	resp = doRequest(t, ts, "POST", "/api/v1/tasks/"+task.ID+"/links", map[string]any{
		"target_id": "some-id",
		"type":      "invalid-type",
	})
	assertStatus(t, resp, http.StatusUnprocessableEntity)
	_ = resp.Body.Close()
}

// --- Collection envelope ---

func TestAPICollectionFormat(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Empty collection
	resp := doRequest(t, ts, "GET", "/api/v1/tasks", nil)
	assertStatus(t, resp, http.StatusOK)

	var body map[string]json.RawMessage
	decodeJSON(t, resp, &body)

	// Must have "items" and "count"
	if _, ok := body["items"]; !ok {
		t.Error("collection missing 'items' field")
	}
	if _, ok := body["count"]; !ok {
		t.Error("collection missing 'count' field")
	}

	// "items" must be an array (even if empty), not null
	var items []json.RawMessage
	if err := json.Unmarshal(body["items"], &items); err != nil {
		t.Errorf("items is not an array: %v", err)
	}
}

// --- Search ---

func TestAPISearch(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create tasks
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "Buy groceries for dinner"})
	_ = resp.Body.Close()
	resp = doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{"title": "Write code review"})
	_ = resp.Body.Close()

	// Search
	resp = doRequest(t, ts, "GET", "/api/v1/tasks/search?q=groceries", nil)
	assertStatus(t, resp, http.StatusOK)
	var searchResp struct {
		Items []TaskJSON `json:"items"`
		Count int        `json:"count"`
	}
	decodeJSON(t, resp, &searchResp)
	if searchResp.Count != 1 {
		t.Errorf("search count = %d, want 1", searchResp.Count)
	}
	if searchResp.Items[0].Title != "Buy groceries for dinner" {
		t.Errorf("search result = %q", searchResp.Items[0].Title)
	}
}

func TestAPISearchMissingQuery(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp := doRequest(t, ts, "GET", "/api/v1/tasks/search", nil)
	assertStatus(t, resp, http.StatusBadRequest)
	_ = resp.Body.Close()
}

// --- Filter by query params ---

func TestAPIListTasksWithFilter(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create tasks with different priorities
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title": "High task", "priority": "high",
	})
	_ = resp.Body.Close()
	resp = doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title": "Low task", "priority": "low",
	})
	_ = resp.Body.Close()

	// Filter by priority=high
	resp = doRequest(t, ts, "GET", "/api/v1/tasks?priority=high", nil)
	assertStatus(t, resp, http.StatusOK)
	var listResp struct {
		Items []TaskJSON `json:"items"`
		Count int        `json:"count"`
	}
	decodeJSON(t, resp, &listResp)
	if listResp.Count != 1 {
		t.Errorf("filtered count = %d, want 1", listResp.Count)
	}
	if listResp.Items[0].Title != "High task" {
		t.Errorf("filtered task = %q", listResp.Items[0].Title)
	}
}

// --- Meta endpoints ---

func TestAPIMetaEndpoints(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create task with tags and context
	resp := doRequest(t, ts, "POST", "/api/v1/tasks", map[string]any{
		"title":    "Tagged task",
		"tags":     []string{"work", "urgent"},
		"contexts": []string{"office"},
	})
	_ = resp.Body.Close()

	// Tags
	resp = doRequest(t, ts, "GET", "/api/v1/meta/tags", nil)
	assertStatus(t, resp, http.StatusOK)
	var tagResp struct {
		Items []string `json:"items"`
		Count int      `json:"count"`
	}
	decodeJSON(t, resp, &tagResp)
	if tagResp.Count != 2 {
		t.Errorf("tag count = %d, want 2", tagResp.Count)
	}

	// Contexts
	resp = doRequest(t, ts, "GET", "/api/v1/meta/contexts", nil)
	assertStatus(t, resp, http.StatusOK)
	var ctxResp struct {
		Items []string `json:"items"`
		Count int      `json:"count"`
	}
	decodeJSON(t, resp, &ctxResp)
	if ctxResp.Count != 1 {
		t.Errorf("context count = %d, want 1", ctxResp.Count)
	}
}

// --- Index rebuild ---

func TestAPIRebuildIndex(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp := doRequest(t, ts, "POST", "/api/v1/index/rebuild", nil)
	assertStatus(t, resp, http.StatusOK)

	var result map[string]int
	decodeJSON(t, resp, &result)
	if _, ok := result["indexed"]; !ok {
		t.Error("rebuild response missing 'indexed' field")
	}
}
