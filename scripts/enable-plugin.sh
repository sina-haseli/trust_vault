#!/bin/bash
# Script to enable Trust Vault Plugin secrets engine

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
PLUGIN_NAME="trust-vault"
MOUNT_PATH="${MOUNT_PATH:-trust-vault}"
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-}"

# Function to print usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Enable Trust Vault Plugin secrets engine"
    echo ""
    echo "Options:"
    echo "  -m, --mount-path PATH   Mount path (default: trust-vault)"
    echo "  -a, --vault-addr ADDR   Vault address (default: http://127.0.0.1:8200)"
    echo "  -t, --vault-token TOKEN Vault token (default: from VAULT_TOKEN env)"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  MOUNT_PATH              Mount path for the secrets engine"
    echo "  VAULT_ADDR              Vault server address"
    echo "  VAULT_TOKEN             Vault authentication token"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -m|--mount-path)
            MOUNT_PATH="$2"
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

echo -e "${GREEN}Trust Vault Plugin - Enable Secrets Engine${NC}"
echo "==========================================="
echo ""

# Check if vault CLI is installed
if ! command -v vault &> /dev/null; then
    echo -e "${RED}Error: vault CLI not found${NC}"
    echo "Please install HashiCorp Vault CLI: https://www.vaultproject.io/downloads"
    exit 1
fi

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

# Check if plugin is registered
echo -e "${YELLOW}Checking if plugin is registered...${NC}"
if ! vault plugin list secret | grep -q "$PLUGIN_NAME"; then
    echo -e "${RED}Error: Plugin '$PLUGIN_NAME' is not registered${NC}"
    echo "Please run './scripts/register-plugin.sh' first"
    exit 1
fi

echo -e "${GREEN}✓ Plugin is registered${NC}"
echo ""

# Check if already enabled
if vault secrets list | grep -q "^$MOUNT_PATH/"; then
    echo -e "${YELLOW}Warning: Secrets engine already enabled at $MOUNT_PATH${NC}"
    read -p "Do you want to disable and re-enable it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Disabling existing secrets engine...${NC}"
        vault secrets disable "$MOUNT_PATH"
        echo -e "${GREEN}✓ Disabled${NC}"
        echo ""
    else
        echo "Keeping existing secrets engine"
        exit 0
    fi
fi

# Enable the secrets engine
echo -e "${YELLOW}Enabling secrets engine at $MOUNT_PATH...${NC}"
vault secrets enable -path="$MOUNT_PATH" "$PLUGIN_NAME"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Secrets engine enabled successfully${NC}"
    echo ""
    echo "Mount Path: $MOUNT_PATH"
    echo "Plugin: $PLUGIN_NAME"
    echo ""
    echo -e "${GREEN}Example usage:${NC}"
    echo ""
    echo "# Create a wallet"
    echo "vault write $MOUNT_PATH/wallets/my-eth-wallet coin_type=60"
    echo ""
    echo "# Get wallet info"
    echo "vault read $MOUNT_PATH/wallets/my-eth-wallet"
    echo ""
    echo "# Get address for a coin type"
    echo "vault read $MOUNT_PATH/wallets/my-eth-wallet/addresses/60"
    echo ""
    echo "# Sign a transaction"
    echo "vault write $MOUNT_PATH/wallets/my-eth-wallet/sign tx_data=@transaction.json"
    echo ""
    echo "# List wallets"
    echo "vault list $MOUNT_PATH/wallets"
    echo ""
    echo "# Delete a wallet"
    echo "vault delete $MOUNT_PATH/wallets/my-eth-wallet"
else
    echo -e "${RED}✗ Failed to enable secrets engine${NC}"
    exit 1
fi
