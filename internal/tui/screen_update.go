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

func viewUpdateConfirmScreen(m Model) tea.View {
	contentWidth := m.ContentWidth()
	current := m.CurrentVersion
	if current == "" {
		current = "0.0.0"
	}
	body := BodyStyle.Render(fmt.Sprintf("Доступна версия %s (текущая: %s). Установить?", m.AvailableVersion, current))
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n" + BodyStyle.Render("  ")
	yesBtn := ButtonStyle.Render("[ Y ] Да")
	noBtn := ButtonStyle.Render("[ N ] Нет")
	if m.UpdateState.ConfirmButtonFoc == UpdateButtonYes {
		yesBtn = ButtonActiveStyle.Render("[ Y ] Да")
	} else {
		noBtn = ButtonActiveStyle.Render("[ N ] Нет")
	}
	// Пробел между кнопками — синий фон, чтобы не было чёрной полосы
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
	// Вся строка футера одним рендером — синий фон до конца
	footerLine := FooterStyle.Render(footerStr + strings.Repeat(" ", pad))
	body += footerLine + "\n" + FooterStyle.Render(strings.Repeat(" ", m.ContentWidth()))
	rendered := FrameWithTitle("  ОБНОВЛЕНИЕ  ", body, m.ContentWidth())
	return tea.NewView(rendered)
}

func updateUpdateConfirmScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.UpdateState.ConfirmButtonFoc == UpdateButtonYes {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Progress.Title = "Update"
				m.Progress.Status = "Скачивание и установка..."
				m.Progress.Percent = -1
				m = ReplaceScreen(m, ScreenProgress)
				return m, runInstallCmd(repo, m.AvailableVersion)
			}
			m = PopScreen(m)
			return m, nil
		case "tab", "right":
			if m.UpdateState.ConfirmButtonFoc == UpdateButtonNo {
				m.UpdateState.ConfirmButtonFoc = UpdateButtonYes
			} else {
				m.UpdateState.ConfirmButtonFoc = UpdateButtonNo
			}
			return m, nil
		case "shift+tab", "left":
			if m.UpdateState.ConfirmButtonFoc == UpdateButtonYes {
				m.UpdateState.ConfirmButtonFoc = UpdateButtonNo
			} else {
				m.UpdateState.ConfirmButtonFoc = UpdateButtonYes
			}
			return m, nil
		case "esc":
			m = PopScreen(m)
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateCheckErrorScreen(m Model) tea.View {
	contentWidth := m.ContentWidth()
	body := BodyStyle.Render("Не удалось проверить обновления:")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	line2 := BodyStyle.Render("  ") + ErrorStyle.Render(m.UpdateState.CheckError)
	if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
		line2 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line2 + "\n\n" + BodyStyle.Render("  ")
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.UpdateState.ErrorButtonFoc == 0 {
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
	rendered := FrameWithTitle("  Ошибка проверки  ", body, contentWidth)
	return tea.NewView(rendered)
}

func updateUpdateCheckErrorScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.UpdateState.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Progress.Title = "Проверка обновлений"
				m.Progress.Status = "Проверка обновлений..."
				m.Progress.Percent = -1
				m = ReplaceScreen(m, ScreenUpdateChecking)
				return m, runCheckUpdateCmd(repo, m.CurrentVersion)
			}
			m.UpdateState.CheckError = ""
			m = PopScreen(m)
			return m, nil
		case "tab", "right":
			m.UpdateState.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.UpdateState.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.UpdateState.CheckError = ""
			m = PopScreen(m)
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateInstallErrorScreen(m Model) tea.View {
	contentWidth := m.ContentWidth()
	body := BodyStyle.Render("Ошибка при установке:")
	if pad := contentWidth - lipgloss.Width(body); pad > 0 {
		body += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += "\n\n"
	line2 := BodyStyle.Render("  ") + ErrorStyle.Render(m.UpdateState.InstallError)
	if pad := contentWidth - lipgloss.Width(line2); pad > 0 {
		line2 += BodyStyle.Render(strings.Repeat(" ", pad))
	}
	body += line2 + "\n\n" + BodyStyle.Render("  ")
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.UpdateState.ErrorButtonFoc == 0 {
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
	rendered := FrameWithTitle("  Ошибка установки  ", body, contentWidth)
	return tea.NewView(rendered)
}

func updateUpdateInstallErrorScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if m.UpdateState.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Progress.Title = "Update"
				m.Progress.Status = "Скачивание и установка..."
				m.Progress.Percent = -1
				m = ReplaceScreen(m, ScreenProgress)
				return m, runInstallCmd(repo, m.AvailableVersion)
			}
			m.UpdateState.InstallError = ""
			m = PopScreen(m)
			return m, nil
		case "tab", "right":
			m.UpdateState.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.UpdateState.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.UpdateState.InstallError = ""
			m = PopScreen(m)
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateSuccessScreen(m Model) tea.View {
	contentWidth := m.ContentWidth()
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
	rendered := FrameWithTitle("  Установлено  ", body, contentWidth)
	return tea.NewView(rendered)
}

func updateUpdateSuccessScreen(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}
