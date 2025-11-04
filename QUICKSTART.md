# Trust Vault Plugin - Quick Start Guide

This guide will help you get started with the Trust Vault Plugin quickly.

## Quick Start (Local Development)

### 1. Build the Plugin

```bash
# Using Make
make build

# Or using build script
./build.sh
```

### 2. Register and Enable

```bash
# Register the plugin with Vault
./scripts/register-plugin.sh

# Enable the secrets engine
./scripts/enable-plugin.sh
```

### 3. Create Your First Wallet

```bash
# Set Vault address and token
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='your-vault-token'

# Create an Ethereum wallet
vault write trust-vault/wallets/my-first-wallet coin_type=60

# View wallet information
vault read trust-vault/wallets/my-first-wallet
```

## Quick Start (Docker)

### 1. Build and Run

```bash
# Build and start Vault with plugin
docker-compose up -d vault

# Wait for initialization
docker-compose logs -f vault
```

### 2. Access Vault

The initialization script will output the root token and unseal key. Save these securely!

```bash
# Set environment variables
export VAULT_ADDR='http://localhost:8200'
export VAULT_TOKEN='<root-token-from-logs>'

# Verify plugin is enabled
vault secrets list
```

### 3. Create a Wallet

```bash
# Create a wallet
vault write trust-vault/wallets/docker-wallet coin_type=60

# Get wallet info
vault read trust-vault/wallets/docker-wallet
```

## Common Operations

### Create Wallets for Different Chains

```bash
# Bitcoin
vault write trust-vault/wallets/btc-wallet coin_type=0

# Ethereum
vault write trust-vault/wallets/eth-wallet coin_type=60

# Solana
vault write trust-vault/wallets/sol-wallet coin_type=501
```

### Import Existing Wallet

```bash
vault write trust-vault/wallets/imported-wallet \
  coin_type=60 \
  mnemonic="your twelve word mnemonic phrase goes here like this example"
```

### Get Address

```bash
# Get default address
vault read trust-vault/wallets/eth-wallet/addresses/60

# Get address with custom derivation path
vault read trust-vault/wallets/eth-wallet/addresses/60 \
  derivation_path="m/44'/60'/0'/0/1"
```

### Sign Transaction

```bash
# Create transaction data file
cat > tx.json << EOF
{
  "to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "value": "1000000000000000000",
  "data": "0x"
}
EOF

# Sign the transaction
vault write trust-vault/wallets/eth-wallet/sign \
  tx_data=@tx.json
```

### List and Delete

```bash
# List all wallets
vault list trust-vault/wallets

# Delete a wallet
vault delete trust-vault/wallets/old-wallet
```

## Troubleshooting

### Vault Not Running

```bash
# Check Vault status
vault status

# If not running, start Vault
vault server -dev
```

### Plugin Not Found

```bash
# Verify plugin is registered
vault plugin list secret | grep trust-vault

# If not registered, run registration script
./scripts/register-plugin.sh
```

### Permission Denied

```bash
# Ensure you're authenticated
vault login

# Check your token has permissions
vault token lookup
```

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Review the [API documentation](docs/API.md) for all endpoints
- Check out [example scripts](examples/) for common use cases
- Review security best practices in the design document

## Support

For issues and questions:
- Check the troubleshooting section in README.md
- Review Vault logs: `vault audit list` and check audit logs
- For Docker: `docker-compose logs vault`

## Coin Type Reference

Common coin types for wallet creation:

| Blockchain          | Coin Type | Symbol |
| ------------------- | --------- | ------ |
| Bitcoin             | 0         | BTC    |
| Ethereum            | 60        | ETH    |
| Solana              | 501       | SOL    |
| Binance Smart Chain | 60        | BNB    |
| Polygon             | 60        | MATIC  |
| Avalanche           | 60        | AVAX   |

For a complete list, refer to [SLIP-0044](https://github.com/satoshilabs/slips/blob/master/slip-0044.md).
