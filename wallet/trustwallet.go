package wallet

// NOTE: This file uses CGO to interface with Trust Wallet Core C library.
// The linter errors you see locally are EXPECTED when Trust Wallet Core headers
// are not installed on your local machine. These errors will NOT affect the
// Docker build, which includes all necessary libraries and headers.
//
// To build successfully, use Docker:
//   docker-compose build build
//
// For local development without Docker, you need to:
//   1. Build Trust Wallet Core from source
//   2. Install headers to /usr/local/include
//   3. Install libraries to /usr/local/lib

// #cgo CFLAGS: -I${SRCDIR}/../../third_party/wallet-core/include -I/usr/local/include
// #cgo LDFLAGS: -L/usr/local/lib -lTrustWalletCore -lwallet_core_rs -lTrezorCrypto -lprotobuf -lstdc++ -lm -lpthread
// #include <TrustWalletCore/TWHDWallet.h>
// #include <TrustWalletCore/TWPrivateKey.h>
// #include <TrustWalletCore/TWPublicKey.h>
// #include <TrustWalletCore/TWCoinType.h>
// #include <TrustWalletCore/TWCurve.h>
// #include <TrustWalletCore/TWAnyAddress.h>
// #include <TrustWalletCore/TWString.h>
// #include <TrustWalletCore/TWData.h>
// #include <TrustWalletCore/TWMnemonic.h>
// #include <stdlib.h>
import "C"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"unsafe"
)

// Coin type constants for supported blockchains
const (
	CoinTypeBitcoin  uint32 = 0   // TWCoinTypeBitcoin
	CoinTypeEthereum uint32 = 60  // TWCoinTypeEthereum
	CoinTypeSolana   uint32 = 501 // TWCoinTypeSolana
)

var (
	ErrInvalidMnemonic     = errors.New("invalid mnemonic phrase")
	ErrInvalidCoinType     = errors.New("invalid coin type")
	ErrKeyGenerationFailed = errors.New("key generation failed")
	ErrSigningFailed       = errors.New("transaction signing failed")
	ErrAddressDerivation   = errors.New("address derivation failed")
)

// WalletKeys contains the key material for a wallet
type WalletKeys struct {
	Mnemonic   string
	PrivateKey []byte
	PublicKey  []byte
	Address    string
}

// TrustWalletCore wraps Trust Wallet Core functionality
type TrustWalletCore struct{}

// NewTrustWalletCore creates a new Trust Wallet Core wrapper instance
func NewTrustWalletCore() *TrustWalletCore {
	return &TrustWalletCore{}
}

// GenerateWallet generates a new HD wallet for the specified coin type
// It creates a new mnemonic phrase and derives keys for the given blockchain
func (twc *TrustWalletCore) GenerateWallet(coinType uint32) (*WalletKeys, error) {
	if !twc.isValidCoinType(coinType) {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCoinType, coinType)
	}

	// Generate a new HD wallet with 128 bits (12 words)
	emptyPassphrase := C.TWStringCreateWithUTF8Bytes(C.CString(""))
	defer C.TWStringDelete(emptyPassphrase)

	wallet := C.TWHDWalletCreate(128, emptyPassphrase)
	if wallet == nil {
		return nil, fmt.Errorf("%w: failed to create HD wallet", ErrKeyGenerationFailed)
	}
	defer C.TWHDWalletDelete(wallet)

	// Get mnemonic
	mnemonicTW := C.TWHDWalletMnemonic(wallet)
	if mnemonicTW == nil {
		return nil, fmt.Errorf("%w: empty mnemonic generated", ErrKeyGenerationFailed)
	}
	defer C.TWStringDelete(mnemonicTW)
	mnemonic := C.GoString(C.TWStringUTF8Bytes(mnemonicTW))

	// Derive key for the specified coin type
	privateKey := C.TWHDWalletGetKeyForCoin(wallet, coinType)
	if privateKey == nil {
		return nil, fmt.Errorf("%w: failed to derive key for coin type %d", ErrKeyGenerationFailed, coinType)
	}
	defer C.TWPrivateKeyDelete(privateKey)

	// Get private key data
	privateKeyData := C.TWPrivateKeyData(privateKey)
	if privateKeyData == nil {
		return nil, fmt.Errorf("%w: failed to get private key data", ErrKeyGenerationFailed)
	}
	defer C.TWDataDelete(privateKeyData)

	privateKeyBytes := C.GoBytes(unsafe.Pointer(C.TWDataBytes(privateKeyData)), C.int(C.TWDataSize(privateKeyData)))

	// Get public key
	publicKey := C.TWPrivateKeyGetPublicKeySecp256k1(privateKey, true)
	if publicKey == nil {
		return nil, fmt.Errorf("%w: failed to derive public key", ErrKeyGenerationFailed)
	}
	defer C.TWPublicKeyDelete(publicKey)

	// Get public key data
	publicKeyData := C.TWPublicKeyData(publicKey)
	if publicKeyData == nil {
		return nil, fmt.Errorf("%w: failed to get public key data", ErrKeyGenerationFailed)
	}
	defer C.TWDataDelete(publicKeyData)

	publicKeyBytes := C.GoBytes(unsafe.Pointer(C.TWDataBytes(publicKeyData)), C.int(C.TWDataSize(publicKeyData)))

	// Derive address for the coin type
	address, err := twc.getAddressForCoinType(publicKey, coinType)
	if err != nil {
		return nil, err
	}

	return &WalletKeys{
		Mnemonic:   mnemonic,
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Address:    address,
	}, nil
}

// ImportWallet imports an existing wallet from a mnemonic phrase
// It validates the mnemonic and derives keys for the specified coin type
func (twc *TrustWalletCore) ImportWallet(mnemonic string, coinType uint32) (*WalletKeys, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("%w: empty mnemonic", ErrInvalidMnemonic)
	}

	if !twc.isValidCoinType(coinType) {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCoinType, coinType)
	}

	// Validate mnemonic
	mnemonicTW := C.TWStringCreateWithUTF8Bytes(C.CString(mnemonic))
	defer C.TWStringDelete(mnemonicTW)

	if !C.TWMnemonicIsValid(mnemonicTW) {
		return nil, fmt.Errorf("%w: mnemonic validation failed", ErrInvalidMnemonic)
	}

	// Import wallet from mnemonic
	emptyPassphrase := C.TWStringCreateWithUTF8Bytes(C.CString(""))
	defer C.TWStringDelete(emptyPassphrase)

	wallet := C.TWHDWalletCreateWithMnemonic(mnemonicTW, emptyPassphrase)
	if wallet == nil {
		return nil, fmt.Errorf("%w: failed to import wallet", ErrInvalidMnemonic)
	}
	defer C.TWHDWalletDelete(wallet)

	// Derive key for the specified coin type
	privateKey := C.TWHDWalletGetKeyForCoin(wallet, coinType)
	if privateKey == nil {
		return nil, fmt.Errorf("%w: failed to derive key for coin type %d", ErrKeyGenerationFailed, coinType)
	}
	defer C.TWPrivateKeyDelete(privateKey)

	// Get private key data
	privateKeyData := C.TWPrivateKeyData(privateKey)
	if privateKeyData == nil {
		return nil, fmt.Errorf("%w: failed to get private key data", ErrKeyGenerationFailed)
	}
	defer C.TWDataDelete(privateKeyData)

	privateKeyBytes := C.GoBytes(unsafe.Pointer(C.TWDataBytes(privateKeyData)), C.int(C.TWDataSize(privateKeyData)))

	// Get public key
	publicKey := C.TWPrivateKeyGetPublicKeySecp256k1(privateKey, true)
	if publicKey == nil {
		return nil, fmt.Errorf("%w: failed to derive public key", ErrKeyGenerationFailed)
	}
	defer C.TWPublicKeyDelete(publicKey)

	// Get public key data
	publicKeyData := C.TWPublicKeyData(publicKey)
	if publicKeyData == nil {
		return nil, fmt.Errorf("%w: failed to get public key data", ErrKeyGenerationFailed)
	}
	defer C.TWDataDelete(publicKeyData)

	publicKeyBytes := C.GoBytes(unsafe.Pointer(C.TWDataBytes(publicKeyData)), C.int(C.TWDataSize(publicKeyData)))

	// Derive address for the coin type
	address, err := twc.getAddressForCoinType(publicKey, coinType)
	if err != nil {
		return nil, err
	}

	return &WalletKeys{
		Mnemonic:   mnemonic,
		PrivateKey: privateKeyBytes,
		PublicKey:  publicKeyBytes,
		Address:    address,
	}, nil
}

// DeriveAddress derives an address for a specific coin type and derivation path
// If derivationPath is empty, it uses the default path for the coin type
func (twc *TrustWalletCore) DeriveAddress(mnemonic string, coinType uint32, derivationPath string) (string, error) {
	if mnemonic == "" {
		return "", fmt.Errorf("%w: empty mnemonic", ErrInvalidMnemonic)
	}

	if !twc.isValidCoinType(coinType) {
		return "", fmt.Errorf("%w: %d", ErrInvalidCoinType, coinType)
	}

	// Validate mnemonic
	mnemonicTW := C.TWStringCreateWithUTF8Bytes(C.CString(mnemonic))
	defer C.TWStringDelete(mnemonicTW)

	if !C.TWMnemonicIsValid(mnemonicTW) {
		return "", fmt.Errorf("%w: mnemonic validation failed", ErrInvalidMnemonic)
	}

	// Import wallet from mnemonic
	emptyPassphrase := C.TWStringCreateWithUTF8Bytes(C.CString(""))
	defer C.TWStringDelete(emptyPassphrase)

	wallet := C.TWHDWalletCreateWithMnemonic(mnemonicTW, emptyPassphrase)
	if wallet == nil {
		return "", fmt.Errorf("%w: failed to import wallet", ErrInvalidMnemonic)
	}
	defer C.TWHDWalletDelete(wallet)

	var privateKey *C.struct_TWPrivateKey
	if derivationPath != "" {
		// Use custom derivation path
		pathTW := C.TWStringCreateWithUTF8Bytes(C.CString(derivationPath))
		defer C.TWStringDelete(pathTW)

		privateKey = C.TWHDWalletGetKey(wallet, coinType, pathTW)
		if privateKey == nil {
			return "", fmt.Errorf("%w: failed to derive key for path %s", ErrAddressDerivation, derivationPath)
		}
	} else {
		// Use default derivation path
		privateKey = C.TWHDWalletGetKeyForCoin(wallet, coinType)
		if privateKey == nil {
			return "", fmt.Errorf("%w: failed to derive key for coin type %d", ErrAddressDerivation, coinType)
		}
	}
	defer C.TWPrivateKeyDelete(privateKey)

	// Get public key
	publicKey := C.TWPrivateKeyGetPublicKeySecp256k1(privateKey, true)
	if publicKey == nil {
		return "", fmt.Errorf("%w: failed to derive public key", ErrAddressDerivation)
	}
	defer C.TWPublicKeyDelete(publicKey)

	// Derive address for the coin type
	address, err := twc.getAddressForCoinType(publicKey, coinType)
	if err != nil {
		return "", err
	}

	return address, nil
}

// SignTransaction signs a transaction using the private key for the specified coin type
// The txData should be the serialized transaction data appropriate for the blockchain
func (twc *TrustWalletCore) SignTransaction(privateKey []byte, coinType uint32, txData []byte) ([]byte, error) {
	if len(privateKey) == 0 {
		return nil, fmt.Errorf("%w: empty private key", ErrSigningFailed)
	}

	if len(txData) == 0 {
		return nil, fmt.Errorf("%w: empty transaction data", ErrSigningFailed)
	}

	if !twc.isValidCoinType(coinType) {
		return nil, fmt.Errorf("%w: %d", ErrInvalidCoinType, coinType)
	}

	// Create private key from bytes
	privateKeyData := C.TWDataCreateWithBytes((*C.uint8_t)(unsafe.Pointer(&privateKey[0])), C.size_t(len(privateKey)))
	if privateKeyData == nil {
		return nil, fmt.Errorf("%w: failed to create private key data", ErrSigningFailed)
	}
	defer C.TWDataDelete(privateKeyData)

	privKey := C.TWPrivateKeyCreateWithData(privateKeyData)
	if privKey == nil {
		return nil, fmt.Errorf("%w: failed to create private key", ErrSigningFailed)
	}
	defer C.TWPrivateKeyDelete(privKey)

	// Create transaction data
	txDataTW := C.TWDataCreateWithBytes((*C.uint8_t)(unsafe.Pointer(&txData[0])), C.size_t(len(txData)))
	if txDataTW == nil {
		return nil, fmt.Errorf("%w: failed to create transaction data", ErrSigningFailed)
	}
	defer C.TWDataDelete(txDataTW)

	// Sign the transaction data using SECP256k1 curve
	signature := C.TWPrivateKeySign(privKey, txDataTW, C.TWCurveSECP256k1)
	if signature == nil {
		return nil, fmt.Errorf("%w: signature generation failed", ErrSigningFailed)
	}
	defer C.TWDataDelete(signature)

	signatureBytes := C.GoBytes(unsafe.Pointer(C.TWDataBytes(signature)), C.int(C.TWDataSize(signature)))
	if len(signatureBytes) == 0 {
		return nil, fmt.Errorf("%w: empty signature generated", ErrSigningFailed)
	}

	return signatureBytes, nil
}

// isValidCoinType checks if the coin type is supported
func (twc *TrustWalletCore) isValidCoinType(coinType uint32) bool {
	// For now, we explicitly support Bitcoin, Ethereum, and Solana
	// In a production implementation, this could be expanded to check against
	// all coin types supported by Trust Wallet Core
	supportedTypes := map[uint32]bool{
		uint32(CoinTypeBitcoin):  true,
		uint32(CoinTypeEthereum): true,
		uint32(CoinTypeSolana):   true,
	}
	return supportedTypes[coinType]
}

// getAddressForCoinType derives the address from a public key for a specific coin type
func (twc *TrustWalletCore) getAddressForCoinType(publicKey *C.struct_TWPublicKey, coinType uint32) (string, error) {
	// Use TWAnyAddress for all coin types
	anyAddress := C.TWAnyAddressCreateWithPublicKey(publicKey, coinType)
	if anyAddress == nil {
		return "", fmt.Errorf("%w: failed to create address for coin type %d", ErrAddressDerivation, coinType)
	}
	defer C.TWAnyAddressDelete(anyAddress)

	addressTW := C.TWAnyAddressDescription(anyAddress)
	if addressTW == nil {
		return "", fmt.Errorf("%w: failed to get address description", ErrAddressDerivation)
	}
	defer C.TWStringDelete(addressTW)

	address := C.GoString(C.TWStringUTF8Bytes(addressTW))
	if address == "" {
		return "", fmt.Errorf("%w: empty address generated", ErrAddressDerivation)
	}

	return address, nil
}

// GetPrivateKeyHex returns the private key as a hexadecimal string
func GetPrivateKeyHex(privateKey []byte) string {
	return hex.EncodeToString(privateKey)
}

// GetPublicKeyHex returns the public key as a hexadecimal string
func GetPublicKeyHex(publicKey []byte) string {
	return hex.EncodeToString(publicKey)
}
