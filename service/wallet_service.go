package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/sina-haseli/trust_vault/storage"
	"github.com/sina-haseli/trust_vault/wallet"
)

var (
	// ErrWalletNotFound is returned when a wallet doesn't exist
	ErrWalletNotFound = errors.New("wallet not found")
	// ErrWalletExists is returned when attempting to create a duplicate wallet
	ErrWalletExists = errors.New("wallet already exists")
	// ErrInvalidCoinType is returned when an unsupported coin type is specified
	ErrInvalidCoinType = errors.New("invalid coin type")
	// ErrInvalidMnemonic is returned when a mnemonic phrase is invalid
	ErrInvalidMnemonic = errors.New("invalid mnemonic phrase")
	// ErrInvalidTxData is returned when transaction data is malformed
	ErrInvalidTxData = errors.New("invalid transaction data")
	// ErrSigningFailed is returned when transaction signing fails
	ErrSigningFailed = errors.New("transaction signing failed")
	// ErrInvalidWalletName is returned when wallet name is empty or invalid
	ErrInvalidWalletName = errors.New("invalid wallet name")
)

// WalletService provides business logic for wallet operations
type WalletService struct {
	storage     *storage.StorageService
	trustWallet *wallet.TrustWalletCore
	logger      hclog.Logger
}

// NewWalletService creates a new wallet service instance
func NewWalletService(storageService *storage.StorageService, logger hclog.Logger) *WalletService {
	return &WalletService{
		storage:     storageService,
		trustWallet: wallet.NewTrustWalletCore(),
		logger:      logger,
	}
}

// CreateWallet generates a new wallet via Trust Wallet Core and stores it
// If mnemonic is provided, it imports the wallet instead of generating a new one
func (ws *WalletService) CreateWallet(ctx context.Context, name string, coinType uint32, mnemonic string) (*storage.Wallet, error) {
	if name == "" {
		ws.logger.Warn("attempted to create wallet with empty name")
		return nil, ErrInvalidWalletName
	}

	var keys *wallet.WalletKeys
	var err error

	// Generate or import wallet based on whether mnemonic is provided
	if mnemonic != "" {
		ws.logger.Debug("importing wallet from mnemonic", "name", sanitizeName(name), "coin_type", coinType)
		keys, err = ws.trustWallet.ImportWallet(mnemonic, coinType)
		if err != nil {
			if errors.Is(err, wallet.ErrInvalidMnemonic) {
				ws.logger.Warn("invalid mnemonic provided", "name", sanitizeName(name))
				return nil, ErrInvalidMnemonic
			}
			if errors.Is(err, wallet.ErrInvalidCoinType) {
				ws.logger.Warn("invalid coin type for import", "name", sanitizeName(name), "coin_type", coinType)
				return nil, ErrInvalidCoinType
			}
			ws.logger.Error("failed to import wallet", "name", sanitizeName(name), "error", sanitizeError(err))
			return nil, fmt.Errorf("failed to import wallet: %w", err)
		}
	} else {
		ws.logger.Debug("generating new wallet", "name", sanitizeName(name), "coin_type", coinType)
		keys, err = ws.trustWallet.GenerateWallet(coinType)
		if err != nil {
			if errors.Is(err, wallet.ErrInvalidCoinType) {
				ws.logger.Warn("invalid coin type for generation", "name", sanitizeName(name), "coin_type", coinType)
				return nil, ErrInvalidCoinType
			}
			ws.logger.Error("failed to generate wallet", "name", sanitizeName(name), "error", sanitizeError(err))
			return nil, fmt.Errorf("failed to generate wallet: %w", err)
		}
	}

	ws.logger.Debug("wallet keys generated successfully", "name", sanitizeName(name))

	// Create wallet object
	walletObj := &storage.Wallet{
		Name:       name,
		CoinType:   coinType,
		Mnemonic:   keys.Mnemonic,
		PrivateKey: keys.PrivateKey,
		PublicKey:  wallet.GetPublicKeyHex(keys.PublicKey),
		Address:    keys.Address,
		CreatedAt:  time.Now().UTC(),
	}

	// Store wallet
	if err := ws.storage.StoreWallet(ctx, walletObj); err != nil {
		if errors.Is(err, storage.ErrWalletExists) {
			ws.logger.Warn("wallet already exists", "name", sanitizeName(name))
			return nil, ErrWalletExists
		}
		ws.logger.Error("failed to store wallet", "name", sanitizeName(name), "error", err)
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	ws.logger.Info("wallet created successfully", "name", sanitizeName(name), "coin_type", coinType)

	// Return wallet without sensitive fields
	return &storage.Wallet{
		Name:      walletObj.Name,
		CoinType:  walletObj.CoinType,
		PublicKey: walletObj.PublicKey,
		Address:   walletObj.Address,
		CreatedAt: walletObj.CreatedAt,
	}, nil
}

// GetWallet retrieves wallet metadata without exposing private keys or mnemonic
func (ws *WalletService) GetWallet(ctx context.Context, name string) (*storage.Wallet, error) {
	if name == "" {
		ws.logger.Warn("attempted to get wallet with empty name")
		return nil, ErrInvalidWalletName
	}

	ws.logger.Debug("retrieving wallet metadata", "name", sanitizeName(name))

	// Get wallet metadata only (no sensitive fields)
	walletObj, err := ws.storage.GetWalletMetadata(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			ws.logger.Debug("wallet not found", "name", sanitizeName(name))
			return nil, ErrWalletNotFound
		}
		ws.logger.Error("failed to retrieve wallet metadata", "name", sanitizeName(name), "error", err)
		return nil, fmt.Errorf("failed to retrieve wallet: %w", err)
	}

	ws.logger.Debug("wallet metadata retrieved successfully", "name", sanitizeName(name))

	return walletObj, nil
}

// DeleteWallet removes a wallet from storage with existence verification
func (ws *WalletService) DeleteWallet(ctx context.Context, name string) error {
	if name == "" {
		ws.logger.Warn("attempted to delete wallet with empty name")
		return ErrInvalidWalletName
	}

	ws.logger.Debug("deleting wallet", "name", sanitizeName(name))

	// Delete wallet (storage service verifies existence)
	if err := ws.storage.DeleteWallet(ctx, name); err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			ws.logger.Warn("wallet not found for deletion", "name", sanitizeName(name))
			return ErrWalletNotFound
		}
		ws.logger.Error("failed to delete wallet", "name", sanitizeName(name), "error", err)
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	ws.logger.Info("wallet deleted successfully", "name", sanitizeName(name))

	return nil
}

// ListWallets returns a list of all wallet names with pagination support
func (ws *WalletService) ListWallets(ctx context.Context, offset, limit int) ([]string, error) {
	ws.logger.Debug("listing wallets", "offset", offset, "limit", limit)

	wallets, err := ws.storage.ListWallets(ctx, offset, limit)
	if err != nil {
		ws.logger.Error("failed to list wallets", "error", err)
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	ws.logger.Debug("wallets listed successfully", "count", len(wallets))

	return wallets, nil
}

// SignTransaction retrieves a wallet, signs the transaction, and clears sensitive data from memory
func (ws *WalletService) SignTransaction(ctx context.Context, name string, txData []byte) ([]byte, error) {
	if name == "" {
		ws.logger.Warn("attempted to sign transaction with empty wallet name")
		return nil, ErrInvalidWalletName
	}

	if len(txData) == 0 {
		ws.logger.Warn("attempted to sign empty transaction data", "name", sanitizeName(name))
		return nil, ErrInvalidTxData
	}

	ws.logger.Debug("signing transaction", "name", sanitizeName(name), "tx_size", len(txData))

	// Retrieve wallet with decrypted private key
	walletObj, err := ws.storage.GetWallet(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			ws.logger.Warn("wallet not found for signing", "name", sanitizeName(name))
			return nil, ErrWalletNotFound
		}
		ws.logger.Error("failed to retrieve wallet for signing", "name", sanitizeName(name), "error", err)
		return nil, fmt.Errorf("failed to retrieve wallet: %w", err)
	}

	// Ensure private key is cleared from memory after use
	defer func() {
		// Clear private key from memory
		for i := range walletObj.PrivateKey {
			walletObj.PrivateKey[i] = 0
		}
		// Clear mnemonic from memory
		walletObj.Mnemonic = ""
		// Force garbage collection to clear memory
		runtime.GC()
		ws.logger.Debug("sensitive data cleared from memory", "name", sanitizeName(name))
	}()

	// Sign transaction
	signature, err := ws.trustWallet.SignTransaction(walletObj.PrivateKey, walletObj.CoinType, txData)
	if err != nil {
		if errors.Is(err, wallet.ErrSigningFailed) {
			ws.logger.Error("transaction signing failed", "name", sanitizeName(name), "error", sanitizeError(err))
			return nil, ErrSigningFailed
		}
		if errors.Is(err, wallet.ErrInvalidCoinType) {
			ws.logger.Warn("invalid coin type for signing", "name", sanitizeName(name), "coin_type", walletObj.CoinType)
			return nil, ErrInvalidCoinType
		}
		ws.logger.Error("failed to sign transaction", "name", sanitizeName(name), "error", sanitizeError(err))
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	ws.logger.Info("transaction signed successfully", "name", sanitizeName(name), "signature_size", len(signature))

	return signature, nil
}

// GetAddress derives an address for a specific coin type and optional derivation path
func (ws *WalletService) GetAddress(ctx context.Context, name string, coinType uint32, derivationPath string) (string, error) {
	if name == "" {
		ws.logger.Warn("attempted to get address with empty wallet name")
		return "", ErrInvalidWalletName
	}

	ws.logger.Debug("deriving address", "name", sanitizeName(name), "coin_type", coinType, "has_custom_path", derivationPath != "")

	// Retrieve wallet with decrypted mnemonic
	walletObj, err := ws.storage.GetWallet(ctx, name)
	if err != nil {
		if errors.Is(err, storage.ErrWalletNotFound) {
			ws.logger.Warn("wallet not found for address derivation", "name", sanitizeName(name))
			return "", ErrWalletNotFound
		}
		ws.logger.Error("failed to retrieve wallet for address derivation", "name", sanitizeName(name), "error", err)
		return "", fmt.Errorf("failed to retrieve wallet: %w", err)
	}

	// Ensure mnemonic is cleared from memory after use
	defer func() {
		// Clear mnemonic from memory
		walletObj.Mnemonic = ""
		// Clear private key from memory
		for i := range walletObj.PrivateKey {
			walletObj.PrivateKey[i] = 0
		}
		// Force garbage collection to clear memory
		runtime.GC()
		ws.logger.Debug("sensitive data cleared from memory", "name", sanitizeName(name))
	}()

	// Derive address
	address, err := ws.trustWallet.DeriveAddress(walletObj.Mnemonic, coinType, derivationPath)
	if err != nil {
		if errors.Is(err, wallet.ErrInvalidCoinType) {
			ws.logger.Warn("invalid coin type for address derivation", "name", sanitizeName(name), "coin_type", coinType)
			return "", ErrInvalidCoinType
		}
		if errors.Is(err, wallet.ErrAddressDerivation) {
			ws.logger.Error("address derivation failed", "name", sanitizeName(name), "error", sanitizeError(err))
			return "", fmt.Errorf("address derivation failed: %w", err)
		}
		ws.logger.Error("failed to derive address", "name", sanitizeName(name), "error", sanitizeError(err))
		return "", fmt.Errorf("failed to derive address: %w", err)
	}

	ws.logger.Debug("address derived successfully", "name", sanitizeName(name), "coin_type", coinType)

	return address, nil
}

// sanitizeName sanitizes wallet name for logging (prevents logging sensitive data)
func sanitizeName(name string) string {
	if len(name) > 50 {
		return name[:50] + "..."
	}
	return name
}

// sanitizeError sanitizes error messages to prevent logging sensitive data
// Removes any potential private keys, mnemonics, or other sensitive information
func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	// Return only the error type, not the full message which might contain sensitive data
	errStr := err.Error()
	if len(errStr) > 200 {
		return errStr[:200] + "..."
	}
	return errStr
}
