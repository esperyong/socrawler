# Deployment Guide for Socrawler

This guide covers deploying socrawler on Ubuntu servers with headless Chrome support.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Method 1: Direct Installation on Ubuntu](#method-1-direct-installation-on-ubuntu)
- [Method 2: Docker Deployment](#method-2-docker-deployment)
- [Method 3: Docker Compose](#method-3-docker-compose)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### System Requirements

- **OS**: Ubuntu 20.04 LTS or later (recommended: Ubuntu 24.04 LTS)
- **RAM**: Minimum 2GB, recommended 4GB+
- **CPU**: 2+ cores recommended
- **Disk**: At least 2GB free space for Chrome + storage for videos

**Note for Ubuntu 24.04 users:** Package names have changed (see [UBUNTU_24.04_NOTES.md](UBUNTU_24.04_NOTES.md)). Our scripts handle this automatically.

### Software Requirements

- Go 1.23.0 or higher (for non-Docker deployment)
- Docker and Docker Compose (for Docker deployment)

---

## Method 1: Direct Installation on Ubuntu

This method installs Chrome and runs the service directly on Ubuntu.

### Step 1: Install Chrome and Dependencies

We provide an automated script to install Chrome:

```bash
# Clone the repository (if not already done)
git clone https://github.com/esperyong/socrawler.git
cd socrawler

# Make the setup script executable (if not already)
chmod +x setup_browser_ubuntu.sh

# Run the setup script
./setup_browser_ubuntu.sh
```

**What the script does:**
- Updates system packages
- Installs all required libraries for Chrome
- Adds Google Chrome repository
- Installs Google Chrome Stable
- Verifies the installation
- Tests headless mode

**Manual Installation (alternative):**

If you prefer manual installation, see [UBUNTU_24.04_NOTES.md](UBUNTU_24.04_NOTES.md) for version-specific instructions.

**For Ubuntu 24.04:**
```bash
# Update packages
sudo apt-get update

# Install dependencies (note: t64 suffix for Ubuntu 24.04)
sudo apt-get install -y \
    wget gnupg ca-certificates \
    fonts-liberation libappindicator3-1 \
    libasound2t64 libatk-bridge2.0-0t64 libatk1.0-0t64 \
    libcups2t64 libdbus-1-3 libgdk-pixbuf2.0-0 \
    libnspr4 libnss3 libx11-xcb1 libxcomposite1 \
    libxdamage1 libxrandr2 xdg-utils libgbm1 \
    libxkbcommon0 libpango-1.0-0 libcairo2

# Add Google Chrome repository (new method)
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | \
    sudo gpg --dearmor -o /usr/share/keyrings/google-chrome-keyring.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome-keyring.gpg] http://dl.google.com/linux/chrome/deb/ stable main" | \
    sudo tee /etc/apt/sources.list.d/google-chrome.list

# Install Chrome
sudo apt-get update
sudo apt-get install -y google-chrome-stable

# Verify installation
google-chrome --version
```

**For Ubuntu 22.04 and earlier:**
```bash
# Use package names without t64 suffix
sudo apt-get install -y \
    wget gnupg ca-certificates \
    fonts-liberation libappindicator3-1 \
    libasound2 libatk-bridge2.0-0 libatk1.0-0 \
    libcups2 libdbus-1-3 libgdk-pixbuf2.0-0 \
    libnspr4 libnss3 libx11-xcb1 libxcomposite1 \
    libxdamage1 libxrandr2 xdg-utils libgbm1 \
    libxkbcommon0 libpango-1.0-0 libcairo2

# Add Chrome repository
wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | sudo apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" | \
    sudo tee /etc/apt/sources.list.d/google-chrome.list

# Install Chrome
sudo apt-get update
sudo apt-get install -y google-chrome-stable
```

### Step 2: Install Go (if not installed)

```bash
# Download Go 1.23.1
wget https://go.dev/dl/go1.23.1.linux-amd64.tar.gz

# Remove old Go installation
sudo rm -rf /usr/local/go

# Extract new version
sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz

# Add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile

# Verify
go version
```

### Step 3: Build and Run Socrawler

```bash
# Build the application
go build -o socrawler .

# Create downloads directory
mkdir -p downloads/sora

# Run the service
./socrawler runserver --headless=true --port 8080
```

### Step 4: Run as a System Service (Optional)

Create a systemd service for automatic startup:

```bash
sudo nano /etc/systemd/system/socrawler.service
```

Add the following content:

```ini
[Unit]
Description=Socrawler - Sora Video Crawler Service
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/path/to/socrawler
ExecStart=/path/to/socrawler/socrawler runserver --headless=true --port 8080
Restart=on-failure
RestartSec=5s

# Environment
Environment="PATH=/usr/local/go/bin:/usr/bin:/bin"
Environment="CHROME_BIN=/usr/bin/google-chrome"

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=socrawler

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable socrawler

# Start the service
sudo systemctl start socrawler

# Check status
sudo systemctl status socrawler

# View logs
sudo journalctl -u socrawler -f
```

---

## Method 2: Docker Deployment

Using Docker is the easiest way to deploy with all dependencies included.

### Step 1: Install Docker

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group (optional, to run without sudo)
sudo usermod -aG docker $USER
newgrp docker

# Verify installation
docker --version
```

### Step 2: Build and Run Container

```bash
# Build the Docker image
docker build -t socrawler:latest .

# Run the container
docker run -d \
  --name socrawler \
  -p 8080:8080 \
  -v $(pwd)/downloads:/app/downloads \
  --shm-size=2gb \
  --security-opt seccomp=unconfined \
  socrawler:latest

# Check logs
docker logs -f socrawler

# Stop container
docker stop socrawler

# Start container
docker start socrawler

# Remove container
docker rm -f socrawler
```

---

## Method 3: Docker Compose

The simplest method for production deployment.

### Step 1: Install Docker Compose

```bash
# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify
docker-compose --version
```

### Step 2: Deploy with Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

---

## Verification

After deployment, verify the service is running:

### 1. Check Health Endpoint

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "socrawler",
  "version": "1.0.0"
}
```

### 2. Test Crawl Endpoint (Short Test)

```bash
curl -X POST http://localhost:8080/api/sora/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "total_duration_seconds": 30,
    "scroll_interval_seconds": 10,
    "save_path": "./downloads/sora"
  }'
```

### 3. Check Downloaded Files

```bash
ls -lah downloads/sora/
```

---

## Troubleshooting

### Chrome Not Found

**Error:** `chrome executable not found`

**Solution:**
```bash
# Check if Chrome is installed
which google-chrome

# If not found, run the setup script again
./setup_browser_ubuntu.sh
```

### Insufficient Shared Memory

**Error:** `Failed to create shared memory`

**Solution for Docker:**
```bash
# Add --shm-size when running container
docker run --shm-size=2gb ...
```

**Solution for Direct Installation:**
```bash
# Check available memory
free -h

# Increase if needed (requires restart)
sudo sysctl -w kernel.shmmax=2147483648
```

### Chrome Crashes

**Error:** `Chrome process crashed`

**Solutions:**

1. **Disable GPU acceleration** (already done in headless mode)
2. **Increase memory limits**
3. **Check for missing dependencies:**
   ```bash
   ldd /usr/bin/google-chrome | grep "not found"
   ```
4. **Run Chrome diagnostics:**
   ```bash
   google-chrome --headless --disable-gpu --dump-dom https://www.google.com
   ```

### Port Already in Use

**Error:** `bind: address already in use`

**Solution:**
```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill the process
sudo kill -9 <PID>

# Or use a different port
./socrawler runserver --port 8081
```

### Permission Denied for Downloads

**Error:** `permission denied: ./downloads/sora`

**Solution:**
```bash
# Create directory with proper permissions
mkdir -p downloads/sora
chmod 755 downloads/sora

# For Docker, ensure volume permissions
sudo chown -R $USER:$USER downloads/
```

### Network Issues

**Error:** `Failed to navigate to Sora page`

**Solutions:**

1. **Check internet connectivity:**
   ```bash
   ping -c 3 google.com
   ```

2. **Check if Sora is accessible:**
   ```bash
   curl -I https://sora.chatgpt.com/
   ```

3. **Check DNS resolution:**
   ```bash
   nslookup sora.chatgpt.com
   ```

### High Memory Usage

Chrome can consume significant memory, especially during long crawls.

**Solutions:**

1. **Limit crawl duration:**
   - Use shorter `total_duration_seconds`
   - Run multiple shorter crawls instead of one long crawl

2. **Set Docker memory limits:**
   ```yaml
   # In docker-compose.yml
   deploy:
     resources:
       limits:
         memory: 4G
   ```

3. **Monitor memory:**
   ```bash
   # For direct installation
   ps aux | grep socrawler
   
   # For Docker
   docker stats socrawler
   ```

---

## Production Recommendations

1. **Use systemd service** for automatic restart on failure
2. **Set up log rotation** to prevent disk space issues
3. **Monitor memory usage** and adjust limits as needed
4. **Use reverse proxy** (nginx/caddy) for HTTPS
5. **Implement rate limiting** to prevent abuse
6. **Regular updates** of Chrome and dependencies
7. **Backup downloaded videos** regularly

---

## Security Considerations

1. **Firewall**: Only expose port 8080 if needed externally
2. **Authentication**: Add authentication middleware for production
3. **HTTPS**: Use reverse proxy with SSL/TLS
4. **Sandboxing**: Chrome runs in sandbox mode by default
5. **Resource limits**: Set memory and CPU limits to prevent DoS

---

## Contact & Support

For issues and questions:
- GitHub Issues: https://github.com/esperyong/socrawler/issues
- Documentation: See README.md

---

**Last Updated:** October 2025

