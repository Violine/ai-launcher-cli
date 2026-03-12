package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/ai-launcher/cli/internal/updater"
)

// MainModel is the main menu screen.
type MainModel struct {
	Shared        *SharedState
	SelectedIndex int
}

// NewMainModel creates a new main menu model.
func NewMainModel(shared *SharedState) *MainModel {
	return &MainModel{Shared: shared}
}

// ID implements ScreenModel.
func (m *MainModel) ID() Screen { return ScreenMain }

// Init implements tea.Model.
func (m *MainModel) Init() tea.Cmd { return nil }

// View implements tea.Model.
func (m *MainModel) View() tea.View {
	s := m.Shared
	versionStr := s.CurrentVersion
	if versionStr == "" {
		versionStr = "0.0.0"
	}
	contentWidth := s.ContentWidth()
	body := SectionStyle.Render("Commands:")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += SectionStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	sel := m.SelectedIndex
	for i, label := range s.Commands {
		num := fmt.Sprintf("%d.", i+1)
		line := BodyStyle.Render("  " + num + " ")
		displayLabel := label
		if i < len(s.CommandNames) && s.CommandNames[i] == "autoupdate" && s.AvailableVersion != "" {
			displayLabel = label + "  " + HelpKeyStyle.Render("— доступно обновление до "+s.AvailableVersion)
		}
		if i == sel {
			line += HighlightStyle.Render(displayLabel)
		} else {
			line += BodyStyle.Render(displayLabel)
		}
		if pad := contentWidth - lipgloss.Width(line); pad > 0 {
			line += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line + "\n"
	}
	footerStr := fmt.Sprintf("↑/↓ or 1-%d: select   Enter: run   F1 Help   F7 Token   F10 Exit", len(s.Commands))
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	footerLine := FooterStyle.Render(footerStr + strings.Repeat(" ", pad))
	body += "\n" + footerLine + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	rendered := FrameWithTitleSubtitle("  AI LAUNCHER  ", "  v "+versionStr, body, contentWidth)
	return tea.NewView(rendered)
}

// Update implements tea.Model.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	n := len(m.Shared.Commands)
	if n == 0 {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "f7":
			if m.Shared.Config != nil {
				m.Shared.Config.APIKey = ""
			}
			return m, PushScreenCmd(ScreenToken)
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
			if idx >= 0 && idx < len(m.Shared.CommandNames) && m.Shared.CommandNames[idx] == "autoupdate" {
				if m.Shared.AvailableVersion != "" {
					return m, PushScreenCmd(ScreenUpdateConfirm)
				}
				repo := updater.DefaultRepo
				if m.Shared.Config != nil && m.Shared.Config.UpdateRepo != "" {
					repo = m.Shared.Config.UpdateRepo
				}
				m.Shared.RunCommandIndex = -1
				return m, tea.Sequence(PushScreenCmd(ScreenUpdateChecking), runCheckUpdateCmd(repo, m.Shared.CurrentVersion))
			}
			m.Shared.RunCommandIndex = m.SelectedIndex
			return m, tea.Quit
		}
	}
	return m, nil
}
