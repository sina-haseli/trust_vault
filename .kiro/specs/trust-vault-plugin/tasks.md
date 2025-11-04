# Implementation Plan

- [ ] 1. Initialize project structure and dependencies









  - Create Go module with `go mod init github.com/sina-haseli/trust_vault`
  - Add HashiCorp Vault SDK dependencies (`github.com/hashicorp/vault/sdk`)
  - Add Trust Wallet Core Go bindings dependency
  - Create directory structure: `cmd/trust-vault/`, `backend/`, `service/`, `storage/`, `wallet/`
  - Create `.gitignore` for Go projects
  - Create `README.md` with project overview and setup instructions
  - _Requirements: 1.1, 1.2_
-

- [x] 2. Implement Trust Wallet Core wrapper





  - Create `wallet/trustwallet.go` with `TrustWalletCore` struct
  - Implement `GenerateWallet()` method to create new HD wallets with mnemonic generation
  - Implement `ImportWallet()` method to restore wallets from existing mnemonics
  - Implement `DeriveAddress()` method for address derivation with custom paths
  - Implement `SignTransaction()` method using Trust Wallet Core's AnySigner
  - Add coin type constants for Bitcoin (0), Ethereum (60), and Solana (501)
  - _Requirements: 2.1, 2.2, 2.5, 3.2, 4.2_

- [ ]* 2.1 Write unit tests for Trust Wallet Core wrapper
  - Test wallet generation for multiple coin types
  - Test mnemonic import with valid and invalid phrases
  - Test address derivation with different paths
  - Test transaction signing with sample transaction data
  - _Requirements: 2.1, 2.2, 3.2, 4.2_

- [x] 3. Implement storage service with encryption





  - Create `storage/storage_service.go` with `StorageService` struct
  - Implement `StoreWallet()` method with encryption of sensitive fields (mnemonic, private key)
  - Implement `GetWallet()` method with decryption
  - Implement `DeleteWallet()` method
  - Implement `ListWallets()` method with pagination support
  - Add helper methods `encryptWallet()` and `decryptWallet()` using Vault's encryption
  - Define `Wallet` struct with JSON tags (excluding sensitive fields from serialization)
  - _Requirements: 5.1, 5.2, 5.5, 6.1, 6.2, 7.1, 7.2_

- [ ]* 3.1 Write unit tests for storage service
  - Test wallet storage and retrieval with in-memory storage
  - Test encryption/decryption of sensitive fields
  - Test list operation with multiple wallets
  - Test delete operation and error handling for non-existent wallets
  - _Requirements: 5.1, 5.5, 6.1, 7.1_

- [x] 4. Implement wallet service business logic





  - Create `service/wallet_service.go` with `WalletService` struct
  - Implement `CreateWallet()` method that generates wallet via Trust Wallet Core and stores it
  - Implement `GetWallet()` method that retrieves wallet metadata (no private keys)
  - Implement `DeleteWallet()` method with existence verification
  - Implement `ListWallets()` method
  - Implement `SignTransaction()` method that retrieves wallet, signs transaction, and clears memory
  - Implement `GetAddress()` method for address derivation
  - Add error definitions: `ErrWalletNotFound`, `ErrWalletExists`, `ErrInvalidCoinType`, etc.
  - _Requirements: 2.1, 2.2, 2.3, 3.1, 3.2, 4.1, 4.2, 5.3, 6.1, 7.1_

- [ ]* 4.1 Write unit tests for wallet service
  - Test wallet creation with valid and invalid coin types
  - Test wallet import with mnemonic
  - Test address retrieval
  - Test transaction signing workflow
  - Test error handling for duplicate wallets and missing wallets
  - _Requirements: 2.1, 2.2, 3.1, 4.1, 4.4_

- [x] 5. Implement Vault backend and path handlers





  - Create `backend/backend.go` with `TrustVaultBackend` struct implementing `logical.Backend`
  - Implement `Factory()` method to initialize backend with wallet service
  - Create `backend/path_wallet.go` with path handler definitions
  - Implement `pathWalletCreate()` for POST `/wallets/:name` endpoint
  - Implement `pathWalletRead()` for GET `/wallets/:name` endpoint
  - Implement `pathWalletDelete()` for DELETE `/wallets/:name` endpoint
  - Implement `pathWalletList()` for LIST `/wallets` endpoint
  - Implement `pathWalletSign()` for POST `/wallets/:name/sign` endpoint
  - Implement `pathWalletAddress()` for GET `/wallets/:name/addresses/:coin` endpoint
  - Add request/response handling with proper field schemas
  - Add HTTP status code mapping for errors
  - _Requirements: 1.2, 1.5, 2.1, 2.3, 3.1, 3.5, 4.1, 4.4, 6.1, 7.1, 7.3, 8.1_

- [ ]* 5.1 Write integration tests for backend
  - Test complete request/response cycle for wallet creation
  - Test wallet read, delete, and list operations
  - Test transaction signing endpoint
  - Test address generation endpoint
  - Test error responses and status codes
  - _Requirements: 1.5, 2.1, 3.1, 4.1, 8.1_

- [x] 6. Implement plugin entry point and main function





  - Create `cmd/trust-vault/main.go` with plugin entry point
  - Implement `main()` function with plugin.Serve() call
  - Configure TLS provider for secure gRPC communication
  - Add backend factory function registration
  - Add command-line flag parsing for plugin configuration
  - _Requirements: 1.1, 1.3_

- [x] 7. Add error handling and logging




  - Implement structured logging throughout all components using Vault's logger
  - Add error wrapping with context for Trust Wallet Core errors
  - Implement sanitization to prevent logging of sensitive data (private keys, mnemonics)
  - Add validation for all user inputs (coin types, wallet names, transaction data)
  - Implement health check endpoint in backend
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_
-

- [ ] 8. Create build and deployment scripts




  - Create `Makefile` with build, test, and install targets
  - Add build script that compiles plugin binary
  - Add script to calculate SHA256 checksum for Vault registration
  - Create `scripts/register-plugin.sh` for Vault plugin registration
  - Create `scripts/enable-plugin.sh` for enabling the secrets engine
  - Add Docker support with `Dockerfile` for containerized Vault with plugin
  - _Requirements: 1.1, 1.2_

- [x] 9. Add documentation and examples




  - Update `README.md` with installation instructions
  - Add API usage examples for all endpoints (create, read, sign, etc.)
  - Document supported coin types and their IDs
  - Add example transaction signing for Bitcoin, Ethereum, and Solana
  - Create `docs/` directory with detailed API reference
  - Add troubleshooting guide for common issues
  - _Requirements: 1.2, 2.4, 4.5_

- [ ] 10. Implement security hardening
  - Add input validation for wallet names to prevent path traversal
  - Implement memory clearing for private keys after use
  - Add rate limiting configuration options
  - Ensure all API responses exclude private keys and mnemonics
  - Add validation for mnemonic phrases before import
  - Implement transaction data validation before signing
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 4.4_
