# Building Trust Vault Plugin

This document describes how to build the Trust Vault plugin with Trust Wallet Core integration.

## Prerequisites

- Docker and Docker Compose installed
- Git (for submodules)

## Building with Docker (Recommended)

The easiest way to build the plugin is using Docker, which handles all the Trust Wallet Core dependencies automatically.

### On Linux/macOS:

```bash
# Make the build script executable
chmod +x build.sh

# Run the build
./build.sh
```

### On Windows:

```cmd
# Run the build script
build.bat
```

Or use docker-compose directly:

```bash
docker-compose run --rm build
```

The compiled binary will be created as `trust-vault-plugin` in the project root.

## What the Docker Build Does

1. Uses the official `trustwallet/wallet-core` Docker image which has prebuilt libraries
2. Copies the Trust Wallet Core libraries and headers
3. Sets up the Go build environment with CGO enabled
4. Compiles the plugin with proper linking to Trust Wallet Core

## Building Locally (Advanced)

To build locally without Docker, you need to:

1. Build Trust Wallet Core from source following their [build instructions](https://developer.trustwallet.com/wallet-core/building)
2. Install the libraries to `/usr/local/lib` (Linux/macOS) or appropriate location
3. Install headers to `/usr/local/include`
4. Set CGO environment variables:
   ```bash
   export CGO_ENABLED=1
   export CGO_CFLAGS="-I/usr/local/include"
   export CGO_LDFLAGS="-L/usr/local/lib -lTrustWalletCore -lTrezorCrypto -lprotobuf -lstdc++ -lm"
   ```
5. Build with Go:
   ```bash
   go build -o trust-vault-plugin cmd/trust-vault/main.go
   ```

## Troubleshooting

### Docker build fails

- Ensure Docker is running
- Try pulling the base image manually: `docker pull trustwallet/wallet-core:latest`
- Check Docker has enough resources allocated (at least 4GB RAM)

### CGO errors

- Make sure CGO_ENABLED=1 is set
- Verify the Trust Wallet Core libraries are in the correct location
- Check that all required system libraries are installed

## Trust Wallet Core

This plugin uses [Trust Wallet Core](https://github.com/trustwallet/wallet-core), which is included as a Git submodule in `third_party/wallet-core`.

The Docker build uses prebuilt binaries from the official Trust Wallet Core Docker image for faster builds.
