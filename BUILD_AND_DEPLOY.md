# Trust Vault Plugin - Build and Deployment Summary

This document provides a quick reference for all build and deployment scripts available in this project.

## Build Scripts

### Cross-Platform

| Script            | Platform    | Description                                     |
| ----------------- | ----------- | ----------------------------------------------- |
| `Makefile`        | Linux/macOS | Complete build automation with multiple targets |
| `build.sh`        | Linux/macOS | Standalone build script with version info       |
| `build.bat`       | Windows     | Docker-based build for Windows                  |
| `build-local.bat` | Windows     | Local build without Docker                      |

### Build Commands

```bash
# Using Make (recommended for Linux/macOS)
make build              # Build plugin
make build-all          # Build for all platforms
make test               # Run tests
make checksum           # Calculate SHA256
make install            # Install to Vault plugin directory

# Using shell script (Linux/macOS)
./build.sh              # Build for current platform

# Using batch script (Windows)
build.bat               # Build using Docker
build-local.bat         # Build locally without Docker
```

## Deployment Scripts

### Plugin Registration

| Script                        | Platform    | Description                |
| ----------------------------- | ----------- | -------------------------- |
| `scripts/register-plugin.sh`  | Linux/macOS | Register plugin with Vault |
| `scripts/register-plugin.bat` | Windows     | Register plugin with Vault |

Usage:
```bash
# Linux/macOS
./scripts/register-plugin.sh

# Windows
scripts\register-plugin.bat

# With custom options
./scripts/register-plugin.sh --plugin-dir /custom/path --vault-addr https://vault.example.com:8200
```

### Secrets Engine Enablement

| Script                      | Platform    | Description           |
| --------------------------- | ----------- | --------------------- |
| `scripts/enable-plugin.sh`  | Linux/macOS | Enable secrets engine |
| `scripts/enable-plugin.bat` | Windows     | Enable secrets engine |

Usage:
```bash
# Linux/macOS
./scripts/enable-plugin.sh

# Windows
scripts\enable-plugin.bat

# With custom mount path
./scripts/enable-plugin.sh --mount-path custom-vault
```

### Testing

| Script            | Platform    | Description             |
| ----------------- | ----------- | ----------------------- |
| `scripts/test.sh` | Linux/macOS | Run tests with coverage |
| `make test`       | Linux/macOS | Run tests via Make      |

Usage:
```bash
# Run tests
./scripts/test.sh

# Or using Make
make test
make test-coverage
```

## Docker Deployment

### Docker Files

| File                 | Description                            |
| -------------------- | -------------------------------------- |
| `Dockerfile`         | Production image with Vault and plugin |
| `Dockerfile.build`   | Build environment for compilation      |
| `docker-compose.yml` | Multi-service orchestration            |
| `.dockerignore`      | Files to exclude from Docker build     |

### Docker Scripts

| Script                    | Description        |
| ------------------------- | ------------------ |
| `scripts/build-docker.sh` | Build Docker image |

### Docker Commands

```bash
# Build image
docker build -t trust-vault:latest .

# Or use script
./scripts/build-docker.sh

# Run with docker-compose
docker-compose up -d vault

# View logs
docker-compose logs -f vault

# Stop
docker-compose down
```

### Docker Configuration

| File                      | Description                         |
| ------------------------- | ----------------------------------- |
| `docker/vault-config.hcl` | Vault configuration for Docker      |
| `docker/init-plugin.sh`   | Initialization script for container |

## Directory Structure

```
trust_vault/
├── bin/                          # Build output (created by build scripts)
├── cmd/
│   └── trust-vault/             # Plugin entry point
├── backend/                      # Vault backend implementation
├── service/                      # Business logic
├── storage/                      # Storage layer
├── wallet/                       # Trust Wallet Core wrapper
├── docker/                       # Docker configuration files
│   ├── vault-config.hcl         # Vault config for Docker
│   └── init-plugin.sh           # Container initialization
├── scripts/                      # Deployment scripts
│   ├── register-plugin.sh       # Register plugin (Linux/macOS)
│   ├── register-plugin.bat      # Register plugin (Windows)
│   ├── enable-plugin.sh         # Enable secrets engine (Linux/macOS)
│   ├── enable-plugin.bat        # Enable secrets engine (Windows)
│   ├── build-docker.sh          # Build Docker image
│   └── test.sh                  # Run tests
├── Makefile                      # Build automation (Linux/macOS)
├── build.sh                      # Build script (Linux/macOS)
├── build.bat                     # Build script (Windows/Docker)
├── build-local.bat              # Build script (Windows/Local)
├── Dockerfile                    # Production Docker image
├── Dockerfile.build             # Build environment
├── docker-compose.yml           # Docker Compose configuration
├── .dockerignore                # Docker ignore file
├── .gitignore                   # Git ignore file
├── README.md                    # Main documentation
├── QUICKSTART.md                # Quick start guide
├── DEPLOYMENT.md                # Detailed deployment guide
└── BUILD_AND_DEPLOY.md          # This file
```

## Quick Reference

### First Time Setup

```bash
# 1. Build the plugin
make build                        # Linux/macOS
build-local.bat                   # Windows

# 2. Register with Vault
./scripts/register-plugin.sh      # Linux/macOS
scripts\register-plugin.bat       # Windows

# 3. Enable secrets engine
./scripts/enable-plugin.sh        # Linux/macOS
scripts\enable-plugin.bat         # Windows
```

### Development Workflow

```bash
# Make changes to code
# ...

# Run tests
make test

# Build
make build

# Reload plugin in Vault
vault plugin reload -plugin=trust-vault
```

### Production Deployment

```bash
# 1. Build for production
make build

# 2. Calculate checksum
make checksum

# 3. Install to Vault
make install

# 4. Register and enable
./scripts/register-plugin.sh
./scripts/enable-plugin.sh
```

### Docker Workflow

```bash
# Build and run
docker-compose up -d vault

# Check status
docker-compose ps

# View logs
docker-compose logs -f vault

# Stop
docker-compose down
```

## Environment Variables

### Build Environment

| Variable      | Description         | Default      |
| ------------- | ------------------- | ------------ |
| `VERSION`     | Plugin version      | `dev`        |
| `GOOS`        | Target OS           | Current OS   |
| `GOARCH`      | Target architecture | Current arch |
| `CGO_ENABLED` | Enable CGO          | `1`          |

### Deployment Environment

| Variable           | Description                | Default                 |
| ------------------ | -------------------------- | ----------------------- |
| `VAULT_ADDR`       | Vault server address       | `http://127.0.0.1:8200` |
| `VAULT_TOKEN`      | Vault authentication token | -                       |
| `VAULT_PLUGIN_DIR` | Plugin directory           | `/etc/vault/plugins`    |
| `MOUNT_PATH`       | Secrets engine mount path  | `trust-vault`           |

## Makefile Targets

| Target               | Description                             |
| -------------------- | --------------------------------------- |
| `make build`         | Build plugin binary                     |
| `make build-all`     | Build for multiple platforms            |
| `make test`          | Run tests                               |
| `make test-coverage` | Run tests with coverage report          |
| `make checksum`      | Calculate SHA256 checksum               |
| `make install`       | Install plugin to Vault (requires sudo) |
| `make clean`         | Clean build artifacts                   |
| `make fmt`           | Format code                             |
| `make lint`          | Run linter                              |
| `make tidy`          | Tidy dependencies                       |
| `make deps`          | Download dependencies                   |
| `make docker-build`  | Build Docker image                      |
| `make docker-run`    | Run Docker container                    |
| `make docker-stop`   | Stop Docker container                   |
| `make dev`           | Build development version               |
| `make help`          | Show help message                       |

## Common Issues and Solutions

### Build Issues

**Issue**: `go: command not found`
```bash
# Install Go
# Linux: sudo apt-get install golang-go
# macOS: brew install go
# Windows: Download from https://golang.org/dl/
```

**Issue**: `make: command not found`
```bash
# Use build.sh instead
./build.sh

# Or install make
# Linux: sudo apt-get install build-essential
# macOS: xcode-select --install
```

### Deployment Issues

**Issue**: `vault: command not found`
```bash
# Install Vault CLI
# Download from https://www.vaultproject.io/downloads
```

**Issue**: `Permission denied` when installing
```bash
# Use sudo
sudo make install

# Or manually with sudo
sudo cp bin/trust-vault-plugin /etc/vault/plugins/
```

**Issue**: `Plugin registration failed`
```bash
# Check SHA256 matches
sha256sum bin/trust-vault-plugin

# Verify plugin directory
ls -la /etc/vault/plugins/

# Check Vault logs
journalctl -u vault -f
```

### Docker Issues

**Issue**: `docker: command not found`
```bash
# Install Docker
# Follow instructions at https://docs.docker.com/get-docker/
```

**Issue**: `Cannot connect to Docker daemon`
```bash
# Start Docker service
sudo systemctl start docker

# Or on macOS/Windows, start Docker Desktop
```

## Additional Resources

- **README.md**: Complete project documentation
- **QUICKSTART.md**: Quick start guide with examples
- **DEPLOYMENT.md**: Detailed deployment guide for various environments
- **Requirements**: `.kiro/specs/trust-vault-plugin/requirements.md`
- **Design**: `.kiro/specs/trust-vault-plugin/design.md`
- **Tasks**: `.kiro/specs/trust-vault-plugin/tasks.md`

## Support

For build and deployment issues:
1. Check this document for common solutions
2. Review the relevant script's help: `./script.sh --help`
3. Check Vault logs for plugin-specific issues
4. Verify environment variables are set correctly
5. Ensure all prerequisites are installed

## Version Information

To check versions:
```bash
# Go version
go version

# Vault version
vault version

# Docker version
docker --version

# Make version
make --version
```

## Next Steps

After successful build and deployment:
1. Read QUICKSTART.md for usage examples
2. Review DEPLOYMENT.md for production best practices
3. Set up monitoring and logging
4. Configure access policies
5. Enable audit logging
6. Implement backup procedures
