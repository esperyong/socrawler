#!/bin/bash

# Chrome diagnostic script
# This script tests if Chrome can run in headless mode with various configurations

set +e  # Don't exit on error

echo "=========================================="
echo "Chrome Headless Diagnostic Tool"
echo "=========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if Chrome is installed
echo -e "${YELLOW}1. Checking Chrome installation...${NC}"
if command -v google-chrome &> /dev/null; then
    CHROME_VERSION=$(google-chrome --version)
    echo -e "${GREEN}✓ Chrome found: ${CHROME_VERSION}${NC}"
else
    echo -e "${RED}✗ Chrome not found${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}2. Checking Chrome binary location...${NC}"
CHROME_BIN=$(which google-chrome)
echo "Chrome binary: $CHROME_BIN"

echo ""
echo -e "${YELLOW}3. Checking required libraries...${NC}"
MISSING_LIBS=$(ldd "$CHROME_BIN" 2>/dev/null | grep "not found" || true)
if [ -z "$MISSING_LIBS" ]; then
    echo -e "${GREEN}✓ All required libraries are present${NC}"
else
    echo -e "${RED}✗ Missing libraries:${NC}"
    echo "$MISSING_LIBS"
fi

echo ""
echo -e "${YELLOW}4. Testing Chrome with different configurations...${NC}"
echo ""

# Test 1: Basic headless mode
echo "Test 1: Basic headless mode"
if timeout 10 google-chrome --headless --disable-gpu --dump-dom https://www.google.com > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Basic headless mode works${NC}"
else
    echo -e "${RED}✗ Basic headless mode failed${NC}"
fi

echo ""
# Test 2: Headless mode with no-sandbox
echo "Test 2: Headless mode with --no-sandbox"
if timeout 10 google-chrome --headless --disable-gpu --no-sandbox --dump-dom https://www.google.com > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Headless mode with --no-sandbox works${NC}"
else
    echo -e "${RED}✗ Headless mode with --no-sandbox failed${NC}"
fi

echo ""
# Test 3: New headless mode (Chrome 109+)
echo "Test 3: New headless mode (--headless=new)"
if timeout 10 google-chrome --headless=new --disable-gpu --no-sandbox --dump-dom https://www.google.com > /dev/null 2>&1; then
    echo -e "${GREEN}✓ New headless mode works${NC}"
else
    echo -e "${RED}✗ New headless mode failed${NC}"
fi

echo ""
# Test 4: With all recommended flags
echo "Test 4: With all recommended flags for server environment"
if timeout 10 google-chrome \
    --headless=new \
    --disable-gpu \
    --no-sandbox \
    --disable-dev-shm-usage \
    --disable-software-rasterizer \
    --disable-extensions \
    --disable-setuid-sandbox \
    --dump-dom https://www.google.com > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Full configuration works${NC}"
else
    echo -e "${RED}✗ Full configuration failed${NC}"
fi

echo ""
echo -e "${YELLOW}5. Environment information...${NC}"
echo "User: $(whoami)"
echo "UID: $(id -u)"
echo "Display: ${DISPLAY:-not set}"
echo "Temp dir: ${TMPDIR:-/tmp}"
echo "Shared memory: $(df -h /dev/shm 2>/dev/null | tail -1 | awk '{print $2}' || echo 'not available')"

echo ""
echo -e "${YELLOW}6. Testing with verbose output...${NC}"
echo "Running Chrome with error output (this will show any errors):"
echo ""
timeout 10 google-chrome \
    --headless=new \
    --disable-gpu \
    --no-sandbox \
    --disable-dev-shm-usage \
    --dump-dom https://www.google.com 2>&1 | head -20

echo ""
echo "=========================================="
echo "Diagnostic Complete"
echo "=========================================="
echo ""
echo -e "${YELLOW}Recommendations:${NC}"
echo ""
echo "1. If all tests failed:"
echo "   - Check if you're running as root (not recommended)"
echo "   - Verify all dependencies are installed"
echo "   - Check system logs: journalctl -xe"
echo ""
echo "2. If only basic mode failed but --no-sandbox works:"
echo "   - This is normal for containers and root users"
echo "   - go-rod will automatically use --no-sandbox when needed"
echo ""
echo "3. For go-rod usage:"
echo "   - go-rod handles Chrome launch automatically"
echo "   - It will detect and use appropriate flags"
echo "   - No manual Chrome configuration needed"
echo ""
echo "4. Test your socrawler application:"
echo "   ./socrawler runserver --debug --headless=true"

