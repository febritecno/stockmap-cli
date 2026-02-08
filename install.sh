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

# Helper: Check if directory is in PATH
is_in_path() {
    case ":$PATH:" in
        *":$1:"*) return 0 ;;
        *) return 1 ;;
    esac
}

# Helper: Detect shell profile
detect_profile() {
    local SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
        bash)
            if [ -f "$HOME/.bashrc" ]; then echo "$HOME/.bashrc";
            elif [ -f "$HOME/.bash_profile" ]; then echo "$HOME/.bash_profile";
            else echo "$HOME/.bashrc"; fi
            ;;
        zsh)
            echo "$HOME/.zshrc"
            ;;
        fish)
            echo "$HOME/.config/fish/config.fish"
            ;;
        *)
            echo "$HOME/.profile"
            ;;
    esac
}

# 1. Determine Installation Directory
# Priority:
# 1. ~/.local/bin (if it exists or we can create it easily, and it avoids sudo)
# 2. /usr/local/bin (Standard system location)

if [ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=false
elif [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
    USE_SUDO=false
else
    # Default fallback - try ~/.local/bin (create if needed)
    mkdir -p "$HOME/.local/bin"
    INSTALL_DIR="$HOME/.local/bin"
    USE_SUDO=false
    
    # If that fails, fallback to system with sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        INSTALL_DIR="/usr/local/bin"
        USE_SUDO=true
    fi
fi

echo -e "${YELLOW}Target installation directory: ${INSTALL_DIR}${NC}"

# 2. Build or Download
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

if command -v go &> /dev/null; then
    echo -e "${GREEN}Go detected. Building from source...${NC}"
    
    # Check if we are in the repo to build locally, otherwise fetch
    if [ -f "go.mod" ] && grep -q "module github.com/febritecno/stockmap-cli" "go.mod"; then
        echo "Building from local source..."
        go build -o "$TMP_DIR/$BINARY_NAME" .
    else
        echo "Downloading and building latest version..."
        GOBIN="$TMP_DIR" go install github.com/${REPO}@latest
        # Handle potential name mismatch if go install uses module name
        if [ -f "$TMP_DIR/stockmap-cli" ]; then
            mv "$TMP_DIR/stockmap-cli" "$TMP_DIR/$BINARY_NAME"
        fi
    fi
else
    echo -e "${YELLOW}Go not found. Downloading pre-built binary...${NC}"
    
    # Detect OS and Architecture
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) 
            echo -e "${RED}Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac

    # Get latest release
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_RELEASE" ]; then
        echo -e "${RED}Could not fetch latest release. Please install manually.${NC}"
        exit 1
    fi
    
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_RELEASE}/${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
    echo "Downloading ${BINARY_NAME} ${LATEST_RELEASE}..."
    
    curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/release.tar.gz"
    tar -xzf "$TMP_DIR/release.tar.gz" -C "$TMP_DIR"
fi

# 3. Move Binary to Install Directory
echo "Installing binary..."

if [ -f "$TMP_DIR/$BINARY_NAME" ]; then
    if [ "$USE_SUDO" = true ]; then
        echo -e "${YELLOW}Sudo required to move binary to $INSTALL_DIR${NC}"
        sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    else
        mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
    fi
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${RED}Error: Binary not found in temp dir. Installation failed.${NC}"
    exit 1
fi

# 4. Auto-Configure Path if needed
if ! is_in_path "$INSTALL_DIR"; then
    PROFILE=$(detect_profile)
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH.${NC}"
    echo -e "Adding to $PROFILE..."
    
    if [ -w "$PROFILE" ]; then
        echo "" >> "$PROFILE"
        echo "# StockMap CLI" >> "$PROFILE"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$PROFILE"
        echo -e "${GREEN}✓ Added to PATH.${NC}"
        echo "Please restart your terminal or run: source $PROFILE"
    else
        echo -e "${RED}Could not write to $PROFILE. Please add this manually:${NC}"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
fi

echo ""
echo -e "${GREEN}✓ StockMap installed successfully to $INSTALL_DIR/$BINARY_NAME${NC}"
echo "Run '$BINARY_NAME' to start."
echo ""
