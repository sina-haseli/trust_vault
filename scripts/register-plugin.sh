#!/bin/bash
# Script to register Trust Vault Plugin with HashiCorp Vault

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
PLUGIN_NAME="trust-vault"
PLUGIN_BINARY="trust-vault-plugin"
PLUGIN_DIR="${VAULT_PLUGIN_DIR:-/etc/vault/plugins}"
BUILD_DIR="bin"
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-}"

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Register Trust Vault Plugin with HashiCorp Vault"
    echo ""
    echo "Options:"
    echo "  -p, --plugin-dir DIR    Plugin directory (default: /etc/vault/plugins)"
    echo "  -a, --vault-addr ADDR   Vault address (default: http://127.0.0.1:8200)"
    echo "  -t, --vault-token TOKEN Vault token (default: from VAULT_TOKEN env)"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  VAULT_PLUGIN_DIR        Plugin directory"
    echo "  VAULT_ADDR              Vault server address"
    echo "  VAULT_TOKEN             Vault authentication token"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--plugin-dir)
            PLUGIN_DIR="$2"
            shift 2
            ;;
        -a|--vault-addr)
            VAULT_ADDR="$2"
            shift 2
            ;;
        -t|--vault-token)
            VAULT_TOKEN="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

echo -e "${GREEN}Trust Vault Plugin Registration${NC}"
echo "=================================="
echo ""

# Check if vault CLI is installed
if ! command -v vault &> /dev/null; then
    echo -e "${RED}Error: vault CLI not found${NC}"
    echo "Please install HashiCorp Vault CLI: https://www.vaultproject.io/downloads"
    exit 1
fi

# Check if plugin binary exists
PLUGIN_PATH="$BUILD_DIR/$PLUGIN_BINARY"
if [ ! -f "$PLUGIN_PATH" ]; then
    echo -e "${RED}Error: Plugin binary not found at $PLUGIN_PATH${NC}"
    echo "Please run 'make build' first"
    exit 1
fi

# Check if SHA256 file exists
SHA256_FILE="$PLUGIN_PATH.sha256"
if [ ! -f "$SHA256_FILE" ]; then
    echo -e "${YELLOW}SHA256 file not found, calculating...${NC}"
    if command -v sha256sum &> /dev/null; then
        sha256sum "$PLUGIN_PATH" | awk '{print $1}' > "$SHA256_FILE"
    elif command -v shasum &> /dev/null; then
        shasum -a 256 "$PLUGIN_PATH" | awk '{print $1}' > "$SHA256_FILE"
    else
        echo -e "${RED}Error: No SHA256 utility found${NC}"
        exit 1
    fi
fi

SHA256=$(cat "$SHA256_FILE")
echo "Plugin binary: $PLUGIN_PATH"
echo "SHA256: $SHA256"
echo ""

# Check Vault connection
echo -e "${YELLOW}Checking Vault connection...${NC}"
export VAULT_ADDR
if [ -n "$VAULT_TOKEN" ]; then
    export VAULT_TOKEN
fi

if ! vault status &> /dev/null; then
    echo -e "${RED}Error: Cannot connect to Vault at $VAULT_ADDR${NC}"
    echo "Please ensure Vault is running and VAULT_ADDR is correct"
    exit 1
fi

echo -e "${GREEN}✓ Connected to Vault${NC}"
echo ""

# Copy plugin to plugin directory
echo -e "${YELLOW}Copying plugin to $PLUGIN_DIR...${NC}"
sudo mkdir -p "$PLUGIN_DIR"
sudo cp "$PLUGIN_PATH" "$PLUGIN_DIR/"
sudo chmod +x "$PLUGIN_DIR/$PLUGIN_BINARY"
echo -e "${GREEN}✓ Plugin copied${NC}"
echo ""

# Register plugin with Vault
echo -e "${YELLOW}Registering plugin with Vault...${NC}"
vault plugin register \
    -sha256="$SHA256" \
    -command="$PLUGIN_BINARY" \
    secret "$PLUGIN_NAME"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Plugin registered successfully${NC}"
    echo ""
    echo "Plugin Name: $PLUGIN_NAME"
    echo "Plugin Type: secret"
    echo "Command: $PLUGIN_BINARY"
    echo ""
    echo -e "${GREEN}Next steps:${NC}"
    echo "1. Enable the plugin: ./scripts/enable-plugin.sh"
    echo "2. Or manually: vault secrets enable -path=$PLUGIN_NAME $PLUGIN_NAME"
else
    echo -e "${RED}✗ Plugin registration failed${NC}"
    exit 1
fi
