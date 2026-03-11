package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

func viewProgressScreen(m Model) tea.View {
	contentWidth := m.ContentWidth()
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
	body := BodyStyle.Render(desc)
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n" + BodyStyle.Render("  ")
	barWidth := 40
	if m.Progress.Percent >= 0 && m.Progress.Percent <= 100 {
		filled := m.Progress.Percent * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		line2 := BodyStyle.Render("  ") + HighlightStyle.Render(bar)
		if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
			line2 += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line2 + "\n\n" + BodyStyle.Render("  ")
		line3 := BodyStyle.Render(fmt.Sprintf("%d%%", m.Progress.Percent))
		if pad := contentWidth - lipgloss.Width(line3); pad > 0 {
			line3 += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line3
	} else {
		bar := strings.Repeat("░", barWidth)
		line2 := BodyStyle.Render("  ") + BodyStyle.Render(bar)
		if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
			line2 += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line2 + "\n\n" + BodyStyle.Render("  ")
		line3 := BodyStyle.Render("Please wait...")
		if pad := contentWidth - lipgloss.Width(line3); pad > 0 {
			line3 += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line3
	}
	body += "\n\n"
	footerStr := ""
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	body += FooterStyle.Render(footerStr+strings.Repeat(" ", pad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	rendered := FrameWithTitle("  "+title+"  ", body, contentWidth)
	return tea.NewView(rendered)
}

func updateProgressScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
