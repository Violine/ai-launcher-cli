package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func viewProgressScreen(m Model) tea.View {
	title := m.Progress.Title
	if title == "" {
		title = "LOADING"
	}
	desc := "Loading..."
	if m.Progress.Status != "" {
		desc = m.Progress.Status
	} else if m.Progress.Title != "" {
		desc = m.Progress.Title
	}
	body := BodyStyle.Render(desc) + "\n\n  "
	barWidth := 40
	if m.Progress.Percent >= 0 && m.Progress.Percent <= 100 {
		filled := m.Progress.Percent * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		body += HighlightStyle.Render(bar) + "\n\n  "
		body += BodyStyle.Render(fmt.Sprintf("%d%%", m.Progress.Percent))
	} else {
		bar := strings.Repeat("░", barWidth)
		body += BodyStyle.Render(bar) + "\n\n  "
		body += BodyStyle.Render("Please wait...")
	}
	body += "\n\n" + FooterStyle.Render("")
	rendered := FrameWithTitle("  "+title+"  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateProgressScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
