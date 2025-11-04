package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/logical"
)

var (
	// ErrWalletNotFound is returned when a wallet doesn't exist
	ErrWalletNotFound = errors.New("wallet not found")
	// ErrWalletExists is returned when attempting to create a duplicate wallet
	ErrWalletExists = errors.New("wallet already exists")
	// ErrEncryptionFailed is returned when encryption operations fail
	ErrEncryptionFailed = errors.New("encryption failed")
	// ErrDecryptionFailed is returned when decryption operations fail
	ErrDecryptionFailed = errors.New("decryption failed")
)

// Wallet represents a cryptocurrency wallet with its metadata and key material
type Wallet struct {
	Name       string    `json:"name"`
	CoinType   uint32    `json:"coin_type"`
	Mnemonic   string    `json:"-"` // Never serialized to JSON
	PrivateKey []byte    `json:"-"` // Never serialized to JSON
	PublicKey  string    `json:"public_key"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at"`
}

// encryptedWallet is the internal representation with encrypted sensitive fields
type encryptedWallet struct {
	Name              string    `json:"name"`
	CoinType          uint32    `json:"coin_type"`
	MnemonicEncrypted string    `json:"mnemonic_encrypted"`
	PrivateKeyEncrypted string  `json:"private_key_encrypted"`
	PublicKey         string    `json:"public_key"`
	Address           string    `json:"address"`
	CreatedAt         time.Time `json:"created_at"`
}

// StorageService handles encrypted storage of wallet data
type StorageService struct {
	storage       logical.Storage
	encryptionKey []byte
	logger        hclog.Logger
}

// NewStorageService creates a new storage service instance
func NewStorageService(storage logical.Storage, encryptionKey []byte, logger hclog.Logger) *StorageService {
	return &StorageService{
		storage:       storage,
		encryptionKey: encryptionKey,
		logger:        logger,
	}
}

// StoreWallet stores a wallet with encryption of sensitive fields
func (ss *StorageService) StoreWallet(ctx context.Context, wallet *Wallet) error {
	if wallet == nil {
		ss.logger.Error("attempted to store nil wallet")
		return errors.New("wallet cannot be nil")
	}

	ss.logger.Debug("storing wallet", "name", sanitizeName(wallet.Name))

	// Check if wallet already exists
	existing, err := ss.storage.Get(ctx, "wallets/"+wallet.Name)
	if err != nil {
		ss.logger.Error("failed to check wallet existence", "name", sanitizeName(wallet.Name), "error", err)
		return fmt.Errorf("failed to check wallet existence: %w", err)
	}
	if existing != nil {
		ss.logger.Warn("wallet already exists", "name", sanitizeName(wallet.Name))
		return ErrWalletExists
	}

	// Encrypt sensitive fields
	encrypted, err := ss.encryptWallet(wallet)
	if err != nil {
		ss.logger.Error("failed to encrypt wallet", "name", sanitizeName(wallet.Name), "error", err)
		return fmt.Errorf("failed to encrypt wallet: %w", err)
	}

	// Store encrypted wallet
	entry, err := logical.StorageEntryJSON("wallets/"+wallet.Name, encrypted)
	if err != nil {
		ss.logger.Error("failed to create storage entry", "name", sanitizeName(wallet.Name), "error", err)
		return fmt.Errorf("failed to create storage entry: %w", err)
	}

	if err := ss.storage.Put(ctx, entry); err != nil {
		ss.logger.Error("failed to store wallet", "name", sanitizeName(wallet.Name), "error", err)
		return fmt.Errorf("failed to store wallet: %w", err)
	}

	ss.logger.Info("wallet stored successfully", "name", sanitizeName(wallet.Name))

	return nil
}

// GetWallet retrieves a wallet and decrypts sensitive fields
func (ss *StorageService) GetWallet(ctx context.Context, name string) (*Wallet, error) {
	if name == "" {
		ss.logger.Warn("attempted to get wallet with empty name")
		return nil, errors.New("wallet name cannot be empty")
	}

	ss.logger.Debug("retrieving wallet", "name", sanitizeName(name))

	entry, err := ss.storage.Get(ctx, "wallets/"+name)
	if err != nil {
		ss.logger.Error("failed to retrieve wallet from storage", "name", sanitizeName(name), "error", err)
		return nil, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if entry == nil {
		ss.logger.Debug("wallet not found", "name", sanitizeName(name))
		return nil, ErrWalletNotFound
	}

	var encrypted encryptedWallet
	if err := json.Unmarshal(entry.Value, &encrypted); err != nil {
		ss.logger.Error("failed to decode wallet", "name", sanitizeName(name), "error", err)
		return nil, fmt.Errorf("failed to decode wallet: %w", err)
	}

	// Decrypt sensitive fields
	wallet, err := ss.decryptWallet(&encrypted)
	if err != nil {
		ss.logger.Error("failed to decrypt wallet", "name", sanitizeName(name), "error", err)
		return nil, fmt.Errorf("failed to decrypt wallet: %w", err)
	}

	ss.logger.Debug("wallet retrieved successfully", "name", sanitizeName(name))

	return wallet, nil
}

// DeleteWallet removes a wallet from storage
func (ss *StorageService) DeleteWallet(ctx context.Context, name string) error {
	if name == "" {
		ss.logger.Warn("attempted to delete wallet with empty name")
		return errors.New("wallet name cannot be empty")
	}

	ss.logger.Debug("deleting wallet", "name", sanitizeName(name))

	// Verify wallet exists before deletion
	entry, err := ss.storage.Get(ctx, "wallets/"+name)
	if err != nil {
		ss.logger.Error("failed to check wallet existence", "name", sanitizeName(name), "error", err)
		return fmt.Errorf("failed to check wallet existence: %w", err)
	}
	if entry == nil {
		ss.logger.Debug("wallet not found for deletion", "name", sanitizeName(name))
		return ErrWalletNotFound
	}

	// Delete the wallet
	if err := ss.storage.Delete(ctx, "wallets/"+name); err != nil {
		ss.logger.Error("failed to delete wallet", "name", sanitizeName(name), "error", err)
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	ss.logger.Info("wallet deleted successfully", "name", sanitizeName(name))

	return nil
}

// ListWallets returns a list of all wallet names with pagination support
func (ss *StorageService) ListWallets(ctx context.Context, offset, limit int) ([]string, error) {
	ss.logger.Debug("listing wallets", "offset", offset, "limit", limit)

	keys, err := ss.storage.List(ctx, "wallets/")
	if err != nil {
		ss.logger.Error("failed to list wallets", "error", err)
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	// Apply pagination
	total := len(keys)
	if offset >= total {
		ss.logger.Debug("offset exceeds total wallets", "offset", offset, "total", total)
		return []string{}, nil
	}

	end := offset + limit
	if limit <= 0 || end > total {
		end = total
	}

	result := keys[offset:end]
	ss.logger.Debug("wallets listed successfully", "count", len(result), "total", total)

	return result, nil
}

// encryptWallet encrypts sensitive fields of a wallet
func (ss *StorageService) encryptWallet(wallet *Wallet) (*encryptedWallet, error) {
	// Encrypt mnemonic
	mnemonicEncrypted, err := ss.encrypt([]byte(wallet.Mnemonic))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to encrypt mnemonic", ErrEncryptionFailed)
	}

	// Encrypt private key
	privateKeyEncrypted, err := ss.encrypt(wallet.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to encrypt private key", ErrEncryptionFailed)
	}

	return &encryptedWallet{
		Name:                wallet.Name,
		CoinType:            wallet.CoinType,
		MnemonicEncrypted:   mnemonicEncrypted,
		PrivateKeyEncrypted: privateKeyEncrypted,
		PublicKey:           wallet.PublicKey,
		Address:             wallet.Address,
		CreatedAt:           wallet.CreatedAt,
	}, nil
}

// decryptWallet decrypts sensitive fields of an encrypted wallet
func (ss *StorageService) decryptWallet(encrypted *encryptedWallet) (*Wallet, error) {
	// Decrypt mnemonic
	mnemonicBytes, err := ss.decrypt(encrypted.MnemonicEncrypted)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt mnemonic", ErrDecryptionFailed)
	}

	// Decrypt private key
	privateKey, err := ss.decrypt(encrypted.PrivateKeyEncrypted)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decrypt private key", ErrDecryptionFailed)
	}

	return &Wallet{
		Name:       encrypted.Name,
		CoinType:   encrypted.CoinType,
		Mnemonic:   string(mnemonicBytes),
		PrivateKey: privateKey,
		PublicKey:  encrypted.PublicKey,
		Address:    encrypted.Address,
		CreatedAt:  encrypted.CreatedAt,
	}, nil
}

// encrypt encrypts data using AES-GCM
func (ss *StorageService) encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(ss.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts data using AES-GCM
func (ss *StorageService) decrypt(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(ss.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetWalletMetadata retrieves wallet metadata without decrypting sensitive fields
func (ss *StorageService) GetWalletMetadata(ctx context.Context, name string) (*Wallet, error) {
	if name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}

	entry, err := ss.storage.Get(ctx, "wallets/"+name)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve wallet: %w", err)
	}
	if entry == nil {
		return nil, ErrWalletNotFound
	}

	var encrypted encryptedWallet
	if err := json.Unmarshal(entry.Value, &encrypted); err != nil {
		return nil, fmt.Errorf("failed to decode wallet: %w", err)
	}

	// Return wallet without decrypting sensitive fields
	return &Wallet{
		Name:      encrypted.Name,
		CoinType:  encrypted.CoinType,
		PublicKey: encrypted.PublicKey,
		Address:   encrypted.Address,
		CreatedAt: encrypted.CreatedAt,
	}, nil
}

// ListWalletsWithMetadata returns wallet metadata for all wallets with pagination
func (ss *StorageService) ListWalletsWithMetadata(ctx context.Context, offset, limit int) ([]*Wallet, error) {
	names, err := ss.ListWallets(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	wallets := make([]*Wallet, 0, len(names))
	for _, name := range names {
		// Remove trailing slash if present
		name = strings.TrimSuffix(name, "/")
		
		wallet, err := ss.GetWalletMetadata(ctx, name)
		if err != nil {
			// Skip wallets that can't be read
			continue
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

// sanitizeName sanitizes wallet name for logging (prevents logging sensitive data)
func sanitizeName(name string) string {
	if len(name) > 50 {
		return name[:50] + "..."
	}
	return name
}
