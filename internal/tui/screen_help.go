package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel is the help screen.
type HelpModel struct {
	Shared *SharedState
}

// NewHelpModel creates a new help screen model.
func NewHelpModel(shared *SharedState) *HelpModel {
	return &HelpModel{Shared: shared}
}

// ID implements ScreenModel.
func (m *HelpModel) ID() Screen { return ScreenHelp }

// Init implements tea.Model.
func (m *HelpModel) Init() tea.Cmd { return nil }

// View implements tea.Model.
func (m *HelpModel) View() tea.View {
	contentWidth := m.Shared.ContentWidth()
	n := len(m.Shared.Commands)
	if n <= 0 {
		n = 4
	}
	pad := func(s string) string { return HelpDescStyle.Render(s) }
	lines := []string{
		HelpKeyStyle.Render("F1") + pad("    ") + HelpDescStyle.Render("Show / hide this help"),
		HelpKeyStyle.Render("F7") + pad("    ") + HelpDescStyle.Render("Open API token screen (from main menu)"),
		HelpKeyStyle.Render("F10") + pad("   ") + HelpDescStyle.Render("Exit application"),
		"",
		HelpKeyStyle.Render(fmt.Sprintf("↑/↓ or 1-%d", n)) + pad("  ") + HelpDescStyle.Render("Select menu item (main screen)"),
		HelpKeyStyle.Render("Enter") + pad("  ") + HelpDescStyle.Render("Run selected command (main) or confirm (token)"),
		"",
		HelpKeyStyle.Render("Tab / ←→") + pad("  ") + HelpDescStyle.Render("Switch between OK / Cancel on token screen"),
		HelpKeyStyle.Render("Esc") + pad("    ") + HelpDescStyle.Render("Cancel / back to menu"),
	}
	body := ""
	for _, line := range lines {
		if line == "" {
			body += HelpDescStyle.Render(strings.Repeat(" ", contentWidth)) + "\n"
			continue
		}
		body += line
		if p := contentWidth - lipgloss.Width(line); p > 0 {
			body += HelpDescStyle.Render(strings.Repeat(" ", p))
		}
		body += "\n"
	}
	footerStr := "F1 or Esc to close help"
	footerPad := contentWidth - lipgloss.Width(footerStr)
	if footerPad < 0 {
		footerPad = 0
	}
	body += "\n" + FooterStyle.Render(footerStr+strings.Repeat(" ", footerPad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	return tea.NewView(FrameWithTitle("  HELP  ", body, contentWidth))
}

// Update implements tea.Model.
func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "f1", "esc":
			return m, PopScreenCmd()
		}
	}
	return m, nil
}
