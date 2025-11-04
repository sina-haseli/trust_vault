#!/bin/bash
# Build Trust Vault Plugin using Dockerfile.build
# This script builds Trust Wallet Core from source and compiles the plugin

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

IMAGE_NAME="${IMAGE_NAME:-trust-vault-builder}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
BUILD_OUTPUT="${BUILD_OUTPUT:-bin/trust-vault-plugin}"

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Trust Vault Plugin Builder${NC}"
echo -e "${BLUE}Using Dockerfile.build${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    echo "Please start Docker Desktop and try again."
    exit 1
fi

# Check if wallet-core submodule is initialized
if [ ! -d "third_party/wallet-core/.git" ]; then
    echo -e "${YELLOW}Initializing git submodule...${NC}"
    git submodule update --init --recursive
fi

# Check if wallet-core directory has content
if [ ! "$(ls -A third_party/wallet-core 2>/dev/null)" ]; then
    echo -e "${RED}Error: wallet-core directory is empty${NC}"
    echo "Please run: git submodule update --init --recursive"
    exit 1
fi

echo -e "${YELLOW}Building Docker image: $IMAGE_NAME:$IMAGE_TAG${NC}"
echo "This will take a while as it builds Trust Wallet Core from source..."
echo ""

# Build the Docker image
if docker build -t "$IMAGE_NAME:$IMAGE_TAG" -f Dockerfile.build .; then
    echo ""
    echo -e "${GREEN}✓ Docker image built successfully${NC}"
    echo ""
else
    echo -e "${RED}✗ Docker build failed${NC}"
    exit 1
fi

# Create output directory
mkdir -p bin

# Run the build container to get the binary
echo -e "${YELLOW}Extracting built binary...${NC}"

# Create a temporary container from the built image
CONTAINER_ID=$(docker create "$IMAGE_NAME:$IMAGE_TAG" 2>/dev/null) || {
    echo -e "${RED}Error: Failed to create container from image${NC}"
    echo "The image may not have built correctly. Check the build logs above."
    exit 1
}

# Extract the binary
if docker cp "$CONTAINER_ID:/workspace/trust-vault-plugin" "$BUILD_OUTPUT" 2>/dev/null; then
    echo -e "${GREEN}✓ Binary extracted successfully${NC}"
else
    # Try to find the binary in the container
    echo -e "${YELLOW}Binary not found in default location, searching...${NC}"
    BINARY_PATH=$(docker exec "$CONTAINER_ID" find /workspace -name "trust-vault-plugin" -type f 2>/dev/null | head -1)
    
    if [ -n "$BINARY_PATH" ]; then
        echo "Found binary at: $BINARY_PATH"
        docker cp "$CONTAINER_ID:$BINARY_PATH" "$BUILD_OUTPUT"
        echo -e "${GREEN}✓ Binary extracted successfully${NC}"
    else
        echo -e "${RED}Error: Could not find binary in container${NC}"
        echo "Container ID: $CONTAINER_ID"
        echo "Listing /workspace contents:"
        docker exec "$CONTAINER_ID" ls -la /workspace 2>/dev/null || true
        echo ""
        echo "You can inspect the container with:"
        echo "  docker run -it --rm $IMAGE_NAME:$IMAGE_TAG /bin/bash"
        docker rm "$CONTAINER_ID" > /dev/null 2>&1 || true
        exit 1
    fi
fi

# Clean up
docker rm "$CONTAINER_ID" > /dev/null 2>&1 || true

# Make binary executable
chmod +x "$BUILD_OUTPUT"

# Calculate SHA256 checksum
if command -v shasum > /dev/null; then
    SHA256=$(shasum -a 256 "$BUILD_OUTPUT" | awk '{print $1}')
    echo "$SHA256" > "$BUILD_OUTPUT.sha256"
    echo -e "${GREEN}✓ SHA256 checksum: $SHA256${NC}"
elif command -v sha256sum > /dev/null; then
    SHA256=$(sha256sum "$BUILD_OUTPUT" | awk '{print $1}')
    echo "$SHA256" > "$BUILD_OUTPUT.sha256"
    echo -e "${GREEN}✓ SHA256 checksum: $SHA256${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Build Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Binary location: $BUILD_OUTPUT"
echo "SHA256 file: $BUILD_OUTPUT.sha256"
echo ""
echo -e "${BLUE}To test the plugin:${NC}"
echo "  docker-compose up -d vault"
echo ""
echo -e "${BLUE}Or build and run manually:${NC}"
echo "  docker build -t trust-vault:latest -f Dockerfile ."
echo "  docker run -d -p 8200:8200 --name trust-vault trust-vault:latest"
echo ""

