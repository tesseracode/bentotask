// Package mcp implements a Model Context Protocol server over stdio.
// MCP allows AI assistants to interact with BentoTask via JSON-RPC 2.0.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tesserabox/bentotask/internal/app"
)

// Server handles MCP JSON-RPC communication over stdio.
type Server struct {
	app   *app.App
	tools map[string]Tool
}

// Tool represents an MCP tool that can be called by the AI.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	handler     func(params map[string]any) (string, error)
}

// JSON-RPC types
type jsonRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id"`
	Result  any           `json:"result,omitempty"`
	Error   *jsonRPCError `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// toolContent is the MCP content block returned from tool calls.
type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewServer creates a new MCP server and registers all tools.
func NewServer(a *app.App) *Server {
	s := &Server{
		app:   a,
		tools: make(map[string]Tool),
	}
	s.registerTools()
	return s
}

// Run starts the MCP server reading from os.Stdin and writing to os.Stdout.
func (s *Server) Run() error {
	return s.RunWithIO(os.Stdin, os.Stdout)
}

// RunWithIO starts the MCP server with custom reader/writer (for testing).
func (s *Server) RunWithIO(reader io.Reader, writer io.Writer) error {
	scanner := bufio.NewScanner(reader)
	// Increase buffer for large messages
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("MCP: invalid JSON-RPC: %v", err)
			s.writeError(writer, nil, -32700, "Parse error")
			continue
		}

		// Notifications have no ID — don't send a response
		if req.ID == nil {
			log.Printf("MCP: notification: %s", req.Method)
			continue
		}

		resp := s.handleRequest(req)
		s.writeResponse(writer, resp)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req jsonRPCRequest) jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)},
		}
	}
}

func (s *Server) handleInitialize(req jsonRPCRequest) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "bentotask",
				"version": "1.0.0",
			},
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
		},
	}
}

func (s *Server) handleToolsList(req jsonRPCRequest) jsonRPCResponse {
	toolList := make([]map[string]any, 0, len(s.tools))
	for _, t := range s.tools {
		toolList = append(toolList, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  map[string]any{"tools": toolList},
	}
}

func (s *Server) handleToolsCall(req jsonRPCRequest) jsonRPCResponse {
	name, _ := req.Params["name"].(string)
	args, _ := req.Params["arguments"].(map[string]any)
	if args == nil {
		args = make(map[string]any)
	}

	tool, ok := s.tools[name]
	if !ok {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonRPCError{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", name)},
		}
	}

	text, err := tool.handler(args)
	if err != nil {
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []toolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				"isError": true,
			},
		}
	}

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"content": []toolContent{{Type: "text", Text: text}},
		},
	}
}

func (s *Server) writeResponse(w io.Writer, resp jsonRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("MCP: failed to marshal response: %v", err)
		return
	}
	_, _ = fmt.Fprintf(w, "%s\n", data)
}

func (s *Server) writeError(w io.Writer, id any, code int, message string) {
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: message},
	}
	s.writeResponse(w, resp)
}

func (s *Server) register(t Tool) {
	s.tools[t.Name] = t
}

// Parameter extraction helpers

func getString(params map[string]any, key string) string {
	v, _ := params[key].(string)
	return v
}

func getInt(params map[string]any, key string) int {
	switch v := params[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func getStringSlice(params map[string]any, key string) []string {
	arr, ok := params[key].([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
