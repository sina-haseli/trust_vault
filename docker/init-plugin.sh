#!/bin/bash
# Initialization script for Trust Vault Plugin in Docker

set -e

# Start Vault in the background
vault server -config=/vault/config/vault.hcl &
VAULT_PID=$!

# Wait for Vault to start
echo "Waiting for Vault to start..."
sleep 5

# Set Vault address
export VAULT_ADDR='http://127.0.0.1:8200'

# Check if Vault is initialized
if ! vault status &> /dev/null; then
    echo "Initializing Vault..."
    
    # Initialize Vault
    INIT_OUTPUT=$(vault operator init -key-shares=1 -key-threshold=1 -format=json)
    
    # Extract keys and token
    UNSEAL_KEY=$(echo "$INIT_OUTPUT" | jq -r '.unseal_keys_b64[0]')
    ROOT_TOKEN=$(echo "$INIT_OUTPUT" | jq -r '.root_token')
    
    # Save to file
    echo "$INIT_OUTPUT" > /vault/data/init-keys.json
    
    echo "Vault initialized!"
    echo "Root Token: $ROOT_TOKEN"
    echo "Unseal Key: $UNSEAL_KEY"
    echo ""
    echo "IMPORTANT: Save these credentials securely!"
    echo "Keys saved to: /vault/data/init-keys.json"
    echo ""
    
    # Unseal Vault
    echo "Unsealing Vault..."
    vault operator unseal "$UNSEAL_KEY"
    
    # Login with root token
    vault login "$ROOT_TOKEN"
    
    # Register and enable plugin
    echo "Registering Trust Vault Plugin..."
    SHA256=$(cat /vault/plugins/trust-vault-plugin.sha256)
    
    vault plugin register \
        -sha256="$SHA256" \
        -command="trust-vault-plugin" \
        secret trust-vault
    
    echo "Enabling Trust Vault Plugin..."
    vault secrets enable -path=trust-vault trust-vault
    
    echo ""
    echo "Trust Vault Plugin is ready!"
    echo "Access Vault UI at: http://localhost:8200"
    echo "Root Token: $ROOT_TOKEN"
else
    echo "Vault is already initialized"
    
    # Check if sealed
    if vault status | grep -q "Sealed.*true"; then
        echo "Vault is sealed. Please unseal manually."
        echo "Unseal key is in: /vault/data/init-keys.json"
    fi
fi

# Keep container running
wait $VAULT_PID
