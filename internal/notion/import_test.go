package notion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tesserabox/bentotask/internal/app"
)

func TestPropertyMapping(t *testing.T) {
	page := Page{
		ID: "test-page",
		Properties: map[string]Property{
			"Name": {
				Type:  "title",
				Title: []RichText{{PlainText: "Buy groceries"}},
			},
			"Priority": {
				Type:   "select",
				Select: &SelectValue{Name: "High"},
			},
			"Due Date": {
				Type: "date",
				Date: &DateValue{Start: "2026-05-10"},
			},
			"Tags": {
				Type:        "multi_select",
				MultiSelect: []SelectValue{{Name: "errands"}, {Name: "home"}},
			},
			"Energy": {
				Type:   "select",
				Select: &SelectValue{Name: "Low"},
			},
			"Duration": {
				Type:   "number",
				Number: floatPtr(30),
			},
		},
	}

	title := extractTitle(page)
	if title != "Buy groceries" {
		t.Errorf("title = %q, want 'Buy groceries'", title)
	}

	opts := mapProperties(page)
	if opts.Priority != "high" {
		t.Errorf("priority = %q, want 'high'", opts.Priority)
	}
	if opts.DueDate != "2026-05-10" {
		t.Errorf("due_date = %q, want '2026-05-10'", opts.DueDate)
	}
	if len(opts.Tags) != 2 || opts.Tags[0] != "errands" {
		t.Errorf("tags = %v, want [errands, home]", opts.Tags)
	}
	if opts.Energy != "low" {
		t.Errorf("energy = %q, want 'low'", opts.Energy)
	}
	if opts.Duration != 30 {
		t.Errorf("duration = %d, want 30", opts.Duration)
	}
}

func TestPriorityMapping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"urgent", "urgent"},
		{"Critical", "urgent"},
		{"P0", "urgent"},
		{"high", "high"},
		{"P1", "high"},
		{"medium", "medium"},
		{"Normal", "medium"},
		{"low", "low"},
		{"P3", "low"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		got := string(mapPriority(tt.input))
		if got != tt.want {
			t.Errorf("mapPriority(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestImportWithMockServer(t *testing.T) {
	// Create a mock Notion API server
	mockResp := DatabaseQueryResponse{
		Results: []Page{
			{
				ID: "page-1",
				Properties: map[string]Property{
					"Name":     {Type: "title", Title: []RichText{{PlainText: "Mock task 1"}}},
					"Priority": {Type: "select", Select: &SelectValue{Name: "High"}},
				},
			},
			{
				ID: "page-2",
				Properties: map[string]Property{
					"Name": {Type: "title", Title: []RichText{{PlainText: "Mock task 2"}}},
					"Tags": {Type: "multi_select", MultiSelect: []SelectValue{{Name: "work"}}},
				},
			},
			{
				ID: "page-3",
				Properties: map[string]Property{
					// No title — should be skipped
					"Notes": {Type: "rich_text", RichText: []RichText{{PlainText: "Just notes"}}},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResp)
	}))
	defer srv.Close()

	// Override base URL for testing
	client := &Client{
		token:      "test-token",
		httpClient: srv.Client(),
	}
	// We need to query the mock server, not the real Notion API
	// Use dry-run mode to test mapping without needing an App
	dataDir := t.TempDir()
	a, err := app.Open(dataDir)
	if err != nil {
		t.Fatalf("open app: %v", err)
	}
	defer func() { _ = a.Close() }()

	// Manually query mock and test mapping
	req, _ := http.NewRequest("POST", srv.URL+"/databases/test-db/query", nil)
	client.setHeaders(req)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var dbResp DatabaseQueryResponse
	_ = json.NewDecoder(resp.Body).Decode(&dbResp)

	if len(dbResp.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(dbResp.Results))
	}

	// Test that title extraction works
	title1 := extractTitle(dbResp.Results[0])
	if title1 != "Mock task 1" {
		t.Errorf("title1 = %q", title1)
	}

	// Test that empty title is detected
	title3 := extractTitle(dbResp.Results[2])
	if title3 != "" {
		t.Errorf("title3 should be empty, got %q", title3)
	}

	// Test dry-run import
	result, err := importFromPages(dbResp.Results, a, true)
	if err != nil {
		t.Fatalf("import error: %v", err)
	}
	if result.Imported != 2 {
		t.Errorf("imported = %d, want 2", result.Imported)
	}
	if result.Skipped != 1 {
		t.Errorf("skipped = %d, want 1", result.Skipped)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
