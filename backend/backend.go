package backend

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/sina-haseli/trust_vault/service"
	"github.com/sina-haseli/trust_vault/storage"
)

// TrustVaultBackend implements the Vault logical.Backend interface
// for the Trust Vault plugin
type TrustVaultBackend struct {
	*framework.Backend
	walletService *service.WalletService
	logger        hclog.Logger
}

// Factory creates and initializes a new TrustVaultBackend instance
// This is called by Vault when the plugin is mounted
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := &TrustVaultBackend{}

	// Initialize logger
	b.logger = conf.Logger

	b.logger.Info("initializing Trust Vault plugin")

	// Generate encryption key for storage (32 bytes for AES-256)
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		b.logger.Error("failed to generate encryption key", "error", err)
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	b.logger.Debug("encryption key generated successfully")

	// Initialize storage service
	storageService := storage.NewStorageService(conf.StorageView, encryptionKey, b.logger)

	// Initialize wallet service
	b.walletService = service.NewWalletService(storageService, b.logger)

	// Configure backend
	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help:        "Trust Vault Plugin provides cryptocurrency wallet management through Trust Wallet Core",
		Paths: []*framework.Path{
			b.pathWalletCreate(),
			b.pathWalletRead(),
			b.pathWalletDelete(),
			b.pathWalletList(),
			b.pathWalletSign(),
			b.pathWalletAddress(),
			b.pathHealth(),
		},
	}

	if err := b.Setup(ctx, conf); err != nil {
		b.logger.Error("failed to setup backend", "error", err)
		return nil, fmt.Errorf("failed to setup backend: %w", err)
	}

	b.logger.Info("Trust Vault plugin initialized successfully")

	return b, nil
}

// pathHealth returns the path configuration for health check endpoint
// GET /trust-vault/health
func (b *TrustVaultBackend) pathHealth() *framework.Path {
	return &framework.Path{
		Pattern: "health$",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.handleHealth,
				Summary:  "Health check endpoint",
			},
		},
		HelpSynopsis:    "Check plugin health status",
		HelpDescription: "Returns the health status of the Trust Vault plugin",
	}
}

// handleHealth handles health check requests
func (b *TrustVaultBackend) handleHealth(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.logger.Debug("health check requested")

	return &logical.Response{
		Data: map[string]interface{}{
			"status":  "healthy",
			"version": "1.0.0",
		},
	}, nil
}
