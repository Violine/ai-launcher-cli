// Package tui implements the TUI using Bubble Tea and Lip Gloss.
package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/ai-launcher/cli/internal/config"
)

// Screen is a TUI screen identifier.
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

// TokenButton is the focused button on the token screen.
type TokenButton int

const (
	TokenButtonOK TokenButton = iota
	TokenButtonCancel
)

// UpdateButton is the focused button on the update confirm screen.
type UpdateButton int

const (
	UpdateButtonNo UpdateButton = iota
	UpdateButtonYes
)

// UpdateAvailableMsg is sent when a newer version is found (from background check).
type UpdateAvailableMsg struct {
	Version string
}

// UpdateCheckErrorMsg is sent when the background update check fails.
type UpdateCheckErrorMsg struct {
	Err error
}

// UpdateCheckResultMsg is sent when an explicit update check completes (user chose Update from menu).
type UpdateCheckResultMsg struct {
	Available string
	Err       error
}

// InstallDoneMsg is sent when updater.Install finishes (from TUI install Cmd).
type InstallDoneMsg struct {
	Err error
}

// TokenScreenState holds state for the token (API key) screen.
type TokenScreenState struct {
	Input    string
	Error    string
	ButtonFoc TokenButton
}

// ProgressState holds state for the progress/loading screen.
type ProgressState struct {
	Title   string
	Status  string
	Percent int
}

// UpdateScreenState holds state for all update-related screens (confirm, check error, install error).
type UpdateScreenState struct {
	ConfirmButtonFoc UpdateButton
	CheckError       string
	InstallError     string
	ErrorButtonFoc   int
}

// Model holds the TUI state.
type Model struct {
	ScreenStack []Screen
	Config      *config.Config
	Commands    []string
	CommandNames []string

	SelectedIndex   int
	RunCommandIndex int

	Width  int
	Height int

	AvailableVersion string
	CurrentVersion   string

	Token       TokenScreenState
	Progress    ProgressState
	UpdateState UpdateScreenState
}

// NewModel creates the initial TUI model.
func NewModel(cfg *config.Config, menuLabels []string, commandNames []string, currentVersion string) Model {
	m := Model{
		Config:           cfg,
		Commands:         menuLabels,
		CommandNames:     commandNames,
		RunCommandIndex:  -1,
		CurrentVersion:   currentVersion,
		Token:            TokenScreenState{ButtonFoc: TokenButtonOK},
		UpdateState:      UpdateScreenState{ConfirmButtonFoc: UpdateButtonNo},
	}
	if cfg == nil || cfg.APIKey == "" {
		m.ScreenStack = []Screen{ScreenToken}
	} else {
		m.ScreenStack = []Screen{ScreenMain}
	}
	return m
}

// NewModelNoConfig creates a model without config (backward compat); starts on token screen.
func NewModelNoConfig(commands []string) Model {
	return NewModel(nil, commands, commands, "0.0.0")
}

// ContentWidth returns the width for frame content so the whole view fits in the terminal.
func (m Model) ContentWidth() int {
	if m.Width <= 0 {
		return FrameWidth
	}
	const extra = 10
	w := m.Width - extra
	if w < 40 {
		w = 40
	}
	if w > FrameWidth {
		w = FrameWidth
	}
	return w
}

// CurrentScreen returns the top of the screen stack, or ScreenMain if stack is empty.
func (m Model) CurrentScreen() Screen {
	if len(m.ScreenStack) == 0 {
		return ScreenMain
	}
	return m.ScreenStack[len(m.ScreenStack)-1]
}

// PushScreen returns a copy of m with s pushed onto the screen stack.
func PushScreen(m Model, s Screen) Model {
	newStack := make([]Screen, len(m.ScreenStack)+1)
	copy(newStack, m.ScreenStack)
	newStack[len(newStack)-1] = s
	m.ScreenStack = newStack
	return m
}

// PopScreen returns a copy of m with the top screen popped; no-op if stack has at most one element.
func PopScreen(m Model) Model {
	if len(m.ScreenStack) <= 1 {
		return m
	}
	newStack := make([]Screen, len(m.ScreenStack)-1)
	copy(newStack, m.ScreenStack[:len(m.ScreenStack)-1])
	m.ScreenStack = newStack
	return m
}

// ReplaceScreen returns a copy of m with the top of the stack replaced by s; no-op if stack is empty.
func ReplaceScreen(m Model, s Screen) Model {
	if len(m.ScreenStack) == 0 {
		return m
	}
	newStack := make([]Screen, len(m.ScreenStack))
	copy(newStack, m.ScreenStack)
	newStack[len(newStack)-1] = s
	m.ScreenStack = newStack
	return m
}

// Init runs once at startup.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and delegates to screen-specific update.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateAvailableMsg:
		m.AvailableVersion = msg.Version
		return m, nil
	case UpdateCheckErrorMsg:
		m.UpdateState.CheckError = ""
		m.UpdateState.ErrorButtonFoc = 0
		if msg.Err != nil {
			m.UpdateState.CheckError = msg.Err.Error()
		}
		m = PushScreen(m, ScreenUpdateCheckError)
		return m, nil
	case UpdateCheckResultMsg:
		if msg.Err != nil {
			m.UpdateState.CheckError = msg.Err.Error()
			m = ReplaceScreen(m, ScreenUpdateCheckError)
			return m, nil
		}
		if msg.Available != "" {
			m.AvailableVersion = msg.Available
			m = ReplaceScreen(m, ScreenUpdateConfirm)
			return m, nil
		}
		m = PopScreen(m)
		return m, nil
	case InstallDoneMsg:
		if msg.Err != nil {
			m.UpdateState.InstallError = msg.Err.Error()
			m.UpdateState.ErrorButtonFoc = 0
			m = ReplaceScreen(m, ScreenUpdateInstallError)
			return m, nil
		}
		m = ReplaceScreen(m, ScreenUpdateSuccess)
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
			if m.CurrentScreen() == ScreenHelp {
				m = PopScreen(m)
				return m, nil
			}
			m = PushScreen(m, ScreenHelp)
			return m, nil
		}
	}
	switch m.CurrentScreen() {
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

// View renders the UI and delegates to screen-specific view.
func (m Model) View() tea.View {
	switch m.CurrentScreen() {
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
