// Package updater реализует самообновление launcher (FR-601–FR-607).
// Проверка по semver, скачивание с Update Server (GitLab Releases/Nexus),
// резервная копия перед обновлением и откат при ошибке.
package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
// При ошибке сети или парсинга версий вызывается onError(err), если не nil.
func CheckInBackground(currentVersion, repo string, onNewVersion func(available string), onError func(err error)) {
	if repo == "" {
		repo = DefaultRepo
	}
	latest, err := LatestRelease(context.Background(), repo)
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return
	}
	if latest == "" {
		return
	}
	newer, err := NewerThan(currentVersion, latest)
	if err != nil {
		if onError != nil {
			onError(err)
		}
		return
	}
	if !newer {
		return
	}
	if onNewVersion != nil {
		onNewVersion(latest)
	}
}

// LatestRelease возвращает последнюю доступную версию (тег) для репозитория.
// repo в формате "owner/repo". Если пустой — используется DefaultRepo.
// Контекст используется для отмены запроса (например, таймаут).
func LatestRelease(ctx context.Context, repo string) (string, error) {
	if repo == "" {
		repo = DefaultRepo
	}
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repo: %s", repo)
	}
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetLatestRelease(ctx, parts[0], parts[1])
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

// progressWriter wraps io.Writer and calls onProgress periodically.
type progressWriter struct {
	w         io.Writer
	written   int64
	total     int64
	onProgress func(written, total int64)
	lastReport int64
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	if n > 0 {
		p.written += int64(n)
		// Report every ~64KB or when we finish
		report := p.written-p.lastReport >= 64*1024 || (p.total > 0 && p.written >= p.total)
		if report {
			p.lastReport = p.written
			p.onProgress(p.written, p.total)
		}
	}
	return n, err
}

// Install выполняет скачивание, backup текущего бинарника и замену (FR-604, FR-605).
// При ошибке — откат (FR-606). Подтверждение запрашивается снаружи (autoupdate).
// onProgress вызывается при скачивании: written и total в байтах (total может быть -1 если неизвестен).
func Install(downloadURL, currentBinaryPath string, onProgress func(written, total int64)) error {
	// Resolve symlinks so we replace the real file
	resolved, err := filepath.EvalSymlinks(currentBinaryPath)
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}
	currentBinaryPath = resolved
	dir := filepath.Dir(currentBinaryPath)

	// 1. Download to temp file in same directory (for rename to work)
	tmpFile, err := os.CreateTemp(dir, "ai-launcher-new-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // remove temp if we leave before renaming to target

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download: HTTP %s", resp.Status)
	}
	total := resp.ContentLength
	var dst io.Writer = tmpFile
	if onProgress != nil {
		dst = &progressWriter{w: tmpFile, total: total, onProgress: onProgress}
	}
	_, err = io.Copy(dst, resp.Body)
	tmpFile.Close()
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	if onProgress != nil && total >= 0 {
		onProgress(total, total)
	}

	// 2. Make executable (Unix)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			return fmt.Errorf("chmod: %w", err)
		}
	}

	// 3. Backup current binary
	backupPath := currentBinaryPath + ".bak"
	if err := os.Rename(currentBinaryPath, backupPath); err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	// 4. Replace with new binary (rollback on failure)
	if err := os.Rename(tmpPath, currentBinaryPath); err != nil {
		if revert := os.Rename(backupPath, currentBinaryPath); revert != nil {
			return fmt.Errorf("replace failed (%v), rollback failed: %v", err, revert)
		}
		return fmt.Errorf("replace: %w", err)
	}

	// Success; keep .bak for user safety (they can delete it)
	return nil
}
