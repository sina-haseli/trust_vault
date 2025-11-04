package main

import (
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	"github.com/sina-haseli/trust_vault/backend"
)

func main() {
	// Create API client metadata for plugin configuration
	apiClientMeta := &api.PluginAPIClientMeta{}

	// Set up command-line flags for plugin configuration
	flags := apiClientMeta.FlagSet()
	if err := flags.Parse(os.Args[1:]); err != nil {
		logger := hclog.New(&hclog.LoggerOptions{})
		logger.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	// Get TLS configuration for secure gRPC communication
	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	// Serve the plugin
	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: backend.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})

	if err != nil {
		logger := hclog.New(&hclog.LoggerOptions{})
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
