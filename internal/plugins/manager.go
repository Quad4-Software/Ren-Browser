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
	GetSetting(key string) (string, error)
	SetSetting(key string, value string) error
}

type InstalledPlugin struct {
	Manifest           Manifest `json:"manifest"`
	Dir                string   `json:"dir"`
	Enabled            bool     `json:"enabled"`
	Error              string   `json:"error,omitempty"`
	GrantedPermissions []string `json:"grantedPermissions,omitempty"`
	Tampered           bool     `json:"tampered,omitempty"`
	IntegrityHash      string   `json:"integrityHash,omitempty"`
}

type PluginNetworkRecorder func(pluginID, method, url string, statusCode, bytes int, durationMs int64, errMsg string)

type Manager struct {
	mu              sync.RWMutex
	reg             *Registry
	store           PluginStateStore
	storage         *Storage
	wasm            *WasmRuntime
	dir             string
	app             *application.App
	devLog          DevLogFunc
	networkRecorder PluginNetworkRecorder
	plugins         map[string]InstalledPlugin
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

func (m *Manager) Store() PluginStateStore {
	return m.store
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
		granted := DefaultGrantedPermissions(manifest)
		var settings RuntimeSettings
		if row, ok := stateMap[manifest.ID]; ok {
			enabled = row.Enabled
			settings = ParseRuntimeSettings(row.SettingsJSON)
			if len(settings.GrantedPermissions) > 0 {
				granted = NormalizeGrantedPermissions(manifest, settings.GrantedPermissions)
			}
		} else {
			settings = RuntimeSettings{GrantedPermissions: granted}
		}

		var integrityErr string
		settings, okIntegrity, integrityErr := m.verifyIntegrityOnLoad(manifest.ID, dir, settings)
		if !okIntegrity {
			enabled = false
		}
		_ = m.saveRuntimeSettings(manifest.ID, enabled, settings)

		m.mu.Lock()
		m.plugins[manifest.ID] = InstalledPlugin{
			Manifest:           manifest,
			Dir:                dir,
			Enabled:            enabled,
			GrantedPermissions: granted,
			Tampered:           settings.Tampered,
			IntegrityHash:      settings.IntegrityHash,
			Error:              integrityErr,
		}
		m.mu.Unlock()
		if !okIntegrity {
			_ = m.store.SetPluginEnabled(manifest.ID, false)
			m.logPlugin(manifest.ID, "integrity", "error", integrityErr, "")
			continue
		}
		if enabled {
			if err := m.enableLocked(manifest.ID); err != nil {
				m.mu.Lock()
				p := m.plugins[manifest.ID]
				p.Error = err.Error()
				p.Enabled = false
				m.plugins[manifest.ID] = p
				m.mu.Unlock()
				_ = m.store.SetPluginEnabled(manifest.ID, false)
				m.logPlugin(manifest.ID, "enable", "error", err.Error(), "")
				m.emitPluginError(manifest.ID, "enable", err.Error())
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
	dir := p.Dir
	manifest := p.Manifest
	m.mu.Unlock()

	settings := m.loadRuntimeSettings(id)
	settings, err := m.establishIntegrityHash(dir, settings)
	if err != nil {
		return err
	}
	if err := m.saveRuntimeSettings(id, true, settings); err != nil {
		return err
	}

	m.mu.Lock()
	p, ok = m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", id)
	}
	m.mu.Unlock()

	if err := m.enableLocked(id); err != nil {
		_ = m.FailPlugin(id, "enable", err)
		return err
	}
	m.mu.Lock()
	p, ok = m.plugins[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("plugin %q not found", id)
	}
	p.Enabled = true
	p.Error = ""
	p.Tampered = false
	p.IntegrityHash = settings.IntegrityHash
	m.plugins[id] = p
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

func (m *Manager) InstallFromDir(src string, granted []string) (InstalledPlugin, error) {
	sig := VerifyDirSignature(src)
	if err := RequireValidSignature(sig); err != nil {
		return InstalledPlugin{}, err
	}
	manifest, err := LoadManifest(src)
	if err != nil {
		return InstalledPlugin{}, err
	}
	if err := ValidatePermissions(manifest.Permissions); err != nil {
		return InstalledPlugin{}, err
	}
	granted = NormalizeGrantedPermissions(manifest, granted)
	dest := filepath.Join(m.dir, manifest.ID)
	if err := os.RemoveAll(dest); err != nil {
		return InstalledPlugin{}, err
	}
	if err := copyDir(src, dest); err != nil {
		return InstalledPlugin{}, err
	}
	settings := RuntimeSettings{GrantedPermissions: granted}
	settings, err = m.establishIntegrityHash(dest, settings)
	if err != nil {
		return InstalledPlugin{}, err
	}
	_ = m.store.UpsertPlugin(manifest.ID, true, settings.JSON())
	m.mu.Lock()
	m.plugins[manifest.ID] = InstalledPlugin{
		Manifest:           manifest,
		Dir:                dest,
		Enabled:            true,
		GrantedPermissions: granted,
		IntegrityHash:      settings.IntegrityHash,
	}
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

func (m *Manager) InstallFromZip(zipPath string, granted []string) (InstalledPlugin, error) {
	sig := VerifyZipSignature(zipPath)
	if err := RequireValidSignature(sig); err != nil {
		return InstalledPlugin{}, err
	}
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
	return m.InstallFromDir(tmpDir, granted)
}

func (m *Manager) InstallFromWasm(wasmPath string, granted []string) (InstalledPlugin, error) {
	info, err := os.Stat(wasmPath)
	if err != nil {
		return InstalledPlugin{}, err
	}
	if info.Size() > maxWasmBundleBytes {
		return InstalledPlugin{}, fmt.Errorf("wasm module too large")
	}
	data, err := os.ReadFile(wasmPath) // #nosec G304 -- path from desktop file picker
	if err != nil {
		return InstalledPlugin{}, err
	}
	sig := VerifyWasmSignature(data)
	if err := RequireValidSignature(sig); err != nil {
		return InstalledPlugin{}, err
	}
	bundle, err := ParseWasmBundle(data)
	if err != nil {
		return InstalledPlugin{}, err
	}
	if err := bundle.ValidateEmbedded(); err != nil {
		return InstalledPlugin{}, err
	}
	if err := ValidatePermissions(bundle.Manifest.Permissions); err != nil {
		return InstalledPlugin{}, err
	}
	granted = NormalizeGrantedPermissions(bundle.Manifest, granted)
	dest := filepath.Join(m.dir, bundle.Manifest.ID)
	if err := os.RemoveAll(dest); err != nil {
		return InstalledPlugin{}, err
	}
	if err := writeWasmBundle(dest, bundle); err != nil {
		return InstalledPlugin{}, err
	}
	settings := RuntimeSettings{GrantedPermissions: granted}
	settings, err = m.establishIntegrityHash(dest, settings)
	if err != nil {
		return InstalledPlugin{}, err
	}
	_ = m.store.UpsertPlugin(bundle.Manifest.ID, true, settings.JSON())
	m.mu.Lock()
	m.plugins[bundle.Manifest.ID] = InstalledPlugin{
		Manifest:           bundle.Manifest,
		Dir:                dest,
		Enabled:            true,
		GrantedPermissions: granted,
		IntegrityHash:      settings.IntegrityHash,
	}
	m.mu.Unlock()
	if err := m.Enable(bundle.Manifest.ID); err != nil {
		return InstalledPlugin{}, err
	}
	p, ok := m.Get(bundle.Manifest.ID)
	if !ok {
		return InstalledPlugin{}, fmt.Errorf("plugin %q missing after install", bundle.Manifest.ID)
	}
	return p, nil
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
	activated := false
	defer func() {
		if !activated {
			m.disableLocked(id)
		}
	}()
	m.reg.RegisterManifest(p.Manifest)
	m.reg.RegisterContributions(id, p.Manifest.Contributes)
	if p.Manifest.Backend != "" {
		wasmPath := filepath.Join(p.Dir, p.Manifest.Backend)
		fetch := func(req WasmHTTPRequest) (WasmHTTPResponse, error) {
			return m.PluginHTTPFetch(id, req)
		}
		wp, err := m.wasm.LoadPluginWithFetch(id, wasmPath, p.Manifest, p.GrantedPermissions, fetch)
		if err != nil {
			return err
		}
		if len(p.Manifest.Contributes.Renderers) > 0 {
			m.reg.RegisterRenderer(wasmRendererAdapter{plugin: wp})
		}
		if wp.hasExport("before_fetch") {
			m.reg.RegisterBeforeFetch(id, m.wrapBeforeFetch(id, wp.BeforeFetchHook()))
		}
		if wp.hasExport("after_fetch") {
			m.reg.RegisterAfterFetch(id, m.wrapAfterFetch(id, wp.AfterFetchHook()))
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
	activated = true
	return nil
}

func (m *Manager) disableLocked(id string) {
	m.reg.UnregisterPlugin(id)
	m.wasm.Unload(id)
}

func (m *Manager) WasmCall(pluginID, exportName, input string) (string, error) {
	p, ok := m.Get(pluginID)
	if !ok || !p.Enabled {
		return "", fmt.Errorf("plugin %q not enabled", pluginID)
	}
	if strings.TrimSpace(p.Manifest.Backend) == "" {
		return "", fmt.Errorf("plugin %q has no wasm backend", pluginID)
	}
	if wasmExportNeedsNetwork(exportName) {
		if err := RequireGrantedPermission(p.GrantedPermissions, p.Manifest, PermNetworkFetch); err != nil {
			return "", err
		}
	}
	wp, ok := m.wasm.Get(pluginID)
	if !ok {
		return "", fmt.Errorf("plugin %q wasm module not loaded", pluginID)
	}
	if wasmExportNeedsNetwork(exportName) {
		beginWasmFetchBudget(pluginID, wasmMaxFetchesPerCall)
		defer endWasmFetchBudget(pluginID)
	}
	out, err := wp.CallExport(exportName, []byte(input))
	if err != nil {
		m.LogPluginError(pluginID, "wasm:"+exportName, err.Error(), "")
		return "", err
	}
	return string(out), nil
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
