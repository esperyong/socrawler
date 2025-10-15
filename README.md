# GoMCP - Go MCP Server Scaffold

A clean, minimal scaffold for building MCP (Model Context Protocol) servers in Go.

## Features

- 🚀 **Ready-to-use MCP server** with JSON-RPC and HTTP support
- 🔧 **Command-line interface** with `runserver` and `version` commands  
- 🛠️ **Extensible architecture** for adding custom tools
- 🌐 **CORS and middleware support** out of the box
- 📝 **Sample hello_world tool** as a starting point
- 🔄 **SSE support** for streaming responses (optional)

## Quick Start

### 1. Install Dependencies

```bash
cd gomcp
go mod tidy
```

### 2. Run the Server

```bash
# Start the server on default port 8080
go run . runserver

# Or specify a custom port
go run . runserver --port 3000

# Enable debug logging
go run . runserver --debug
```

### 3. Test the MCP Connection

Use the MCP Inspector to test your server:

```bash
npx @modelcontextprotocol/inspector
```

Then connect to: `http://localhost:8080/mcp`

## Available Commands

```bash
# Start the MCP server
go run . runserver [--port 8080] [--debug]

# Show version information  
go run . version

# Show help
go run . --help
```

## Project Structure

```
gomcp/
├── main.go              # CLI entry point with cobra commands
├── app_server.go        # Core server implementation
├── streamable_http.go   # MCP protocol handler (JSON-RPC over HTTP)
├── routes.go            # HTTP routing setup
├── middleware.go        # CORS and error handling middleware
├── types.go             # MCP and JSON-RPC type definitions
├── tools.go             # MCP tool implementations
├── go.mod               # Go module definition
└── README.md            # This file
```

## Adding New MCP Tools

1. **Define your tool in `tools.go`:**

```go
func (s *AppServer) handleMyCustomTool(ctx context.Context, args map[string]interface{}) *MCPToolResult {
    // Extract arguments
    param1, _ := args["param1"].(string)
    
    // Your tool logic here
    result := fmt.Sprintf("Processing: %s", param1)
    
    return &MCPToolResult{
        Content: []MCPContent{{
            Type: "text", 
            Text: result,
        }},
        IsError: false,
    }
}
```

2. **Register the tool in `streamable_http.go`:**

Add to `processToolsList()`:
```go
{
    "name":        "my_custom_tool",
    "description": "Description of what your tool does",
    "inputSchema": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Description of param1",
            },
        },
        "required": []string{"param1"},
    },
},
```

Add to `processToolCall()`:
```go
case "my_custom_tool":
    result = s.handleMyCustomTool(ctx, toolArgs)
```

## Endpoints

- **`GET /health`** - Health check endpoint
- **`POST /mcp`** - Main MCP endpoint for JSON-RPC requests
- **`GET /mcp`** - SSE endpoint for streaming (optional)

## Configuration

The server can be configured through command-line flags:

- `--port, -p`: Server port (default: 8080)
- `--debug, -d`: Enable debug logging
- `--help, -h`: Show help information

## Example MCP Tools

### Hello World Tool

The scaffold includes a sample `hello_world` tool that demonstrates:
- Parameter parsing
- Response formatting
- Error handling
- Logging

Test it with:
```json
{
  "name": "hello_world",
  "arguments": {
    "name": "Developer",
    "message": "Welcome"
  }
}
```

## Development Tips

1. **Logging**: Use `logrus` for structured logging
2. **Context**: Always pass context for cancellation support
3. **Error Handling**: Return proper `MCPToolResult` with `IsError: true` for errors
4. **Validation**: Validate input parameters in your tool handlers
5. **Testing**: Use the MCP Inspector for interactive testing

## Building for Production

```bash
# Build binary
go build -o gomcp-server .

# Run the binary
./gomcp-server runserver --port 8080
```

## License

This scaffold is provided as-is for educational and development purposes.

---

Built with ❤️ using Go and the MCP protocol specification.
