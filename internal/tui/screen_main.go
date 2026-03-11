package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/ai-launcher/cli/internal/updater"
)

func viewMainScreen(m Model) tea.View {
	versionStr := m.CurrentVersion
	if versionStr == "" {
		versionStr = "0.0.0"
	}
	contentWidth := m.ContentWidth()
	body := SectionStyle.Render("Commands:")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += SectionStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	sel := m.SelectedIndex
	for i, label := range m.Commands {
		num := fmt.Sprintf("%d.", i+1)
		// Префикс "  1. " целиком в синем фоне, чтобы не было чёрных полос между цифрой и текстом
		line := BodyStyle.Render("  " + num + " ")
		displayLabel := label
		if i < len(m.CommandNames) && m.CommandNames[i] == "autoupdate" && m.AvailableVersion != "" {
			displayLabel = label + "  " + HelpKeyStyle.Render("— доступно обновление до "+m.AvailableVersion)
		}
		if i == sel {
			line += HighlightStyle.Render(displayLabel)
		} else {
			line += BodyStyle.Render(displayLabel)
		}
		// Добиваем строку до полной ширины синим фоном, чтобы справа не было чёрного
		if pad := contentWidth - lipgloss.Width(line); pad > 0 {
			line += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line + "\n"
	}
	footerStr := fmt.Sprintf("↑/↓ or 1-%d: select   Enter: run   F1 Help   F7 Token   F10 Exit", len(m.Commands))
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	// Вся строка футера одним рендером — синий фон до конца, без чёрных клеток после "F10 Exit"
	footerLine := FooterStyle.Render(footerStr + strings.Repeat(" ", pad))
	body += "\n" + footerLine
	body += "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
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
