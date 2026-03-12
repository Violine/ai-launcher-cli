package tui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/ai-launcher/cli/internal/updater"
)

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

// --- UpdateConfirmModel ---

type UpdateConfirmModel struct {
	Shared        *SharedState
	ConfirmButtonFoc UpdateButton
}

func NewUpdateConfirmModel(shared *SharedState) *UpdateConfirmModel {
	return &UpdateConfirmModel{Shared: shared, ConfirmButtonFoc: UpdateButtonYes}
}

func (m *UpdateConfirmModel) ID() Screen { return ScreenUpdateConfirm }
func (m *UpdateConfirmModel) Init() tea.Cmd { return nil }

func (m *UpdateConfirmModel) View() tea.View {
	s := m.Shared
	contentWidth := s.ContentWidth()
	current := s.CurrentVersion
	if current == "" {
		current = "0.0.0"
	}
	body := BodyStyle.Render(fmt.Sprintf("Доступна версия %s (текущая: %s). Установить?", s.AvailableVersion, current))
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n" + BodyStyle.Render("  ")
	yesBtn := ButtonStyle.Render("[ Y ] Да")
	noBtn := ButtonStyle.Render("[ N ] Нет")
	if m.ConfirmButtonFoc == UpdateButtonYes {
		yesBtn = ButtonActiveStyle.Render("[ Y ] Да")
	} else {
		noBtn = ButtonActiveStyle.Render("[ N ] Нет")
	}
	body += yesBtn + BodyStyle.Render("  ") + noBtn
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n"
	footerStr := "Tab: переключить   Enter: подтвердить   Esc: отмена"
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	body += FooterStyle.Render(footerStr+strings.Repeat(" ", pad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	return tea.NewView(FrameWithTitle("  ОБНОВЛЕНИЕ  ", body, contentWidth))
}

func (m *UpdateConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.ConfirmButtonFoc == UpdateButtonYes {
				repo := updater.DefaultRepo
				if m.Shared.Config != nil && m.Shared.Config.UpdateRepo != "" {
					repo = m.Shared.Config.UpdateRepo
				}
				return m, tea.Sequence(ReplaceScreenCmd(ScreenProgress), runInstallCmd(repo, m.Shared.AvailableVersion))
			}
			return m, PopScreenCmd()
		case "tab", "right":
			if m.ConfirmButtonFoc == UpdateButtonNo {
				m.ConfirmButtonFoc = UpdateButtonYes
			} else {
				m.ConfirmButtonFoc = UpdateButtonNo
			}
			return m, nil
		case "shift+tab", "left":
			if m.ConfirmButtonFoc == UpdateButtonYes {
				m.ConfirmButtonFoc = UpdateButtonNo
			} else {
				m.ConfirmButtonFoc = UpdateButtonYes
			}
			return m, nil
		case "esc":
			return m, PopScreenCmd()
		}
	}
	return m, nil
}

// --- UpdateCheckErrorModel ---

type UpdateCheckErrorModel struct {
	Shared        *SharedState
	ErrorButtonFoc int
}

func NewUpdateCheckErrorModel(shared *SharedState) *UpdateCheckErrorModel {
	return &UpdateCheckErrorModel{Shared: shared}
}

func (m *UpdateCheckErrorModel) ID() Screen { return ScreenUpdateCheckError }
func (m *UpdateCheckErrorModel) Init() tea.Cmd { return nil }

func (m *UpdateCheckErrorModel) View() tea.View {
	contentWidth := m.Shared.ContentWidth()
	body := BodyStyle.Render("Не удалось проверить обновления:")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	line2 := BodyStyle.Render("  ") + ErrorStyle.Render(m.Shared.CheckError)
	if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
		line2 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line2 + "\n\n" + BodyStyle.Render("  ")
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.ErrorButtonFoc == 0 {
		retryBtn = ButtonActiveStyle.Render("[ Retry ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	line3 := retryBtn + BodyStyle.Render("  ") + cancelBtn
	if pad := contentWidth - lipgloss.Width(line3); pad > 0 {
		line3 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line3 + "\n\n"
	footerStr := "Tab / ←→: переключить   Enter: подтвердить   Esc: в меню"
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	body += FooterStyle.Render(footerStr+strings.Repeat(" ", pad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	return tea.NewView(FrameWithTitle("  Ошибка проверки  ", body, contentWidth))
}

func (m *UpdateCheckErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Shared.Config != nil && m.Shared.Config.UpdateRepo != "" {
					repo = m.Shared.Config.UpdateRepo
				}
				m.Shared.CheckError = ""
				return m, tea.Sequence(ReplaceScreenCmd(ScreenUpdateChecking), runCheckUpdateCmd(repo, m.Shared.CurrentVersion))
			}
			m.Shared.CheckError = ""
			return m, PopScreenCmd()
		case "tab", "right":
			m.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.Shared.CheckError = ""
			return m, PopScreenCmd()
		}
	}
	return m, nil
}

// --- UpdateInstallErrorModel ---

type UpdateInstallErrorModel struct {
	Shared        *SharedState
	ErrorButtonFoc int
}

func NewUpdateInstallErrorModel(shared *SharedState) *UpdateInstallErrorModel {
	return &UpdateInstallErrorModel{Shared: shared}
}

func (m *UpdateInstallErrorModel) ID() Screen { return ScreenUpdateInstallError }
func (m *UpdateInstallErrorModel) Init() tea.Cmd { return nil }

func (m *UpdateInstallErrorModel) View() tea.View {
	contentWidth := m.Shared.ContentWidth()
	body := BodyStyle.Render("Ошибка при установке:")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	line2 := BodyStyle.Render("  ") + ErrorStyle.Render(m.Shared.InstallError)
	if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
		line2 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line2 + "\n\n" + BodyStyle.Render("  ")
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.ErrorButtonFoc == 0 {
		retryBtn = ButtonActiveStyle.Render("[ Retry ]")
	} else {
		cancelBtn = ButtonActiveStyle.Render("[ Cancel ]")
	}
	line3 := retryBtn + BodyStyle.Render("  ") + cancelBtn
	if pad := contentWidth - lipgloss.Width(line3); pad > 0 {
		line3 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line3 + "\n\n"
	footerStr := "Tab / ←→: переключить   Enter: подтвердить   Esc: в меню"
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	body += FooterStyle.Render(footerStr+strings.Repeat(" ", pad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	return tea.NewView(FrameWithTitle("  Ошибка установки  ", body, contentWidth))
}

func (m *UpdateInstallErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Shared.Config != nil && m.Shared.Config.UpdateRepo != "" {
					repo = m.Shared.Config.UpdateRepo
				}
				m.Shared.InstallError = ""
				return m, tea.Sequence(ReplaceScreenCmd(ScreenProgress), runInstallCmd(repo, m.Shared.AvailableVersion))
			}
			m.Shared.InstallError = ""
			return m, PopScreenCmd()
		case "tab", "right":
			m.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.Shared.InstallError = ""
			return m, PopScreenCmd()
		}
	}
	return m, nil
}

// --- UpdateSuccessModel ---

type UpdateSuccessModel struct {
	Shared *SharedState
}

func NewUpdateSuccessModel(shared *SharedState) *UpdateSuccessModel {
	return &UpdateSuccessModel{Shared: shared}
}

func (m *UpdateSuccessModel) ID() Screen { return ScreenUpdateSuccess }
func (m *UpdateSuccessModel) Init() tea.Cmd { return nil }

func (m *UpdateSuccessModel) View() tea.View {
	contentWidth := m.Shared.ContentWidth()
	body := BodyStyle.Render("Обновление установлено.")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	line2 := BodyStyle.Render("  ") + BodyStyle.Render("Перезапустите приложение для использования новой версии.")
	if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
		line2 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line2 + "\n\n"
	footerStr := "Нажмите Enter для выхода"
	pad := contentWidth - lipgloss.Width(footerStr)
	if pad < 0 {
		pad = 0
	}
	body += FooterStyle.Render(footerStr+strings.Repeat(" ", pad)) + "\n" + FooterStyle.Render(strings.Repeat(" ", contentWidth))
	return tea.NewView(FrameWithTitle("  Установлено  ", body, contentWidth))
}

func (m *UpdateSuccessModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}
