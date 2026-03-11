// Package configgen provides the plugin for generating config for Claude/OpenCode.
package configgen

import (
	"context"

	"github.com/ai-launcher/cli/pkg/plugin"
)

// Plugin implements plugin.Plugin for config generation.
type Plugin struct{}

// Name returns the command name.
func (Plugin) Name() string {
	return "configgen"
}

// Run runs the config generator. Not implemented yet.
func (Plugin) Run(ctx context.Context) error {
	_ = ctx
	// TODO: generate config for Claude/OpenCode
	return nil
}

// New returns the configgen plugin instance.
func New() plugin.Plugin {
	return &Plugin{}
}
