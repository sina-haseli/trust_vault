# Trust Wallet Core Wrapper

This package provides a Go wrapper around the Trust Wallet Core C library for cryptocurrency wallet operations.

## Prerequisites

The Trust Wallet Core C library must be built and installed on your system before this package can be compiled.

### Building Trust Wallet Core

1. Clone the Trust Wallet Core repository:
```bash
git clone https://github.com/trustwallet/wallet-core.git
cd wallet-core
```

2. Install dependencies (varies by platform):

**Linux:**
```bash
sudo apt-get install build-essential cmake ninja-build libboost-all-dev
```

**macOS:**
```bash
brew install cmake ninja boost
```

**Windows:**
- Install Visual Studio with C++ support
- Install CMake
- Install Boost

3. Build the library:
```bash
./bootstrap.sh
cmake -H. -Bbuild -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

4. Install the library:
```bash
sudo cmake --install build
```

### Setting up CGO

After installing Trust Wallet Core, you may need to set CGO environment variables:

**Linux/macOS:**
```bash
export CGO_CFLAGS="-I/usr/local/include"
export CGO_LDFLAGS="-L/usr/local/lib -lTrustWalletCore"
```

**Windows:**
```powershell
$env:CGO_CFLAGS="-IC:\TrustWalletCore\include"
$env:CGO_LDFLAGS="-LC:\TrustWalletCore\lib -lTrustWalletCore"
```

## Usage

Once Trust Wallet Core is installed, you can use this package:

```go
package main

import (
    "fmt"
    "github.com/sina-haseli/trust_vault/wallet"
)

func main() {
    twc := wallet.NewTrustWalletCore()
    
    // Generate a new Ethereum wallet
    keys, err := twc.GenerateWallet(wallet.CoinTypeEthereum)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Mnemonic: %s\n", keys.Mnemonic)
    fmt.Printf("Address: %s\n", keys.Address)
}
```

## Supported Coin Types

- Bitcoin (0)
- Ethereum (60)
- Solana (501)

## API

### GenerateWallet(coinType uint32) (*WalletKeys, error)
Generates a new HD wallet with a 12-word mnemonic phrase.

### ImportWallet(mnemonic string, coinType uint32) (*WalletKeys, error)
Imports an existing wallet from a mnemonic phrase.

### DeriveAddress(mnemonic string, coinType uint32, derivationPath string) (string, error)
Derives an address for a specific coin type and optional custom derivation path.

### SignTransaction(privateKey []byte, coinType uint32, txData []byte) ([]byte, error)
Signs transaction data using the private key.

## Notes

- This implementation uses CGO to interface with the Trust Wallet Core C library
- The library must be properly installed and accessible to the Go compiler
- Private keys and mnemonics should be handled securely and never logged or exposed
