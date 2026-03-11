// Package tui implements the TUI using Bubble Tea and Lip Gloss.
package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/config"
)

// Screen is the current TUI screen.
type Screen int

const (
	ScreenMain Screen = iota
	ScreenToken
)

// Button for token screen OK/Cancel.
type TokenButton int

const (
	TokenButtonOK TokenButton = iota
	TokenButtonCancel
)

// Model holds the TUI state.
type Model struct {
	Screen   Screen
	Config   *config.Config
	Commands []string

	// Token screen state
	TokenInput      string
	TokenError      string
	TokenButtonFoc  TokenButton
}

// NewModel creates the initial TUI model. If cfg has no API key, starts on token screen.
// Use NewModelWithConfig when config is available (e.g. from main).
func NewModel(cfg *config.Config, commands []string) Model {
	m := Model{
		Config:          cfg,
		Commands:        commands,
		TokenButtonFoc: TokenButtonOK,
	}
	if cfg == nil || cfg.APIKey == "" {
		m.Screen = ScreenToken
	} else {
		m.Screen = ScreenMain
	}
	return m
}

// NewModelNoConfig creates a model without config (backward compat); starts on token screen.
func NewModelNoConfig(commands []string) Model {
	return NewModel(nil, commands)
}

// Init runs once at startup.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages (delegates to screen-specific logic; main screen and quit handled here).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	// Delegate to screen
	if m.Screen == ScreenToken {
		return updateTokenScreen(m, msg)
	}
	return updateMainScreen(m, msg)
}

// View renders the UI.
func (m Model) View() tea.View {
	if m.Screen == ScreenToken {
		return viewTokenScreen(m)
	}
	return viewMainScreen(m)
}

func viewMainScreen(m Model) tea.View {
	s := "  AI Launcher\n\n"
	s += "  Commands:\n"
	for _, name := range m.Commands {
		s += fmt.Sprintf("    %s\n", name)
	}
	s += "\n  Press q to quit.\n"
	return tea.NewView(s)
}

func viewTokenScreen(m Model) tea.View {
	// Placeholder; next commit will add Lip Gloss layout
	s := "  API TOKEN (placeholder)\n"
	return tea.NewView(s)
}

func updateMainScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	// F7 will be added in a later commit
	return m, nil
}

func updateTokenScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	// Input and OK/Cancel will be added in later commits
	return m, nil
}
