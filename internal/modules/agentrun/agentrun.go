// Package agentrun provides the plugin for running AI agents (exec).
package agentrun

import (
	"context"

	"github.com/ai-launcher/cli/pkg/plugin"
)

// Plugin implements plugin.Plugin for running AI agents.
type Plugin struct{}

// Name returns the command name.
func (Plugin) Name() string {
	return "agentrun"
}

// Run runs the agent. Not implemented yet.
func (Plugin) Run(ctx context.Context) error {
	_ = ctx
	// TODO: use executor (subprocess) to run AI agents
	return nil
}

// New returns the agentrun plugin instance.
func New() plugin.Plugin {
	return &Plugin{}
}
