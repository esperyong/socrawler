package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// StreamableHTTPHandler handles Streamable HTTP protocol for MCP requests
func (s *AppServer) StreamableHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Mcp-Session-Id")

		// Handle OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle different methods
		switch r.Method {
		case "GET":
			// GET requests for SSE connections (optional feature)
			s.handleSSEConnection(w, r)
		case "POST":
			// POST requests handle JSON-RPC
			s.handleJSONRPCRequest(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleSSEConnection handles SSE connections (optional, for server push)
func (s *AppServer) handleSSEConnection(w http.ResponseWriter, r *http.Request) {
	// Check if SSE is supported
	if !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		http.Error(w, "SSE not requested", http.StatusBadRequest)
		return
	}

	// Set SSE response headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send initialization message
	fmt.Fprintf(w, "event: open\n")
	fmt.Fprintf(w, "data: {\"type\":\"connection\",\"status\":\"connected\"}\n\n")

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Keep connection open (in real use, you can push notifications here)
	<-r.Context().Done()
}

// handleJSONRPCRequest handles JSON-RPC requests
func (s *AppServer) handleJSONRPCRequest(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendStreamableError(w, nil, -32700, "Parse error")
		return
	}
	defer r.Body.Close()

	// Parse JSON-RPC request
	var request JSONRPCRequest
	if err := json.Unmarshal(body, &request); err != nil {
		s.sendStreamableError(w, nil, -32700, "Parse error")
		return
	}

	logrus.WithField("method", request.Method).Info("Received Streamable HTTP request")

	// Check Accept header to determine if client supports SSE
	acceptSSE := strings.Contains(r.Header.Get("Accept"), "text/event-stream")

	// Process request
	response := s.processJSONRPCRequest(&request, r.Context())

	// Send response based on client capabilities
	if acceptSSE && s.isStreamableMethod(request.Method) {
		s.sendSSEResponse(w, response)
	} else {
		// Use regular JSON response
		s.sendJSONResponse(w, response)
	}
}

// processJSONRPCRequest processes JSON-RPC requests and returns responses
func (s *AppServer) processJSONRPCRequest(request *JSONRPCRequest, ctx context.Context) *JSONRPCResponse {
	switch request.Method {
	case "initialize":
		return s.processInitialize(request)
	case "initialized":
		// Client confirms initialization complete
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  map[string]interface{}{},
			ID:      request.ID,
		}
	case "ping":
		// Handle ping requests
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  map[string]interface{}{},
			ID:      request.ID,
		}
	case "tools/list":
		return s.processToolsList(request)
	case "tools/call":
		return s.processToolCall(ctx, request)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
			ID: request.ID,
		}
	}
}

// processInitialize handles initialization requests
func (s *AppServer) processInitialize(request *JSONRPCRequest) *JSONRPCResponse {
	result := map[string]interface{}{
		"protocolVersion": "2025-03-26", // Use new protocol version
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "gomcp-scaffold",
			"version": "1.0.0",
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      request.ID,
	}
}

// processToolsList handles tool list requests
func (s *AppServer) processToolsList(request *JSONRPCRequest) *JSONRPCResponse {
	tools := []map[string]interface{}{
		{
			"name":        "hello_world",
			"description": "A simple hello world tool for testing MCP functionality",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name to greet (optional)",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Custom message (optional)",
					},
				},
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result: map[string]interface{}{
			"tools": tools,
		},
		ID: request.ID,
	}
}

// processToolCall handles tool calls
func (s *AppServer) processToolCall(ctx context.Context, request *JSONRPCRequest) *JSONRPCResponse {
	// Parse parameters
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
			},
			ID: request.ID,
		}
	}

	toolName, _ := params["name"].(string)
	toolArgs, _ := params["arguments"].(map[string]interface{})

	var result *MCPToolResult

	switch toolName {
	case "hello_world":
		result = s.handleHelloWorld(ctx, toolArgs)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", toolName),
			},
			ID: request.ID,
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      request.ID,
	}
}

// isStreamableMethod determines if a method supports streaming responses
func (s *AppServer) isStreamableMethod(_ string) bool {
	// Currently, our methods don't need streaming responses
	// You can add streaming support for specific methods here in the future
	return false
}

// sendJSONResponse sends regular JSON responses
func (s *AppServer) sendJSONResponse(w http.ResponseWriter, response *JSONRPCResponse) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logrus.WithError(err).Error("Failed to encode response")
	}
}

// sendSSEResponse sends SSE responses
func (s *AppServer) sendSSEResponse(w http.ResponseWriter, response *JSONRPCResponse) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Convert response to JSON
	data, err := json.Marshal(response)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal SSE response")
		return
	}

	// Send SSE format response
	fmt.Fprintf(w, "data: %s\n\n", string(data))

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// sendStreamableError sends error responses
func (s *AppServer) sendStreamableError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := &JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
	s.sendJSONResponse(w, response)
}
