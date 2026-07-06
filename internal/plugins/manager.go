// SPDX-License-Identifier: MIT
package plugins

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"renbrowser/internal/db"
	"renbrowser/internal/paths"
)

const (
	maxZipBytes        = 32 * 1024 * 1024
	maxZipFiles        = 256
	maxZipUncompressed = 64 * 1024 * 1024
)

type PluginStateStore interface {
	UpsertPlugin(id string, enabled bool, settingsJSON string) error
	GetPlugin(id string) (db.PluginRow, error)
	ListPlugins() ([]db.PluginRow, error)
	SetPluginEnabled(id string, enabled bool) error
	DeletePlugin(id string) error
	GetPluginSetting(pluginID, key string) (string, error)
	SetPluginSetting(pluginID, key, value string) error
}

type InstalledPlugin struct {
	Manifest Manifest `json:"manifest"`
	Dir      string   `json:"dir"`
	Enabled  bool     `json:"enabled"`
	Error    string   `json:"error,omitempty"`
}

type Manager struct {
	mu      sync.RWMutex
	reg     *Registry
	store   PluginStateStore
	storage *Storage
	wasm    *WasmRuntime
	dir     string
	app     *application.App
	plugins map[string]InstalledPlugin
}

func NewManager(store PluginStateStore) *Manager {
	reg := NewRegistry()
	return &Manager{
		reg:     reg,
		store:   store,
		storage: NewStorage(store),
		wasm:    NewWasmRuntime(),
		dir:     paths.Join(".renbrowser", "plugins"),
		plugins: make(map[string]InstalledPlugin),
	}
}

func (m *Manager) Registry() *Registry {
	return m.reg
}

func (m *Manager) Storage() *Storage {
	return m.storage
}

func (m *Manager) PluginsDir() string {
	return m.dir
}

func (m *Manager) SetPluginsDirForTest(dir string) {
	m.dir = dir
}

func (m *Manager) SetApp(app *application.App) {
	m.mu.Lock()
	m.app = app
	m.mu.Unlock()
}

func (m *Manager) LoadAll() error {
	if err := os.MkdirAll(m.dir, 0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return err
	}
	states, _ := m.store.ListPlugins()
	stateMap := make(map[string]db.PluginRow, len(states))
	for _, row := range states {
		stateMap[row.ID] = row
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(m.dir, entry.Name())
		manifest, err := LoadManifest(dir)
		if err != nil {
			m.mu.Lock()
			m.plugins[entry.Name()] = InstalledPlugin{Dir: dir, Enabled: false, Error: err.Error()}
			m.mu.Unlock()
			continue
		}
		enabled := true
		if row, ok := stateMap[manifest.ID]; ok {
			enabled = row.Enabled
		} else {
			_ = m.store.UpsertPlugin(manifest.ID, true, "{}")
		}
		m.mu.Lock()
		m.plugins[manifest.ID] = InstalledPlugin{Manifest: manifest, Dir: dir, Enabled: enabled}
		m.mu.Unlock()
		if enabled {
			if err := m.enableLocked(manifest.ID); err != nil {
				m.mu.Lock()
				p := m.plugins[manifest.ID]
				p.Error = err.Error()
				p.Enabled = false
				m.plugins[manifest.ID] = p
				m.mu.Unlock()
				_ = m.store.SetPluginEnabled(manifest.ID, false)
			}
		}
	}
	return nil
}

func (m *Manager) List() []InstalledPlugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]InstalledPlugin, 0, len(m.plugins))
	for _, p := range m.plugins {
		out = append(out, p)
	}
	return out
}

func (m *Manager) Get(id string) (InstalledPlugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[id]
	return p, ok
}

func (m *Manager) Enable(id string) error {
	m.mu.Lock()
	p, ok := m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", id)
	}
	if err := m.enableLocked(id); err != nil {
		m.mu.Unlock()
		return err
	}
	p.Enabled = true
	p.Error = ""
	m.plugins[id] = p
	manifest := p.Manifest
	m.mu.Unlock()
	_ = m.store.SetPluginEnabled(id, true)
	m.emitLoaded(manifest)
	return nil
}

func (m *Manager) Disable(id string) error {
	m.mu.Lock()
	p, ok := m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", id)
	}
	m.disableLocked(id)
	p.Enabled = false
	m.plugins[id] = p
	manifest := p.Manifest
	m.mu.Unlock()
	_ = m.store.SetPluginEnabled(id, false)
	m.emitUnloaded(manifest)
	return nil
}

func (m *Manager) Uninstall(id string) error {
	m.mu.Lock()
	p, ok := m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", id)
	}
	m.disableLocked(id)
	dir := p.Dir
	manifest := p.Manifest
	delete(m.plugins, id)
	m.mu.Unlock()
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	_ = m.store.DeletePlugin(id)
	m.emitUnloaded(manifest)
	return nil
}

func (m *Manager) InstallFromDir(src string) (InstalledPlugin, error) {
	manifest, err := LoadManifest(src)
	if err != nil {
		return InstalledPlugin{}, err
	}
	if err := ValidatePermissions(manifest.Permissions); err != nil {
		return InstalledPlugin{}, err
	}
	dest := filepath.Join(m.dir, manifest.ID)
	if err := os.RemoveAll(dest); err != nil {
		return InstalledPlugin{}, err
	}
	if err := copyDir(src, dest); err != nil {
		return InstalledPlugin{}, err
	}
	_ = m.store.UpsertPlugin(manifest.ID, true, "{}")
	m.mu.Lock()
	m.plugins[manifest.ID] = InstalledPlugin{Manifest: manifest, Dir: dest, Enabled: true}
	m.mu.Unlock()
	if err := m.Enable(manifest.ID); err != nil {
		return InstalledPlugin{}, err
	}
	p, ok := m.Get(manifest.ID)
	if !ok {
		return InstalledPlugin{}, fmt.Errorf("plugin %q missing after install", manifest.ID)
	}
	return p, nil
}

func (m *Manager) InstallFromZip(zipPath string) (InstalledPlugin, error) {
	info, err := os.Stat(zipPath)
	if err != nil {
		return InstalledPlugin{}, err
	}
	if info.Size() > maxZipBytes {
		return InstalledPlugin{}, fmt.Errorf("zip file too large")
	}
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return InstalledPlugin{}, err
	}
	defer reader.Close()
	if len(reader.File) > maxZipFiles {
		return InstalledPlugin{}, fmt.Errorf("zip has too many files")
	}
	var total int64
	var manifestData []byte
	tmpDir, err := os.MkdirTemp("", "renplugin-*")
	if err != nil {
		return InstalledPlugin{}, err
	}
	defer os.RemoveAll(tmpDir)
	for _, f := range reader.File {
		if f.UncompressedSize64 > uint64(maxZipUncompressed) {
			return InstalledPlugin{}, fmt.Errorf("zip uncompressed size too large")
		}
		total += int64(f.UncompressedSize64) // #nosec G115 -- bounded by check above
		if total > maxZipUncompressed {
			return InstalledPlugin{}, fmt.Errorf("zip uncompressed size too large")
		}
		clean := filepath.Clean(strings.ReplaceAll(f.Name, `\`, "/"))
		if clean == "" || clean == "." {
			continue
		}
		dest, err := safeZipJoin(tmpDir, clean)
		if err != nil {
			return InstalledPlugin{}, err
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0o750); err != nil {
				return InstalledPlugin{}, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			return InstalledPlugin{}, err
		}
		if err := extractZipFile(f, dest); err != nil {
			return InstalledPlugin{}, err
		}
		if filepath.Base(clean) == ManifestFileName && manifestData == nil {
			manifestData, _ = os.ReadFile(dest) // #nosec G304 -- path cleaned and under temp extract dir
		}
	}
	if len(manifestData) == 0 {
		return InstalledPlugin{}, fmt.Errorf("manifest not found in zip")
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return InstalledPlugin{}, err
	}
	if err := ValidateManifest(manifest); err != nil {
		return InstalledPlugin{}, err
	}
	return m.InstallFromDir(tmpDir)
}

func (m *Manager) InvokeCommand(pluginID, commandID string, args map[string]string) error {
	m.mu.RLock()
	p, ok := m.plugins[pluginID]
	m.mu.RUnlock()
	if !ok || !p.Enabled {
		return fmt.Errorf("plugin %q not enabled", pluginID)
	}
	for _, cmd := range p.Manifest.Contributes.Commands {
		if cmd.ID == commandID {
			m.mu.RLock()
			app := m.app
			m.mu.RUnlock()
			if app != nil {
				app.Event.Emit("plugin:"+pluginID+":command:"+commandID, args)
			}
			return nil
		}
	}
	return fmt.Errorf("command %q not found", commandID)
}

func (m *Manager) EnabledIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []string
	for id, p := range m.plugins {
		if p.Enabled {
			out = append(out, id)
		}
	}
	return out
}

func (m *Manager) PluginRoot(id string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[id]
	if !ok {
		return "", false
	}
	return p.Dir, true
}

func (m *Manager) enableLocked(id string) error {
	p, ok := m.plugins[id]
	if !ok {
		return fmt.Errorf("plugin %q not found", id)
	}
	if err := ValidatePermissions(p.Manifest.Permissions); err != nil {
		return err
	}
	m.reg.RegisterManifest(p.Manifest)
	m.reg.RegisterContributions(id, p.Manifest.Contributes)
	if p.Manifest.Backend != "" {
		wasmPath := filepath.Join(p.Dir, p.Manifest.Backend)
		renderer, err := m.wasm.LoadRenderer(id, wasmPath, p.Manifest)
		if err != nil {
			return err
		}
		m.reg.RegisterRenderer(renderer)
	}
	if HasPermission(p.Manifest, PermNetworkFetch) {
		if hook := m.wasm.BeforeFetchHook(id); hook != nil {
			m.reg.RegisterBeforeFetch(id, hook)
		}
		if hook := m.wasm.AfterFetchHook(id); hook != nil {
			m.reg.RegisterAfterFetch(id, hook)
		}
	}
	for _, scheme := range p.Manifest.Contributes.URLSchemes {
		contrib := scheme
		m.reg.RegisterScheme(id, contrib, func(rawURL string) (SchemeResult, bool) {
			return SchemeResult{
				DelegateFrontend: true,
				PluginID:         id,
				Handler:          contrib.Handler,
				URL:              rawURL,
			}, true
		})
	}
	return nil
}

func (m *Manager) disableLocked(id string) {
	m.reg.UnregisterPlugin(id)
	m.wasm.Unload(id)
}

func (m *Manager) emitLoaded(manifest Manifest) {
	m.mu.RLock()
	app := m.app
	m.mu.RUnlock()
	if app != nil {
		app.Event.Emit("plugin:loaded", manifest)
	}
}

func (m *Manager) emitUnloaded(manifest Manifest) {
	m.mu.RLock()
	app := m.app
	m.mu.RUnlock()
	if app != nil {
		app.Event.Emit("plugin:unloaded", manifest)
	}
}

func (m *Manager) EmitEvent(pluginID, event string, data any) {
	m.mu.RLock()
	app := m.app
	m.mu.RUnlock()
	if app != nil {
		app.Event.Emit("plugin:"+pluginID+":"+event, data)
	}
}

func safeZipJoin(root, name string) (string, error) {
	rootClean := filepath.Clean(root)
	clean := filepath.Clean(strings.ReplaceAll(name, `\`, "/"))
	if clean == "" || clean == "." {
		return rootClean, nil
	}
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("zip path traversal: %s", name)
	}
	target := filepath.Join(rootClean, clean)
	rel, err := filepath.Rel(rootClean, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("zip path traversal: %s", name)
	}
	return target, nil
}

func extractZipFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()) // #nosec G304
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc) // #nosec G110 -- zip size limits enforced before extract
	return err
}

func copyDir(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		data, err := os.ReadFile(path) // #nosec G304 G122 -- install copies user-selected plugin tree
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600) // #nosec G703 G306 -- target under validated dest root
	})
}

func (m *Manager) Close() error {
	m.mu.Lock()
	wasm := m.wasm
	m.mu.Unlock()
	if wasm == nil {
		return nil
	}
	return wasm.Close()
}
