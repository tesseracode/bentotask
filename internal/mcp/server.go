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
	app       *app.App
	tools     map[string]Tool
	resources map[string]Resource
	prompts   map[string]Prompt
}

// Tool represents an MCP tool that can be called by the AI.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
	handler     func(params map[string]any) (string, error)
}

// Resource represents read-only data the AI can browse.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
	handler     func() (string, error)
}

// Prompt represents a reusable prompt template.
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	handler     func(args map[string]string) ([]PromptMessage, error)
}

// PromptArgument describes an argument to a prompt.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// PromptMessage is a message returned by a prompt.
type PromptMessage struct {
	Role    string        `json:"role"`
	Content PromptContent `json:"content"`
}

// PromptContent is the content of a prompt message.
type PromptContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
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

// NewServer creates a new MCP server and registers all tools, resources, and prompts.
func NewServer(a *app.App) *Server {
	s := &Server{
		app:       a,
		tools:     make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
	}
	s.registerTools()
	s.registerResources()
	s.registerPrompts()
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
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
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
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
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

func (s *Server) registerResource(r Resource) {
	s.resources[r.URI] = r
}

func (s *Server) registerPrompt(p Prompt) {
	s.prompts[p.Name] = p
}

func (s *Server) handleResourcesList(req jsonRPCRequest) jsonRPCResponse {
	list := make([]map[string]any, 0, len(s.resources))
	for _, r := range s.resources {
		list = append(list, map[string]any{
			"uri":         r.URI,
			"name":        r.Name,
			"description": r.Description,
			"mimeType":    r.MimeType,
		})
	}
	return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"resources": list}}
}

func (s *Server) handleResourcesRead(req jsonRPCRequest) jsonRPCResponse {
	uri, _ := req.Params["uri"].(string)
	r, ok := s.resources[uri]
	if !ok {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCError{Code: -32602, Message: fmt.Sprintf("unknown resource: %s", uri)}}
	}
	text, err := r.handler()
	if err != nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCError{Code: -32603, Message: err.Error()}}
	}
	return jsonRPCResponse{
		JSONRPC: "2.0", ID: req.ID,
		Result: map[string]any{
			"contents": []map[string]string{{"uri": uri, "mimeType": r.MimeType, "text": text}},
		},
	}
}

func (s *Server) handlePromptsList(req jsonRPCRequest) jsonRPCResponse {
	list := make([]map[string]any, 0, len(s.prompts))
	for _, p := range s.prompts {
		list = append(list, map[string]any{
			"name":        p.Name,
			"description": p.Description,
			"arguments":   p.Arguments,
		})
	}
	return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"prompts": list}}
}

func (s *Server) handlePromptsGet(req jsonRPCRequest) jsonRPCResponse {
	name, _ := req.Params["name"].(string)
	p, ok := s.prompts[name]
	if !ok {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCError{Code: -32602, Message: fmt.Sprintf("unknown prompt: %s", name)}}
	}
	args := make(map[string]string)
	if rawArgs, ok := req.Params["arguments"].(map[string]any); ok {
		for k, v := range rawArgs {
			if sv, ok := v.(string); ok {
				args[k] = sv
			}
		}
	}
	messages, err := p.handler(args)
	if err != nil {
		return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonRPCError{Code: -32603, Message: err.Error()}}
	}
	return jsonRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"messages": messages}}
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
