#!/bin/bash

# Test script for Sora video crawler API
# This script sends a test request to crawl Sora videos

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Sora Video Crawler API${NC}"
echo "========================================"
echo ""

# Test parameters - short duration for testing
DURATION=60  # 60 seconds (1 minute)
INTERVAL=10  # Scroll every 10 seconds
SAVE_PATH="./downloads/sora"

echo "Test parameters:"
echo "  Duration: ${DURATION} seconds"
echo "  Scroll interval: ${INTERVAL} seconds"
echo "  Save path: ${SAVE_PATH}"
echo ""

# Send the request
echo -e "${YELLOW}Sending request to http://localhost:8080/api/sora/crawl...${NC}"
echo ""

curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d "{
    \"total_duration_seconds\": ${DURATION},
    \"scroll_interval_seconds\": ${INTERVAL},
    \"save_path\": \"${SAVE_PATH}\"
  }" | python3 -m json.tool

echo ""
echo -e "${GREEN}Test completed!${NC}"
echo ""
echo "Check the ${SAVE_PATH} directory for downloaded files."

