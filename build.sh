#!/bin/bash
# Build script for Trust Vault Plugin

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
PLUGIN_NAME="trust-vault-plugin"
BUILD_DIR="bin"
VERSION=${VERSION:-"dev"}
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo -e "${GREEN}Building Trust Vault Plugin${NC}"
echo "Version: $VERSION"
echo "Commit: $COMMIT_HASH"
echo "Build Time: $BUILD_TIME"
echo ""

# Create build directory
mkdir -p "$BUILD_DIR"

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X main.Version=$VERSION"
LDFLAGS="$LDFLAGS -X main.CommitHash=$COMMIT_HASH"
LDFLAGS="$LDFLAGS -X main.BuildTime=$BUILD_TIME"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

OUTPUT_FILE="$BUILD_DIR/$PLUGIN_NAME"
if [ "$OS" = "windows" ]; then
    OUTPUT_FILE="$OUTPUT_FILE.exe"
fi

echo -e "${YELLOW}Building for $OS/$ARCH...${NC}"

# Build the plugin
go build -v -ldflags="$LDFLAGS" -o "$OUTPUT_FILE" ./cmd/trust-vault

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful: $OUTPUT_FILE${NC}"
    
    # Calculate SHA256
    echo ""
    echo -e "${YELLOW}Calculating SHA256 checksum...${NC}"
    
    if command -v sha256sum &> /dev/null; then
        SHA256=$(sha256sum "$OUTPUT_FILE" | awk '{print $1}')
    elif command -v shasum &> /dev/null; then
        SHA256=$(shasum -a 256 "$OUTPUT_FILE" | awk '{print $1}')
    else
        echo -e "${RED}Warning: No SHA256 utility found${NC}"
        SHA256="unknown"
    fi
    
    echo "$SHA256" > "$OUTPUT_FILE.sha256"
    echo -e "${GREEN}SHA256: $SHA256${NC}"
    echo ""
    
    # Display file info
    FILE_SIZE=$(ls -lh "$OUTPUT_FILE" | awk '{print $5}')
    echo "Binary size: $FILE_SIZE"
    echo "Location: $OUTPUT_FILE"
    echo ""
    echo -e "${GREEN}Build complete!${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi
