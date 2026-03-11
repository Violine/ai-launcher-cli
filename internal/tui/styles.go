// Styles for TUI per spec 5.1 (MS-DOS/Norton Commander style).
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

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
	// Root: blue background
	RootStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorBackground)).
			Foreground(lipgloss.Color(ColorText))

	// Title and frame borders: yellow
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			Bold(true)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color(ColorTitleFrame)).
			Padding(0, 1)

	// Body text: white
	BodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText))

	// Highlight (selected row, active field): cyan on black
	HighlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Background(lipgloss.Color(ColorHighlightBg))

	// Buttons: gray bg, black text
	ButtonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(ColorButton)).
			Foreground(lipgloss.Color(ColorButtonText)).
			Padding(0, 2)

	// Selected button: red background
	ButtonActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(ColorButtonActive)).
				Foreground(lipgloss.Color(ColorText)).
				Padding(0, 2)

	// Footer hint (F-keys)
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			MarginTop(1)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorButtonActive))
)

// FrameWithTitle renders a bordered block with title (for token/main screens).
func FrameWithTitle(title, body string) string {
	t := TitleStyle.Render(title)
	inner := BorderStyle.Render(body)
	return RootStyle.Render(lipgloss.JoinVertical(lipgloss.Left, t, "", inner))
}
