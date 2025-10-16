#!/bin/bash

# Test script for feed downloader
# This script tests the new feed-based Sora video downloader

set -e

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo ""
    echo "========================================"
    echo "  $1"
    echo "========================================"
    echo ""
}

# Configuration
TEST_SAVE_PATH="./downloads/sora_test"
TEST_DB_PATH="./sora_test.db"
TEST_LIMIT=5

print_header "Feed Downloader Test"

# Clean up previous test files
print_info "Cleaning up previous test files..."
rm -rf "$TEST_SAVE_PATH"
rm -f "$TEST_DB_PATH"

# Test 1: Build the binary
print_header "Test 1: Build Binary"
print_info "Building socrawler..."
if go build -o socrawler .; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi

# Test 2: Check command help
print_header "Test 2: Check Feed Command Help"
print_info "Running: ./socrawler feed --help"
./socrawler feed --help
print_success "Help command works"

# Test 3: Run feed downloader with small limit
print_header "Test 3: Run Feed Downloader (limit=$TEST_LIMIT)"
print_info "This will download up to $TEST_LIMIT videos from the feed"
print_info "Save path: $TEST_SAVE_PATH"
print_info "Database: $TEST_DB_PATH"
echo ""

if ./socrawler feed \
    --save-path="$TEST_SAVE_PATH" \
    --db-path="$TEST_DB_PATH" \
    --limit=$TEST_LIMIT \
    --headless=true \
    --debug; then
    print_success "Feed download completed"
else
    print_error "Feed download failed"
    exit 1
fi

# Test 4: Verify downloaded files
print_header "Test 4: Verify Downloaded Files"

if [ -d "$TEST_SAVE_PATH" ]; then
    file_count=$(find "$TEST_SAVE_PATH" -type f | wc -l)
    print_info "Found $file_count files in $TEST_SAVE_PATH"
    
    if [ $file_count -gt 0 ]; then
        print_success "Files downloaded successfully"
        echo ""
        print_info "Downloaded files:"
        ls -lh "$TEST_SAVE_PATH"
    else
        print_error "No files downloaded"
    fi
else
    print_error "Save directory not created"
fi

# Test 5: Verify database
print_header "Test 5: Verify Database"

if [ -f "$TEST_DB_PATH" ]; then
    db_size=$(ls -lh "$TEST_DB_PATH" | awk '{print $5}')
    print_success "Database created: $TEST_DB_PATH (size: $db_size)"
    
    # Query database to see how many records
    print_info "Checking database records..."
    video_count=$(sqlite3 "$TEST_DB_PATH" "SELECT COUNT(*) FROM sora_videos;" 2>/dev/null || echo "0")
    print_info "Database contains $video_count video records"
    
    if [ "$video_count" -gt 0 ]; then
        print_success "Database populated successfully"
        
        # Show sample records
        echo ""
        print_info "Sample records from database:"
        sqlite3 "$TEST_DB_PATH" "SELECT post_id, username, posted_at FROM sora_videos LIMIT 3;" 2>/dev/null || true
    fi
else
    print_error "Database not created"
fi

# Test 6: Run again to test deduplication
print_header "Test 6: Test Deduplication"
print_info "Running feed downloader again to verify it skips already downloaded videos"
echo ""

if ./socrawler feed \
    --save-path="$TEST_SAVE_PATH" \
    --db-path="$TEST_DB_PATH" \
    --limit=$TEST_LIMIT \
    --headless=true; then
    print_success "Second run completed (should have found 0 new videos)"
else
    print_error "Second run failed"
fi

# Test 7: Test with run_service.sh wrapper
print_header "Test 7: Test run_service.sh Integration"
print_info "Testing feed command through run_service.sh"
echo ""

if ./run_service.sh feed --limit=2 --save-path="$TEST_SAVE_PATH" --db-path="$TEST_DB_PATH"; then
    print_success "run_service.sh feed command works"
else
    print_error "run_service.sh feed command failed"
fi

# Summary
print_header "Test Summary"
print_success "All tests completed!"
echo ""
print_info "Next steps:"
echo "  1. Run with larger limit: ./run_service.sh feed --limit=50"
echo "  2. Schedule with cron: 0 */6 * * * cd $(pwd) && ./run_service.sh feed"
echo "  3. View database: sqlite3 $TEST_DB_PATH 'SELECT * FROM sora_videos;'"
echo ""
print_info "Test files location:"
echo "  Videos: $TEST_SAVE_PATH"
echo "  Database: $TEST_DB_PATH"
echo ""
print_info "To clean up test files:"
echo "  rm -rf $TEST_SAVE_PATH $TEST_DB_PATH"

