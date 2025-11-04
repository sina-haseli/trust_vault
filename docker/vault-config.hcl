# Vault configuration for Docker deployment

# Storage backend
storage "file" {
  path = "/vault/data"
}

# Listener configuration
listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 1
}

# Plugin directory
plugin_directory = "/vault/plugins"

# API address
api_addr = "http://0.0.0.0:8200"

# UI
ui = true

# Disable mlock for Docker
disable_mlock = true

# Log level
log_level = "info"
