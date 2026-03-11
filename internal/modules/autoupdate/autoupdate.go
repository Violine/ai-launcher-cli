// Package autoupdate provides the plugin for checking and applying launcher updates.
package autoupdate

import (
	"context"

	"github.com/ai-launcher/cli/internal/updater"
	"github.com/ai-launcher/cli/pkg/plugin"
)

// Version is the current launcher version (set at build time or default).
var Version = "0.0.0"

// Plugin implements plugin.Plugin for self-update checks.
type Plugin struct{}

// Name returns the command name.
func (Plugin) Name() string {
	return "autoupdate"
}

// Run checks for updates. Used when user runs "ai-launcher autoupdate".
// For FR-601 (check at startup in background), main starts a goroutine that calls updater.CheckInBackground.
func (Plugin) Run(ctx context.Context) error {
	_ = ctx
	updater.CheckInBackground(Version, nil)
	return nil
}

// New returns the autoupdate plugin instance.
func New() plugin.Plugin {
	return &Plugin{}
}
