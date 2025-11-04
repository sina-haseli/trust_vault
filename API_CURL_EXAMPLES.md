# Trust Vault API - cURL Examples

This document provides cURL examples for all Trust Vault Plugin API endpoints, similar to DQ Vault.

## Prerequisites

1. Vault must be running and unsealed
2. Plugin must be registered and enabled
3. You need a valid Vault token

### Get Vault Token

**Option 1: From Vault initialization (first time setup)**

```bash
# Initialize Vault (if not already initialized)
docker exec trust-vault sh -c "export VAULT_ADDR='http://127.0.0.1:8200' && vault operator init -key-shares=1 -key-threshold=1"

# Extract the root token from the output
# Look for: "Initial Root Token: hvs.xxxxx"
# Save it:
export VAULT_TOKEN="hvs.xxxxx"  # Replace with your actual token
export VAULT_ADDR="http://127.0.0.1:8200"
```

**Option 2: Get token from container (if already initialized)**

```bash
# Get token from container's saved file
docker exec trust-vault sh -c "grep ROOT_TOKEN /tmp/vault-keys.txt | cut -d= -f2-"

# Or set it directly:
export VAULT_TOKEN=$(docker exec trust-vault sh -c "grep ROOT_TOKEN /tmp/vault-keys.txt | cut -d= -f2-")
export VAULT_ADDR="http://127.0.0.1:8200"
```

**Option 3: Create a new token (recommended for production)**

```bash
# Create a token with policy (replace ROOT_TOKEN with your root token)
curl --location 'http://127.0.0.1:8200/v1/auth/token/create' \
--header 'X-Vault-Token: hvs.YourRootTokenHere' \
--header 'Content-Type: application/json' \
--data '{
    "policies": ["trust-vault-policy"],
    "ttl": "24h"
}'
```

## API Endpoints

### 1. Create Wallet

Creates a new cryptocurrency wallet or imports an existing one from a mnemonic phrase.

**Endpoint:** `POST /v1/trust-vault/wallets/:name`

**cURL Example:**

```bash
# Set your Vault token (get it from initialization or container)
export VAULT_TOKEN="hvs.YourActualTokenHere"  # Replace with real token
export VAULT_ADDR="http://127.0.0.1:8200"

# Create a new wallet (generates new mnemonic)
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-btc-wallet" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data '{
    "coin_type": 0
}'

# Import existing wallet with mnemonic
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/imported-wallet" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data '{
    "coin_type": 0,
    "mnemonic": "mean sheriff enter shadow social awake rent love oil core vocal fresh"
}'
```

**Response:**
```json
{
    "request_id": "...",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "name": "my-btc-wallet",
        "coin_type": 0,
        "address": "bc1qteql8gzaz74jx2q0wedqr3vz4mwyyq0v23zrfy",
        "public_key": "02e3b29818ffb10fd1fc8ca46a813fbc6ea25a48ec2192158fa14792797b89663c",
        "created_at": "2025-11-04T18:51:07Z"
    }
}
```

**Coin Types (BIP-44):**
- `0` - Bitcoin
- `60` - Ethereum
- `501` - Solana
- `195` - Tron
- See [SLIP-44](https://github.com/satoshilabs/slips/blob/master/slip-0044.md) for full list

---

### 2. Generate Address

Retrieves an address for a specific coin type with optional custom derivation path.

**Endpoint:** `GET /v1/trust-vault/wallets/:name/addresses/:coin`

**cURL Examples:**

```bash
# Set your Vault token first
export VAULT_TOKEN="hvs.YourActualTokenHere"  # Replace with real token
export VAULT_ADDR="http://127.0.0.1:8200"

# Get Bitcoin address (default derivation path: m/84'/0'/0'/0/0)
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-btc-wallet/addresses/0" \
--header "X-Vault-Token: $VAULT_TOKEN"

# Get Ethereum address
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet/addresses/60" \
--header "X-Vault-Token: $VAULT_TOKEN"

# Get address with custom derivation path (query parameter)
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-btc-wallet/addresses/0?derivation_path=m/84%27/0%27/0%27/0/1" \
--header "X-Vault-Token: $VAULT_TOKEN"
```

**Note:** In URLs, single quotes in derivation paths must be URL-encoded as `%27` (e.g., `m/44'/0'/0'` becomes `m/44%27/0%27/0%27`)

**Response:**
```json
{
    "request_id": "...",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "address": "bc1qteql8gzaz74jx2q0wedqr3vz4mwyyq0v23zrfy",
        "coin_type": 0
    }
}
```

---

### 3. Sign Transaction

Signs a transaction using the wallet's private key.

**Endpoint:** `PUT /v1/trust-vault/wallets/:name/sign`

**cURL Example:**

```bash
# Set your Vault token first
export VAULT_TOKEN="hvs.YourActualTokenHere"  # Replace with real token
export VAULT_ADDR="http://127.0.0.1:8200"

# Sign Bitcoin transaction
# First, prepare your transaction data (hex format)
TX_DATA="0100000001..." # Your raw transaction hex

# Base64 encode the transaction data
TX_DATA_B64=$(echo -n "$TX_DATA" | xxd -r -p | base64)

curl --location --request PUT "$VAULT_ADDR/v1/trust-vault/wallets/my-btc-wallet/sign" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data "{
    \"tx_data\": \"$TX_DATA_B64\"
}"

# Sign Ethereum transaction
# Example: Raw transaction data (hex)
ETH_TX_DATA="02f86c808504a817c80082520894742d35cc6634c0532925a3b844bc9e7595f0beb880de0b6b3a764000080c001a0..." # Your transaction hex
ETH_TX_DATA_B64=$(echo -n "$ETH_TX_DATA" | xxd -r -p | base64)

curl --location --request PUT "$VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet/sign" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data "{
    \"tx_data\": \"$ETH_TX_DATA_B64\"
}"

# Sign Tron transaction (similar to DQ Vault example)
TRON_TX_DATA="0a029c1c2208aabd84db65119e0440b8f58fefd1325a68080112640a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412330a1541726105df7e5011296e69639f7066eb0278d4c51712154110a16244dee80d6e373339f578d21a019630d12b1880ade20470fea98cefd132"
TRON_TX_DATA_B64=$(echo -n "$TRON_TX_DATA" | xxd -r -p | base64)

curl --location --request PUT "$VAULT_ADDR/v1/trust-vault/wallets/my-tron-wallet/sign" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data "{
    \"tx_data\": \"$TRON_TX_DATA_B64\"
}"
```

**Response:**
```json
{
    "request_id": "...",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "signed_tx": "base64-encoded-signed-transaction"
    }
}
```

---

### 4. Get Wallet Information

Retrieves wallet metadata (without exposing private keys).

**Endpoint:** `GET /v1/trust-vault/wallets/:name`

**cURL Example:**

```bash
# Set your Vault token first
export VAULT_TOKEN="hvs.YourActualTokenHere"  # Replace with real token
export VAULT_ADDR="http://127.0.0.1:8200"

curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-btc-wallet" \
--header "X-Vault-Token: $VAULT_TOKEN"
```

**Response:**
```json
{
    "request_id": "...",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "name": "my-btc-wallet",
        "coin_type": 0,
        "address": "bc1qteql8gzaz74jx2q0wedqr3vz4mwyyq0v23zrfy",
        "public_key": "02e3b29818ffb10fd1fc8ca46a813fbc6ea25a48ec2192158fa14792797b89663c",
        "created_at": "2025-11-04T18:51:07Z"
    }
}
```

---

### 5. List All Wallets

Lists all wallet names.

**Endpoint:** `LIST /v1/trust-vault/wallets`

**cURL Example:**

```bash
# Set your Vault token first
export VAULT_TOKEN="hvs.YourActualTokenHere"  # Replace with real token
export VAULT_ADDR="http://127.0.0.1:8200"

curl --location --request LIST "$VAULT_ADDR/v1/trust-vault/wallets" \
--header "X-Vault-Token: $VAULT_TOKEN"
```

**Response:**
```json
{
    "request_id": "...",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "keys": [
            "my-btc-wallet",
            "my-eth-wallet",
            "my-tron-wallet"
        ]
    }
}
```

---

### 6. Delete Wallet

Deletes a wallet from Vault.

**Endpoint:** `DELETE /v1/trust-vault/wallets/:name`

**cURL Example:**

```bash
# Set your Vault token first
export VAULT_TOKEN="hvs.YourActualTokenHere"  # Replace with real token
export VAULT_ADDR="http://127.0.0.1:8200"

curl --location --request DELETE "$VAULT_ADDR/v1/trust-vault/wallets/my-btc-wallet" \
--header "X-Vault-Token: $VAULT_TOKEN"
```

**Response:**
```json
{
    "request_id": "...",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": null
}
```

---

## Complete Example Workflow

Here's a complete example similar to DQ Vault:

```bash
# 1. Get your Vault token (choose one method)

# Method A: Get from container if already initialized
export VAULT_TOKEN=$(docker exec trust-vault sh -c "grep ROOT_TOKEN /tmp/vault-keys.txt | cut -d= -f2-")
export VAULT_ADDR="http://127.0.0.1:8200"

# Method B: Initialize and get token (if first time)
# docker exec trust-vault sh -c "export VAULT_ADDR='http://127.0.0.1:8200' && vault operator init -key-shares=1 -key-threshold=1"
# Extract "Initial Root Token: hvs.xxxxx" from output and set:
# export VAULT_TOKEN="hvs.xxxxx"
# export VAULT_ADDR="http://127.0.0.1:8200"

# Verify token works
echo "Using token: $VAULT_TOKEN"

# 2. Create/Import a wallet
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-wallet" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data '{
    "coin_type": 0,
    "mnemonic": "mean sheriff enter shadow social awake rent love oil core vocal fresh"
}'

# 3. Generate Bitcoin address
curl --location "$VAULT_ADDR/v1/trust-vault/wallets/my-wallet/addresses/0" \
--header "X-Vault-Token: $VAULT_TOKEN"

# Expected response:
# {"data":{"address":"bc1qteql8gzaz74jx2q0wedqr3vz4mwyyq0v23zrfy","coin_type":0}}

# 4. Sign a transaction (example with hex data)
TX_HEX="0100000001..."  # Your raw transaction hex
TX_B64=$(echo -n "$TX_HEX" | xxd -r -p | base64)

curl --location --request PUT "$VAULT_ADDR/v1/trust-vault/wallets/my-wallet/sign" \
--header "X-Vault-Token: $VAULT_TOKEN" \
--header 'Content-Type: application/json' \
--data "{
    \"tx_data\": \"$TX_B64\"
}"

# Response:
# {"data":{"signed_tx":"base64-encoded-signed-transaction"}}
```

---

## Error Responses

All errors follow this format:

```json
{
    "errors": [
        "error message here"
    ]
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `404` - Wallet not found
- `500` - Internal server error
- `503` - Vault is sealed

---

## Notes

1. **Transaction Data Format**: Transaction data must be provided as base64-encoded raw transaction bytes (hex format converted to bytes, then base64).

2. **Derivation Paths**: 
   - Default Bitcoin (Native SegWit): `m/84'/0'/0'/0/0`
   - Default Ethereum: `m/44'/60'/0'/0/0`
   - Custom paths can be specified in the address endpoint

3. **Security**: Never expose your Vault token in production. Use proper authentication and token management.

4. **URL Encoding**: When using derivation paths in URLs, single quotes (`'`) must be URL-encoded as `%27`.

---

## Comparison with DQ Vault

| DQ Vault | Trust Vault |
|----------|-------------|
| `POST /v1/dq/register` | `POST /v1/trust-vault/wallets/:name` |
| `GET /v1/dq/address?uuid=...&path=...&coinType=...` | `GET /v1/trust-vault/wallets/:name/addresses/:coin` |
| `POST /v1/dq/signature` | `PUT /v1/trust-vault/wallets/:name/sign` |

Trust Vault uses wallet names instead of UUIDs, and includes coin type in the URL path for better RESTful design.

