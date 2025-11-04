# Trust Vault Plugin - Deployment Guide

This guide covers deploying the Trust Vault Plugin in various environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Local Development](#local-development)
- [Production Deployment](#production-deployment)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Security Considerations](#security-considerations)
- [Monitoring and Maintenance](#monitoring-and-maintenance)

## Prerequisites

### System Requirements

- Go 1.25 or later
- HashiCorp Vault 1.12 or later
- Trust Wallet Core library
- Linux, macOS, or Windows
- Minimum 512MB RAM
- 100MB disk space

### Required Tools

```bash
# Vault CLI
vault version

# Go
go version

# Make (optional)
make --version

# Docker (for containerized deployment)
docker --version
docker-compose --version
```

## Local Development

### 1. Build from Source

```bash
# Clone repository
git clone https://github.com/sina-haseli/trust_vault.git
cd trust_vault

# Install dependencies
go mod download

# Build plugin
make build

# Or use build script
./build.sh
```

### 2. Start Vault in Dev Mode

```bash
# Start Vault dev server
vault server -dev -dev-root-token-id="root" -dev-plugin-dir="./bin"
```

### 3. Register Plugin

In a new terminal:

```bash
# Set environment
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='root'

# Register plugin
./scripts/register-plugin.sh

# Enable secrets engine
./scripts/enable-plugin.sh
```

## Production Deployment

### 1. Prepare Environment

```bash
# Create Vault user
sudo useradd -r -d /var/lib/vault -s /bin/false vault

# Create directories
sudo mkdir -p /etc/vault/plugins
sudo mkdir -p /var/lib/vault/data
sudo mkdir -p /var/log/vault

# Set permissions
sudo chown -R vault:vault /var/lib/vault
sudo chown -R vault:vault /var/log/vault
```

### 2. Build and Install Plugin

```bash
# Build for production
make build

# Install plugin
sudo cp bin/trust-vault-plugin /etc/vault/plugins/
sudo chmod +x /etc/vault/plugins/trust-vault-plugin
sudo chown vault:vault /etc/vault/plugins/trust-vault-plugin

# Calculate SHA256
sha256sum bin/trust-vault-plugin | awk '{print $1}' > plugin.sha256
```

### 3. Configure Vault

Create `/etc/vault/vault.hcl`:

```hcl
storage "file" {
  path = "/var/lib/vault/data"
}

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_cert_file = "/etc/vault/tls/vault.crt"
  tls_key_file  = "/etc/vault/tls/vault.key"
}

plugin_directory = "/etc/vault/plugins"

api_addr = "https://vault.example.com:8200"
cluster_addr = "https://vault.example.com:8201"

ui = true

log_level = "info"
log_file = "/var/log/vault/vault.log"
```

### 4. Start Vault Service

```bash
# Create systemd service
sudo tee /etc/systemd/system/vault.service > /dev/null <<EOF
[Unit]
Description=HashiCorp Vault
Documentation=https://www.vaultproject.io/docs/
Requires=network-online.target
After=network-online.target

[Service]
User=vault
Group=vault
ExecStart=/usr/local/bin/vault server -config=/etc/vault/vault.hcl
ExecReload=/bin/kill -HUP $MAINPID
KillMode=process
Restart=on-failure
RestartSec=5
LimitNOFILE=65536
LimitMEMLOCK=infinity

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl enable vault
sudo systemctl start vault
sudo systemctl status vault
```

### 5. Initialize Vault

```bash
# Initialize Vault
vault operator init -key-shares=5 -key-threshold=3

# Save the unseal keys and root token securely!

# Unseal Vault (repeat with 3 different keys)
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>

# Login
vault login <root-token>
```

### 6. Register and Enable Plugin

```bash
# Register plugin
SHA256=$(cat plugin.sha256)
vault plugin register \
  -sha256="$SHA256" \
  -command="trust-vault-plugin" \
  secret trust-vault

# Enable secrets engine
vault secrets enable -path=trust-vault trust-vault

# Verify
vault secrets list
```

## Docker Deployment

### Single Container

```bash
# Build image
docker build -t trust-vault:latest .

# Run container
docker run -d \
  --name trust-vault \
  -p 8200:8200 \
  --cap-add=IPC_LOCK \
  -v vault-data:/vault/data \
  -v vault-logs:/vault/logs \
  trust-vault:latest

# Check logs
docker logs -f trust-vault
```

### Docker Compose

```bash
# Start services
docker-compose up -d vault

# View logs
docker-compose logs -f vault

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Docker Compose with Custom Config

Create `docker-compose.override.yml`:

```yaml
version: '3.8'

services:
  vault:
    environment:
      VAULT_ADDR: "https://vault.example.com:8200"
    volumes:
      - ./custom-config.hcl:/vault/config/custom.hcl:ro
      - ./tls:/vault/tls:ro
    ports:
      - "443:8200"
```

## Kubernetes Deployment

### 1. Create Namespace

```bash
kubectl create namespace vault
```

### 2. Build and Push Image

```bash
# Build image
docker build -t your-registry/trust-vault:v1.0.0 .

# Push to registry
docker push your-registry/trust-vault:v1.0.0
```

### 3. Deploy with Helm

```bash
# Add HashiCorp Helm repository
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update

# Create values file
cat > values.yaml <<EOF
server:
  image:
    repository: "your-registry/trust-vault"
    tag: "v1.0.0"
  
  extraVolumes:
    - type: secret
      name: vault-tls
      path: /vault/tls
  
  standalone:
    enabled: true
    config: |
      ui = true
      
      listener "tcp" {
        address = "[::]:8200"
        cluster_address = "[::]:8201"
        tls_cert_file = "/vault/tls/tls.crt"
        tls_key_file = "/vault/tls/tls.key"
      }
      
      storage "file" {
        path = "/vault/data"
      }
      
      plugin_directory = "/vault/plugins"
EOF

# Install Vault
helm install vault hashicorp/vault \
  --namespace vault \
  --values values.yaml
```

### 4. Initialize and Configure

```bash
# Port forward
kubectl port-forward -n vault vault-0 8200:8200

# Initialize
vault operator init

# Unseal
vault operator unseal <key>

# Register plugin
vault plugin register \
  -sha256="<sha256>" \
  -command="trust-vault-plugin" \
  secret trust-vault

# Enable
vault secrets enable -path=trust-vault trust-vault
```

## Security Considerations

### TLS Configuration

Always use TLS in production:

```hcl
listener "tcp" {
  address       = "0.0.0.0:8200"
  tls_cert_file = "/etc/vault/tls/vault.crt"
  tls_key_file  = "/etc/vault/tls/vault.key"
  tls_min_version = "tls12"
}
```

### Access Control

Create policies for different roles:

```bash
# Create policy for wallet operators
vault policy write wallet-operator - <<EOF
path "trust-vault/wallets/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF

# Create policy for read-only access
vault policy write wallet-reader - <<EOF
path "trust-vault/wallets/*" {
  capabilities = ["read", "list"]
}
EOF

# Create token with policy
vault token create -policy=wallet-operator
```

### Audit Logging

Enable audit logging:

```bash
# Enable file audit
vault audit enable file file_path=/var/log/vault/audit.log

# Enable syslog audit
vault audit enable syslog tag="vault" facility="AUTH"
```

### Backup Strategy

```bash
# Backup Vault data
tar -czf vault-backup-$(date +%Y%m%d).tar.gz /var/lib/vault/data

# Backup to S3
aws s3 cp vault-backup-$(date +%Y%m%d).tar.gz s3://your-backup-bucket/
```

## Monitoring and Maintenance

### Health Checks

```bash
# Check Vault status
vault status

# Check plugin status
vault plugin list secret | grep trust-vault

# Check secrets engine
vault secrets list
```

### Metrics

Enable Prometheus metrics:

```hcl
telemetry {
  prometheus_retention_time = "30s"
  disable_hostname = true
}
```

### Log Rotation

Configure logrotate for Vault logs:

```bash
sudo tee /etc/logrotate.d/vault > /dev/null <<EOF
/var/log/vault/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 vault vault
    sharedscripts
    postrotate
        systemctl reload vault > /dev/null 2>&1 || true
    endscript
}
EOF
```

### Upgrading Plugin

```bash
# Build new version
make build

# Stop Vault (or reload)
sudo systemctl stop vault

# Replace plugin
sudo cp bin/trust-vault-plugin /etc/vault/plugins/

# Calculate new SHA256
sha256sum bin/trust-vault-plugin | awk '{print $1}' > plugin.sha256

# Start Vault
sudo systemctl start vault

# Reload plugin
SHA256=$(cat plugin.sha256)
vault plugin reload -plugin=trust-vault -sha256="$SHA256"
```

## Troubleshooting

### Plugin Not Loading

```bash
# Check plugin permissions
ls -la /etc/vault/plugins/trust-vault-plugin

# Check Vault logs
journalctl -u vault -f

# Verify SHA256
sha256sum /etc/vault/plugins/trust-vault-plugin
```

### Connection Issues

```bash
# Check Vault is listening
netstat -tlnp | grep 8200

# Test connection
curl -k https://localhost:8200/v1/sys/health

# Check firewall
sudo ufw status
```

### Performance Issues

```bash
# Check resource usage
top -p $(pgrep vault)

# Check disk space
df -h /var/lib/vault

# Review metrics
vault read sys/metrics
```

## Best Practices

1. **Always use TLS in production**
2. **Enable audit logging**
3. **Implement proper access policies**
4. **Regular backups of Vault data**
5. **Monitor plugin health and metrics**
6. **Keep plugin and Vault versions up to date**
7. **Use separate environments for dev/staging/prod**
8. **Implement disaster recovery procedures**
9. **Regular security audits**
10. **Document your deployment configuration**

## Support

For deployment issues:
- Check Vault logs: `/var/log/vault/vault.log`
- Review plugin logs in Vault output
- Verify network connectivity
- Check file permissions
- Review Vault documentation: https://www.vaultproject.io/docs
