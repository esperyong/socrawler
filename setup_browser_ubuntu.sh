#!/bin/bash

# Setup script for installing Chrome browser on Ubuntu server for headless operation
# This script installs Chrome and all required dependencies for go-rod

set -e

echo "=========================================="
echo "Installing Chrome for Headless Operation"
echo "=========================================="
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if running as root
#if [ "$EUID" -eq 0 ]; then
#    echo -e "${RED}Please do not run this script as root. Use sudo when prompted.${NC}"
#    exit 1
#fi

echo -e "${YELLOW}Step 1: Updating package list...${NC}"
sudo apt-get update

echo ""
echo -e "${YELLOW}Step 2: Installing required dependencies...${NC}"
# Install required libraries for Chrome
sudo apt-get install -y \
    wget \
    gnupg \
    ca-certificates \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libdbus-1-3 \
    libgdk-pixbuf2.0-0 \
    libnspr4 \
    libnss3 \
    libx11-xcb1 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    xdg-utils \
    libgbm1 \
    libxkbcommon0 \
    libpango-1.0-0 \
    libcairo2

echo ""
echo -e "${YELLOW}Step 3: Adding Google Chrome repository...${NC}"
# Add Google's signing key
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -

# Add Chrome repository
if [ ! -f /etc/apt/sources.list.d/google-chrome.list ]; then
    echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | sudo tee /etc/apt/sources.list.d/google-chrome.list
fi

echo ""
echo -e "${YELLOW}Step 4: Installing Google Chrome (stable)...${NC}"
sudo apt-get update
sudo apt-get install -y google-chrome-stable

echo ""
echo -e "${YELLOW}Step 5: Verifying Chrome installation...${NC}"
if command -v google-chrome &> /dev/null; then
    CHROME_VERSION=$(google-chrome --version)
    echo -e "${GREEN}✓ Chrome installed successfully: ${CHROME_VERSION}${NC}"
else
    echo -e "${RED}✗ Chrome installation failed${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 6: Testing headless mode...${NC}"
# Test Chrome in headless mode
if google-chrome --headless --disable-gpu --dump-dom https://www.google.com > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Chrome headless mode is working${NC}"
else
    echo -e "${RED}✗ Chrome headless mode test failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}=========================================="
echo "Installation Complete!"
echo "==========================================${NC}"
echo ""
echo "Chrome is now ready for headless operation."
echo "You can now run your socrawler service with:"
echo ""
echo "  ./socrawler runserver --headless=true"
echo ""
echo "Chrome location: $(which google-chrome)"
echo "Chrome version: $(google-chrome --version)"

