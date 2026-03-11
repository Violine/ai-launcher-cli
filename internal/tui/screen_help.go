package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func viewHelpScreen(m Model) tea.View {
	n := len(m.Commands)
	if n <= 0 {
		n = 4
	}
	lines := []string{
		HelpKeyStyle.Render("F1") + "    " + HelpDescStyle.Render("Show / hide this help"),
		HelpKeyStyle.Render("F7") + "    " + HelpDescStyle.Render("Open API token screen (from main menu)"),
		HelpKeyStyle.Render("F10") + "   " + HelpDescStyle.Render("Exit application"),
		"",
		HelpKeyStyle.Render(fmt.Sprintf("↑/↓ or 1-%d", n)) + "  " + HelpDescStyle.Render("Select menu item (main screen)"),
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
