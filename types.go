package main

// HTTP API Response Types

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

// JSON-RPC Types

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
	ID      any    `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      any           `json:"id"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCP Types

// MCPToolCall represents an MCP tool call
type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolResult represents the result of an MCP tool call
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents MCP content
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Sora Crawler Types

// SoraCrawlRequest represents a request to crawl Sora videos
type SoraCrawlRequest struct {
	TotalDurationSeconds  int    `json:"total_duration_seconds"`  // Total duration in seconds (default: 300)
	ScrollIntervalSeconds int    `json:"scroll_interval_seconds"` // Scroll interval in seconds (default: 20)
	SavePath              string `json:"save_path"`               // Save path (default: ./downloads/sora)
}

// SoraCrawlResponse represents the response from crawling Sora videos
type SoraCrawlResponse struct {
	Videos          []string `json:"videos"`           // List of downloaded video paths
	Thumbnails      []string `json:"thumbnails"`       // List of downloaded thumbnail paths
	TotalVideos     int      `json:"total_videos"`     // Total number of videos downloaded
	TotalThumbnails int      `json:"total_thumbnails"` // Total number of thumbnails downloaded
	DurationSeconds int      `json:"duration_seconds"` // Actual duration in seconds
}
