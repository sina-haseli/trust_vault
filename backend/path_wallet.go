package backend

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/sina-haseli/trust_vault/service"
	"github.com/sina-haseli/trust_vault/storage"
)

// pathWalletCreate returns the path configuration for creating wallets
// POST /trust-vault/wallets/:name
func (b *TrustVaultBackend) pathWalletCreate() *framework.Path {
	return &framework.Path{
		Pattern: "wallets/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Unique name for the wallet",
				Required:    true,
			},
			"coin_type": {
				Type:        framework.TypeInt,
				Description: "Coin type (e.g., 0=Bitcoin, 60=Ethereum, 501=Solana)",
				Required:    true,
			},
			"mnemonic": {
				Type:        framework.TypeString,
				Description: "Optional mnemonic phrase for importing an existing wallet",
				Required:    false,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.handleWalletCreate,
				Summary:  "Create a new cryptocurrency wallet",
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.handleWalletCreate,
				Summary:  "Create a new cryptocurrency wallet",
			},
		},
		HelpSynopsis:    "Create a new cryptocurrency wallet for the specified blockchain",
		HelpDescription: "Creates a new HD wallet using Trust Wallet Core. If a mnemonic is provided, it imports the wallet; otherwise, it generates a new one.",
	}
}

// handleWalletCreate handles wallet creation requests
func (b *TrustVaultBackend) handleWalletCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	
	// Validate wallet name
	if err := validateWalletName(name); err != nil {
		b.logger.Warn("invalid wallet name provided", "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	coinTypeRaw, ok := data.GetOk("coin_type")
	if !ok {
		b.logger.Warn("coin_type not provided in wallet creation request")
		return logical.ErrorResponse("coin_type is required"), nil
	}
	coinType := uint32(coinTypeRaw.(int))

	// Validate coin type
	if err := validateCoinType(coinType); err != nil {
		b.logger.Warn("invalid coin type provided", "coin_type", coinType, "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	mnemonic := data.Get("mnemonic").(string)
	
	// Log operation (without sensitive data)
	if mnemonic != "" {
		b.logger.Info("importing wallet", "name", sanitizeWalletName(name), "coin_type", coinType)
	} else {
		b.logger.Info("creating new wallet", "name", sanitizeWalletName(name), "coin_type", coinType)
	}

	// Create wallet
	wallet, err := b.walletService.CreateWallet(ctx, name, coinType, mnemonic)
	if err != nil {
		b.logger.Error("failed to create wallet", "name", sanitizeWalletName(name), "coin_type", coinType, "error", err)
		return b.handleError(err)
	}

	b.logger.Info("wallet created successfully", "name", sanitizeWalletName(name), "coin_type", coinType, "address", wallet.Address)

	// Return wallet metadata (no sensitive data)
	return &logical.Response{
		Data: map[string]interface{}{
			"name":       wallet.Name,
			"coin_type":  wallet.CoinType,
			"address":    wallet.Address,
			"public_key": wallet.PublicKey,
			"created_at": wallet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// pathWalletRead returns the path configuration for reading wallet metadata
// GET /trust-vault/wallets/:name
func (b *TrustVaultBackend) pathWalletRead() *framework.Path {
	return &framework.Path{
		Pattern: "wallets/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the wallet to retrieve",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleWalletRead,
				Summary:  "Read wallet metadata",
			},
		},
		HelpSynopsis:    "Retrieve wallet metadata without exposing private keys",
		HelpDescription: "Returns wallet information including address and public key, but never exposes private keys or mnemonic phrases.",
	}
}

// handleWalletRead handles wallet read requests
func (b *TrustVaultBackend) handleWalletRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	
	// Validate wallet name
	if err := validateWalletName(name); err != nil {
		b.logger.Warn("invalid wallet name provided for read", "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	b.logger.Debug("reading wallet metadata", "name", sanitizeWalletName(name))

	// Get wallet metadata
	wallet, err := b.walletService.GetWallet(ctx, name)
	if err != nil {
		b.logger.Error("failed to read wallet", "name", sanitizeWalletName(name), "error", err)
		return b.handleError(err)
	}

	b.logger.Debug("wallet metadata retrieved successfully", "name", sanitizeWalletName(name))

	// Return wallet metadata (no sensitive data)
	return &logical.Response{
		Data: map[string]interface{}{
			"name":       wallet.Name,
			"coin_type":  wallet.CoinType,
			"address":    wallet.Address,
			"public_key": wallet.PublicKey,
			"created_at": wallet.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// pathWalletDelete returns the path configuration for deleting wallets
// DELETE /trust-vault/wallets/:name
func (b *TrustVaultBackend) pathWalletDelete() *framework.Path {
	return &framework.Path{
		Pattern: "wallets/" + framework.GenericNameRegex("name"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the wallet to delete",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.handleWalletDelete,
				Summary:  "Delete a wallet",
			},
		},
		HelpSynopsis:    "Delete a wallet and all associated key material",
		HelpDescription: "Permanently removes a wallet from storage. This operation cannot be undone.",
	}
}

// handleWalletDelete handles wallet deletion requests
func (b *TrustVaultBackend) handleWalletDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	
	// Validate wallet name
	if err := validateWalletName(name); err != nil {
		b.logger.Warn("invalid wallet name provided for deletion", "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	b.logger.Info("deleting wallet", "name", sanitizeWalletName(name))

	// Delete wallet
	if err := b.walletService.DeleteWallet(ctx, name); err != nil {
		b.logger.Error("failed to delete wallet", "name", sanitizeWalletName(name), "error", err)
		return b.handleError(err)
	}

	b.logger.Info("wallet deleted successfully", "name", sanitizeWalletName(name))

	return &logical.Response{
		Data: map[string]interface{}{
			"deleted": true,
		},
	}, nil
}

// pathWalletList returns the path configuration for listing wallets
// LIST /trust-vault/wallets
func (b *TrustVaultBackend) pathWalletList() *framework.Path {
	return &framework.Path{
		Pattern: "wallets/?$",
		Fields: map[string]*framework.FieldSchema{
			"offset": {
				Type:        framework.TypeInt,
				Description: "Pagination offset (default: 0)",
				Required:    false,
				Default:     0,
			},
			"limit": {
				Type:        framework.TypeInt,
				Description: "Maximum number of wallets to return (default: 100, 0 for all)",
				Required:    false,
				Default:     100,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.handleWalletList,
				Summary:  "List all wallets",
			},
		},
		HelpSynopsis:    "List all wallet names",
		HelpDescription: "Returns a list of all wallet names stored in the plugin. Supports pagination.",
	}
}

// handleWalletList handles wallet list requests
func (b *TrustVaultBackend) handleWalletList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	offset := data.Get("offset").(int)
	limit := data.Get("limit").(int)

	// Validate pagination parameters
	if offset < 0 {
		b.logger.Warn("invalid offset provided", "offset", offset)
		return logical.ErrorResponse("offset must be non-negative"), nil
	}
	if limit < 0 {
		b.logger.Warn("invalid limit provided", "limit", limit)
		return logical.ErrorResponse("limit must be non-negative"), nil
	}

	b.logger.Debug("listing wallets", "offset", offset, "limit", limit)

	// List wallets
	wallets, err := b.walletService.ListWallets(ctx, offset, limit)
	if err != nil {
		b.logger.Error("failed to list wallets", "error", err)
		return b.handleError(err)
	}

	b.logger.Debug("wallets listed successfully", "count", len(wallets))

	return logical.ListResponse(wallets), nil
}

// pathWalletSign returns the path configuration for signing transactions
// POST /trust-vault/wallets/:name/sign
func (b *TrustVaultBackend) pathWalletSign() *framework.Path {
	return &framework.Path{
		Pattern: "wallets/" + framework.GenericNameRegex("name") + "/sign$",
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the wallet to use for signing",
				Required:    true,
			},
			"tx_data": {
				Type:        framework.TypeString,
				Description: "Base64-encoded transaction data to sign",
				Required:    true,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.handleWalletSign,
				Summary:  "Sign a transaction",
			},
		},
		HelpSynopsis:    "Sign a transaction using the wallet's private key",
		HelpDescription: "Signs transaction data using Trust Wallet Core. The transaction data should be base64-encoded and formatted according to the blockchain's requirements.",
	}
}

// handleWalletSign handles transaction signing requests
func (b *TrustVaultBackend) handleWalletSign(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	
	// Validate wallet name
	if err := validateWalletName(name); err != nil {
		b.logger.Warn("invalid wallet name provided for signing", "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	txDataEncoded := data.Get("tx_data").(string)
	if txDataEncoded == "" {
		b.logger.Warn("tx_data not provided in signing request")
		return logical.ErrorResponse("tx_data is required"), nil
	}

	// Validate transaction data length
	if len(txDataEncoded) > 1024*1024 { // 1MB limit
		b.logger.Warn("transaction data too large", "size", len(txDataEncoded))
		return logical.ErrorResponse("transaction data exceeds maximum size of 1MB"), nil
	}

	// Decode base64 transaction data
	txData, err := base64.StdEncoding.DecodeString(txDataEncoded)
	if err != nil {
		b.logger.Warn("invalid base64 transaction data", "error", err)
		return logical.ErrorResponse("invalid tx_data: must be base64-encoded"), nil
	}

	// Validate decoded transaction data
	if len(txData) == 0 {
		b.logger.Warn("empty transaction data after decoding")
		return logical.ErrorResponse("transaction data cannot be empty"), nil
	}

	b.logger.Info("signing transaction", "name", sanitizeWalletName(name), "tx_size", len(txData))

	// Sign transaction
	signature, err := b.walletService.SignTransaction(ctx, name, txData)
	if err != nil {
		b.logger.Error("failed to sign transaction", "name", sanitizeWalletName(name), "error", err)
		return b.handleError(err)
	}

	b.logger.Info("transaction signed successfully", "name", sanitizeWalletName(name), "signature_size", len(signature))

	// Return base64-encoded signature
	return &logical.Response{
		Data: map[string]interface{}{
			"signed_tx": base64.StdEncoding.EncodeToString(signature),
		},
	}, nil
}

// pathWalletAddress returns the path configuration for deriving addresses
// GET /trust-vault/wallets/:name/addresses/:coin
func (b *TrustVaultBackend) pathWalletAddress() *framework.Path {
	return &framework.Path{
		Pattern: "wallets/" + framework.GenericNameRegex("name") + "/addresses/" + framework.GenericNameRegex("coin"),
		Fields: map[string]*framework.FieldSchema{
			"name": {
				Type:        framework.TypeString,
				Description: "Name of the wallet",
				Required:    true,
			},
			"coin": {
				Type:        framework.TypeString,
				Description: "Coin type (e.g., 0=Bitcoin, 60=Ethereum, 501=Solana)",
				Required:    true,
			},
			"derivation_path": {
				Type:        framework.TypeString,
				Description: "Optional custom derivation path (e.g., m/44'/60'/0'/0/0)",
				Required:    false,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleWalletAddress,
				Summary:  "Derive an address for a specific coin type",
			},
		},
		HelpSynopsis:    "Derive a cryptocurrency address from the wallet",
		HelpDescription: "Derives an address for the specified coin type using the wallet's mnemonic. Optionally accepts a custom derivation path.",
	}
}

// handleWalletAddress handles address derivation requests
func (b *TrustVaultBackend) handleWalletAddress(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	name := data.Get("name").(string)
	
	// Validate wallet name
	if err := validateWalletName(name); err != nil {
		b.logger.Warn("invalid wallet name provided for address derivation", "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	coinStr := data.Get("coin").(string)
	if coinStr == "" {
		b.logger.Warn("coin type not provided in address derivation request")
		return logical.ErrorResponse("coin type is required"), nil
	}

	// Parse coin type
	var coinType uint32
	if _, err := fmt.Sscanf(coinStr, "%d", &coinType); err != nil {
		b.logger.Warn("invalid coin type format", "coin", coinStr, "error", err)
		return logical.ErrorResponse("invalid coin type: must be a number"), nil
	}

	// Validate coin type
	if err := validateCoinType(coinType); err != nil {
		b.logger.Warn("invalid coin type provided", "coin_type", coinType, "error", err)
		return logical.ErrorResponse(err.Error()), nil
	}

	derivationPath := data.Get("derivation_path").(string)
	
	// Validate derivation path if provided
	if derivationPath != "" {
		if err := validateDerivationPath(derivationPath); err != nil {
			b.logger.Warn("invalid derivation path", "path", derivationPath, "error", err)
			return logical.ErrorResponse(err.Error()), nil
		}
	}

	b.logger.Debug("deriving address", "name", sanitizeWalletName(name), "coin_type", coinType, "has_custom_path", derivationPath != "")

	// Derive address
	address, err := b.walletService.GetAddress(ctx, name, coinType, derivationPath)
	if err != nil {
		b.logger.Error("failed to derive address", "name", sanitizeWalletName(name), "coin_type", coinType, "error", err)
		return b.handleError(err)
	}

	b.logger.Debug("address derived successfully", "name", sanitizeWalletName(name), "coin_type", coinType, "address", address)

	return &logical.Response{
		Data: map[string]interface{}{
			"address":   address,
			"coin_type": coinType,
		},
	}, nil
}

// handleError maps service errors to appropriate HTTP responses
func (b *TrustVaultBackend) handleError(err error) (*logical.Response, error) {
	switch {
	case errors.Is(err, service.ErrWalletNotFound), errors.Is(err, storage.ErrWalletNotFound):
		resp := logical.ErrorResponse("wallet not found")
		resp.Data["http_status_code"] = 404
		return resp, nil
	case errors.Is(err, service.ErrWalletExists), errors.Is(err, storage.ErrWalletExists):
		resp := logical.ErrorResponse("wallet already exists")
		resp.Data["http_status_code"] = 409
		return resp, nil
	case errors.Is(err, service.ErrInvalidCoinType):
		return logical.ErrorResponse("invalid coin type"), nil
	case errors.Is(err, service.ErrInvalidMnemonic):
		return logical.ErrorResponse("invalid mnemonic phrase"), nil
	case errors.Is(err, service.ErrInvalidTxData):
		return logical.ErrorResponse("invalid transaction data"), nil
	case errors.Is(err, service.ErrSigningFailed):
		return logical.ErrorResponse("transaction signing failed"), nil
	case errors.Is(err, service.ErrInvalidWalletName):
		return logical.ErrorResponse("invalid wallet name"), nil
	default:
		return nil, fmt.Errorf("internal error: %w", err)
	}
}

// validateWalletName validates wallet name to prevent path traversal and ensure valid format
func validateWalletName(name string) error {
	if name == "" {
		return errors.New("wallet name is required")
	}

	if len(name) > 255 {
		return errors.New("wallet name exceeds maximum length of 255 characters")
	}

	// Check for path traversal attempts
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return errors.New("wallet name contains invalid characters")
	}

	// Check for control characters
	for _, r := range name {
		if r < 32 || r == 127 {
			return errors.New("wallet name contains invalid control characters")
		}
	}

	return nil
}

// validateCoinType validates that the coin type is supported
func validateCoinType(coinType uint32) error {
	// Supported coin types: Bitcoin (0), Ethereum (60), Solana (501)
	supportedTypes := map[uint32]bool{
		0:   true, // Bitcoin
		60:  true, // Ethereum
		501: true, // Solana
	}

	if !supportedTypes[coinType] {
		return fmt.Errorf("unsupported coin type: %d (supported: 0=Bitcoin, 60=Ethereum, 501=Solana)", coinType)
	}

	return nil
}

// validateDerivationPath validates the derivation path format
func validateDerivationPath(path string) error {
	if path == "" {
		return nil
	}

	if len(path) > 100 {
		return errors.New("derivation path exceeds maximum length of 100 characters")
	}

	// Basic validation: must start with m/ and contain only valid characters
	if !strings.HasPrefix(path, "m/") {
		return errors.New("derivation path must start with 'm/'")
	}

	// Check for valid characters (numbers, /, ', m)
	for _, r := range path {
		if !((r >= '0' && r <= '9') || r == '/' || r == '\'' || r == 'm') {
			return fmt.Errorf("derivation path contains invalid character: %c", r)
		}
	}

	return nil
}

// sanitizeWalletName sanitizes wallet name for logging (truncate if too long)
func sanitizeWalletName(name string) string {
	if len(name) > 50 {
		return name[:50] + "..."
	}
	return name
}
