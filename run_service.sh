#!/bin/bash

# Service runner script for socrawler
# This script manages the socrawler service lifecycle: start, stop, status, and test

set -e

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="socrawler"
SERVICE_BIN="./socrawler"
SERVICE_PORT="8080"
SERVICE_HOST="localhost"
PID_FILE="./${SERVICE_NAME}.pid"
LOG_FILE="./debug.log"
CURL_LOG_FILE="./curl_response.log"

# Default parameters
HEADLESS="true"
DEBUG="true"

# Test parameters
TEST_DURATION=60
TEST_INTERVAL=10
TEST_SAVE_PATH="./downloads/sora"

# Feed parameters
FEED_SAVE_PATH="./downloads/sora"
FEED_DB_PATH="./sora.db"
FEED_LIMIT=50

# Goldcast upload parameters
GOLDCAST_API_KEY="${GOLDCAST_API_KEY:-ucHZBRJ1.w8njpEorJlDgjp0ESnw0qSyOkN6V6VUe}"
GOLDCAST_API_URL="${GOLDCAST_API_URL:-https://financial.xiaoyequ9.com/api/v1/external/media/upload}"
GOLDCAST_LIMIT=0

# OSS parameters (required for Goldcast upload)
OSS_ACCESS_KEY_ID="${OSS_ACCESS_KEY_ID:-}"
OSS_ACCESS_KEY_SECRET="${OSS_ACCESS_KEY_SECRET:-}"
OSS_BUCKET_NAME="${OSS_BUCKET_NAME:-dreammedias}"
OSS_ENDPOINT="${OSS_ENDPOINT:-oss-cn-beijing.aliyuncs.com}"
OSS_REGION="${OSS_REGION:-cn-beijing}"

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if service is running
is_running() {
    if [ -f "$PID_FILE" ]; then
        local pid=$(cat "$PID_FILE")
        if ps -p "$pid" > /dev/null 2>&1; then
            return 0
        else
            # PID file exists but process is not running
            rm -f "$PID_FILE"
            return 1
        fi
    fi
    return 1
}

# Function to get service PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    else
        echo ""
    fi
}

# Function to check if service binary exists
check_binary() {
    if [ ! -f "$SERVICE_BIN" ]; then
        print_error "Service binary not found: $SERVICE_BIN"
        print_info "Please build the service first: go build -o socrawler ."
        exit 1
    fi
    
    if [ ! -x "$SERVICE_BIN" ]; then
        print_error "Service binary is not executable: $SERVICE_BIN"
        print_info "Making it executable..."
        chmod +x "$SERVICE_BIN"
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local max_attempts=30
    local attempt=0
    
    print_info "Waiting for service to be ready..."
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "http://${SERVICE_HOST}:${SERVICE_PORT}/health" > /dev/null 2>&1; then
            print_success "Service is ready!"
            return 0
        fi
        
        attempt=$((attempt + 1))
        echo -n "."
        sleep 1
    done
    
    echo ""
    print_error "Service failed to start within ${max_attempts} seconds"
    print_info "Check logs: tail -f $LOG_FILE"
    return 1
}

# Function to start the service
start_service() {
    print_info "Starting ${SERVICE_NAME} service..."
    
    # Check if already running
    if is_running; then
        local pid=$(get_pid)
        print_warning "Service is already running (PID: $pid)"
        return 0
    fi
    
    # Check binary
    check_binary
    
    # Create downloads directory if it doesn't exist
    mkdir -p "$TEST_SAVE_PATH"
    
    # Rotate old log file if it exists and is large
    if [ -f "$LOG_FILE" ] && [ $(stat -f%z "$LOG_FILE" 2>/dev/null || stat -c%s "$LOG_FILE" 2>/dev/null) -gt 10485760 ]; then
        print_info "Rotating large log file..."
        mv "$LOG_FILE" "${LOG_FILE}.old"
    fi
    
    # Start the service
    print_info "Command: $SERVICE_BIN runserver --debug=$DEBUG --headless=$HEADLESS --port=$SERVICE_PORT"
    
    nohup $SERVICE_BIN runserver \
        --debug=$DEBUG \
        --headless=$HEADLESS \
        --port=$SERVICE_PORT \
        > "$LOG_FILE" 2>&1 &
    
    local pid=$!
    echo $pid > "$PID_FILE"
    
    print_success "Service started (PID: $pid)"
    print_info "Log file: $LOG_FILE"
    
    # Wait for service to be ready
    if wait_for_service; then
        return 0
    else
        # Service failed to start, clean up
        stop_service
        return 1
    fi
}

# Function to stop the service
stop_service() {
    print_info "Stopping ${SERVICE_NAME} service..."
    
    if ! is_running; then
        print_warning "Service is not running"
        # Clean up stale PID file if exists
        rm -f "$PID_FILE"
        return 0
    fi
    
    local pid=$(get_pid)
    print_info "Sending SIGTERM to process $pid..."
    
    # Try graceful shutdown first
    kill -TERM "$pid" 2>/dev/null || true
    
    # Wait for process to stop
    local attempt=0
    local max_attempts=10
    
    while [ $attempt -lt $max_attempts ]; do
        if ! ps -p "$pid" > /dev/null 2>&1; then
            print_success "Service stopped gracefully"
            rm -f "$PID_FILE"
            return 0
        fi
        
        attempt=$((attempt + 1))
        echo -n "."
        sleep 1
    done
    
    echo ""
    print_warning "Service did not stop gracefully, forcing shutdown..."
    kill -KILL "$pid" 2>/dev/null || true
    rm -f "$PID_FILE"
    print_success "Service stopped forcefully"
}

# Function to restart the service
restart_service() {
    print_info "Restarting ${SERVICE_NAME} service..."
    stop_service
    sleep 2
    start_service
}

# Function to show service status
show_status() {
    echo "=========================================="
    echo "  ${SERVICE_NAME} Service Status"
    echo "=========================================="
    echo ""
    
    if is_running; then
        local pid=$(get_pid)
        print_success "Service is RUNNING"
        echo "  PID: $pid"
        echo "  Port: $SERVICE_PORT"
        echo "  Log: $LOG_FILE"
        echo ""
        
        # Show process info
        echo "Process info:"
        ps -p "$pid" -o pid,ppid,user,%cpu,%mem,etime,command | tail -n +2
        echo ""
        
        # Try to get service version/health
        if curl -s "http://${SERVICE_HOST}:${SERVICE_PORT}/health" > /dev/null 2>&1; then
            print_success "Health check: PASSED"
        else
            print_warning "Health check: FAILED (service may be starting)"
        fi
    else
        print_error "Service is NOT RUNNING"
        
        if [ -f "$LOG_FILE" ]; then
            echo ""
            echo "Last 10 lines of log:"
            echo "----------------------------------------"
            tail -n 10 "$LOG_FILE"
        fi
    fi
    
    echo ""
}

# Function to view logs
view_logs() {
    if [ ! -f "$LOG_FILE" ]; then
        print_warning "Log file not found: $LOG_FILE"
        return 1
    fi
    
    print_info "Viewing logs (Ctrl+C to exit)..."
    tail -f "$LOG_FILE"
}

# Function to test the service
test_service() {
    print_info "Testing ${SERVICE_NAME} service..."
    echo ""
    
    # Check if service is running
    if ! is_running; then
        print_error "Service is not running. Please start it first."
        return 1
    fi
    
    # Check if service is ready
    if ! curl -s "http://${SERVICE_HOST}:${SERVICE_PORT}/health" > /dev/null 2>&1; then
        print_error "Service health check failed. Service may not be ready."
        return 1
    fi
    
    print_success "Service is ready, sending test request..."
    echo ""
    echo "Test parameters:"
    echo "  Duration: ${TEST_DURATION} seconds"
    echo "  Scroll interval: ${TEST_INTERVAL} seconds"
    echo "  Save path: ${TEST_SAVE_PATH}"
    echo ""
    
    # Create save directory
    mkdir -p "$TEST_SAVE_PATH"
    
    # Send the request in background
    print_info "Sending crawl request (running in background)..."
    
    curl -X POST "http://${SERVICE_HOST}:${SERVICE_PORT}/api/sora/crawl" \
        -H "Content-Type: application/json" \
        -d "{
            \"total_duration_seconds\": ${TEST_DURATION},
            \"scroll_interval_seconds\": ${TEST_INTERVAL},
            \"save_path\": \"${TEST_SAVE_PATH}\"
        }" > "$CURL_LOG_FILE" 2>&1 &
    
    local curl_pid=$!
    
    print_success "Request sent (PID: $curl_pid)"
    print_info "Response will be saved to: $CURL_LOG_FILE"
    echo ""
    
    # Wait a moment for the request to be processed
    sleep 2
    
    print_info "You can monitor the progress with:"
    echo "  - Service logs: tail -f $LOG_FILE"
    echo "  - API response: tail -f $CURL_LOG_FILE"
    echo "  - Downloaded files: ls -lh $TEST_SAVE_PATH"
    echo ""
    
    # Ask if user wants to view logs
    print_info "Starting log viewer (Ctrl+C to exit)..."
    sleep 1
    tail -f "$LOG_FILE"
}

# Function to clean up
cleanup() {
    print_info "Cleaning up..."
    
    # Stop service if running
    if is_running; then
        stop_service
    fi
    
    # Remove log files
    if [ -f "$LOG_FILE" ]; then
        rm -f "$LOG_FILE"
        print_success "Removed log file: $LOG_FILE"
    fi
    
    if [ -f "$CURL_LOG_FILE" ]; then
        rm -f "$CURL_LOG_FILE"
        print_success "Removed curl log: $CURL_LOG_FILE"
    fi
    
    # Remove PID file
    if [ -f "$PID_FILE" ]; then
        rm -f "$PID_FILE"
        print_success "Removed PID file: $PID_FILE"
    fi
    
    print_success "Cleanup completed"
}

# Function to run feed sync (default feed command)
run_feed_sync() {
    print_info "Running feed sync (fetch + download)..."
    echo ""
    
    # Check binary
    check_binary
    
    # Create save directory
    mkdir -p "$FEED_SAVE_PATH"
    
    print_info "Feed sync parameters:"
    echo "  Save path: $FEED_SAVE_PATH"
    echo "  Database: $FEED_DB_PATH"
    echo "  Limit: $FEED_LIMIT"
    echo "  Headless: $HEADLESS"
    echo ""
    
    # Run the feed sync command
    print_info "Command: $SERVICE_BIN feed sync --save-path=$FEED_SAVE_PATH --db-path=$FEED_DB_PATH --limit=$FEED_LIMIT --headless=$HEADLESS"
    
    $SERVICE_BIN feed sync \
        --save-path="$FEED_SAVE_PATH" \
        --db-path="$FEED_DB_PATH" \
        --limit=$FEED_LIMIT \
        --headless=$HEADLESS
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_success "Feed sync completed successfully!"
        
        # Show downloaded files
        if [ -d "$FEED_SAVE_PATH" ]; then
            echo ""
            print_info "Downloaded files (latest 10):"
            ls -lt "$FEED_SAVE_PATH" | head -n 11
        fi
        
        # Show database info
        if [ -f "$FEED_DB_PATH" ]; then
            echo ""
            print_info "Database: $FEED_DB_PATH"
            ls -lh "$FEED_DB_PATH"
        fi
    else
        print_error "Feed sync failed with exit code: $exit_code"
        return $exit_code
    fi
}

# Function to run feed fetch (fetch only, save to file)
run_feed_fetch() {
    local output="${FEED_OUTPUT:-feed.json}"
    
    print_info "Fetching feed to file..."
    echo ""
    
    # Check binary
    check_binary
    
    print_info "Feed fetch parameters:"
    echo "  Output file: $output"
    echo "  Headless: $HEADLESS"
    echo ""
    
    # Run the feed fetch command
    print_info "Command: $SERVICE_BIN feed fetch --output=$output --headless=$HEADLESS"
    
    $SERVICE_BIN feed fetch \
        --output="$output" \
        --headless=$HEADLESS
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_success "Feed fetched and saved to: $output"
        
        if [ -f "$output" ]; then
            local size=$(ls -lh "$output" | awk '{print $5}')
            print_info "File size: $size"
        fi
    else
        print_error "Feed fetch failed with exit code: $exit_code"
        return $exit_code
    fi
}

# Function to run feed download (download from saved feed file)
run_feed_download() {
    local input="${FEED_INPUT:-feed.json}"
    
    print_info "Downloading videos from saved feed file..."
    echo ""
    
    # Check binary
    check_binary
    
    # Check if input file exists
    if [ ! -f "$input" ]; then
        print_error "Input feed file not found: $input"
        return 1
    fi
    
    # Create save directory
    mkdir -p "$FEED_SAVE_PATH"
    
    print_info "Feed download parameters:"
    echo "  Input file: $input"
    echo "  Save path: $FEED_SAVE_PATH"
    echo "  Database: $FEED_DB_PATH"
    echo "  Limit: $FEED_LIMIT"
    echo ""
    
    # Run the feed download command
    print_info "Command: $SERVICE_BIN feed download --input=$input --save-path=$FEED_SAVE_PATH --db-path=$FEED_DB_PATH --limit=$FEED_LIMIT"
    
    $SERVICE_BIN feed download \
        --input="$input" \
        --save-path="$FEED_SAVE_PATH" \
        --db-path="$FEED_DB_PATH" \
        --limit=$FEED_LIMIT
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_success "Feed download completed successfully!"
        
        # Show downloaded files
        if [ -d "$FEED_SAVE_PATH" ]; then
            echo ""
            print_info "Downloaded files (latest 10):"
            ls -lt "$FEED_SAVE_PATH" | head -n 11
        fi
    else
        print_error "Feed download failed with exit code: $exit_code"
        return $exit_code
    fi
}

# Function to run feed export (export database to JSON)
run_feed_export() {
    local output="${FEED_OUTPUT:-}"
    local limit="${FEED_EXPORT_LIMIT:-0}"
    
    print_info "Exporting videos from database..."
    echo ""
    
    # Check binary
    check_binary
    
    # Check if database exists
    if [ ! -f "$FEED_DB_PATH" ]; then
        print_error "Database file not found: $FEED_DB_PATH"
        return 1
    fi
    
    print_info "Feed export parameters:"
    echo "  Database: $FEED_DB_PATH"
    echo "  Output: ${output:-stdout}"
    echo "  Limit: ${limit} (0 = all)"
    echo ""
    
    # Build command
    local cmd="$SERVICE_BIN feed export --db-path=$FEED_DB_PATH --limit=$limit"
    
    if [ -n "$output" ]; then
        cmd="$cmd --output=$output"
        print_info "Command: $cmd"
        
        $SERVICE_BIN feed export \
            --db-path="$FEED_DB_PATH" \
            --limit=$limit \
            --output="$output"
        
        local exit_code=$?
        
        if [ $exit_code -eq 0 ]; then
            print_success "Videos exported to: $output"
            
            if [ -f "$output" ]; then
                local size=$(ls -lh "$output" | awk '{print $5}')
                print_info "File size: $size"
            fi
        else
            print_error "Feed export failed with exit code: $exit_code"
            return $exit_code
        fi
    else
        print_info "Command: $cmd (output to stdout)"
        
        $SERVICE_BIN feed export \
            --db-path="$FEED_DB_PATH" \
            --limit=$limit
    fi
}

# Function to run Goldcast upload
run_goldcast_upload() {
    print_info "Uploading videos to Goldcast..."
    echo ""
    
    # Check binary
    check_binary
    
    # Check if database exists
    if [ ! -f "$FEED_DB_PATH" ]; then
        print_error "Database file not found: $FEED_DB_PATH"
        print_info "Please run feed-sync first to download videos"
        return 1
    fi
    
    # Check OSS credentials (required)
    if [ -z "$OSS_ACCESS_KEY_ID" ] || [ -z "$OSS_ACCESS_KEY_SECRET" ]; then
        print_error "OSS credentials are required!"
        echo ""
        print_info "Please provide OSS credentials via:"
        echo "  1. Command-line: --oss-access-key-id=xxx --oss-access-key-secret=xxx"
        echo "  2. Environment: export OSS_ACCESS_KEY_ID=xxx && export OSS_ACCESS_KEY_SECRET=xxx"
        echo ""
        print_info "Optional (have defaults):"
        echo "  --oss-bucket-name=$OSS_BUCKET_NAME"
        echo "  --oss-endpoint=$OSS_ENDPOINT"
        echo "  --oss-region=$OSS_REGION"
        return 1
    fi
    
    print_info "Goldcast upload parameters:"
    echo "  Database: $FEED_DB_PATH"
    echo "  API URL: $GOLDCAST_API_URL"
    echo "  Limit: $GOLDCAST_LIMIT (0 = all)"
    echo "  OSS Bucket: $OSS_BUCKET_NAME"
    echo "  OSS Endpoint: $OSS_ENDPOINT"
    echo ""
    
    # Build command
    local cmd="$SERVICE_BIN feed uploadgoldcast --db-path=$FEED_DB_PATH --limit=$GOLDCAST_LIMIT"
    cmd="$cmd --oss-access-key-id=$OSS_ACCESS_KEY_ID --oss-access-key-secret=***"
    cmd="$cmd --oss-bucket-name=$OSS_BUCKET_NAME --oss-endpoint=$OSS_ENDPOINT --oss-region=$OSS_REGION"
    
    # Add API key if provided
    if [ -n "$GOLDCAST_API_KEY" ]; then
        cmd="$cmd --api-key=$GOLDCAST_API_KEY"
    fi
    
    # Add API URL if provided
    if [ -n "$GOLDCAST_API_URL" ]; then
        cmd="$cmd --api-url=$GOLDCAST_API_URL"
    fi
    
    print_info "Command: $cmd"
    
    # Run the upload command
    $SERVICE_BIN feed uploadgoldcast \
        --db-path="$FEED_DB_PATH" \
        --limit=$GOLDCAST_LIMIT \
        --api-key="$GOLDCAST_API_KEY" \
        --api-url="$GOLDCAST_API_URL" \
        --oss-access-key-id="$OSS_ACCESS_KEY_ID" \
        --oss-access-key-secret="$OSS_ACCESS_KEY_SECRET" \
        --oss-bucket-name="$OSS_BUCKET_NAME" \
        --oss-endpoint="$OSS_ENDPOINT" \
        --oss-region="$OSS_REGION"
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_success "Goldcast upload completed!"
    else
        print_error "Goldcast upload failed with exit code: $exit_code"
        return $exit_code
    fi
}

# Function to build the service
build_service() {
    print_info "Building ${SERVICE_NAME} service..."
    
    # Check if go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        print_info "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    
    # Display Go version
    local go_version=$(go version)
    print_info "Using: $go_version"
    
    # Build the binary
    print_info "Running: go build -o $SERVICE_BIN ."
    
    if go build -o "$SERVICE_BIN" .; then
        print_success "Build completed successfully!"
        print_info "Binary: $SERVICE_BIN"
        
        # Make it executable
        chmod +x "$SERVICE_BIN"
        
        # Show binary info
        if [ -f "$SERVICE_BIN" ]; then
            local size=$(ls -lh "$SERVICE_BIN" | awk '{print $5}')
            print_info "Binary size: $size"
        fi
    else
        print_error "Build failed!"
        return 1
    fi
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build                Build the service binary
    start                Start the service
    stop                 Stop the service
    restart              Restart the service
    status               Show service status
    logs                 View service logs (tail -f)
    test                 Start service and run a test crawl
    feed-sync            Fetch feed and download videos (default feed behavior)
    feed-fetch           Fetch feed and save to file (no downloads)
    feed-download        Download videos from saved feed file
    feed-export          Export downloaded videos as JSON
    feed-uploadgoldcast  Upload videos to Goldcast media CMS
    cleanup              Stop service and remove log files
    help                 Show this help message

Options:
    --headless=<true|false>     Run browser in headless mode (default: true)
    --debug=<true|false>        Enable debug logging (default: true)
    --port=<port>               Service port (default: 8080)
    --duration=<seconds>        Test crawl duration (default: 60)
    --interval=<seconds>        Test scroll interval (default: 10)
    --limit=<number>            Feed download limit (default: 50)
    --db-path=<path>            Database path for feed (default: ./sora.db)
    --save-path=<path>          Save path for downloads (default: ./downloads/sora)
    --input=<path>              Input feed file for feed-download (default: feed.json)
    --output=<path>             Output file for feed-fetch/feed-export
    --export-limit=<number>     Limit for feed-export (default: 0 = all)
    --api-key=<key>             Goldcast API key (env: GOLDCAST_API_KEY)
    --api-url=<url>             Goldcast API URL (env: GOLDCAST_API_URL)
    --upload-limit=<number>     Upload limit for Goldcast (default: 0 = all)
    --oss-access-key-id=<key>   OSS Access Key ID (env: OSS_ACCESS_KEY_ID, required)
    --oss-access-key-secret=<s> OSS Access Key Secret (env: OSS_ACCESS_KEY_SECRET, required)
    --oss-bucket-name=<name>    OSS Bucket Name (env: OSS_BUCKET_NAME, default: dreammedias)
    --oss-endpoint=<endpoint>   OSS Endpoint (env: OSS_ENDPOINT, default: oss-cn-beijing.aliyuncs.com)
    --oss-region=<region>       OSS Region (env: OSS_REGION, default: cn-beijing)

Examples:
    # Build the service binary
    $0 build

    # Start the service
    $0 start

    # Start with custom port
    $0 start --port=9090

    # Start in non-headless mode for debugging
    $0 start --headless=false

    # Check service status
    $0 status

    # Run a test crawl
    $0 test

    # Run a longer test (5 minutes)
    $0 test --duration=300 --interval=15

    # View logs
    $0 logs

    # Stop the service
    $0 stop

    # Feed commands:
    
    # Sync (fetch + download, recommended for regular use)
    $0 feed-sync
    $0 feed-sync --limit=100

    # Fetch feed only (save for later inspection/debugging)
    $0 feed-fetch --output=my_feed.json

    # Download from saved feed file (useful for testing)
    $0 feed-download --input=my_feed.json --limit=10

    # Export database videos as JSON
    $0 feed-export --output=videos.json --export-limit=50
    $0 feed-export --export-limit=100  # Output to stdout

    # Upload videos to Goldcast (OSS credentials required)
    $0 feed-uploadgoldcast --oss-access-key-id=YOUR_KEY_ID --oss-access-key-secret=YOUR_SECRET
    $0 feed-uploadgoldcast --upload-limit=10 --oss-access-key-id=KEY --oss-access-key-secret=SECRET
    
    # Or use environment variables
    export OSS_ACCESS_KEY_ID=YOUR_KEY_ID
    export OSS_ACCESS_KEY_SECRET=YOUR_SECRET
    $0 feed-uploadgoldcast

    # Run feed sync without headless mode (for debugging)
    $0 feed-sync --headless=false --limit=10

    # Clean up everything
    $0 cleanup

EOF
}

# Parse command line arguments
COMMAND="${1:-help}"
shift || true

while [ $# -gt 0 ]; do
    case "$1" in
        --headless=*)
            HEADLESS="${1#*=}"
            ;;
        --debug=*)
            DEBUG="${1#*=}"
            ;;
        --port=*)
            SERVICE_PORT="${1#*=}"
            ;;
        --duration=*)
            TEST_DURATION="${1#*=}"
            ;;
        --interval=*)
            TEST_INTERVAL="${1#*=}"
            ;;
        --limit=*)
            FEED_LIMIT="${1#*=}"
            ;;
        --db-path=*)
            FEED_DB_PATH="${1#*=}"
            ;;
        --save-path=*)
            FEED_SAVE_PATH="${1#*=}"
            ;;
        --input=*)
            FEED_INPUT="${1#*=}"
            ;;
        --output=*)
            FEED_OUTPUT="${1#*=}"
            ;;
        --export-limit=*)
            FEED_EXPORT_LIMIT="${1#*=}"
            ;;
        --api-key=*)
            GOLDCAST_API_KEY="${1#*=}"
            ;;
        --api-url=*)
            GOLDCAST_API_URL="${1#*=}"
            ;;
        --upload-limit=*)
            GOLDCAST_LIMIT="${1#*=}"
            ;;
        --oss-access-key-id=*)
            OSS_ACCESS_KEY_ID="${1#*=}"
            ;;
        --oss-access-key-secret=*)
            OSS_ACCESS_KEY_SECRET="${1#*=}"
            ;;
        --oss-bucket-name=*)
            OSS_BUCKET_NAME="${1#*=}"
            ;;
        --oss-endpoint=*)
            OSS_ENDPOINT="${1#*=}"
            ;;
        --oss-region=*)
            OSS_REGION="${1#*=}"
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
    shift
done

# Execute command
case "$COMMAND" in
    build)
        build_service
        ;;
    start)
        start_service
        ;;
    stop)
        stop_service
        ;;
    restart)
        restart_service
        ;;
    status)
        show_status
        ;;
    logs)
        view_logs
        ;;
    test)
        # Start service if not running
        if ! is_running; then
            start_service || exit 1
        fi
        test_service
        ;;
    feed|feed-sync)
        run_feed_sync
        ;;
    feed-fetch)
        run_feed_fetch
        ;;
    feed-download)
        run_feed_download
        ;;
    feed-export)
        run_feed_export
        ;;
    feed-uploadgoldcast)
        run_goldcast_upload
        ;;
    cleanup)
        cleanup
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        echo ""
        show_usage
        exit 1
        ;;
esac

