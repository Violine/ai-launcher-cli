// Command ai-launcher is the main entry point for the AI Launcher CLI.
// It loads config, registers plugins, and dispatches to the requested command by name.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/config"
	"github.com/ai-launcher/cli/internal/modules/agentrun"
	"github.com/ai-launcher/cli/internal/modules/autoupdate"
	"github.com/ai-launcher/cli/internal/modules/configgen"
	"github.com/ai-launcher/cli/internal/modules/mcpupdate"
	"github.com/ai-launcher/cli/internal/tui"
	"github.com/ai-launcher/cli/internal/updater"
	"github.com/ai-launcher/cli/pkg/plugin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	// FR-601: check for updates in background at startup (non-blocking)
	go updater.CheckInBackground(autoupdate.Version, func(available string) {
		fmt.Fprintf(os.Stderr, "Update available: %s (current: %s). Run 'ai-launcher autoupdate' to install.\n", available, autoupdate.Version)
	})

	plugins := []plugin.Plugin{
		configgen.New(),
		mcpupdate.New(),
		agentrun.New(),
		autoupdate.New(),
	}
	byName := make(map[string]plugin.Plugin)
	for _, p := range plugins {
		byName[p.Name()] = p
	}

	// No args: start TUI (agreed default)
	if len(os.Args) < 2 {
		commandNames := make([]string, 0, len(plugins))
		for _, p := range plugins {
			commandNames = append(commandNames, p.Name())
		}
		p := tea.NewProgram(tui.NewModelNoConfig(commandNames))
		if _, err := p.Run(); err != nil {
			// No TTY (e.g. pipe, IDE): show usage and exit 0 instead of failing
			if strings.Contains(err.Error(), "TTY") || strings.Contains(err.Error(), "tty") {
				fmt.Println("Usage: ai-launcher [<command>]")
				fmt.Println("Commands:")
				for _, name := range commandNames {
					fmt.Printf("  %s\n", name)
				}
				return
			}
			fmt.Fprintf(os.Stderr, "tui: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]
	p, ok := byName[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}

	ctx := context.WithValue(context.Background(), plugin.ConfigKey, cfg)
	if err := p.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
		os.Exit(1)
	}
}
