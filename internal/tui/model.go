// Package tui implements the TUI using Bubble Tea and Lip Gloss.
package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/config"
	"github.com/ai-launcher/cli/internal/updater"
)

// Screen is the current TUI screen.
type Screen int

const (
	ScreenMain   Screen = iota
	ScreenToken
	ScreenHelp
	ScreenProgress
	ScreenUpdateConfirm
	ScreenUpdateChecking
	ScreenUpdateCheckError
	ScreenUpdateInstallError
	ScreenUpdateSuccess
)

// Button for token screen OK/Cancel.
type TokenButton int

const (
	TokenButtonOK TokenButton = iota
	TokenButtonCancel
)

type UpdateButton int

const (
	UpdateButtonNo UpdateButton = iota
	UpdateButtonYes
)

// UpdateAvailableMsg is sent when a newer version is found (from background check).
// Model sets AvailableVersion so the main menu can show "update available" next to the item.
type UpdateAvailableMsg struct {
	Version string
}

type UpdateCheckErrorMsg struct {
	Err error
}

type UpdateCheckResultMsg struct {
	Available string
	Err       error
}

type InstallDoneMsg struct {
	Err error
}

// Model holds the TUI state.
type Model struct {
	Screen   Screen
	Config   *config.Config
	Commands []string // display labels for menu
	CommandNames []string // plugin names in same order (for Run on Enter)

	// Main screen: selected menu index (0-based)
	SelectedIndex int
	// After quit: if >= 0, main should run CommandNames[RunCommandIndex]
	RunCommandIndex int

	// When Screen == ScreenHelp, return to this screen on F1/Esc
	PrevScreen Screen

	// Token screen state
	TokenInput     string
	TokenError     string
	TokenButtonFoc TokenButton

	// Terminal size (from tea.WindowSizeMsg); 0 until first message
	Width  int
	Height int

	// Progress screen (ScreenProgress): title, status line, percent (-1 = indeterminate)
	ProgressTitle   string
	ProgressStatus  string
	ProgressPercent int // 0–100 or -1 for indeterminate

	// Set when background check finds a newer version; show next to "AI Launcher Update" in menu
	AvailableVersion string
	CurrentVersion   string
	UpdateButtonFoc UpdateButton
	CheckError      string
	InstallError    string
	// Error screens (CheckError / InstallError): 0 = Retry, 1 = Cancel
	ErrorButtonFoc int
}

// NewModel creates the initial TUI model. currentVersion is shown on update confirm (e.g. autoupdate.Version).
func NewModel(cfg *config.Config, menuLabels []string, commandNames []string, currentVersion string) Model {
	m := Model{
		Config:           cfg,
		Commands:         menuLabels,
		CommandNames:     commandNames,
		TokenButtonFoc:   TokenButtonOK,
		UpdateButtonFoc:  UpdateButtonNo,
		RunCommandIndex:  -1,
		CurrentVersion:   currentVersion,
	}
	if cfg == nil || cfg.APIKey == "" {
		m.Screen = ScreenToken
	} else {
		m.Screen = ScreenMain
	}
	return m
}

// NewModelNoConfig creates a model without config (backward compat); starts on token screen.
// commands is used as both display labels and command names.
func NewModelNoConfig(commands []string) Model {
	return NewModel(nil, commands, commands, "0.0.0")
}

// ContentWidth returns the width for frame content so the whole view fits in the terminal.
// RootStyle: 2 cols left/right; BorderStyle: 2 padding + 1 border each side → 2+2+1+1+2+2 = 10 extra.
// So contentWidth = terminal width − 10 (min 40, max FrameWidth).
func (m Model) ContentWidth() int {
	if m.Width <= 0 {
		return FrameWidth
	}
	const extra = 10 // root padding 2+2, border 1+1, inner padding 2+2
	w := m.Width - extra
	if w < 40 {
		w = 40
	}
	if w > FrameWidth {
		w = FrameWidth
	}
	return w
}

// Init runs once at startup. WindowSizeMsg is sent by the runtime when the terminal is ready.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages (delegates to screen-specific logic; main screen and quit handled here).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateAvailableMsg:
		m.AvailableVersion = msg.Version
		return m, nil
	case UpdateCheckErrorMsg:
		m.CheckError = ""
		m.ErrorButtonFoc = 0
		if msg.Err != nil {
			m.CheckError = msg.Err.Error()
		}
		m.Screen = ScreenUpdateCheckError
		return m, nil
	case UpdateCheckResultMsg:
		m.Screen = ScreenMain
		if msg.Err != nil {
			m.CheckError = msg.Err.Error()
			m.Screen = ScreenUpdateCheckError
			return m, nil
		}
		if msg.Available != "" {
			m.AvailableVersion = msg.Available
			m.Screen = ScreenUpdateConfirm
		}
		return m, nil
	case InstallDoneMsg:
		if msg.Err != nil {
			m.InstallError = msg.Err.Error()
			m.ErrorButtonFoc = 0
			m.Screen = ScreenUpdateInstallError
			return m, nil
		}
		m.Screen = ScreenUpdateSuccess
		return m, nil
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "f10":
			return m, tea.Quit
		case "f1":
			if m.Screen == ScreenHelp {
				m.Screen = m.PrevScreen
				return m, nil
			}
			m.PrevScreen = m.Screen
			m.Screen = ScreenHelp
			return m, nil
		}
	}
	// Delegate to screen
	switch m.Screen {
	case ScreenHelp:
		return updateHelpScreen(m, msg)
	case ScreenToken:
		return updateTokenScreen(m, msg)
	case ScreenProgress:
		return updateProgressScreen(m, msg)
	case ScreenUpdateConfirm:
		return updateUpdateConfirmScreen(m, msg)
	case ScreenUpdateChecking:
		return updateProgressScreen(m, msg)
	case ScreenUpdateCheckError:
		return updateUpdateCheckErrorScreen(m, msg)
	case ScreenUpdateInstallError:
		return updateUpdateInstallErrorScreen(m, msg)
	case ScreenUpdateSuccess:
		return updateUpdateSuccessScreen(m, msg)
	default:
		return updateMainScreen(m, msg)
	}
}

// View renders the UI.
func (m Model) View() tea.View {
	switch m.Screen {
	case ScreenHelp:
		return viewHelpScreen(m)
	case ScreenToken:
		return viewTokenScreen(m)
	case ScreenProgress:
		return viewProgressScreen(m)
	case ScreenUpdateConfirm:
		return viewUpdateConfirmScreen(m)
	case ScreenUpdateChecking:
		return viewProgressScreen(m)
	case ScreenUpdateCheckError:
		return viewUpdateCheckErrorScreen(m)
	case ScreenUpdateInstallError:
		return viewUpdateInstallErrorScreen(m)
	case ScreenUpdateSuccess:
		return viewUpdateSuccessScreen(m)
	default:
		return viewMainScreen(m)
	}
}

func viewMainScreen(m Model) tea.View {
	// Render menu manually (no list component) to avoid duplication when selection changes
	versionStr := m.CurrentVersion
	if versionStr == "" {
		versionStr = "0.0.0"
	}
	body := SectionStyle.Render("Commands:")
	body += "\n\n"
	sel := m.SelectedIndex
	for i, label := range m.Commands {
		num := fmt.Sprintf("%d.", i+1)
		line := "  " + BodyStyle.Render(num) + " "
		// For autoupdate item, show explicit "доступно обновление до X.Y.Z" when update is available
		displayLabel := label
		if i < len(m.CommandNames) && m.CommandNames[i] == "autoupdate" && m.AvailableVersion != "" {
			displayLabel = label + "  " + HelpKeyStyle.Render("— доступно обновление до "+m.AvailableVersion)
		}
		if i == sel {
			line += HighlightStyle.Render(displayLabel)
		} else {
			line += BodyStyle.Render(displayLabel)
		}
		body += line + "\n"
	}
	body += "\n" + FooterStyle.Render("↑/↓ or 1-9: select   Enter: run   F1 Help   F7 Token   F10 Exit")
	rendered := FrameWithTitleSubtitle("  AI LAUNCHER  ", "  v "+versionStr, body, m.ContentWidth())
	return tea.NewView(rendered)
}

func viewTokenScreen(m Model) tea.View {
	body := BodyStyle.Render("Enter your API token to continue:")
	body += "\n\n  "
	// Input line: mask with * (FR-102); highlight only the content, not padding — no black empty cells
	fieldWidth := m.ContentWidth() - 6
	if fieldWidth < 20 {
		fieldWidth = 20
	}
	mask := strings.Repeat("*", len(m.TokenInput))
	if mask == "" {
		mask = " "
	}
	// Only the actual "field" (space + mask + space) gets HighlightStyle; padding uses background
	content := " " + mask + " "
	body += HighlightStyle.Render(content)
	if len(content) < fieldWidth {
		body += BodyStyle.Render(strings.Repeat(" ", fieldWidth-len(content)))
	}
	body += "\n\n  "
	// Buttons [ OK ] [ Cancel ]
	okBtn := ButtonStyle.Render("[ OK ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.TokenButtonFoc == TokenButtonOK {
		okBtn = ButtonActiveStyle.Render("[ OK ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	body += okBtn + "  " + cancelBtn + "\n"
	if m.TokenError != "" {
		body += "\n  " + ErrorStyle.Render(m.TokenError) + "\n"
	}
	body += "\n" + FooterStyle.Render("F1 Help   Tab: switch button   Enter: confirm   Esc: back   F10 Exit")
	rendered := FrameWithTitle("  API TOKEN  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func viewHelpScreen(m Model) tea.View {
	lines := []string{
		HelpKeyStyle.Render("F1") + "    " + HelpDescStyle.Render("Show / hide this help"),
		HelpKeyStyle.Render("F7") + "    " + HelpDescStyle.Render("Open API token screen (from main menu)"),
		HelpKeyStyle.Render("F10") + "   " + HelpDescStyle.Render("Exit application"),
		"",
		HelpKeyStyle.Render("↑/↓ or 1-9") + "  " + HelpDescStyle.Render("Select menu item (main screen)"),
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

// viewProgressScreen renders the loading/download screen (ТЗ 5.2.2, FR-202).
// ProgressPercent 0–100 = progress bar; -1 = indeterminate.
func viewProgressScreen(m Model) tea.View {
	title := m.ProgressTitle
	if title == "" {
		title = "LOADING"
	}
	desc := "Loading..."
	if m.ProgressStatus != "" {
		desc = m.ProgressStatus
	} else if m.ProgressTitle != "" {
		desc = m.ProgressTitle
	}
	body := BodyStyle.Render(desc) + "\n\n  "
	barWidth := 40
	if m.ProgressPercent >= 0 && m.ProgressPercent <= 100 {
		filled := m.ProgressPercent * barWidth / 100
		if filled > barWidth {
			filled = barWidth
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		body += HighlightStyle.Render(bar) + "\n\n  "
		body += BodyStyle.Render(fmt.Sprintf("%d%%", m.ProgressPercent))
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

func updateMainScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	n := len(m.Commands)
	if n == 0 {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "f7":
			// FR-107: switch to token screen (reset/change key)
			m.Screen = ScreenToken
			m.TokenInput = ""
			m.TokenError = ""
			if m.Config != nil {
				m.Config.APIKey = ""
			}
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
				// Full update flow in TUI; do not quit
				if m.AvailableVersion != "" {
					m.Screen = ScreenUpdateConfirm
					return m, nil
				}
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenUpdateChecking
				m.ProgressTitle = "Проверка обновлений"
				m.ProgressStatus = "Проверка обновлений..."
				m.ProgressPercent = -1
				return m, runCheckUpdateCmd(repo, m.CurrentVersion)
			}
			m.RunCommandIndex = m.SelectedIndex
			return m, tea.Quit
		}
	}
	return m, nil
}

func updateTokenScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		s := msg.String()
		switch s {
		case "enter":
			if m.TokenButtonFoc == TokenButtonOK {
				// Validate and save (FR-103, FR-104, FR-105/106)
				if err := config.ValidateAPIKey(m.TokenInput); err != nil {
					m.TokenError = err.Error()
					return m, nil
				}
				cfg := m.Config
				if cfg == nil {
					cfg = &config.Config{}
				}
				cfg.APIKey = strings.TrimSpace(m.TokenInput)
				if err := config.Save(cfg); err != nil {
					m.TokenError = err.Error()
					return m, nil
				}
				m.Config = cfg
				m.Screen = ScreenMain
				m.TokenError = ""
				m.TokenInput = ""
				return m, nil
			}
			// Cancel: return to main menu (do not quit)
			m.Screen = ScreenMain
			m.TokenError = ""
			m.TokenInput = ""
			return m, nil
		case "tab", "right":
			m.TokenButtonFoc = TokenButtonCancel
			m.TokenError = ""
			return m, nil
		case "shift+tab", "left":
			m.TokenButtonFoc = TokenButtonOK
			m.TokenError = ""
			return m, nil
		case "esc":
			// Esc: return to main menu (same as Cancel)
			m.Screen = ScreenMain
			m.TokenError = ""
			m.TokenInput = ""
			return m, nil
		case "backspace":
			if len(m.TokenInput) > 0 {
				m.TokenInput = m.TokenInput[:len(m.TokenInput)-1]
				m.TokenError = ""
			}
			return m, nil
		default:
			// Single rune: append to input (FR-101)
			if len(s) == 1 && s[0] >= 32 && s[0] != 127 {
				m.TokenInput += s
				m.TokenError = ""
				return m, nil
			}
		}
	}
	return m, nil
}

func runCheckUpdateCmd(repo, currentVersion string) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				msg = UpdateCheckResultMsg{Err: fmt.Errorf("проверка обновлений: %v", r)}
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		latest, err := updater.LatestRelease(ctx, repo)
		if err != nil {
			return UpdateCheckResultMsg{Err: err}
		}
		if latest == "" {
			return UpdateCheckResultMsg{}
		}
		newer, err := updater.NewerThan(currentVersion, latest)
		if err != nil {
			return UpdateCheckResultMsg{Err: err}
		}
		if !newer {
			return UpdateCheckResultMsg{}
		}
		return UpdateCheckResultMsg{Available: latest}
	}
}

func runInstallCmd(repo, version string) tea.Cmd {
	return func() (msg tea.Msg) {
		defer func() {
			if r := recover(); r != nil {
				msg = InstallDoneMsg{Err: fmt.Errorf("установка: %v", r)}
			}
		}()
		currentPath, err := os.Executable()
		if err != nil {
			return InstallDoneMsg{Err: fmt.Errorf("executable path: %w", err)}
		}
		downloadURL := updater.DownloadURL(repo, version)
		if err := updater.Install(downloadURL, currentPath, nil); err != nil {
			return InstallDoneMsg{Err: err}
		}
		return InstallDoneMsg{}
	}
}

func viewUpdateConfirmScreen(m Model) tea.View {
	current := m.CurrentVersion
	if current == "" {
		current = "0.0.0"
	}
	body := BodyStyle.Render(fmt.Sprintf("Доступна версия %s (текущая: %s). Установить?", m.AvailableVersion, current))
	body += "\n\n  "
	noBtn := ButtonStyle.Render("[ N ] Нет")
	yesBtn := ButtonStyle.Render("[ Y ] Да")
	if m.UpdateButtonFoc == UpdateButtonNo {
		noBtn = ButtonActiveStyle.Render("[ N ] Нет")
	} else {
		yesBtn = ButtonActiveStyle.Render("[ Y ] Да")
	}
	body += noBtn + "  " + yesBtn + "\n"
	body += "\n" + FooterStyle.Render("Tab: переключить   Enter: подтвердить   Esc: отмена")
	rendered := FrameWithTitle("  ОБНОВЛЕНИЕ  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateUpdateConfirmScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.UpdateButtonFoc == UpdateButtonYes {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenProgress
				m.ProgressTitle = "Update"
				m.ProgressStatus = "Скачивание и установка..."
				m.ProgressPercent = -1
				return m, runInstallCmd(repo, m.AvailableVersion)
			}
			m.Screen = ScreenMain
			return m, nil
		case "tab", "right":
			if m.UpdateButtonFoc == UpdateButtonNo {
				m.UpdateButtonFoc = UpdateButtonYes
			} else {
				m.UpdateButtonFoc = UpdateButtonNo
			}
			return m, nil
		case "shift+tab", "left":
			if m.UpdateButtonFoc == UpdateButtonYes {
				m.UpdateButtonFoc = UpdateButtonNo
			} else {
				m.UpdateButtonFoc = UpdateButtonYes
			}
			return m, nil
		case "esc":
			m.Screen = ScreenMain
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateCheckErrorScreen(m Model) tea.View {
	body := BodyStyle.Render("Не удалось проверить обновления:")
	body += "\n\n  " + ErrorStyle.Render(m.CheckError)
	body += "\n\n  "
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.ErrorButtonFoc == 0 {
		retryBtn = ButtonActiveStyle.Render("[ Retry ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	body += retryBtn + "  " + cancelBtn
	body += "\n\n" + FooterStyle.Render("Tab / ←→: переключить   Enter: подтвердить   Esc: в меню")
	rendered := FrameWithTitle("  Ошибка проверки  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateUpdateCheckErrorScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenUpdateChecking
				m.ProgressTitle = "Проверка обновлений"
				m.ProgressStatus = "Проверка обновлений..."
				m.ProgressPercent = -1
				return m, runCheckUpdateCmd(repo, m.CurrentVersion)
			}
			m.Screen = ScreenMain
			m.CheckError = ""
			return m, nil
		case "tab", "right":
			m.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.Screen = ScreenMain
			m.CheckError = ""
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateInstallErrorScreen(m Model) tea.View {
	body := BodyStyle.Render("Ошибка при установке:")
	body += "\n\n  " + ErrorStyle.Render(m.InstallError)
	body += "\n\n  "
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.ErrorButtonFoc == 0 {
		retryBtn = ButtonActiveStyle.Render("[ Retry ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	body += retryBtn + "  " + cancelBtn
	body += "\n\n" + FooterStyle.Render("Tab / ←→: переключить   Enter: подтвердить   Esc: в меню")
	rendered := FrameWithTitle("  Ошибка установки  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateUpdateInstallErrorScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenProgress
				m.ProgressTitle = "Update"
				m.ProgressStatus = "Скачивание и установка..."
				m.ProgressPercent = -1
				return m, runInstallCmd(repo, m.AvailableVersion)
			}
			m.Screen = ScreenMain
			m.InstallError = ""
			return m, nil
		case "tab", "right":
			m.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.Screen = ScreenMain
			m.InstallError = ""
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateSuccessScreen(m Model) tea.View {
	body := BodyStyle.Render("Обновление установлено.")
	body += "\n\n  " + BodyStyle.Render("Перезапустите приложение для использования новой версии.")
	body += "\n\n" + FooterStyle.Render("Нажмите Enter для выхода")
	rendered := FrameWithTitle("  Установлено  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateUpdateSuccessScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}
