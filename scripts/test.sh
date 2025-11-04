#!/bin/bash
# Test script for Trust Vault Plugin

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running Trust Vault Plugin Tests${NC}"
echo "=================================="
echo ""

# Run unit tests
echo -e "${YELLOW}Running unit tests...${NC}"
go test -v -race -coverprofile=coverage.out ./...

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All tests passed${NC}"
    echo ""
    
    # Generate coverage report
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -html=coverage.out -o coverage.html
    
    # Display coverage summary
    echo ""
    echo -e "${GREEN}Coverage Summary:${NC}"
    go tool cover -func=coverage.out | tail -n 1
    echo ""
    echo "Detailed coverage report: coverage.html"
else
    echo ""
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
fi
