package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

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
	current := m.CurrentVersion
	if current == "" {
		current = "0.0.0"
	}
	body := BodyStyle.Render(fmt.Sprintf("Доступна версия %s (текущая: %s). Установить?", m.AvailableVersion, current))
	body += "\n\n  "
	noBtn := ButtonStyle.Render("[ N ] Нет")
	yesBtn := ButtonStyle.Render("[ Y ] Да")
	if m.UpdateState.ConfirmButtonFoc == UpdateButtonNo {
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
			if m.UpdateState.ConfirmButtonFoc == UpdateButtonYes {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenProgress
				m.Progress.Title = "Update"
				m.Progress.Status = "Скачивание и установка..."
				m.Progress.Percent = -1
				return m, runInstallCmd(repo, m.AvailableVersion)
			}
			m.Screen = ScreenMain
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
			m.Screen = ScreenMain
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateCheckErrorScreen(m Model) tea.View {
	body := BodyStyle.Render("Не удалось проверить обновления:")
	body += "\n\n  " + ErrorStyle.Render(m.UpdateState.CheckError)
	body += "\n\n  "
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.UpdateState.ErrorButtonFoc == 0 {
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
			if m.UpdateState.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenUpdateChecking
				m.Progress.Title = "Проверка обновлений"
				m.Progress.Status = "Проверка обновлений..."
				m.Progress.Percent = -1
				return m, runCheckUpdateCmd(repo, m.CurrentVersion)
			}
			m.Screen = ScreenMain
			m.UpdateState.CheckError = ""
			return m, nil
		case "tab", "right":
			m.UpdateState.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.UpdateState.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.Screen = ScreenMain
			m.UpdateState.CheckError = ""
			return m, nil
		}
	}
	return m, nil
}

func viewUpdateInstallErrorScreen(m Model) tea.View {
	body := BodyStyle.Render("Ошибка при установке:")
	body += "\n\n  " + ErrorStyle.Render(m.UpdateState.InstallError)
	body += "\n\n  "
	retryBtn := ButtonStyle.Render("[ Retry ]")
	cancelBtn := ButtonStyle.Render("[ Cancel ]")
	if m.UpdateState.ErrorButtonFoc == 0 {
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
			if m.UpdateState.ErrorButtonFoc == 0 {
				repo := updater.DefaultRepo
				if m.Config != nil && m.Config.UpdateRepo != "" {
					repo = m.Config.UpdateRepo
				}
				m.Screen = ScreenProgress
				m.Progress.Title = "Update"
				m.Progress.Status = "Скачивание и установка..."
				m.Progress.Percent = -1
				return m, runInstallCmd(repo, m.AvailableVersion)
			}
			m.Screen = ScreenMain
			m.UpdateState.InstallError = ""
			return m, nil
		case "tab", "right":
			m.UpdateState.ErrorButtonFoc = 1
			return m, nil
		case "shift+tab", "left":
			m.UpdateState.ErrorButtonFoc = 0
			return m, nil
		case "esc":
			m.Screen = ScreenMain
			m.UpdateState.InstallError = ""
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
