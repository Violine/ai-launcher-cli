// Package mcpupdate provides the plugin for updating MCP packages via GitLab.
package mcpupdate

import (
	"context"

	"github.com/ai-launcher/cli/pkg/plugin"
)

// Plugin implements plugin.Plugin for MCP updates.
type Plugin struct{}

// Name returns the command name.
func (Plugin) Name() string {
	return "mcpupdate"
}

// Run runs the MCP updater. Not implemented yet.
func (Plugin) Run(ctx context.Context) error {
	_ = ctx
	// TODO: use internal/mcp and internal/gitlab to update MCP packages
	return nil
}

// New returns the mcpupdate plugin instance.
func New() plugin.Plugin {
	return &Plugin{}
}
