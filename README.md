# Socrawler - Sora Video Crawler

A Go-based service for crawling Sora videos from sora.chatgpt.com with REST API and MCP (Model Context Protocol) support.

## Features

- 🎥 **Sora Video Crawler** - Automatically scroll and download videos/thumbnails from Sora
- 🚀 **REST API** - Simple HTTP API for triggering crawls
- 🔧 **Command-line interface** with `runserver` and `version` commands  
- 🛠️ **Browser Automation** - Uses headless Chrome for dynamic content scraping
- 🌐 **CORS and middleware support** out of the box
- 📝 **MCP Integration** - Optional MCP tools for AI agent integration

## Quick Start

### 1. Install Dependencies

```bash
cd socrawler
go mod tidy
```

### 2. Run the Server

```bash
# Start the server on default port 8080
go run . runserver

# Or specify a custom port
go run . runserver --port 3000

# Enable debug logging and disable headless mode (to see browser)
go run . runserver --debug --headless=false

# For production, use headless mode (default)
go run . runserver --headless=true
```

### 3. Test the Sora Crawler API

Use the included test script:

```bash
./test_sora_crawl.sh
```

Or use curl directly:

```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 60,
    "scroll_interval_seconds": 10,
    "save_path": "./downloads/sora"
  }'
```

### 4. Test the MCP Connection (Optional)

Use the MCP Inspector to test your server:

```bash
npx @modelcontextprotocol/inspector
```

Then connect to: `http://localhost:8080/mcp`

## Available Commands

```bash
# Start the server
go run . runserver [--port 8080] [--debug] [--headless=true]

# Show version information  
go run . version

# Show help
go run . --help
```

## API Endpoints

### POST /api/sora/crawl

Crawls Sora videos and thumbnails from sora.chatgpt.com.

**Request Body:**
```json
{
  "total_duration_seconds": 300,    // Total crawl duration (default: 300s)
  "scroll_interval_seconds": 20,    // Scroll interval (default: 20s)
  "save_path": "./downloads/sora"   // Save directory (default: ./downloads/sora)
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "videos": ["path1.mp4", "path2.mp4"],
    "thumbnails": ["thumb1.webp", "thumb2.webp"],
    "total_videos": 2,
    "total_thumbnails": 2,
    "duration_seconds": 300
  },
  "message": "Crawl completed successfully"
}
```

### GET /health

Health check endpoint.

## Project Structure

```
socrawler/
├── main.go              # CLI entry point with cobra commands
├── app_server.go        # Core server implementation
├── service.go           # Sora service business logic
├── handlers.go          # API request handlers
├── routes.go            # HTTP routing setup
├── middleware.go        # CORS and error handling middleware
├── types.go             # Type definitions
├── streamable_http.go   # MCP protocol handler (JSON-RPC over HTTP)
├── tools.go             # MCP tool implementations
├── browser/             # Browser automation
│   └── browser.go       # Clean browser initialization
├── configs/             # Configuration
│   └── browser.go       # Headless mode config
├── sora/                # Sora crawler implementation
│   ├── crawler.go       # Browser scrolling and URL interception
│   ├── downloader.go    # Media file downloader
│   └── types.go         # Internal types
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

## How It Works

1. **Browser Automation**: Uses go-rod to control a headless Chrome browser
2. **Dynamic Scrolling**: Automatically scrolls the Sora page at configured intervals
3. **Network Interception**: Captures video and thumbnail URLs from network requests to `videos.openai.com`
4. **Media Detection**: 
   - Videos: URLs containing `.mp4`
   - Thumbnails: URLs containing `.webp` or `thumbnail`
5. **Download**: Downloads and saves media files with unique SHA256-based filenames

## Configuration

The server can be configured through command-line flags:

- `--port, -p`: Server port (default: 8080)
- `--debug, -d`: Enable debug logging
- `--headless`: Run browser in headless mode (default: true)
- `--help, -h`: Show help information

## Usage Examples

### Quick Test (1 minute)
```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 60,
    "scroll_interval_seconds": 10,
    "save_path": "./downloads/sora"
  }'
```

### Production Crawl (5 minutes)
```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 300,
    "scroll_interval_seconds": 20,
    "save_path": "./downloads/sora"
  }'
```

### Long Crawl (30 minutes)
```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 1800,
    "scroll_interval_seconds": 30,
    "save_path": "./downloads/sora"
  }'
```

## Development Tips

1. **Logging**: Use `logrus` for structured logging
2. **Context**: Always pass context for cancellation support
3. **Error Handling**: Proper error responses with status codes
4. **Browser Mode**: Use `--headless=false` for debugging
5. **Testing**: Start with short durations for testing

## Building for Production

```bash
# Build binary
go build -o socrawler .

# Run the binary
./socrawler runserver --port 8080 --headless=true
```

## Requirements

- Go 1.23.0 or higher
- Chrome/Chromium browser (automatically managed by go-rod)
- Internet connection (for accessing sora.chatgpt.com)

## Troubleshooting

### Browser not launching
- Make sure Chrome/Chromium is installed
- Try running with `--headless=false` to see browser activity
- Check logs with `--debug` flag

### No videos downloaded
- Ensure Sora page is accessible
- Increase `total_duration_seconds` to allow more time
- Check network connectivity
- Verify save_path is writable

### Memory usage
- The crawler keeps videos in memory before writing
- For long crawls, monitor system resources
- Consider shorter crawl durations or restart periodically

## License

This project is provided as-is for educational and development purposes.

---

Built with ❤️ using Go, go-rod, and browser automation.
