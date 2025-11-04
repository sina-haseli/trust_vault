#!/bin/bash
# Development Docker container - for iterative development and testing
# Mounts your local code so you can edit and test in real-time

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

IMAGE_NAME="${IMAGE_NAME:-trust-vault-builder:latest}"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Trust Vault Development Environment${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker image exists
if ! docker image inspect "$IMAGE_NAME" &> /dev/null; then
    echo -e "${YELLOW}Docker image not found. Building it first...${NC}"
    echo "This will take 30-60 minutes on first build..."
    make docker-build-build
    echo ""
fi

echo -e "${GREEN}Starting development container...${NC}"
echo ""
echo "Your local project is mounted at /workspace"
echo "Changes you make locally will be reflected immediately"
echo ""
echo -e "${YELLOW}Quick commands:${NC}"
echo "  go run test_wallet.go              - Run wallet tests"
echo "  go build ./wallet                  - Build wallet package"
echo "  go build -o test-plugin ./cmd/trust-vault - Build plugin"
echo "  go test ./...                      - Run all tests"
echo "  go vet ./...                       - Check for issues"
echo ""
echo -e "${YELLOW}Example workflow:${NC}"
echo "  1. Edit wallet/trustwallet.go in your editor"
echo "  2. Run: go run test_wallet.go"
echo "  3. See results immediately"
echo ""
echo -e "${BLUE}Starting interactive shell...${NC}"
echo ""

# Run interactive container
docker run -it --rm \
    -v "$PROJECT_DIR:/workspace" \
    -w /workspace \
    "$IMAGE_NAME" \
    bash

