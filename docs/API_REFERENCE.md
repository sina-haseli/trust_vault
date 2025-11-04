# Trust Vault Plugin API Reference

This document provides detailed information about all API endpoints available in the Trust Vault Plugin.

## Table of Contents

- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Create Wallet](#create-wallet)
  - [Get Wallet](#get-wallet)
  - [Delete Wallet](#delete-wallet)
  - [List Wallets](#list-wallets)
  - [Sign Transaction](#sign-transaction)
  - [Get Address](#get-address)
- [Error Responses](#error-responses)
- [Coin Types](#coin-types)

## Authentication

All API requests must include a valid Vault token. Set the token using:

```bash
export VAULT_TOKEN="your-vault-token"
```

Or include it in the request header:

```
X-Vault-Token: your-vault-token
```

## Endpoints

### Create Wallet

Creates a new cryptocurrency wallet or imports an existing one from a mnemonic phrase.

**Endpoint:** `POST /trust-vault/wallets/:name`

**Parameters:**

| Parameter | Type    | Required | Description                                                 |
| --------- | ------- | -------- | ----------------------------------------------------------- |
| name      | string  | Yes      | Unique identifier for the wallet (path parameter)           |
| coin_type | integer | Yes      | BIP-44 coin type (e.g., 0 for Bitcoin, 60 for Ethereum)     |
| mnemonic  | string  | No       | 12 or 24-word mnemonic phrase for importing existing wallet |

**Request Example (CLI):**

```bash
# Create new wallet
vault write trust-vault/wallets/my-eth-wallet coin_type=60

# Import existing wallet
vault write trust-vault/wallets/imported-wallet \
  coin_type=60 \
  mnemonic="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
```

**Request Example (HTTP):**

```bash
curl -X POST \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -d '{"coin_type": 60}' \
  $VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet
```

**Response:**

```json
{
  "request_id": "abc123...",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "name": "my-eth-wallet",
    "coin_type": 60,
    "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "public_key": "0x04a8d5c...",
    "created_at": "2025-11-04T10:30:00Z"
  }
}
```

**Status Codes:**

- `200` - Wallet created successfully
- `400` - Invalid coin type or mnemonic
- `409` - Wallet with this name already exists
- `500` - Internal server error

---

### Get Wallet

Retrieves wallet information without exposing private keys or mnemonic.

**Endpoint:** `GET /trust-vault/wallets/:name`

**Parameters:**

| Parameter | Type   | Required | Description                        |
| --------- | ------ | -------- | ---------------------------------- |
| name      | string | Yes      | Wallet identifier (path parameter) |

**Request Example (CLI):**

```bash
vault read trust-vault/wallets/my-eth-wallet
```

**Request Example (HTTP):**

```bash
curl -X GET \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  $VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet
```

**Response:**

```json
{
  "request_id": "def456...",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "name": "my-eth-wallet",
    "coin_type": 60,
    "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "public_key": "0x04a8d5c...",
    "created_at": "2025-11-04T10:30:00Z"
  }
}
```

**Status Codes:**

- `200` - Wallet found
- `404` - Wallet not found
- `500` - Internal server error

---

### Delete Wallet

Permanently deletes a wallet and all associated key material.

**Endpoint:** `DELETE /trust-vault/wallets/:name`

**Parameters:**

| Parameter | Type   | Required | Description                        |
| --------- | ------ | -------- | ---------------------------------- |
| name      | string | Yes      | Wallet identifier (path parameter) |

**Request Example (CLI):**

```bash
vault delete trust-vault/wallets/my-eth-wallet
```

**Request Example (HTTP):**

```bash
curl -X DELETE \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  $VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet
```

**Response:**

```json
{
  "request_id": "ghi789...",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": null
}
```

**Status Codes:**

- `204` - Wallet deleted successfully
- `404` - Wallet not found
- `500` - Internal server error

---

### List Wallets

Returns a list of all wallet names.

**Endpoint:** `LIST /trust-vault/wallets`

**Parameters:** None

**Request Example (CLI):**

```bash
vault list trust-vault/wallets
```

**Request Example (HTTP):**

```bash
curl -X LIST \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  $VAULT_ADDR/v1/trust-vault/wallets
```

**Response:**

```json
{
  "request_id": "jkl012...",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "keys": [
      "my-eth-wallet",
      "my-btc-wallet",
      "my-sol-wallet"
    ]
  }
}
```

**Status Codes:**

- `200` - List retrieved successfully
- `500` - Internal server error

---

### Sign Transaction

Signs a transaction using the wallet's private key.

**Endpoint:** `POST /trust-vault/wallets/:name/sign`

**Parameters:**

| Parameter | Type   | Required | Description                        |
| --------- | ------ | -------- | ---------------------------------- |
| name      | string | Yes      | Wallet identifier (path parameter) |
| tx_data   | string | Yes      | Base64-encoded transaction data    |

**Request Example (CLI):**

```bash
# Prepare transaction data
echo '{"to":"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb","value":"1000000000000000000","data":"0x"}' | base64 > tx.b64

# Sign transaction
vault write trust-vault/wallets/my-eth-wallet/sign tx_data=@tx.b64
```

**Request Example (HTTP):**

```bash
curl -X POST \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -d '{"tx_data": "eyJ0byI6IjB4NzQyZDM1Q2M2NjM0QzA1MzI5MjVhM2I4NDRCYzllNzU5NWYwYkViIiwidmFsdWUiOiIxMDAwMDAwMDAwMDAwMDAwMDAwIiwiZGF0YSI6IjB4In0="}' \
  $VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet/sign
```

**Response:**

```json
{
  "request_id": "mno345...",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "signed_tx": "0xf86c808504a817c800825208947..."
  }
}
```

**Status Codes:**

- `200` - Transaction signed successfully
- `400` - Invalid transaction data
- `404` - Wallet not found
- `500` - Signing failed

---

### Get Address

Retrieves an address for a specific coin type, optionally using a custom derivation path.

**Endpoint:** `GET /trust-vault/wallets/:name/addresses/:coin`

**Parameters:**

| Parameter       | Type    | Required | Description                                     |
| --------------- | ------- | -------- | ----------------------------------------------- |
| name            | string  | Yes      | Wallet identifier (path parameter)              |
| coin            | integer | Yes      | Coin type (path parameter)                      |
| derivation_path | string  | No       | Custom BIP-44 derivation path (query parameter) |

**Request Example (CLI):**

```bash
# Get default address
vault read trust-vault/wallets/my-eth-wallet/addresses/60

# Get address with custom derivation path
vault read trust-vault/wallets/my-eth-wallet/addresses/60 \
  derivation_path="m/44'/60'/0'/0/1"
```

**Request Example (HTTP):**

```bash
# Default address
curl -X GET \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  $VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet/addresses/60

# Custom derivation path
curl -X GET \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  "$VAULT_ADDR/v1/trust-vault/wallets/my-eth-wallet/addresses/60?derivation_path=m/44'/60'/0'/0/1"
```

**Response:**

```json
{
  "request_id": "pqr678...",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "coin_type": 60,
    "derivation_path": "m/44'/60'/0'/0/0"
  }
}
```

**Status Codes:**

- `200` - Address retrieved successfully
- `400` - Invalid coin type or derivation path
- `404` - Wallet not found
- `500` - Internal server error

---

## Error Responses

All error responses follow this format:

```json
{
  "errors": [
    "descriptive error message"
  ]
}
```

### Common Error Messages

| Error Message                | Cause                   | Solution                                             |
| ---------------------------- | ----------------------- | ---------------------------------------------------- |
| `wallet not found`           | Wallet doesn't exist    | Verify wallet name and create if needed              |
| `wallet already exists`      | Duplicate wallet name   | Use a different name or delete existing wallet       |
| `invalid coin type`          | Unsupported coin type   | Check supported coin types list                      |
| `invalid mnemonic phrase`    | Malformed mnemonic      | Verify mnemonic has 12 or 24 words                   |
| `invalid transaction data`   | Malformed tx_data       | Ensure proper JSON and base64 encoding               |
| `transaction signing failed` | Trust Wallet Core error | Check transaction format for the specific blockchain |

---

## Coin Types

The plugin uses BIP-44 coin types. Here are the most common ones:

| Coin Type | Blockchain          | Symbol |
| --------- | ------------------- | ------ |
| 0         | Bitcoin             | BTC    |
| 2         | Litecoin            | LTC    |
| 3         | Dogecoin            | DOGE   |
| 60        | Ethereum            | ETH    |
| 60        | Binance Smart Chain | BSC    |
| 60        | Polygon             | MATIC  |
| 145       | Bitcoin Cash        | BCH    |
| 354       | Polkadot            | DOT    |
| 501       | Solana              | SOL    |
| 714       | Binance Chain       | BNB    |
| 1815      | Cardano             | ADA    |

For a complete list, refer to [SLIP-0044](https://github.com/satoshilabs/slips/blob/master/slip-0044.md).

---

## Rate Limiting

The plugin respects Vault's rate limiting configuration. If you encounter rate limit errors, adjust your Vault configuration or implement client-side throttling.

## Access Control

Use Vault policies to control access to wallet operations:

```hcl
# Allow wallet creation and reading
path "trust-vault/wallets/*" {
  capabilities = ["create", "read"]
}

# Allow transaction signing
path "trust-vault/wallets/*/sign" {
  capabilities = ["create", "update"]
}

# Deny wallet deletion
path "trust-vault/wallets/*" {
  capabilities = ["deny"]
}
```
