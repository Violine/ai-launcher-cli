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

// --- Navigation: message-based ---

// PushScreenMsg pushes a screen onto the stack.
type PushScreenMsg struct {
	Screen Screen
}

// PopScreenMsg pops the current screen.
type PopScreenMsg struct{}

// ReplaceScreenMsg replaces the top of the stack with a screen.
type ReplaceScreenMsg struct {
	Screen Screen
}

// PushScreenCmd returns a Cmd that emits PushScreenMsg.
func PushScreenCmd(s Screen) tea.Cmd {
	return func() tea.Msg { return PushScreenMsg{Screen: s} }
}

// PopScreenCmd returns a Cmd that emits PopScreenMsg.
func PopScreenCmd() tea.Cmd {
	return func() tea.Msg { return PopScreenMsg{} }
}

// ReplaceScreenCmd returns a Cmd that emits ReplaceScreenMsg.
func ReplaceScreenCmd(s Screen) tea.Cmd {
	return func() tea.Msg { return ReplaceScreenMsg{Screen: s} }
}

// --- Shared state (config, versions, size) ---

// SharedState holds cross-cutting state used by multiple screens.
type SharedState struct {
	Config       *config.Config
	Commands     []string
	CommandNames []string

	Width  int
	Height int

	AvailableVersion string
	CurrentVersion   string
	RunCommandIndex  int

	// Update flow: set by root before pushing error screens
	CheckError   string
	InstallError string
}

// ContentWidth returns the width for frame content so the whole view fits in the terminal.
func (s *SharedState) ContentWidth() int {
	if s.Width <= 0 {
		return FrameWidth
	}
	const extra = 10
	w := s.Width - extra
	if w < 40 {
		w = 40
	}
	if w > FrameWidth {
		w = FrameWidth
	}
	return w
}

// --- Screen model interface ---

// ScreenModel is a tea.Model that represents one screen and knows its ID.
type ScreenModel interface {
	tea.Model
	ID() Screen
}

// --- Root model ---

// RootModel is the root tea.Model: navigation stack + shared state + screen registry.
type RootModel struct {
	Shared    *SharedState
	Stack     []ScreenModel
	NewScreen func(Screen) ScreenModel
}

// NewModel creates the initial TUI model (RootModel). Compatible with tea.NewProgram.
func NewModel(cfg *config.Config, menuLabels []string, commandNames []string, currentVersion string) *RootModel {
	shared := &SharedState{
		Config:           cfg,
		Commands:         menuLabels,
		CommandNames:     commandNames,
		RunCommandIndex:  -1,
		CurrentVersion:   currentVersion,
	}
	root := &RootModel{
		Shared:    shared,
		Stack:     nil,
		NewScreen: nil,
	}
	root.NewScreen = newScreenFactory(root)
	var first Screen
	if cfg == nil || cfg.APIKey == "" {
		first = ScreenToken
	} else {
		first = ScreenMain
	}
	root.Stack = []ScreenModel{root.NewScreen(first)}
	return root
}

// newScreenFactory returns a function that builds screen models (used by RootModel).
func newScreenFactory(root *RootModel) func(Screen) ScreenModel {
	return func(s Screen) ScreenModel {
		switch s {
		case ScreenMain:
			return NewMainModel(root.Shared)
		case ScreenToken:
			return NewTokenModel(root.Shared)
		case ScreenHelp:
			return NewHelpModel(root.Shared)
		case ScreenProgress:
			return NewProgressModel(root.Shared, "", "")
		case ScreenUpdateConfirm:
			return NewUpdateConfirmModel(root.Shared)
		case ScreenUpdateChecking:
			return NewProgressModelForUpdateChecking(root.Shared)
		case ScreenUpdateCheckError:
			return NewUpdateCheckErrorModel(root.Shared)
		case ScreenUpdateInstallError:
			return NewUpdateInstallErrorModel(root.Shared)
		case ScreenUpdateSuccess:
			return NewUpdateSuccessModel(root.Shared)
		default:
			return NewMainModel(root.Shared)
		}
	}
}

// NewModelNoConfig creates a model without config (backward compat); starts on token screen.
func NewModelNoConfig(commands []string) *RootModel {
	return NewModel(nil, commands, commands, "0.0.0")
}

// RunCommandIndex returns the selected command index after quit (for main.go).
func (m *RootModel) RunCommandIndex() int { return m.Shared.RunCommandIndex }

// CommandNames returns command names (for main.go).
func (m *RootModel) CommandNames() []string { return m.Shared.CommandNames }

func (m *RootModel) current() ScreenModel {
	if len(m.Stack) == 0 {
		return nil
	}
	return m.Stack[len(m.Stack)-1]
}

func (m *RootModel) pushScreen(s Screen) {
	m.Stack = append(m.Stack, m.NewScreen(s))
}

func (m *RootModel) popScreen() {
	if len(m.Stack) > 1 {
		m.Stack = m.Stack[:len(m.Stack)-1]
	}
}

func (m *RootModel) replaceScreen(s Screen) {
	if len(m.Stack) == 0 {
		m.pushScreen(s)
		return
	}
	m.Stack[len(m.Stack)-1] = m.NewScreen(s)
}

// Init implements tea.Model.
func (m *RootModel) Init() tea.Cmd {
	cur := m.current()
	if cur == nil {
		return nil
	}
	return cur.Init()
}

// Update implements tea.Model.
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case UpdateAvailableMsg:
		m.Shared.AvailableVersion = msg.Version
		return m, nil
	case UpdateCheckErrorMsg:
		m.Shared.CheckError = ""
		if msg.Err != nil {
			m.Shared.CheckError = msg.Err.Error()
		}
		m.pushScreen(ScreenUpdateCheckError)
		return m, nil
	case UpdateCheckResultMsg:
		if msg.Err != nil {
			m.Shared.CheckError = msg.Err.Error()
			m.replaceScreen(ScreenUpdateCheckError)
			return m, nil
		}
		if msg.Available != "" {
			m.Shared.AvailableVersion = msg.Available
			m.replaceScreen(ScreenUpdateConfirm)
			return m, nil
		}
		m.popScreen()
		return m, nil
	case InstallDoneMsg:
		if msg.Err != nil {
			m.Shared.InstallError = msg.Err.Error()
			m.replaceScreen(ScreenUpdateInstallError)
			return m, nil
		}
		m.replaceScreen(ScreenUpdateSuccess)
		return m, nil
	case tea.WindowSizeMsg:
		m.Shared.Width = msg.Width
		m.Shared.Height = msg.Height
		return m, nil
	case PushScreenMsg:
		m.pushScreen(msg.Screen)
		return m, nil
	case PopScreenMsg:
		m.popScreen()
		return m, nil
	case ReplaceScreenMsg:
		m.replaceScreen(msg.Screen)
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "f10":
			return m, tea.Quit
		case "f1":
			if cur := m.current(); cur != nil && cur.ID() == ScreenHelp {
				m.popScreen()
				return m, nil
			}
			m.pushScreen(ScreenHelp)
			return m, nil
		}
	}

	cur := m.current()
	if cur == nil {
		return m, nil
	}
	updated, cmd := cur.Update(msg)
	if updated != nil {
		if sm, ok := updated.(ScreenModel); ok {
			m.Stack[len(m.Stack)-1] = sm
		}
	}
	return m, cmd
}

// View implements tea.Model.
func (m *RootModel) View() tea.View {
	cur := m.current()
	if cur == nil {
		return tea.NewView("")
	}
	return cur.View()
}
