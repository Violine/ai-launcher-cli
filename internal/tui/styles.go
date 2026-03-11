// Styles for TUI per spec 5.1 (MS-DOS/Norton Commander style).
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// FrameWidth is the minimum width of the content area on all screens (consistent layout).
// Wide enough for footer hints (e.g. "↑/↓ or 1-4: select   Enter: run   F1 Help   F7 Token   F10 Exit").
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
			Background(lipgloss.Color(ColorBackground)).
			Padding(0, 2)

	// ContentBox ensures body has fixed width and blue fill so no black gaps between lines.
	ContentBoxStyle = lipgloss.NewStyle().
			Width(FrameWidth).
			MaxWidth(FrameWidth).
			Background(lipgloss.Color(ColorBackground))

	// Body text: white on blue so content area stays filled
	BodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorBackground))

	// Section heading (e.g. "Commands:")
	SectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			Bold(true).
			Background(lipgloss.Color(ColorBackground))

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

	// Footer hint (F-keys): on blue
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorTitleFrame)).
			Background(lipgloss.Color(ColorBackground)).
			MarginTop(1).
			Bold(false)

	// Error message: on blue
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorButtonActive)).
			Bold(true).
			Background(lipgloss.Color(ColorBackground))

	// Help screen: key description, on blue
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHighlight)).
			Bold(true).
			Background(lipgloss.Color(ColorBackground))
	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			Background(lipgloss.Color(ColorBackground))

	// Version under the title: not bold, dimmer (visually smaller)
	VersionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Bold(false)
)

// FrameWithTitle renders a bordered block with title (for token/main/help screens).
// width is the content area width (e.g. from Model.ContentWidth()); if <= 0, FrameWidth is used.
func FrameWithTitle(title, body string, width int) string {
	return FrameWithTitleSubtitle(title, "", body, width)
}

// FrameWithTitleSubtitle renders title, optional subtitle (e.g. version) directly under it, then the bordered body.
// Title and subtitle are given the same blue background as root so the whole header area is filled.
func FrameWithTitleSubtitle(title, subtitle, body string, width int) string {
	if width <= 0 {
		width = FrameWidth
	}
	// Bordered box width = content width + inner padding 2+2 + border 1+1
	fullWidth := width + 6
	titleBlock := TitleStyle.Copy().
		Background(lipgloss.Color(ColorBackground)).
		Width(fullWidth).
		Render(title)
	blocks := []string{titleBlock}
	if subtitle != "" {
		blocks = append(blocks, VersionStyle.Copy().
			Background(lipgloss.Color(ColorBackground)).
			Width(fullWidth).
			Render(subtitle))
	}
	blocks = append(blocks, "", BorderStyle.Render(ContentBoxStyle.Copy().Width(width).MaxWidth(width).Render(body)))
	return RootStyle.Render(lipgloss.JoinVertical(lipgloss.Left, blocks...))
}
