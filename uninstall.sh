#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BINARY_NAME="stockmap"
BINARY_NAME_CLI="stockmap-cli"

echo -e "${BLUE}StockMap Uninstaller${NC}"
echo ""

# Common installation paths
LOCATIONS=(
    "/usr/local/bin"
    "$HOME/.local/bin"
    "$HOME/go/bin"
    "$(go env GOPATH 2>/dev/null)/bin"
)

FOUND=0

for DIR in "${LOCATIONS[@]}"; do
    # Skip empty paths
    [ -z "$DIR" ] && continue

    # Check for stockmap
    if [ -f "$DIR/$BINARY_NAME" ]; then
        echo -e "${YELLOW}Found $BINARY_NAME in $DIR${NC}"
        rm -f "$DIR/$BINARY_NAME"
        echo -e "${GREEN}Removed $DIR/$BINARY_NAME${NC}"
        FOUND=1
    fi

    # Check for stockmap-cli (go install default)
    if [ -f "$DIR/$BINARY_NAME_CLI" ]; then
        echo -e "${YELLOW}Found $BINARY_NAME_CLI in $DIR${NC}"
        rm -f "$DIR/$BINARY_NAME_CLI"
        echo -e "${GREEN}Removed $DIR/$BINARY_NAME_CLI${NC}"
        FOUND=1
    fi
done

if [ $FOUND -eq 0 ]; then
    echo -e "${YELLOW}No installation of stockmap found in standard locations.${NC}"
    echo "If you installed it elsewhere, please remove it manually."
else
    echo ""
    echo -e "${GREEN}✓ Uninstall complete.${NC}"
    echo ""
    echo "Note: You may want to remove the PATH configuration from your shell profile"
    echo "      (e.g. .zshrc, .bashrc) if it was added manually."
fi
