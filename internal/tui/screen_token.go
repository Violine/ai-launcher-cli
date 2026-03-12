package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/ai-launcher/cli/internal/config"
)

// TokenModel is the API token input screen.
type TokenModel struct {
	Shared   *SharedState
	Input    string
	Error    string
	ButtonFoc TokenButton
}

// NewTokenModel creates a new token screen model.
func NewTokenModel(shared *SharedState) *TokenModel {
	return &TokenModel{Shared: shared, ButtonFoc: TokenButtonOK}
}

// ID implements ScreenModel.
func (m *TokenModel) ID() Screen { return ScreenToken }

// Init implements tea.Model.
func (m *TokenModel) Init() tea.Cmd { return nil }

// View implements tea.Model.
func (m *TokenModel) View() tea.View {
	contentWidth := m.Shared.ContentWidth()
	body := BodyStyle.Render("Enter your API token to continue:")
	body += "\n\n" + BodyStyle.Render("  ")
	fieldWidth := contentWidth - 6
	if fieldWidth < 20 {
		fieldWidth = 20
	}
	mask := strings.Repeat("*", len(m.Input))
	if mask == "" {
		mask = " "
	}
	content := " " + mask + " "
	body += HighlightStyle.Render(content)
	if len(content) < fieldWidth {
		body += BodyStyle.Render(strings.Repeat(" ", fieldWidth-len(content)))
	}
	body += "\n\n" + BodyStyle.Render("  ")
	okBtn := ButtonStyle.Render("[ OK ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.ButtonFoc == TokenButtonOK {
		okBtn = ButtonActiveStyle.Render("[ OK ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	body += okBtn + BodyStyle.Render("  ") + cancelBtn + "\n"
	if m.Error != "" {
		body += "\n" + BodyStyle.Render("  ") + ErrorStyle.Render(m.Error) + "\n"
	}
	footerStr := "F1 Help   Tab: switch button   Enter: confirm   Esc: back   F10 Exit"
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	body += "\n" + FooterStyle.Render(footerStr+strings.Repeat(" ", pad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	return tea.NewView(FrameWithTitle("  API TOKEN  ", body, contentWidth))
}

// Update implements tea.Model.
func (m *TokenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := msg.String()
		switch s {
		case "enter":
			if m.ButtonFoc == TokenButtonOK {
				if err := config.ValidateAPIKey(m.Input); err != nil {
					m.Error = err.Error()
					return m, nil
				}
				cfg := m.Shared.Config
				if cfg == nil {
					cfg = &config.Config{}
				}
				cfg.APIKey = strings.TrimSpace(m.Input)
				if err := config.Save(cfg); err != nil {
					m.Error = err.Error()
					return m, nil
				}
				m.Shared.Config = cfg
				m.Error = ""
				m.Input = ""
				return m, PopScreenCmd()
			}
			m.Error = ""
			m.Input = ""
			return m, PopScreenCmd()
		case "tab", "right":
			m.ButtonFoc = TokenButtonCancel
			m.Error = ""
			return m, nil
		case "shift+tab", "left":
			m.ButtonFoc = TokenButtonOK
			m.Error = ""
			return m, nil
		case "esc":
			m.Error = ""
			m.Input = ""
			return m, PopScreenCmd()
		case "backspace":
			if len(m.Input) > 0 {
				m.Input = m.Input[:len(m.Input)-1]
				m.Error = ""
			}
			return m, nil
		default:
			if len(s) == 1 && s[0] >= 32 && s[0] != 127 {
				m.Input += s
				m.Error = ""
				return m, nil
			}
		}
	}
	return m, nil
}
