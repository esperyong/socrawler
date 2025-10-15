#!/bin/bash

# Test script to verify deduplication functionality
# This will run the crawler twice and check that duplicates are not re-downloaded

set -e

echo "Testing Sora Crawler Deduplication"
echo "===================================="
echo ""

# Clean up previous test results (optional - comment out to test with existing files)
# rm -rf downloads/sora/*
# echo "Cleaned up previous downloads"
# echo ""

# First crawl - should download new files
echo "First crawl (30 seconds)..."
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 30,
    "scroll_interval_seconds": 5,
    "save_path": "./downloads/sora"
  }' 2>/dev/null | jq .

echo ""
echo "Waiting 5 seconds before second crawl..."
sleep 5

# Second crawl - should skip duplicates
echo ""
echo "Second crawl (30 seconds)..."
echo "Watch for 'File already exists, skipping download' messages in the logs"
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 30,
    "scroll_interval_seconds": 5,
    "save_path": "./downloads/sora"
  }' 2>/dev/null | jq .

echo ""
echo "Test completed!"
echo ""
echo "Check the file structure:"
echo "ls -la downloads/sora/ | head -20"
ls -la downloads/sora/ | head -20

echo ""
echo "Sample folder structure (first folder):"
FIRST_FOLDER=$(ls downloads/sora/ | grep -v debug | head -1)
if [ -n "$FIRST_FOLDER" ]; then
    echo "downloads/sora/$FIRST_FOLDER:"
    ls -la downloads/sora/$FIRST_FOLDER/
fi

