#!/bin/bash

# This script builds the application for different platforms.

# Exit on error
set -e

# Get the version from the git tag
VERSION=$(git describe --tags --abbrev=0)

# Create a release directory
mkdir -p release

# Define the platforms to build for
PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64"

for PLATFORM in $PLATFORMS
do
    # Split the platform into OS and architecture
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}

    # Set the output file name
    OUTPUT_NAME="stockmap"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="stockmap.exe"
    fi

    # Build the application
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-X 'stockmap/cmd.version=$VERSION'" -o "release/$OUTPUT_NAME" main.go

    # Create a tarball
    tar -czf "release/stockmap_${GOOS}_${GOARCH}.tar.gz" -C release "$OUTPUT_NAME"

    # Remove the binary
    rm "release/$OUTPUT_NAME"
done

echo "Build complete. Binaries are in the release directory."
