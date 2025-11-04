# Trust Vault Plugin

A HashiCorp Vault secrets engine plugin that integrates Trust Wallet Core library to provide cryptocurrency wallet functionality. The plugin enables secure generation of cryptocurrency addresses and transaction signing for multiple blockchain networks through Vault's API.

## Features

- **Multi-Chain Support**: Bitcoin, Ethereum, Solana, and other networks supported by Trust Wallet Core
- **Secure Key Management**: Private keys encrypted at rest using Vault's encryption
- **HD Wallet Support**: Hierarchical Deterministic wallet generation and import
- **Transaction Signing**: Sign transactions without exposing private keys
- **Address Derivation**: Generate addresses with custom derivation paths
- **RESTful API**: Full integration with Vault's HTTP API

## Supported Blockchain Networks

The plugin supports all blockchain networks available in Trust Wallet Core. Here are the most commonly used:

| Blockchain          | Symbol | Coin Type | Address Format      |
| ------------------- | ------ | --------- | ------------------- |
| Bitcoin             | BTC    | 0         | P2PKH, P2SH, Bech32 |
| Ethereum            | ETH    | 60        | 0x... (EIP-55)      |
| Solana              | SOL    | 501       | Base58              |
| Binance Smart Chain | BSC    | 60        | 0x... (same as ETH) |
| Polygon             | MATIC  | 60        | 0x... (same as ETH) |
| Bitcoin Cash        | BCH    | 145       | CashAddr            |
| Litecoin            | LTC    | 2         | L... or M...        |
| Dogecoin            | DOGE   | 3         | D...                |
| Cardano             | ADA    | 1815      | addr1...            |
| Polkadot            | DOT    | 354       | 1...                |

For a complete list of supported networks, see the [Trust Wallet Core documentation](https://developer.trustwallet.com/wallet-core/integration-guide/blockchain-support).

## Prerequisites

- Go 1.25 or later
- HashiCorp Vault 1.12 or later
- Trust Wallet Core library
- Make (optional, for using Makefile)

## Building

### Using Docker (Recommended)

Build with Trust Wallet Core from source:

```bash
# Build using Dockerfile.build (includes Trust Wallet Core)
make docker-build-build
```

The binary will be created at `bin/trust-vault-plugin`.

### Using Make

```bash
# Build the plugin (requires Trust Wallet Core installed locally)
make build

# Build for multiple platforms (Linux and macOS)
make build-all

# Run tests
make test

# Generate coverage report
make test-coverage

# Calculate SHA256 checksum
make checksum
```

### Testing in Docker

Test your code without installing Trust Wallet Core locally:

```bash
# Start interactive development container
make docker-dev

# Inside container, test your code:
go build ./wallet
go test ./...
# Create test files as needed
```

### Using build script

```bash
# Build for current platform
./build.sh

# Build for specific platform
GOOS=linux GOARCH=amd64 ./build.sh
```

### Manual build

```bash
# Create build directory
mkdir -p bin

# Build the plugin
go build -o bin/trust-vault-plugin ./cmd/trust-vault

# Calculate SHA256
sha256sum bin/trust-vault-plugin | awk '{print $1}' > bin/trust-vault-plugin.sha256
```

## Installation

### Option 1: Using Make

```bash
# Build and install (requires sudo)
make install
```

### Option 2: Using registration script

```bash
# Build the plugin
make build

# Register with Vault
./scripts/register-plugin.sh

# Enable the secrets engine
./scripts/enable-plugin.sh
```

### Option 3: Manual installation

```bash
# Copy plugin to Vault's plugin directory
sudo cp bin/trust-vault-plugin /etc/vault/plugins/
sudo chmod +x /etc/vault/plugins/trust-vault-plugin

# Calculate SHA256
SHA256=$(sha256sum bin/trust-vault-plugin | awk '{print $1}')

# Register plugin with Vault
vault plugin register \
  -sha256="$SHA256" \
  -command="trust-vault-plugin" \
  secret trust-vault

# Enable the secrets engine
vault secrets enable -path=trust-vault trust-vault
```

## Docker Deployment

### Using Docker Compose

```bash
# Build and start Vault with plugin
docker-compose up -d vault

# View logs
docker-compose logs -f vault

# Stop
docker-compose down
```

### Using Docker directly

```bash
# Build image
docker build -t trust-vault:latest .

# Run container
docker run -d \
  -p 8200:8200 \
  --name trust-vault \
  --cap-add=IPC_LOCK \
  trust-vault:latest

# Access Vault
export VAULT_ADDR='http://localhost:8200'
```

## Usage

### Create a Wallet

```bash
# Create an Ethereum wallet
vault write trust-vault/wallets/my-eth-wallet coin_type=60

# Create a Bitcoin wallet
vault write trust-vault/wallets/my-btc-wallet coin_type=0

# Import wallet from mnemonic
vault write trust-vault/wallets/imported-wallet \
  coin_type=60 \
  mnemonic="word1 word2 word3 ... word12"
```

### Get Wallet Information

```bash
# Get wallet details (no private keys)
vault read trust-vault/wallets/my-eth-wallet
```

### Get Address

```bash
# Get address for a specific coin type
vault read trust-vault/wallets/my-eth-wallet/addresses/60

# Get address with custom derivation path
vault read trust-vault/wallets/my-eth-wallet/addresses/60 \
  derivation_path="m/44'/60'/0'/0/1"
```

### Sign Transaction

```bash
# Sign a transaction
vault write trust-vault/wallets/my-eth-wallet/sign \
  tx_data=@transaction.json

# Example transaction data (base64 encoded)
echo '{"to":"0x...","value":"1000000000000000000","data":"0x"}' | base64 > transaction.json
vault write trust-vault/wallets/my-eth-wallet/sign \
  tx_data=@transaction.json
```

### List Wallets

```bash
# List all wallets
vault list trust-vault/wallets
```

### Delete Wallet

```bash
# Delete a wallet
vault delete trust-vault/wallets/my-eth-wallet
```

## API Endpoints

| Method | Path                                         | Description               |
| ------ | -------------------------------------------- | ------------------------- |
| POST   | `/trust-vault/wallets/:name`                 | Create a new wallet       |
| GET    | `/trust-vault/wallets/:name`                 | Get wallet information    |
| DELETE | `/trust-vault/wallets/:name`                 | Delete a wallet           |
| LIST   | `/trust-vault/wallets`                       | List all wallets          |
| POST   | `/trust-vault/wallets/:name/sign`            | Sign a transaction        |
| GET    | `/trust-vault/wallets/:name/addresses/:coin` | Get address for coin type |

## Configuration

### Vault Configuration

Add to your Vault configuration file:

```hcl
plugin_directory = "/etc/vault/plugins"
```

### Plugin Configuration

```bash
# Configure plugin settings
vault write trust-vault/config \
  max_wallets=1000 \
  default_coin_type=60
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./service -run TestCreateWallet
```

### Code Formatting

```bash
# Format code
make fmt

# Run linter
make lint
```

### Development Build

```bash
# Build with debug symbols
make dev
```

## Security Considerations

- Private keys are encrypted at rest using Vault's encryption
- Keys are decrypted only in memory during operations
- Memory is cleared immediately after use
- No logging of sensitive data (private keys, mnemonics)
- Input validation for all user inputs
- Leverage Vault's policy system for access control

## Troubleshooting

### Plugin not found

```bash
# Verify plugin is in the correct directory
ls -la /etc/vault/plugins/trust-vault-plugin

# Check plugin registration
vault plugin list secret | grep trust-vault
```

### SHA256 mismatch

```bash
# Recalculate and re-register
make checksum
./scripts/register-plugin.sh
```

### Connection refused

```bash
# Check Vault is running
vault status

# Verify VAULT_ADDR
echo $VAULT_ADDR
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Submit a pull request

## License

See LICENSE file for details.

## References

- [HashiCorp Vault Plugin System](https://www.vaultproject.io/docs/plugins)
- [Trust Wallet Core](https://github.com/trustwallet/wallet-core)
- [Vault Plugin Development](https://www.vaultproject.io/docs/plugins/plugin-development)
