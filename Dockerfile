# Multi-stage Dockerfile for Trust Vault Plugin with HashiCorp Vault

# Stage 1: Build the plugin
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make bash

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the plugin
RUN make build

# Calculate SHA256
RUN sha256sum bin/trust-vault-plugin | awk '{print $1}' > bin/trust-vault-plugin.sha256

# Stage 2: Create Vault image with plugin
FROM hashicorp/vault:latest

# Install bash for scripts
USER root
RUN apk add --no-cache bash

# Create plugin directory
RUN mkdir -p /vault/plugins

# Copy plugin binary from builder
COPY --from=builder /build/bin/trust-vault-plugin /vault/plugins/
COPY --from=builder /build/bin/trust-vault-plugin.sha256 /vault/plugins/

# Set permissions
RUN chmod +x /vault/plugins/trust-vault-plugin

# Copy configuration
COPY docker/vault-config.hcl /vault/config/vault.hcl

# Copy initialization script
COPY docker/init-plugin.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/init-plugin.sh

# Expose Vault port
EXPOSE 8200

# Set working directory
WORKDIR /vault

# Switch back to vault user
USER vault

# Use custom entrypoint that initializes the plugin
ENTRYPOINT ["/usr/local/bin/init-plugin.sh"]
