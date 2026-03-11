// Package tui implements the TUI using Bubble Tea and Lip Gloss.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/config"
)

// Screen is the current TUI screen.
type Screen int

const (
	ScreenMain  Screen = iota
	ScreenToken
	ScreenHelp
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

	// When Screen == ScreenHelp, return to this screen on F1/Esc
	PrevScreen Screen

	// Token screen state
	TokenInput     string
	TokenError     string
	TokenButtonFoc TokenButton
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
		case "q", "ctrl+c", "f10":
			return m, tea.Quit
		case "f1":
			if m.Screen == ScreenHelp {
				m.Screen = m.PrevScreen
				return m, nil
			}
			m.PrevScreen = m.Screen
			m.Screen = ScreenHelp
			return m, nil
		}
	}
	// Delegate to screen
	switch m.Screen {
	case ScreenHelp:
		return updateHelpScreen(m, msg)
	case ScreenToken:
		return updateTokenScreen(m, msg)
	default:
		return updateMainScreen(m, msg)
	}
}

// View renders the UI.
func (m Model) View() tea.View {
	switch m.Screen {
	case ScreenHelp:
		return viewHelpScreen(m)
	case ScreenToken:
		return viewTokenScreen(m)
	default:
		return viewMainScreen(m)
	}
}

func viewMainScreen(m Model) tea.View {
	body := SectionStyle.Render("Commands:")
	body += "\n\n"
	for _, name := range m.Commands {
		body += "  " + BodyStyle.Render(name) + "\n"
	}
	body += "\n" + FooterStyle.Render("F1 Help   F7 Token   F10 Exit")
	rendered := FrameWithTitle("  AI LAUNCHER  ", body)
	return tea.NewView(rendered)
}

func viewTokenScreen(m Model) tea.View {
	body := BodyStyle.Render("Enter your API token to continue:")
	body += "\n\n  "
	// Input line: mask with * (FR-102), fixed width to match frame
	fieldWidth := FrameWidth - 6
	if fieldWidth < 20 {
		fieldWidth = 20
	}
	mask := strings.Repeat("*", len(m.TokenInput))
	if mask == "" {
		mask = " "
	}
	if len(mask) < fieldWidth {
		mask += strings.Repeat(" ", fieldWidth-len(mask))
	} else {
		mask = mask[:fieldWidth]
	}
	body += HighlightStyle.Render(" "+mask+" ") + "\n\n  "
	// Buttons [ OK ] [ Cancel ]
	okBtn := ButtonStyle.Render("[ OK ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.TokenButtonFoc == TokenButtonOK {
		okBtn = ButtonActiveStyle.Render("[ OK ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	body += okBtn + "  " + cancelBtn + "\n"
	if m.TokenError != "" {
		body += "\n  " + ErrorStyle.Render(m.TokenError) + "\n"
	}
	body += "\n" + FooterStyle.Render("F1 Help   Tab: switch button   Enter: confirm   Esc: back   F10 Exit")
	rendered := FrameWithTitle("  API TOKEN  ", body)
	return tea.NewView(rendered)
}

func viewHelpScreen(m Model) tea.View {
	lines := []string{
		HelpKeyStyle.Render("F1") + "    " + HelpDescStyle.Render("Show / hide this help"),
		HelpKeyStyle.Render("F7") + "    " + HelpDescStyle.Render("Open API token screen (from main menu)"),
		HelpKeyStyle.Render("F10") + "   " + HelpDescStyle.Render("Exit application"),
		"",
		HelpKeyStyle.Render("Tab / ←→") + "  " + HelpDescStyle.Render("Switch between OK / Cancel on token screen"),
		HelpKeyStyle.Render("Enter") + "  " + HelpDescStyle.Render("Confirm (OK) or go back (Cancel)"),
		HelpKeyStyle.Render("Esc") + "    " + HelpDescStyle.Render("Cancel / back to menu"),
	}
	body := ""
	for _, line := range lines {
		body += line + "\n"
	}
	body += "\n" + FooterStyle.Render("F1 or Esc to close help")
	rendered := FrameWithTitle("  HELP  ", body)
	return tea.NewView(rendered)
}

func updateHelpScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "f1", "esc":
			m.Screen = m.PrevScreen
			return m, nil
		}
	}
	return m, nil
}

func updateMainScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "f7":
			// FR-107: switch to token screen (reset/change key)
			m.Screen = ScreenToken
			m.TokenInput = ""
			m.TokenError = ""
			if m.Config != nil {
				m.Config.APIKey = ""
			}
			return m, nil
		}
	}
	return m, nil
}

func updateTokenScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := msg.String()
		switch s {
		case "enter":
			if m.TokenButtonFoc == TokenButtonOK {
				// Validate and save (FR-103, FR-104, FR-105/106)
				if err := config.ValidateAPIKey(m.TokenInput); err != nil {
					m.TokenError = err.Error()
					return m, nil
				}
				cfg := m.Config
				if cfg == nil {
					cfg = &config.Config{}
				}
				cfg.APIKey = strings.TrimSpace(m.TokenInput)
				if err := config.Save(cfg); err != nil {
					m.TokenError = err.Error()
					return m, nil
				}
				m.Config = cfg
				m.Screen = ScreenMain
				m.TokenError = ""
				m.TokenInput = ""
				return m, nil
			}
			// Cancel: return to main menu (do not quit)
			m.Screen = ScreenMain
			m.TokenError = ""
			m.TokenInput = ""
			return m, nil
		case "tab", "right":
			m.TokenButtonFoc = TokenButtonCancel
			m.TokenError = ""
			return m, nil
		case "shift+tab", "left":
			m.TokenButtonFoc = TokenButtonOK
			m.TokenError = ""
			return m, nil
		case "esc":
			// Esc: return to main menu (same as Cancel)
			m.Screen = ScreenMain
			m.TokenError = ""
			m.TokenInput = ""
			return m, nil
		case "backspace":
			if len(m.TokenInput) > 0 {
				m.TokenInput = m.TokenInput[:len(m.TokenInput)-1]
				m.TokenError = ""
			}
			return m, nil
		default:
			// Single rune: append to input (FR-101)
			if len(s) == 1 && s[0] >= 32 && s[0] != 127 {
				m.TokenInput += s
				m.TokenError = ""
				return m, nil
			}
		}
	}
	return m, nil
}
