// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/adrg/xdg"
)

const downloadDirSettingKey = "downloadDir"
const downloadHistoryKey = "downloadHistory"
const maxDownloadHistory = 100

type DownloadItem struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modifiedAt"`
}

func defaultDownloadDir() string {
	if dir := strings.TrimSpace(xdg.UserDirs.Download); dir != "" {
		return filepath.Clean(dir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Clean(filepath.Join(home, "Downloads"))
}

func isTempDownloadDir(dir string) bool {
	dir = filepath.Clean(dir)
	temp := filepath.Clean(os.TempDir())
	if dir == temp {
		return true
	}
	rel, err := filepath.Rel(temp, dir)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func isAcceptableDownloadDir(dir string) bool {
	dir = strings.TrimSpace(dir)
	if dir == "" || isTempDownloadDir(dir) {
		return false
	}
	info, err := os.Stat(dir)
	if err != nil {
		return os.IsNotExist(err)
	}
	return info.IsDir()
}

func (s *BrowserService) persistDefaultDownloadDir() string {
	dir := defaultDownloadDir()
	_ = s.store.SetSetting(downloadDirSettingKey, dir)
	return dir
}

func (s *BrowserService) GetDownloadDir() string {
	raw, err := s.store.GetSetting(downloadDirSettingKey)
	dir := strings.TrimSpace(raw)
	if err == nil && isAcceptableDownloadDir(dir) {
		return filepath.Clean(dir)
	}
	return s.persistDefaultDownloadDir()
}

func (s *BrowserService) SetDownloadDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if !isAcceptableDownloadDir(dir) {
		dir = defaultDownloadDir()
	}
	dir = filepath.Clean(dir)
	_ = s.store.SetSetting(downloadDirSettingKey, dir)
	return dir
}

func (s *BrowserService) PickDownloadDir() (string, error) {
	if s.app == nil {
		return "", errors.New("directory picker unavailable in server mode")
	}
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("Select download folder").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	if result == "" {
		return s.GetDownloadDir(), nil
	}
	return s.SetDownloadDir(result), nil
}

func (s *BrowserService) DownloadToDir(rawURL string) (string, error) {
	dir := s.GetDownloadDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	name := downloadNameFromURL(rawURL)
	dest := uniqueFilePath(filepath.Join(dir, name))
	if err := s.SaveDownload(rawURL, dest); err != nil {
		return "", err
	}
	s.recordDownload(dest)
	return dest, nil
}

func (s *BrowserService) SaveTextToDownloadDir(filename, content string) (string, error) {
	dir := s.GetDownloadDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	name := sanitizeDownloadFilename(filename)
	dest := uniqueFilePath(filepath.Join(dir, name))
	if err := os.WriteFile(dest, []byte(content), 0o600); err != nil { //nolint:gosec // user download path
		return "", err
	}
	s.recordDownload(dest)
	return dest, nil
}

func (s *BrowserService) ListDownloads() []DownloadItem {
	items := s.loadDownloadHistory()
	out := make([]DownloadItem, 0, len(items))
	for _, item := range items {
		info, err := os.Stat(item.Path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			continue
		}
		out = append(out, DownloadItem{
			Name:       filepath.Base(item.Path),
			Path:       item.Path,
			Size:       info.Size(),
			ModifiedAt: info.ModTime().Unix(),
		})
	}
	return out
}

func (s *BrowserService) OpenDownloadPath(path string) error {
	if err := s.validateDownloadPath(path); err != nil {
		return err
	}
	return openPath(path)
}

func (s *BrowserService) ShowDownloadDir() error {
	dir := s.GetDownloadDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return openPath(dir)
}

func (s *BrowserService) recordDownload(path string) {
	path = filepath.Clean(path)
	items := s.loadDownloadHistory()
	filtered := make([]DownloadItem, 0, len(items)+1)
	for _, item := range items {
		if filepath.Clean(item.Path) != path {
			filtered = append(filtered, item)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	entry := DownloadItem{
		Name:       filepath.Base(path),
		Path:       path,
		Size:       info.Size(),
		ModifiedAt: info.ModTime().Unix(),
	}
	filtered = append([]DownloadItem{entry}, filtered...)
	if len(filtered) > maxDownloadHistory {
		filtered = filtered[:maxDownloadHistory]
	}
	raw, err := json.Marshal(filtered)
	if err != nil {
		return
	}
	_ = s.store.SetSetting(downloadHistoryKey, string(raw))
}

func (s *BrowserService) loadDownloadHistory() []DownloadItem {
	raw, err := s.store.GetSetting(downloadHistoryKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var items []DownloadItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ModifiedAt > items[j].ModifiedAt
	})
	return items
}

func (s *BrowserService) validateDownloadPath(path string) error {
	path = filepath.Clean(path)
	dir := filepath.Clean(s.GetDownloadDir())
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return errors.New("download path is outside the download folder")
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func openPath(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start() // #nosec G204 -- path validated by validateDownloadPath before OpenDownloadPath
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start() // #nosec G204 -- path validated by validateDownloadPath before OpenDownloadPath
	default:
		return exec.Command("xdg-open", path).Start() // #nosec G204 -- path validated by validateDownloadPath before OpenDownloadPath
	}
}

func downloadNameFromURL(rawURL string) string {
	if rawURL == "editor:" {
		return "editor.mu"
	}
	if rawURL == "config:" {
		return "reticulum.conf"
	}
	if rawURL == "about:" {
		return "about.html"
	}
	if rawURL == "license:" {
		return "LICENSE"
	}
	path := rawURL
	if _, after, ok := strings.Cut(rawURL, ":/"); ok {
		path = after
	}
	if q := strings.IndexAny(path, "?`"); q >= 0 {
		path = path[:q]
	}
	leaf := filepath.Base(strings.TrimSuffix(path, "/"))
	if leaf != "" && leaf != "." && leaf != "/" {
		return sanitizeDownloadFilename(leaf)
	}
	return "download.bin"
}

func sanitizeDownloadFilename(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" || name == "." {
		return "download.bin"
	}
	var b strings.Builder
	for _, r := range name {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\x00':
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "download.bin"
	}
	return out
}

func uniqueFilePath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	dir := filepath.Dir(path)
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s (%d)%s", base, i, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, os.Getpid(), ext))
}
