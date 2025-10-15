# Socrawler - Sora Video Crawler

A Go-based service for crawling Sora videos from sora.chatgpt.com with REST API and MCP (Model Context Protocol) support.

## Features

- 🎥 **Sora Video Crawler** - Automatically scroll and download videos/thumbnails from Sora
- 🚀 **REST API** - Simple HTTP API for triggering crawls
- 🔧 **Command-line interface** with `runserver` and `version` commands  
- 🛠️ **Browser Automation** - Uses headless Chrome for dynamic content scraping
- 🌐 **CORS and middleware support** out of the box
- 📝 **MCP Integration** - Optional MCP tools for AI agent integration
- 🔒 **Anti-Detection** - Built-in stealth mode to bypass Cloudflare and other anti-bot protections

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

## Anti-Detection & Cloudflare Bypass

The crawler includes built-in anti-detection features to bypass Cloudflare and other anti-bot protections:

### Features

1. **Stealth Mode** - Uses `go-rod/stealth` library to hide automation signatures
2. **Custom User Agent** - Mimics real Chrome browser on macOS
3. **JavaScript Overrides** - Masks `navigator.webdriver` and other detection points
4. **Realistic Browser Behavior** - Smooth scrolling and natural timing

### How It Works

The crawler automatically applies stealth techniques:

```go
// In browser/browser.go
- Sets realistic User-Agent
- Injects stealth.js to hide automation
- Overrides navigator.webdriver
- Masks Chrome automation flags
- Simulates real browser properties
```

### Testing Anti-Detection

To verify the anti-detection is working:

1. Run with visible browser (non-headless mode):
```bash
go run . runserver --headless=false --debug
```

2. Watch the browser navigate to Sora - it should pass Cloudflare verification automatically

3. Check logs for:
```
level=info msg="Applying stealth mode to page..."
level=info msg="Page loaded, waiting for initial content..."
```

### If Cloudflare Still Blocks

If you still see Cloudflare verification pages:

1. **Try non-headless mode first** - Some sites detect headless browsers:
```bash
go run . runserver --headless=false
```

2. **Add delays** - Increase wait times to appear more human:
```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 120,
    "scroll_interval_seconds": 15,
    "save_path": "./downloads/sora"
  }'
```

3. **Check your IP** - Some IPs may be rate-limited or blocked by Cloudflare

4. **Use residential proxy** - For production use, consider using a proxy service

## Debugging & Troubleshooting

### Enhanced Logging

The crawler now includes detailed logging to help diagnose issues:

1. **Network Request Statistics**: Tracks total requests, OpenAI-related requests, and media requests
2. **Page Information**: Logs page title, URL, and HTML length
3. **Login Detection**: Warns if the page appears to require authentication
4. **Screenshot Capture**: Saves initial page screenshot to `{save_path}/debug_initial_page.png`
5. **Real-time Status**: Reports current video/thumbnail count and network statistics during crawling

### Example Debug Output

```
time="..." level=info msg="Applying stealth mode to page..."
time="..." level=info msg="Page title: Sora, URL: https://sora.chatgpt.com/"
time="..." level=info msg="Saved initial page screenshot to: ./downloads/sora/debug_initial_page.png"
time="..." level=debug msg="Page HTML length: 45678 bytes"
time="..." level=debug msg="Detected media-related URL: https://videos.openai.com/..."
time="..." level=info msg="Found video: https://videos.openai.com/video.mp4"
time="..." level=info msg="Current status: videos=5, thumbnails=5, total_requests=234, openai_requests=45, media_requests=10"
```

### Common Issues

**Cloudflare Verification Page**

If you see a Cloudflare "Verifying you are human" page:
1. Try running in non-headless mode: `--headless=false`
2. The stealth mode should handle this automatically
3. Check the screenshot to confirm what page is being loaded
4. See the "Anti-Detection & Cloudflare Bypass" section above

**No videos found (videos=0, thumbnails=0)**

Possible causes:
1. **Authentication Required**: Sora may require login. Check the logs for "Page may require login" warning
2. **Changed URL Structure**: The video URLs may have changed. Check debug logs for "Detected media-related URL" entries
3. **Network Issues**: Check if `openai_requests` and `media_requests` are > 0
4. **Insufficient Wait Time**: Try increasing `scroll_interval_seconds` or `total_duration_seconds`

**Debugging Steps:**
1. Enable debug logging: `go run . runserver --debug`
2. Check the screenshot: `{save_path}/debug_initial_page.png`
3. Review network statistics in the logs
4. Look for "Detected media-related URL" entries to see what URLs are being captured

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

### For Ubuntu Server Deployment

See **[DEPLOYMENT.md](DEPLOYMENT.md)** for detailed instructions on:
- Installing Chrome on Ubuntu servers
- Docker deployment
- Systemd service setup
- Production configurations

**Quick Ubuntu Setup:**
```bash
# Run the automated setup script
./setup_browser_ubuntu.sh
```

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


# 查看帮助
./run_service.sh help

# 启动服务
./run_service.sh start

# 使用自定义端口启动
./run_service.sh start --port=9090

# 非 headless 模式启动（用于调试）
./run_service.sh start --headless=false

# 查看服务状态
./run_service.sh status

# 运行测试（自动启动服务并发送测试请求）
./run_service.sh test

# 运行更长的测试（5分钟，每15秒滚动）
./run_service.sh test --duration=300 --interval=15

# 查看实时日志
./run_service.sh logs

# 停止服务
./run_service.sh stop

# 重启服务
./run_service.sh restart

# 清理所有内容
./run_service.sh cleanup