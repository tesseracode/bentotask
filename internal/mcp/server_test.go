package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/tesserabox/bentotask/internal/app"
)

func setupTestMCP(t *testing.T) (io.Writer, *bufio.Scanner) {
	t.Helper()
	dataDir := t.TempDir()
	a, err := app.Open(dataDir)
	if err != nil {
		t.Fatalf("open app: %v", err)
	}
	t.Cleanup(func() { _ = a.Close() })

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	srv := NewServer(a)
	go func() {
		_ = srv.RunWithIO(stdinR, stdoutW)
		_ = stdoutW.Close()
	}()

	return stdinW, bufio.NewScanner(stdoutR)
}

func sendRequest(t *testing.T, w io.Writer, id any, method string, params map[string]any) {
	t.Helper()
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
	}
	if params != nil {
		req["params"] = params
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	if err != nil {
		t.Fatalf("write request: %v", err)
	}
}

func readResponse(t *testing.T, scanner *bufio.Scanner) map[string]any {
	t.Helper()
	if !scanner.Scan() {
		t.Fatalf("no response (err: %v)", scanner.Err())
	}
	var resp map[string]any
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v\nraw: %s", err, scanner.Text())
	}
	return resp
}

func TestMCPInitialize(t *testing.T) {
	w, scanner := setupTestMCP(t)

	sendRequest(t, w, 1, "initialize", nil)
	resp := readResponse(t, scanner)

	if resp["id"] != float64(1) {
		t.Errorf("id = %v, want 1", resp["id"])
	}

	result, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("result not a map: %v", resp["result"])
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("protocolVersion = %v", result["protocolVersion"])
	}

	serverInfo, _ := result["serverInfo"].(map[string]any)
	if serverInfo["name"] != "bentotask" {
		t.Errorf("serverInfo.name = %v", serverInfo["name"])
	}

	caps, _ := result["capabilities"].(map[string]any)
	if _, ok := caps["tools"]; !ok {
		t.Error("capabilities missing tools")
	}
}

func TestMCPToolsList(t *testing.T) {
	w, scanner := setupTestMCP(t)

	// Initialize first
	sendRequest(t, w, 1, "initialize", nil)
	_ = readResponse(t, scanner)

	// List tools
	sendRequest(t, w, 2, "tools/list", nil)
	resp := readResponse(t, scanner)

	result, _ := resp["result"].(map[string]any)
	tools, _ := result["tools"].([]any)

	if len(tools) != 18 {
		t.Errorf("expected 18 tools, got %d", len(tools))
		for _, tool := range tools {
			m, _ := tool.(map[string]any)
			t.Logf("  tool: %s", m["name"])
		}
	}

	// Verify add_task is present with required schema
	found := false
	for _, tool := range tools {
		m, _ := tool.(map[string]any)
		if m["name"] == "add_task" {
			found = true
			schema, _ := m["inputSchema"].(map[string]any)
			if schema["type"] != "object" {
				t.Error("add_task schema type should be object")
			}
			required, _ := schema["required"].([]any)
			if len(required) == 0 || required[0] != "title" {
				t.Error("add_task should require title")
			}
		}
	}
	if !found {
		t.Error("add_task tool not found")
	}
}

func TestMCPAddTask(t *testing.T) {
	w, scanner := setupTestMCP(t)

	sendRequest(t, w, 1, "initialize", nil)
	_ = readResponse(t, scanner)

	sendRequest(t, w, 2, "tools/call", map[string]any{
		"name": "add_task",
		"arguments": map[string]any{
			"title":    "Buy groceries",
			"priority": "high",
		},
	})
	resp := readResponse(t, scanner)

	result, _ := resp["result"].(map[string]any)
	content, _ := result["content"].([]any)
	if len(content) == 0 {
		t.Fatal("no content in response")
	}
	block, _ := content[0].(map[string]any)
	text, _ := block["text"].(string)

	if !strings.Contains(text, "Buy groceries") {
		t.Errorf("response should contain task title: %s", text)
	}
	if !strings.Contains(text, "priority: high") {
		t.Errorf("response should contain priority: %s", text)
	}
}

func TestMCPListTasks(t *testing.T) {
	w, scanner := setupTestMCP(t)

	sendRequest(t, w, 1, "initialize", nil)
	_ = readResponse(t, scanner)

	// Add tasks first
	sendRequest(t, w, 2, "tools/call", map[string]any{
		"name": "add_task", "arguments": map[string]any{"title": "Task A"},
	})
	_ = readResponse(t, scanner)

	sendRequest(t, w, 3, "tools/call", map[string]any{
		"name": "add_task", "arguments": map[string]any{"title": "Task B"},
	})
	_ = readResponse(t, scanner)

	// List
	sendRequest(t, w, 4, "tools/call", map[string]any{
		"name": "list_tasks", "arguments": map[string]any{},
	})
	resp := readResponse(t, scanner)

	result, _ := resp["result"].(map[string]any)
	content, _ := result["content"].([]any)
	block, _ := content[0].(map[string]any)
	text, _ := block["text"].(string)

	if !strings.Contains(text, "Found 2 tasks") {
		t.Errorf("expected 2 tasks in list: %s", text)
	}
	if !strings.Contains(text, "Task A") {
		t.Errorf("list should contain Task A: %s", text)
	}
}

func TestMCPSuggest(t *testing.T) {
	w, scanner := setupTestMCP(t)

	sendRequest(t, w, 1, "initialize", nil)
	_ = readResponse(t, scanner)

	sendRequest(t, w, 2, "tools/call", map[string]any{
		"name": "add_task", "arguments": map[string]any{"title": "Write report", "priority": "high", "duration": 30},
	})
	_ = readResponse(t, scanner)

	sendRequest(t, w, 3, "tools/call", map[string]any{
		"name":      "suggest",
		"arguments": map[string]any{"time": 60, "energy": "medium", "count": 3},
	})
	resp := readResponse(t, scanner)

	result, _ := resp["result"].(map[string]any)
	content, _ := result["content"].([]any)
	block, _ := content[0].(map[string]any)
	text, _ := block["text"].(string)

	if !strings.Contains(text, "Write report") {
		t.Errorf("suggest should include the task: %s", text)
	}
	if !strings.Contains(text, "score:") {
		t.Errorf("suggest should include score: %s", text)
	}
}

func TestMCPHabitLifecycle(t *testing.T) {
	w, scanner := setupTestMCP(t)

	sendRequest(t, w, 1, "initialize", nil)
	_ = readResponse(t, scanner)

	// Create habit
	sendRequest(t, w, 2, "tools/call", map[string]any{
		"name": "add_habit", "arguments": map[string]any{"title": "Meditate"},
	})
	resp := readResponse(t, scanner)
	result, _ := resp["result"].(map[string]any)
	content, _ := result["content"].([]any)
	block, _ := content[0].(map[string]any)
	text, _ := block["text"].(string)

	if !strings.Contains(text, "Meditate") {
		t.Errorf("add_habit response: %s", text)
	}

	// Extract habit ID from list
	sendRequest(t, w, 3, "tools/call", map[string]any{
		"name": "list_habits", "arguments": map[string]any{},
	})
	resp = readResponse(t, scanner)
	result, _ = resp["result"].(map[string]any)
	content, _ = result["content"].([]any)
	block, _ = content[0].(map[string]any)
	listText, _ := block["text"].(string)

	// Extract the 8-char ID from "[01ABCDEF] Meditate"
	idx := strings.Index(listText, "[")
	if idx < 0 {
		t.Fatalf("no ID in list: %s", listText)
	}
	endIdx := strings.Index(listText[idx:], "]")
	habitID := listText[idx+1 : idx+endIdx]

	// Log completion
	sendRequest(t, w, 4, "tools/call", map[string]any{
		"name": "log_habit", "arguments": map[string]any{"id": habitID},
	})
	resp = readResponse(t, scanner)
	result, _ = resp["result"].(map[string]any)
	content, _ = result["content"].([]any)
	block, _ = content[0].(map[string]any)
	logText, _ := block["text"].(string)

	if !strings.Contains(logText, "Meditate") {
		t.Errorf("log_habit response: %s", logText)
	}

	// Get stats
	sendRequest(t, w, 5, "tools/call", map[string]any{
		"name": "habit_stats", "arguments": map[string]any{"id": habitID},
	})
	resp = readResponse(t, scanner)
	result, _ = resp["result"].(map[string]any)
	content, _ = result["content"].([]any)
	block, _ = content[0].(map[string]any)
	statsText, _ := block["text"].(string)

	if !strings.Contains(statsText, "1 day streak") {
		t.Errorf("habit_stats should show streak of 1: %s", statsText)
	}
	if !strings.Contains(statsText, "1 total completions") {
		t.Errorf("habit_stats should show 1 total: %s", statsText)
	}
}

func TestMCPNotificationNoResponse(t *testing.T) {
	w, scanner := setupTestMCP(t)

	// Send initialize (has ID — gets response)
	sendRequest(t, w, 1, "initialize", nil)
	_ = readResponse(t, scanner)

	// Send notification (no ID) — should NOT get a response
	// We write a notification then a real request — we should only get 1 response
	notif := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	_, _ = fmt.Fprintf(w, "%s\n", notif)

	// Send a real request after the notification
	sendRequest(t, w, 2, "tools/list", nil)
	resp := readResponse(t, scanner)

	// The response should be for ID 2 (tools/list), not the notification
	if resp["id"] != float64(2) {
		t.Errorf("response id = %v, want 2 (notification should have been skipped)", resp["id"])
	}
}
