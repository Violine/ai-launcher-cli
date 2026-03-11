// Styles for TUI per spec 5.1 (MS-DOS/Norton Commander style).
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// FrameWidth is the minimum width of the content area on all screens (consistent layout).
// Wide enough for footer hints (e.g. "↑/↓ or 1-9: select   Enter: run   F1 Help   F7 Token   F10 Exit").
const FrameWidth = 72

// Colors from spec 5.1
const (
	ColorBackground  = "#0000AA" // blue
	ColorText        = "#FFFFFF" // white
	ColorTitleFrame  = "#FFFF55" // yellow
	ColorHighlight   = "#00AAAA" // cyan
	ColorHighlightBg = "#000000" // black
	ColorButton      = "#808080" // gray
	ColorButtonText  = "#000000" // black
	ColorButtonActive = "#FF0000" // red
)

var (
	// Root: blue background, padding around the whole frame
	RootStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBackground)).
			Foreground(lipgloss.Color(ColorText)).
			Padding(1, 2)

	// Title and frame borders: yellow, rounded
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			Bold(true).
			MarginBottom(0)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorTitleFrame)).
			Padding(0, 2)

	// ContentBox ensures body has fixed width so all screens look the same.
	// Use ContentBoxWidth(w) for dynamic width (e.g. terminal width).
	ContentBoxStyle = lipgloss.NewStyle().
			Width(FrameWidth).
			MaxWidth(FrameWidth)

	// Body text: white
	BodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	// Section heading (e.g. "Commands:")
	SectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			Bold(true)

	// Highlight (selected row, active field): cyan on black
	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Background(lipgloss.Color(ColorHighlightBg)).
			Padding(0, 1)

	// Buttons: gray bg, black text
	ButtonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorButton)).
			Foreground(lipgloss.Color(ColorButtonText)).
			Padding(0, 2).
			Bold(true)

	// Selected button: red background
	ButtonActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorButtonActive)).
				Foreground(lipgloss.Color(ColorText)).
				Padding(0, 2).
				Bold(true)

	// Footer hint (F-keys)
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			MarginTop(1).
			Bold(false)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorButtonActive)).
			Bold(true)

	// Help screen: key description
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Bold(true)
	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))
)

// FrameWithTitle renders a bordered block with title (for token/main/help screens).
// width is the content area width (e.g. from Model.ContentWidth()); if <= 0, FrameWidth is used.
func FrameWithTitle(title, body string, width int) string {
	if width <= 0 {
		width = FrameWidth
	}
	titleRendered := TitleStyle.Render(title)
	contentStyle := ContentBoxStyle.Copy().Width(width).MaxWidth(width)
	bodyBoxed := contentStyle.Render(body)
	inner := BorderStyle.Render(bodyBoxed)
	return RootStyle.Render(lipgloss.JoinVertical(lipgloss.Left, titleRendered, "", inner))
}
