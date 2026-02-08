#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="febritecno/stockmap-cli"
BINARY_NAME="stockmap"
INSTALL_DIR="/usr/local/bin"

echo -e "${BLUE}"
echo "  _____ _             _    __  __             "
echo " / ____| |           | |  |  \/  |            "
echo "| (___ | |_ ___   ___| | _| \  / | __ _ _ __  "
echo " \___ \| __/ _ \ / __| |/ / |\/| |/ _\` | '_ \ "
echo " ____) | || (_) | (__|   <| |  | | (_| | |_) |"
echo "|_____/ \__\___/ \___|_|\_\_|  |_|\__,_| .__/ "
echo "                                       | |    "
echo "                                       |_|    "
echo -e "${NC}"
echo "Stock Screener for Terminal"
echo ""

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${YELLOW}Detected: ${OS}/${ARCH}${NC}"

# Check if Go is installed
if command -v go &> /dev/null; then
    echo -e "${GREEN}Go detected. Installing via go install...${NC}"
    go install github.com/${REPO}@latest
    echo -e "${GREEN}✓ Installed successfully!${NC}"
    echo ""
    echo "Run 'stockmap' to start the application."
    exit 0
fi

# Otherwise, download pre-built binary
echo -e "${YELLOW}Go not found. Downloading pre-built binary...${NC}"

# Get latest release
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_RELEASE" ]; then
    echo -e "${RED}Could not fetch latest release. Please install manually.${NC}"
    echo "Visit: https://github.com/${REPO}/releases"
    exit 1
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}_${OS}_${ARCH}.tar.gz"

echo "Downloading ${BINARY_NAME} ${LATEST_RELEASE}..."

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/${BINARY_NAME}.tar.gz"

if [ ! -f "$TMP_DIR/${BINARY_NAME}.tar.gz" ]; then
    echo -e "${RED}Download failed. Please install manually.${NC}"
    exit 1
fi

tar -xzf "$TMP_DIR/${BINARY_NAME}.tar.gz" -C "$TMP_DIR"

# Install
echo -e "${YELLOW}Installing to ${INSTALL_DIR}...${NC}"

if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/${BINARY_NAME}" "$INSTALL_DIR/"
else
    echo -e "${YELLOW}Requesting sudo access...${NC}"
    sudo mv "$TMP_DIR/${BINARY_NAME}" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/${BINARY_NAME}"

echo ""
echo -e "${GREEN}✓ StockMap installed successfully!${NC}"
echo ""
echo "Run 'stockmap' to start the application."
echo ""
