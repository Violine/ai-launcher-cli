// Package autoupdate provides the plugin for checking and applying launcher updates.
package autoupdate

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ai-launcher/cli/internal/config"
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

// Run checks for updates. If a newer version exists, prompts to install and runs updater.Install.
func (Plugin) Run(ctx context.Context) error {
	repo := ""
	if c, ok := ctx.Value(plugin.ConfigKey).(*config.Config); ok && c != nil {
		repo = c.UpdateRepo
	}
	if repo == "" {
		repo = updater.DefaultRepo
	}

	latest, err := updater.LatestRelease(ctx, repo)
	if err != nil {
		return fmt.Errorf("check update: %w", err)
	}
	if latest == "" {
		fmt.Println("No release found.")
		return nil
	}

	newer, err := updater.NewerThan(Version, latest)
	if err != nil {
		return fmt.Errorf("compare versions: %w", err)
	}
	if !newer {
		fmt.Printf("Already up to date: %s\n", Version)
		return nil
	}

	downloadURL := updater.DownloadURL(repo, latest)
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executable path: %w", err)
	}

	fmt.Printf("Update available: %s (current: %s)\n", latest, Version)
	fmt.Printf("Install %s? [y/N]: ", latest)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return nil
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Println("Downloading and installing...")
	onProgress := func(written, total int64) {
		if total <= 0 {
			fmt.Fprintf(os.Stderr, "\r  Downloading... %d KB     ", written/1024)
			return
		}
		pct := int64(0)
		if total > 0 {
			pct = written * 100 / total
		}
		barWidth := 24
		filled := int(pct) * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
		fmt.Fprintf(os.Stderr, "\r  [%s] %d%%   ", bar, pct)
	}
	if err := updater.Install(downloadURL, currentPath, onProgress); err != nil {
		return fmt.Errorf("install: %w", err)
	}
	fmt.Fprint(os.Stderr, "\r")
	fmt.Printf("Installed %s. Restart the application to use the new version.\n", latest)
	return nil
}

// New returns the autoupdate plugin instance.
func New() plugin.Plugin {
	return &Plugin{}
}
