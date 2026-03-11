// Package plugin defines the interface for CLI command plugins.
// Plugins are registered in cmd/ai-launcher and invoked by name.
package plugin

import "context"

// Plugin is the interface that all command modules must implement.
type Plugin interface {
	// Name returns the command name used to invoke the plugin (e.g. "configgen", "mcpupdate").
	Name() string
	// Run executes the plugin. Config can be passed via context using ConfigKey.
	Run(ctx context.Context) error
}

// configKey is the context key for the application config.
// The value should be the same type as internal/config.Config (passed from cmd).
type configKey struct{}

// ConfigKey is the key to pass config through context: context.WithValue(ctx, plugin.ConfigKey, cfg).
// In Run(), retrieve with: cfg := ctx.Value(plugin.ConfigKey).(*config.Config)
var ConfigKey = &configKey{}
