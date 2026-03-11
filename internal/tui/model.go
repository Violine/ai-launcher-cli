// Package tui implements the TUI using Bubble Tea and Lip Gloss.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/list"

	"github.com/ai-launcher/cli/internal/config"
)

// Screen is the current TUI screen.
type Screen int

const (
	ScreenMain Screen = iota
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
	Commands []string // display labels for menu
	CommandNames []string // plugin names in same order (for Run on Enter)

	// Main screen: selected menu index (0-based)
	SelectedIndex int
	// After quit: if >= 0, main should run CommandNames[RunCommandIndex]
	RunCommandIndex int

	// When Screen == ScreenHelp, return to this screen on F1/Esc
	PrevScreen Screen

	// Token screen state
	TokenInput     string
	TokenError     string
	TokenButtonFoc TokenButton

	// Terminal size (from tea.WindowSizeMsg); 0 until first message
	Width  int
	Height int
}

// NewModel creates the initial TUI model. If cfg has no API key, starts on token screen.
// menuLabels and commandNames must be the same length and same order as plugins.
func NewModel(cfg *config.Config, menuLabels []string, commandNames []string) Model {
	m := Model{
		Config:          cfg,
		Commands:        menuLabels,
		CommandNames:    commandNames,
		TokenButtonFoc:  TokenButtonOK,
		RunCommandIndex: -1,
	}
	if cfg == nil || cfg.APIKey == "" {
		m.Screen = ScreenToken
	} else {
		m.Screen = ScreenMain
	}
	return m
}

// NewModelNoConfig creates a model without config (backward compat); starts on token screen.
// commands is used as both display labels and command names.
func NewModelNoConfig(commands []string) Model {
	return NewModel(nil, commands, commands)
}

// ContentWidth returns the width to use for frame content (terminal width minus margins, at least FrameWidth).
func (m Model) ContentWidth() int {
	if m.Width <= 0 {
		return FrameWidth
	}
	// RootStyle has Padding(1, 2) → 2 cols left + 2 right
	w := m.Width - 4
	if w < FrameWidth {
		return FrameWidth
	}
	return w
}

// Init runs once at startup. WindowSizeMsg is sent by the runtime when the terminal is ready.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages (delegates to screen-specific logic; main screen and quit handled here).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
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
	// Build menu list with lipgloss list (enumerator + item styles)
	items := make([]any, len(m.Commands))
	for i, c := range m.Commands {
		items[i] = c
	}
	enumeratorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorTitleFrame)).
		MarginRight(1)
	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		MarginRight(1)
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHighlight)).
		Background(lipgloss.Color(ColorHighlightBg)).
		MarginRight(1)
	sel := m.SelectedIndex
	l := list.New(items...).
		Enumerator(list.Arabic).
		EnumeratorStyle(enumeratorStyle).
		ItemStyleFunc(func(_ list.Items, i int) lipgloss.Style {
			if i == sel {
				return selectedStyle
			}
			return itemStyle
		})

	body := SectionStyle.Render("Commands:")
	body += "\n\n" + l.String()
	body += "\n" + FooterStyle.Render("↑/↓ or 1-9: select   Enter: run   F1 Help   F7 Token   F10 Exit")
	rendered := FrameWithTitle("  AI LAUNCHER  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func viewTokenScreen(m Model) tea.View {
	body := BodyStyle.Render("Enter your API token to continue:")
	body += "\n\n  "
	// Input line: mask with * (FR-102), width from content area
	fieldWidth := m.ContentWidth() - 6
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
	rendered := FrameWithTitle("  API TOKEN  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func viewHelpScreen(m Model) tea.View {
	lines := []string{
		HelpKeyStyle.Render("F1") + "    " + HelpDescStyle.Render("Show / hide this help"),
		HelpKeyStyle.Render("F7") + "    " + HelpDescStyle.Render("Open API token screen (from main menu)"),
		HelpKeyStyle.Render("F10") + "   " + HelpDescStyle.Render("Exit application"),
		"",
		HelpKeyStyle.Render("↑/↓ or 1-9") + "  " + HelpDescStyle.Render("Select menu item (main screen)"),
		HelpKeyStyle.Render("Enter") + "  " + HelpDescStyle.Render("Run selected command (main) or confirm (token)"),
		"",
		HelpKeyStyle.Render("Tab / ←→") + "  " + HelpDescStyle.Render("Switch between OK / Cancel on token screen"),
		HelpKeyStyle.Render("Esc") + "    " + HelpDescStyle.Render("Cancel / back to menu"),
	}
	body := ""
	for _, line := range lines {
		body += line + "\n"
	}
	body += "\n" + FooterStyle.Render("F1 or Esc to close help")
	rendered := FrameWithTitle("  HELP  ", body, m.ContentWidth())
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
	n := len(m.Commands)
	if n == 0 {
		return m, nil
	}
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
		case "up", "k":
			m.SelectedIndex--
			if m.SelectedIndex < 0 {
				m.SelectedIndex = n - 1
			}
			return m, nil
		case "down", "j":
			m.SelectedIndex++
			if m.SelectedIndex >= n {
				m.SelectedIndex = 0
			}
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < n {
				m.SelectedIndex = idx
			}
			return m, nil
		case "enter":
			m.RunCommandIndex = m.SelectedIndex
			return m, tea.Quit
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
