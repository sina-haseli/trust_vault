# Trust Vault Plugin Design Document

## Overview

The Trust Vault Plugin is a HashiCorp Vault secrets engine plugin that provides cryptocurrency wallet management capabilities through the Trust Wallet Core library. The plugin follows Vault's plugin architecture, implementing the logical backend interface to handle wallet creation, address generation, and transaction signing for multiple blockchain networks.

The design leverages:
- **HashiCorp Vault Plugin System**: Go plugin architecture with gRPC communication
- **Trust Wallet Core**: C++ library with Go bindings for multi-chain wallet operations
- **Vault Storage Backend**: Encrypted storage for wallet key material

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Vault Core                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │           HTTP/API Layer                         │  │
│  └────────────────┬─────────────────────────────────┘  │
│                   │                                     │
│  ┌────────────────▼─────────────────────────────────┐  │
│  │         Plugin System (gRPC)                     │  │
│  └────────────────┬─────────────────────────────────┘  │
└───────────────────┼─────────────────────────────────────┘
                    │ gRPC
┌───────────────────▼─────────────────────────────────────┐
│          Trust Vault Plugin Process                     │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │         Backend (framework.Backend)            │    │
│  │  - Path Handlers                               │    │
│  │  - Request Routing                             │    │
│  └──────┬──────────────────────────┬──────────────┘    │
│         │                          │                    │
│  ┌──────▼──────────┐      ┌───────▼────────────┐      │
│  │  Wallet Service │      │  Storage Service   │      │
│  │  - Create       │◄─────┤  - Encrypt/Decrypt │      │
│  │  - Sign         │      │  - CRUD Operations │      │
│  │  - GetAddress   │      └────────────────────┘      │
│  └──────┬──────────┘                                   │
│         │                                               │
│  ┌──────▼──────────────────────────────────┐          │
│  │   Trust Wallet Core (CGO Bindings)      │          │
│  │   - Key Generation                       │          │
│  │   - Address Derivation                   │          │
│  │   - Transaction Signing                  │          │
│  └──────────────────────────────────────────┘          │
└──────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

**Wallet Creation Flow:**
```
Client → Vault API → Plugin gRPC → Backend Handler → Wallet Service
  → Trust Wallet Core (generate keys) → Storage Service (encrypt & store)
  → Response (wallet path + address)
```

**Transaction Signing Flow:**
```
Client → Vault API → Plugin gRPC → Backend Handler → Storage Service (decrypt)
  → Wallet Service → Trust Wallet Core (sign) → Response (signed tx)
```

## Components and Interfaces

### 1. Plugin Entry Point

**File**: `cmd/trust-vault/main.go`

```go
func main() {
    apiClientMeta := &api.PluginAPIClientMeta{}
    flags := apiClientMeta.FlagSet()
    flags.Parse(os.Args[1:])

    tlsConfig := apiClientMeta.GetTLSConfig()
    tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

    backend := &TrustVaultBackend{}
    
    err := plugin.Serve(&plugin.ServeOpts{
        BackendFactoryFunc: backend.Factory,
        TLSProviderFunc:    tlsProviderFunc,
    })
    
    if err != nil {
        log.Fatal(err)
    }
}
```

### 2. Backend Implementation

**File**: `backend/backend.go`

The backend implements Vault's `logical.Backend` interface:

```go
type TrustVaultBackend struct {
    *framework.Backend
    walletService *WalletService
}

func (b *TrustVaultBackend) Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
    b := &TrustVaultBackend{}
    
    b.Backend = &framework.Backend{
        BackendType: logical.TypeLogical,
        Paths: []*framework.Path{
            b.pathWalletCreate(),
            b.pathWalletRead(),
            b.pathWalletDelete(),
            b.pathWalletList(),
            b.pathWalletSign(),
            b.pathWalletAddress(),
        },
    }
    
    b.walletService = NewWalletService(conf.StorageView)
    
    return b, nil
}
```

### 3. Path Handlers

**File**: `backend/path_wallet.go`

API endpoints structure:
- `POST /trust-vault/wallets/:name` - Create wallet
- `GET /trust-vault/wallets/:name` - Get wallet info (no keys)
- `DELETE /trust-vault/wallets/:name` - Delete wallet
- `LIST /trust-vault/wallets` - List all wallets
- `POST /trust-vault/wallets/:name/sign` - Sign transaction
- `GET /trust-vault/wallets/:name/addresses/:coin` - Get address for coin type

```go
func (b *TrustVaultBackend) pathWalletCreate() *framework.Path {
    return &framework.Path{
        Pattern: "wallets/" + framework.GenericNameRegex("name"),
        Fields: map[string]*framework.FieldSchema{
            "name": {Type: framework.TypeString, Description: "Wallet name"},
            "coin_type": {Type: framework.TypeInt, Description: "Coin type (e.g., 0=Bitcoin, 60=Ethereum)"},
            "mnemonic": {Type: framework.TypeString, Description: "Optional mnemonic for import"},
        },
        Operations: map[logical.Operation]framework.OperationHandler{
            logical.CreateOperation: &framework.PathOperation{
                Callback: b.handleWalletCreate,
            },
        },
    }
}
```

### 4. Wallet Service

**File**: `service/wallet_service.go`

Core business logic for wallet operations:

```go
type WalletService struct {
    storage StorageService
}

type Wallet struct {
    Name         string    `json:"name"`
    CoinType     uint32    `json:"coin_type"`
    Mnemonic     string    `json:"-"` // Never serialized
    PrivateKey   []byte    `json:"-"` // Never serialized
    PublicKey    string    `json:"public_key"`
    Address      string    `json:"address"`
    CreatedAt    time.Time `json:"created_at"`
}

func (ws *WalletService) CreateWallet(ctx context.Context, name string, coinType uint32, mnemonic string) (*Wallet, error)
func (ws *WalletService) GetWallet(ctx context.Context, name string) (*Wallet, error)
func (ws *WalletService) DeleteWallet(ctx context.Context, name string) error
func (ws *WalletService) ListWallets(ctx context.Context) ([]string, error)
func (ws *WalletService) SignTransaction(ctx context.Context, name string, txData []byte) ([]byte, error)
func (ws *WalletService) GetAddress(ctx context.Context, name string, coinType uint32, derivationPath string) (string, error)
```

### 5. Trust Wallet Core Integration

**File**: `wallet/trustwallet.go`

Wrapper around Trust Wallet Core Go bindings:

```go
import (
    "github.com/trustwallet/wallet-core/wallet-core"
)

type TrustWalletCore struct{}

func (twc *TrustWalletCore) GenerateWallet(coinType uint32) (*WalletKeys, error) {
    // Generate mnemonic
    wallet := core.NewHDWallet(128, "") // 128 bits = 12 words
    mnemonic := wallet.Mnemonic()
    
    // Derive key for coin type
    privateKey := wallet.GetKeyForCoin(core.CoinType(coinType))
    publicKey := privateKey.GetPublicKeySecp256k1(true)
    
    // Get address
    address := core.CoinTypeConfigurationGetAccountURL(core.CoinType(coinType), publicKey.Data())
    
    return &WalletKeys{
        Mnemonic:   mnemonic,
        PrivateKey: privateKey.Data(),
        PublicKey:  publicKey.Data(),
        Address:    address,
    }, nil
}

func (twc *TrustWalletCore) ImportWallet(mnemonic string, coinType uint32) (*WalletKeys, error)

func (twc *TrustWalletCore) SignTransaction(privateKey []byte, coinType uint32, txData []byte) ([]byte, error) {
    // Use Trust Wallet Core's AnySigner
    input := // Parse txData into protobuf SigningInput for the coin type
    output := core.AnySigner.Sign(input, core.CoinType(coinType))
    return output.Encoded(), nil
}

func (twc *TrustWalletCore) DeriveAddress(privateKey []byte, coinType uint32, derivationPath string) (string, error)
```

### 6. Storage Service

**File**: `storage/storage_service.go`

Handles encrypted storage of wallet data:

```go
type StorageService struct {
    storage logical.Storage
}

func (ss *StorageService) StoreWallet(ctx context.Context, wallet *Wallet) error {
    // Encrypt sensitive fields
    encrypted := ss.encryptWallet(wallet)
    
    entry, err := logical.StorageEntryJSON("wallets/"+wallet.Name, encrypted)
    if err != nil {
        return err
    }
    
    return ss.storage.Put(ctx, entry)
}

func (ss *StorageService) GetWallet(ctx context.Context, name string) (*Wallet, error) {
    entry, err := ss.storage.Get(ctx, "wallets/"+name)
    if err != nil {
        return nil, err
    }
    if entry == nil {
        return nil, ErrWalletNotFound
    }
    
    var wallet Wallet
    if err := entry.DecodeJSON(&wallet); err != nil {
        return nil, err
    }
    
    // Decrypt sensitive fields
    return ss.decryptWallet(&wallet), nil
}

func (ss *StorageService) DeleteWallet(ctx context.Context, name string) error
func (ss *StorageService) ListWallets(ctx context.Context) ([]string, error)

// Uses Vault's encryption for sensitive data
func (ss *StorageService) encryptWallet(wallet *Wallet) *Wallet
func (ss *StorageService) decryptWallet(wallet *Wallet) *Wallet
```

## Data Models

### Wallet Storage Schema

```json
{
  "name": "my-eth-wallet",
  "coin_type": 60,
  "mnemonic_encrypted": "vault:v1:encrypted_mnemonic_data",
  "private_key_encrypted": "vault:v1:encrypted_private_key_data",
  "public_key": "0x04...",
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "created_at": "2025-11-04T10:30:00Z"
}
```

### API Request/Response Models

**Create Wallet Request:**
```json
{
  "coin_type": 60,
  "mnemonic": "optional existing mnemonic phrase..."
}
```

**Create Wallet Response:**
```json
{
  "name": "my-eth-wallet",
  "coin_type": 60,
  "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "public_key": "0x04...",
  "created_at": "2025-11-04T10:30:00Z"
}
```

**Sign Transaction Request:**
```json
{
  "tx_data": "base64_encoded_transaction_data"
}
```

**Sign Transaction Response:**
```json
{
  "signed_tx": "base64_encoded_signed_transaction"
}
```

## Error Handling

### Error Types

```go
var (
    ErrWalletNotFound = errors.New("wallet not found")
    ErrWalletExists = errors.New("wallet already exists")
    ErrInvalidCoinType = errors.New("invalid coin type")
    ErrInvalidMnemonic = errors.New("invalid mnemonic phrase")
    ErrInvalidTxData = errors.New("invalid transaction data")
    ErrSigningFailed = errors.New("transaction signing failed")
)
```

### HTTP Status Code Mapping

- 200: Success
- 400: Bad Request (invalid input, malformed data)
- 404: Not Found (wallet doesn't exist)
- 409: Conflict (wallet already exists)
- 500: Internal Server Error (Trust Wallet Core errors, storage errors)

### Error Response Format

```json
{
  "errors": ["descriptive error message"]
}
```

## Testing Strategy

### Unit Tests

1. **Wallet Service Tests** (`service/wallet_service_test.go`)
   - Test wallet creation with valid/invalid coin types
   - Test wallet import with valid/invalid mnemonics
   - Test address derivation
   - Test transaction signing with mock Trust Wallet Core
   - Test error handling for missing wallets

2. **Storage Service Tests** (`storage/storage_service_test.go`)
   - Test encryption/decryption of wallet data
   - Test CRUD operations with in-memory storage
   - Test list operations with pagination
   - Test concurrent access scenarios

3. **Trust Wallet Core Wrapper Tests** (`wallet/trustwallet_test.go`)
   - Test key generation for different coin types
   - Test mnemonic import
   - Test address derivation with different paths
   - Test transaction signing for major chains (BTC, ETH, SOL)

### Integration Tests

1. **Backend Integration Tests** (`backend/backend_test.go`)
   - Test full request/response cycle through path handlers
   - Test plugin initialization and cleanup
   - Test storage backend integration
   - Test error propagation through layers

2. **End-to-End Tests** (`e2e/plugin_test.go`)
   - Test plugin registration with Vault
   - Test complete wallet lifecycle (create, use, delete)
   - Test transaction signing workflow
   - Test multi-wallet scenarios

### Test Data

- Use deterministic mnemonics for reproducible tests
- Test with multiple coin types: Bitcoin (0), Ethereum (60), Solana (501)
- Use real transaction data from testnets for signing tests

## Security Considerations

1. **Key Material Protection**
   - Private keys and mnemonics encrypted at rest using Vault's encryption
   - Keys decrypted only in memory during operations
   - Memory cleared immediately after use (use `memguard` or similar)
   - No logging of sensitive data

2. **Access Control**
   - Leverage Vault's policy system for path-based access control
   - Separate read/write permissions for wallet operations
   - Audit logging through Vault's audit backend

3. **Input Validation**
   - Validate all coin types against Trust Wallet Core supported types
   - Validate mnemonic phrases before import
   - Validate transaction data structure before signing
   - Sanitize wallet names to prevent path traversal

4. **Dependencies**
   - Pin Trust Wallet Core version for reproducible builds
   - Regular security updates for dependencies
   - Use Go modules for dependency management

## Build and Deployment

### Build Process

```bash
# Install Trust Wallet Core dependencies
# Build plugin binary
go build -o trust-vault-plugin cmd/trust-vault/main.go

# Calculate SHA256 for Vault registration
shasum -a 256 trust-vault-plugin
```

### Vault Registration

```bash
# Copy plugin to Vault's plugin directory
cp trust-vault-plugin /etc/vault/plugins/

# Register plugin
vault plugin register \
  -sha256=<sha256sum> \
  -command=trust-vault-plugin \
  secret trust-vault

# Enable plugin
vault secrets enable -path=trust-vault trust-vault
```

### Configuration

Plugin configuration through Vault:

```bash
vault write trust-vault/config \
  max_wallets=1000 \
  default_coin_type=60
```

## Performance Considerations

1. **Caching**: Cache decrypted wallets in memory with TTL for frequently accessed wallets
2. **Connection Pooling**: Reuse Trust Wallet Core instances where possible
3. **Concurrent Operations**: Use goroutines for parallel wallet operations with proper synchronization
4. **Storage Optimization**: Index wallet names for fast lookups

## Future Enhancements

1. **Multi-signature Support**: Implement multi-sig wallet creation and signing
2. **Batch Operations**: Support batch address generation and transaction signing
3. **Webhook Notifications**: Notify external systems of wallet operations
4. **Backup/Restore**: Export/import wallet collections
5. **Hardware Wallet Integration**: Support for hardware wallet signing
