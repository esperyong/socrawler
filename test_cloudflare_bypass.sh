#!/bin/bash

# 测试 Cloudflare 绕过功能
# 使用非 headless 模式以便观察浏览器行为

echo "========================================="
echo "Testing Cloudflare Bypass with Stealth Mode"
echo "========================================="
echo ""

# 检查服务器是否已经在运行
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null ; then
    echo "⚠️  Server is already running on port 8080"
    echo "   Using existing server..."
    echo ""
else
    echo "🚀 Starting server in non-headless mode..."
    echo "   (You will see a browser window open)"
    echo ""
    
    # 启动服务器（非 headless 模式，debug 日志）
    ./socrawler runserver --headless=false --debug &
    SERVER_PID=$!
    
    echo "   Server PID: $SERVER_PID"
    echo "   Waiting for server to start..."
    sleep 3
    echo ""
fi

# 测试健康检查
echo "📡 Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
echo "   Response: $HEALTH_RESPONSE"
echo ""

# 发送爬取请求
echo "🎬 Starting Sora crawl with stealth mode..."
echo "   Duration: 60 seconds"
echo "   Scroll interval: 10 seconds"
echo "   Watch the browser window - it should bypass Cloudflare automatically!"
echo ""

curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 60,
    "scroll_interval_seconds": 10,
    "save_path": "./downloads/sora"
  }' \
  -w "\n\n📊 HTTP Status: %{http_code}\n" \
  -s | jq '.'

echo ""
echo "========================================="
echo "Test completed!"
echo ""
echo "📝 Check the logs above for:"
echo "   - 'Navigating to Sora page with stealth mode enabled...'"
echo "   - 'Page title: Sora, URL: https://sora.chatgpt.com/'"
echo "   - Video and thumbnail counts"
echo ""
echo "🖼️  Check screenshot:"
echo "   ./downloads/sora/debug_initial_page.png"
echo ""
echo "If you see Cloudflare verification page in the screenshot,"
echo "the bypass didn't work. Otherwise, it succeeded!"
echo "========================================="

# 如果我们启动了服务器，询问是否要停止
if [ ! -z "$SERVER_PID" ]; then
    echo ""
    read -p "Stop the server? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID
        echo "Server stopped."
    else
        echo "Server still running (PID: $SERVER_PID)"
        echo "To stop it later: kill $SERVER_PID"
    fi
fi

