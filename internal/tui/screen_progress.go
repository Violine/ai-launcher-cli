package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel is the progress/loading screen (also used for update checking).
type ProgressModel struct {
	Shared  *SharedState
	Id      Screen
	Title   string
	Status  string
	Percent int
}

// NewProgressModel creates a new progress screen model. If title/status are empty, defaults are used.
func NewProgressModel(shared *SharedState, title, status string) *ProgressModel {
	if title == "" {
		title = "LOADING"
	}
	if status == "" {
		status = "Loading..."
	}
	return &ProgressModel{Shared: shared, Id: ScreenProgress, Title: title, Status: status, Percent: -1}
}

// NewProgressModelForUpdateChecking creates a progress model for the update-checking screen.
func NewProgressModelForUpdateChecking(shared *SharedState) *ProgressModel {
	return &ProgressModel{
		Shared:  shared,
		Id:      ScreenUpdateChecking,
		Title:   "Проверка обновлений",
		Status:  "Проверка обновлений...",
		Percent: -1,
	}
}

// ID implements ScreenModel.
func (m *ProgressModel) ID() Screen { return m.Id }

// Init implements tea.Model.
func (m *ProgressModel) Init() tea.Cmd { return nil }

// View implements tea.Model.
func (m *ProgressModel) View() tea.View {
	contentWidth := m.Shared.ContentWidth()
	desc := m.Status
	body := BodyStyle.Render(desc)
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n" + BodyStyle.Render("  ")
	barWidth := 40
	if m.Percent >= 0 && m.Percent <= 100 {
		filled := m.Percent * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		line2 := BodyStyle.Render("  ") + HighlightStyle.Render(bar)
		if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
			line2 += BodyStyle.Render(strings.Repeat(" ", pad))
		}
		body += line2 + "\n\n" + BodyStyle.Render("  ")
		line3 := BodyStyle.Render(fmt.Sprintf("%d%%", m.Percent))
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
	return tea.NewView(FrameWithTitle("  "+m.Title+"  ", body, contentWidth))
}

// Update implements tea.Model.
func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
