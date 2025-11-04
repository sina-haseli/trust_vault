#!/bin/bash
# Build Docker image with Trust Vault Plugin

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

IMAGE_NAME="${IMAGE_NAME:-trust-vault}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

echo -e "${GREEN}Building Docker image: $IMAGE_NAME:$IMAGE_TAG${NC}"
echo ""

# Build the Docker image
echo -e "${YELLOW}Building image...${NC}"
docker build -t "$IMAGE_NAME:$IMAGE_TAG" -f Dockerfile .

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Docker image built successfully${NC}"
    echo ""
    echo "Image: $IMAGE_NAME:$IMAGE_TAG"
    echo ""
    echo -e "${GREEN}To run the container:${NC}"
    echo "docker run -d -p 8200:8200 --name trust-vault $IMAGE_NAME:$IMAGE_TAG"
    echo ""
    echo -e "${GREEN}Or use docker-compose:${NC}"
    echo "docker-compose up -d vault"
else
    echo -e "${RED}✗ Docker build failed${NC}"
    exit 1
fi
