// Package updater реализует самообновление launcher (FR-601–FR-607).
// Проверка по semver, скачивание с Update Server (GitLab Releases/Nexus),
// резервная копия перед обновлением и откат при ошибке.
package updater

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v62/github"
)

const (
	// DefaultRepo в формате "owner/repo" для GitHub. Для GitLab подставляется через UpdateServerURL.
	DefaultRepo = "ai-launcher/cli"
)

// CheckInBackground проверяет обновления (FR-601). Вызывать из main в горутине.
// Если доступна версия новее currentVersion, вызывается onNewVersion(available).
// При ошибке сети или парсинга версий просто выходит без вызова.
func CheckInBackground(currentVersion, repo string, onNewVersion func(available string)) {
	if repo == "" {
		repo = DefaultRepo
	}
	latest, err := LatestRelease(repo)
	if err != nil {
		return
	}
	if latest == "" {
		return
	}
	newer, err := NewerThan(currentVersion, latest)
	if err != nil || !newer {
		return
	}
	if onNewVersion != nil {
		onNewVersion(latest)
	}
}

// LatestRelease возвращает последнюю доступную версию (тег) для репозитория.
// repo в формате "owner/repo". Если пустой — используется DefaultRepo.
// Для GitLab нужна отдельная реализация через UpdateServerURL (ТЗ 6.3).
func LatestRelease(repo string) (string, error) {
	if repo == "" {
		repo = DefaultRepo
	}
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo: %s", repo)
	}
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), parts[0], parts[1])
	if err != nil {
		return "", err
	}
	if release.TagName == nil {
		return "", nil
	}
	return *release.TagName, nil
}

// NewerThan сравнивает текущую версию с доступной по semver (FR-602).
// Возвращает true, если available новее current.
func NewerThan(current, available string) (bool, error) {
	cur, err := semver.NewVersion(strings.TrimPrefix(current, "v"))
	if err != nil {
		return false, err
	}
	avail, err := semver.NewVersion(strings.TrimPrefix(available, "v"))
	if err != nil {
		return false, err
	}
	return avail.GreaterThan(cur), nil
}

// DownloadURL формирует URL скачивания бинарника для текущей платформы.
// Для GitHub Releases: .../releases/download/vX.Y.Z/ai-launcher_{GOOS}_{GOARCH}[.exe]
func DownloadURL(repo, version string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	ext := ""
	if os == "windows" {
		ext = ".exe"
	}
	// GitHub pattern
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/ai-launcher_%s_%s%s",
		repo, version, os, arch, ext)
}

// Install выполняет скачивание, backup текущего бинарника и замену (FR-604, FR-605).
// При ошибке — откат (FR-606). Требует подтверждения снаружи.
func Install(downloadURL, currentBinaryPath string) error {
	// TODO: скачать в temp, проверить checksum (NFR-303), backup, replace, rollback on error
	_ = downloadURL
	_ = currentBinaryPath
	return fmt.Errorf("updater.Install: not implemented")
}
