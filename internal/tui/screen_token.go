package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/config"
)

func viewTokenScreen(m Model) tea.View {
	body := BodyStyle.Render("Enter your API token to continue:")
	body += "\n\n  "
	fieldWidth := m.ContentWidth() - 6
	if fieldWidth < 20 {
		fieldWidth = 20
	}
	mask := strings.Repeat("*", len(m.Token.Input))
	if mask == "" {
		mask = " "
	}
	content := " " + mask + " "
	body += HighlightStyle.Render(content)
	if len(content) < fieldWidth {
		body += BodyStyle.Render(strings.Repeat(" ", fieldWidth-len(content)))
	}
	body += "\n\n  "
	okBtn := ButtonStyle.Render("[ OK ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.Token.ButtonFoc == TokenButtonOK {
		okBtn = ButtonActiveStyle.Render("[ OK ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	body += okBtn + "  " + cancelBtn + "\n"
	if m.Token.Error != "" {
		body += "\n  " + ErrorStyle.Render(m.Token.Error) + "\n"
	}
	body += "\n" + FooterStyle.Render("F1 Help   Tab: switch button   Enter: confirm   Esc: back   F10 Exit")
	rendered := FrameWithTitle("  API TOKEN  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateTokenScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := msg.String()
		switch s {
		case "enter":
			if m.Token.ButtonFoc == TokenButtonOK {
				if err := config.ValidateAPIKey(m.Token.Input); err != nil {
					m.Token.Error = err.Error()
					return m, nil
				}
				cfg := m.Config
				if cfg == nil {
					cfg = &config.Config{}
				}
				cfg.APIKey = strings.TrimSpace(m.Token.Input)
				if err := config.Save(cfg); err != nil {
					m.Token.Error = err.Error()
					return m, nil
				}
				m.Config = cfg
				m.Token.Error = ""
				m.Token.Input = ""
				m = PopScreen(m)
				return m, nil
			}
			m.Token.Error = ""
			m.Token.Input = ""
			m = PopScreen(m)
			return m, nil
		case "tab", "right":
			m.Token.ButtonFoc = TokenButtonCancel
			m.Token.Error = ""
			return m, nil
		case "shift+tab", "left":
			m.Token.ButtonFoc = TokenButtonOK
			m.Token.Error = ""
			return m, nil
		case "esc":
			m.Token.Error = ""
			m.Token.Input = ""
			m = PopScreen(m)
			return m, nil
		case "backspace":
			if len(m.Token.Input) > 0 {
				m.Token.Input = m.Token.Input[:len(m.Token.Input)-1]
				m.Token.Error = ""
			}
			return m, nil
		default:
			if len(s) == 1 && s[0] >= 32 && s[0] != 127 {
				m.Token.Input += s
				m.Token.Error = ""
				return m, nil
			}
		}
	}
	return m, nil
}
