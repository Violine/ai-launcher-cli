package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/updater"
)

func viewMainScreen(m Model) tea.View {
	versionStr := m.CurrentVersion
	if versionStr == "" {
		versionStr = "0.0.0"
	}
	body := SectionStyle.Render("Commands:")
	body += "\n\n"
	sel := m.SelectedIndex
	for i, label := range m.Commands {
		num := fmt.Sprintf("%d.", i+1)
		line := "  " + BodyStyle.Render(num) + " "
		displayLabel := label
		if i < len(m.CommandNames) && m.CommandNames[i] == "autoupdate" && m.AvailableVersion != "" {
			displayLabel = label + "  " + HelpKeyStyle.Render("— доступно обновление до "+m.AvailableVersion)
		}
		if i == sel {
			line += HighlightStyle.Render(displayLabel)
		} else {
			line += BodyStyle.Render(displayLabel)
		}
		body += line + "\n"
	}
	body += "\n" + FooterStyle.Render(fmt.Sprintf("↑/↓ or 1-%d: select   Enter: run   F1 Help   F7 Token   F10 Exit", len(m.Commands)))
	rendered := FrameWithTitleSubtitle("  AI LAUNCHER  ", "  v "+versionStr, body, m.ContentWidth())
	return tea.NewView(rendered)
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
			m.Token.Input = ""
			m.Token.Error = ""
			if m.Config != nil {
				m.Config.APIKey = ""
			}
			m = PushScreen(m, ScreenToken)
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
			idx := m.SelectedIndex
			if idx >= 0 && idx < len(m.CommandNames) && m.CommandNames[idx] == "autoupdate" {
				if m.AvailableVersion != "" {
					m = PushScreen(m, ScreenUpdateConfirm)
					return m, nil
				}
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Progress.Title = "Проверка обновлений"
				m.Progress.Status = "Проверка обновлений..."
				m.Progress.Percent = -1
				m = PushScreen(m, ScreenUpdateChecking)
				return m, runCheckUpdateCmd(repo, m.CurrentVersion)
			}
			m.RunCommandIndex = m.SelectedIndex
			return m, tea.Quit
		}
	}
	return m, nil
}
