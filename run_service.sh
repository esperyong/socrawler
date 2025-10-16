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

# Function to run feed downloader
run_feed() {
    print_info "Running feed downloader..."
    echo ""
    
    # Check binary
    check_binary
    
    # Create save directory
    mkdir -p "$FEED_SAVE_PATH"
    
    print_info "Feed parameters:"
    echo "  Save path: $FEED_SAVE_PATH"
    echo "  Database: $FEED_DB_PATH"
    echo "  Limit: $FEED_LIMIT"
    echo "  Headless: $HEADLESS"
    echo ""
    
    # Run the feed command
    print_info "Command: $SERVICE_BIN feed --save-path=$FEED_SAVE_PATH --db-path=$FEED_DB_PATH --limit=$FEED_LIMIT --headless=$HEADLESS"
    
    $SERVICE_BIN feed \
        --save-path="$FEED_SAVE_PATH" \
        --db-path="$FEED_DB_PATH" \
        --limit=$FEED_LIMIT \
        --headless=$HEADLESS
    
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_success "Feed download completed successfully!"
        
        # Show downloaded files
        if [ -d "$FEED_SAVE_PATH" ]; then
            echo ""
            print_info "Downloaded files:"
            ls -lh "$FEED_SAVE_PATH" | tail -n 10
        fi
        
        # Show database info
        if [ -f "$FEED_DB_PATH" ]; then
            echo ""
            print_info "Database: $FEED_DB_PATH"
            ls -lh "$FEED_DB_PATH"
        fi
    else
        print_error "Feed download failed with exit code: $exit_code"
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
    build       Build the service binary
    start       Start the service
    stop        Stop the service
    restart     Restart the service
    status      Show service status
    logs        View service logs (tail -f)
    test        Start service and run a test crawl
    feed        Run feed downloader to get new videos
    cleanup     Stop service and remove log files
    help        Show this help message

Options:
    --headless=<true|false>     Run browser in headless mode (default: true)
    --debug=<true|false>        Enable debug logging (default: true)
    --port=<port>               Service port (default: 8080)
    --duration=<seconds>        Test crawl duration (default: 60)
    --interval=<seconds>        Test scroll interval (default: 10)
    --limit=<number>            Feed download limit (default: 50)
    --db-path=<path>            Database path for feed (default: ./sora.db)
    --save-path=<path>          Save path for downloads (default: ./downloads/sora)

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

    # Run feed downloader
    $0 feed

    # Run feed with custom limit
    $0 feed --limit=100

    # Run feed without headless mode (for debugging)
    $0 feed --headless=false --limit=10

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
    feed)
        run_feed
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

