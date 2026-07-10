// SPDX-License-Identifier: MIT
package app

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"quad4/reticulum-go/pkg/debug"
	"quad4/reticulum-go/pkg/transport"

	"renbrowser/internal/apperrors"
	"renbrowser/internal/cache"
	"renbrowser/internal/content"
	"renbrowser/internal/fonts"
	"renbrowser/internal/limits"
	"renbrowser/internal/micron"
	"renbrowser/internal/micronwasm"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins"
	"renbrowser/internal/rns"
	"renbrowser/internal/serverlog"
	"renbrowser/internal/store"
)

type PageResponse struct {
	URL         string `json:"url"`
	NodeHash    string `json:"nodeHash"`
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	HTML        string `json:"html"`
	Raw         string `json:"raw"`
	BinaryB64   string `json:"binaryB64,omitempty"`
	PageFG      string `json:"pageFg,omitempty"`
	PageBG      string `json:"pageBg,omitempty"`
	DurationMs  int64  `json:"durationMs"`
	FromCache   bool   `json:"fromCache"`
	CachedAt    int64  `json:"cachedAt,omitempty"`
	Hops        int    `json:"hops"`
	Interface   string `json:"interface,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorKind   string `json:"errorKind,omitempty"`
}

type DevLogEntry struct {
	Time    int64  `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type NetworkEntry struct {
	Time       int64  `json:"time"`
	URL        string `json:"url"`
	NodeHash   string `json:"nodeHash"`
	Path       string `json:"path"`
	DurationMs int64  `json:"durationMs"`
	Bytes      int    `json:"bytes"`
	FromCache  bool   `json:"fromCache"`
	Hops       int    `json:"hops"`
	Interface  string `json:"interface,omitempty"`
	Error      string `json:"error,omitempty"`
	Source     string `json:"source,omitempty"`
	PluginID   string `json:"pluginId,omitempty"`
	Method     string `json:"method,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
}

type HistoryState struct {
	CanGoBack    bool   `json:"canGoBack"`
	CanGoForward bool   `json:"canGoForward"`
	Current      string `json:"current"`
}

type TabSnapshot = store.TabSnapshot

type HistoryEntry = store.HistoryEntry

type ThemeSettings struct {
	Mode            string            `json:"mode"`
	Accent          string            `json:"accent"`
	FontFamily      string            `json:"fontFamily"`
	FontSize        int               `json:"fontSize"`
	CustomTokens    map[string]string `json:"customTokens"`
	CompactToolbar  bool              `json:"compactToolbar"`
	OverlaySidebars bool              `json:"overlaySidebars"`
}

type BrowserService struct {
	mu                     sync.RWMutex
	stack                  *rns.Stack
	app                    *application.App
	store                  *store.Store
	storePath              string
	profileName            string
	publicMode             bool
	resetWindow            bool
	pageCache              *cache.PageCache
	devLogs                []DevLogEntry
	networkLog             []NetworkEntry
	theme                  ThemeSettings
	history                []string
	histIdx                int
	lastPage               PageResponse
	plugins                *plugins.Manager
	downloads              *downloadManager
	shuttingDown           bool
	shutdownOnce           sync.Once
	downloadsReconcileOnce sync.Once

	windowPersistOnce  sync.Once
	windowCaptureMu    sync.Mutex
	windowCaptureTimer *time.Timer
}

type ServiceOptions struct {
	ProfilePath   string
	ProfileName   string
	PublicMode    bool
	ResetWindow   bool
	ExportProfile string
	ImportProfile string
}

func NewBrowserService(stack *rns.Stack, app *application.App) (*BrowserService, error) {
	return NewBrowserServiceWithOptions(stack, app, ServiceOptions{})
}

func NewBrowserServiceWithOptions(stack *rns.Stack, app *application.App, opts ServiceOptions) (*BrowserService, error) {
	dbPath := opts.ProfilePath
	if dbPath == "" {
		dbPath = store.ProfilePath(opts.ProfileName)
	}
	profileName := store.SanitizeProfileName(opts.ProfileName)
	if profileName == "" {
		profileName = "default"
	}

	st, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	svc := &BrowserService{
		stack:       stack,
		app:         app,
		store:       st,
		storePath:   dbPath,
		profileName: profileName,
		publicMode:  opts.PublicMode,
		resetWindow: opts.ResetWindow,
		pageCache:   cache.NewPageCache(128),
		downloads:   newDownloadManager(),
	}
	svc.downloads.onChange = func(items []ActiveDownload) {
		svc.persistDownloadRecovery(items)
		if svc.app != nil {
			svc.app.Event.Emit("downloads:active", items)
		}
	}
	if err := svc.runStartupProfileIO(opts.ExportProfile, opts.ImportProfile); err != nil {
		_ = st.Close()
		return nil, err
	}
	svc.theme = svc.loadThemeSettings()
	if stack != nil {
		svc.bindPersistence()
	}
	_ = svc.GetDownloadDir()
	if stack != nil {
		svc.reconcileDownloadsOnce()
	}
	return svc, nil
}

func (s *BrowserService) AttachStack(stack *rns.Stack) {
	if stack == nil {
		return
	}
	s.mu.Lock()
	s.stack = stack
	s.mu.Unlock()
	s.bindPersistence()
	s.reconcileDownloadsOnce()
}

func (s *BrowserService) reconcileDownloadsOnce() {
	s.downloadsReconcileOnce.Do(func() {
		s.reconcileDownloadsOnStartup()
	})
}

func (s *BrowserService) bindPersistence() {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return
	}
	stack.Handler().SetOnAnnounce(func(node nomadnet.Node) {
		_ = s.store.UpsertNode(node)
		if s.app != nil {
			s.app.Event.Emit("node:discovered", s.ListNodes())
		}
	})
}

func (s *BrowserService) Store() *store.Store {
	return s.store
}

func (s *BrowserService) PluginManager() *plugins.Manager {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plugins
}

func (s *BrowserService) SetApp(app *application.App) {
	s.mu.Lock()
	s.app = app
	s.mu.Unlock()
	s.ensureWindowPersistence()
	s.bindPersistence()
	s.emitStoreHealth()
}

func (s *BrowserService) StartReticulum() error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return errors.New("reticulum not initialized")
	}
	err := stack.Start()
	s.log("info", "Reticulum start", errString(err))
	if err == nil && s.app != nil {
		s.app.Event.Emit("rns:status", "online")
	}
	return err
}

func (s *BrowserService) StopReticulum() error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nil
	}
	err := stack.Stop()
	s.log("info", "Reticulum stop", errString(err))
	if s.app != nil {
		s.app.Event.Emit("rns:status", "offline")
	}
	return err
}

func (s *BrowserService) GetStatus() rns.Status {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return rns.Status{}
	}
	return stack.Status()
}

func (s *BrowserService) ConfigPath() string {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return ""
	}
	return stack.ConfigPath()
}

func (s *BrowserService) ListInterfaces() []rns.InterfaceInfo {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nil
	}
	return stack.ListInterfaces()
}

func (s *BrowserService) SetInterfaceEnabled(name string, enabled bool) error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return errors.New("reticulum not initialized")
	}
	err := stack.SetInterfaceEnabled(name, enabled)
	if err != nil {
		return err
	}
	s.log("info", "interface updated", fmt.Sprintf("%s enabled=%v", name, enabled))
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return nil
}

func (s *BrowserService) DeleteInterface(name string) error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return errors.New("reticulum not initialized")
	}
	err := stack.DeleteInterface(name)
	if err != nil {
		return err
	}
	s.log("info", "interface deleted", name)
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return nil
}

func (s *BrowserService) SetEnableTransport(enabled bool) error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return errors.New("reticulum not initialized")
	}
	if err := stack.SetEnableTransport(enabled); err != nil {
		return err
	}
	s.log("info", "transport updated", fmt.Sprintf("enabled=%v", enabled))
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return nil
}

func (s *BrowserService) SetShareInstance(enabled bool) error {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return errors.New("reticulum not initialized")
	}
	if err := stack.SetShareInstance(enabled); err != nil {
		return err
	}
	s.log("info", "share instance updated", fmt.Sprintf("enabled=%v", enabled))
	if s.app != nil {
		s.app.Event.Emit("rns:status", "reload")
	}
	return nil
}

func (s *BrowserService) SetLogLevel(level int) int {
	if level < debug.DebugCritical {
		level = debug.DebugCritical
	}
	if level > debug.DebugAll {
		level = debug.DebugAll
	}
	debug.SetDebugLevel(level)
	debug.Init()
	s.log("info", "log level", fmt.Sprintf("%d", level))
	return level
}

func (s *BrowserService) ListNodes() []nomadnet.Node {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	stored, err := s.store.ListNodes()
	var out []nomadnet.Node
	if stack == nil {
		if err == nil {
			out = stored
		}
	} else {
		live := stack.Handler().List()
		if err != nil || len(stored) == 0 {
			out = live
		} else {
			out = mergeNodes(stored, live)
		}
	}
	if len(out) > 0 {
		sort.Slice(out, func(i, j int) bool {
			return out[i].LastSeen > out[j].LastSeen
		})
	}
	return enrichNodeHops(s, out)
}

func enrichNodeHops(s *BrowserService, nodes []nomadnet.Node) []nomadnet.Node {
	for i := range nodes {
		hops := s.hopsForNode(nodes[i].Hash)
		if hops < 0 {
			continue
		}
		if hops > 255 {
			hops = 255
		}
		nodes[i].Hops = uint8(hops)
	}
	return nodes
}

func mergeNodes(stored, live []nomadnet.Node) []nomadnet.Node {
	liveMap := make(map[string]nomadnet.Node, len(live))
	for _, n := range live {
		liveMap[strings.ToLower(n.Hash)] = n
	}
	seen := make(map[string]bool, len(stored)+len(live))
	out := make([]nomadnet.Node, 0, len(stored)+len(live))

	for _, n := range live {
		key := strings.ToLower(n.Hash)
		seen[key] = true
		out = append(out, n)
	}
	for _, n := range stored {
		key := strings.ToLower(n.Hash)
		if _, ok := liveMap[key]; ok {
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, n)
	}
	return out
}

func (s *BrowserService) ListSystemFonts() []string {
	return fonts.ListSystemFonts()
}

func (s *BrowserService) GetBrowsingHistory(limit int) []HistoryEntry {
	if s.publicMode {
		return []HistoryEntry{}
	}
	rows, err := s.store.BrowsingHistory(limit)
	if err != nil {
		return []HistoryEntry{}
	}
	return rows
}

func (s *BrowserService) ClearBrowsingHistory() error {
	if s.publicMode {
		return nil
	}
	return s.store.ClearBrowsingHistory()
}

func (s *BrowserService) GetKeybinds() KeybindSettings {
	raw, err := s.store.GetSetting(keybindsSettingKey)
	if err != nil {
		return DefaultKeybinds()
	}
	settings, err := decodeKeybinds(raw)
	if err != nil {
		return DefaultKeybinds()
	}
	return settings
}

func (s *BrowserService) SetKeybinds(settings KeybindSettings) KeybindSettings {
	settings = mergeKeybinds(settings)
	encoded, err := encodeKeybinds(settings)
	if err != nil {
		return DefaultKeybinds()
	}
	_ = s.store.SetSetting(keybindsSettingKey, encoded)
	return settings
}

func (s *BrowserService) GetFavorites() []string {
	if s.publicMode {
		return []string{}
	}
	return s.store.Favorites()
}

func (s *BrowserService) AddFavorite(url string) []string {
	if s.publicMode {
		return []string{url}
	}
	return s.store.AddFavorite(url)
}

func (s *BrowserService) RemoveFavorite(url string) []string {
	if s.publicMode {
		return []string{}
	}
	return s.store.RemoveFavorite(url)
}

func (s *BrowserService) GetRecent() []string {
	return s.store.Recent()
}

func (s *BrowserService) Navigate(rawURL string) PageResponse {
	return s.navigate(rawURL, true, false)
}

func (s *BrowserService) OpenURL(rawURL string) PageResponse {
	return s.navigate(rawURL, false, false)
}

func (s *BrowserService) NavigateFresh(rawURL string) PageResponse {
	return s.navigate(rawURL, true, true)
}

func (s *BrowserService) OpenFreshURL(rawURL string) PageResponse {
	return s.navigate(rawURL, false, true)
}

func (s *BrowserService) navigate(rawURL string, pushHistory, skipCache bool) PageResponse {
	s.mu.RLock()
	shutting := s.shuttingDown
	s.mu.RUnlock()
	if shutting {
		resp := PageResponse{URL: rawURL}
		applyPageError(&resp, "application shutting down", nil)
		return resp
	}

	s.mu.RLock()
	manager := s.plugins
	s.mu.RUnlock()
	if manager == nil {
		if isAboutURL(rawURL) {
			return s.aboutPage(pushHistory)
		}
		if isLicenseURL(rawURL) {
			return s.licensePage(pushHistory)
		}
		if isEditorURL(rawURL) {
			return s.editorPage(pushHistory)
		}
		if isConfigURL(rawURL) {
			return s.configPage(pushHistory)
		}
		if isSettingsURL(rawURL) {
			return s.settingsPage(pushHistory)
		}
	} else if resp, ok := s.handlePluginScheme(rawURL, pushHistory); ok {
		return resp
	}

	if isDocumentURL(rawURL) {
		return s.documentPage(rawURL, pushHistory)
	}

	parsed, err := nomadnet.ParseURL(rawURL)
	if err != nil {
		resp := PageResponse{URL: rawURL}
		applyPageError(&resp, err.Error(), nil)
		s.log("error", "parse url", err.Error())
		s.recordNetwork(resp)
		return resp
	}

	url := nomadnet.FormatURLWithRequest(parsed.NodeHash, parsed.Path, parsed.Request)
	if pushHistory {
		s.pushHistory(url)
		if !s.publicMode {
			title := parsed.NodeHash
			s.mu.RLock()
			stack := s.stack
			s.mu.RUnlock()
			if stack != nil {
				title = historyTitle(parsed.NodeHash, stack.Handler())
			}
			_ = s.store.AddHistory(url, title, parsed.NodeHash)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if !skipCache && s.GetBrowserPrefs().PageCacheEnabled {
		if entry, ok := s.pageCache.Get(parsed.NodeHash, parsed.Path, parsed.Request); ok {
			if entry.IsStale(cache.DefaultPageCacheMaxAge) {
				if fresh := s.tryRefreshCachedPage(ctx, url, parsed, entry); fresh != nil {
					return *fresh
				}
				return s.pageResponseFromCache(url, parsed, entry)
			}
			return s.pageResponseFromCache(url, parsed, entry)
		}
	}

	fetchPath, fetchReq, cancelled := s.runFetchHooks(parsed.NodeHash, parsed.Path, parsed.Request)
	if cancelled {
		resp := PageResponse{URL: url, NodeHash: parsed.NodeHash, Path: parsed.Path}
		applyPageError(&resp, "fetch cancelled by extension", nil)
		s.setLastPage(resp)
		s.recordNetwork(resp)
		if s.app != nil {
			s.app.Event.Emit("page:error", resp)
		}
		return resp
	}

	fetch := s.fetchWithRetry(ctx, parsed.NodeHash, fetchPath, fetchReq)
	resp := PageResponse{
		URL:         url,
		NodeHash:    fetch.NodeHash,
		Path:        fetch.Path,
		ContentType: fetch.ContentType,
		DurationMs:  fetch.DurationMs,
		Hops:        fetch.Hops,
		Interface:   fetch.Interface,
	}
	if fetch.Error != "" {
		applyPageError(&resp, fetch.Error, fetch.Body)
		resp.Hops = fetch.Hops
		if resp.Hops < 0 {
			resp.Hops = s.hopsForNode(parsed.NodeHash)
		}
		s.log("error", "fetch page", fetch.Error)
		s.setLastPage(resp)
		s.recordNetwork(resp)
		if s.app != nil {
			s.app.Event.Emit("page:error", resp)
		}
		return resp
	}

	if len(fetch.Body) > 0 && len(fetch.Body) < 256 {
		if kind, detail := apperrors.ClassifyFetch("", fetch.Body); kind == apperrors.KindNotFound || kind == apperrors.KindInternal {
			applyPageError(&resp, detail, fetch.Body)
			s.log("error", "page response", detail)
			s.setLastPage(resp)
			s.recordNetwork(resp)
			if s.app != nil {
				s.app.Event.Emit("page:error", resp)
			}
			return resp
		}
	}

	rendered := content.Render(fetch.Path, fetch.Body, fetch.NodeHash)
	resp.HTML = rendered.HTML
	resp.Raw = rendered.Raw
	resp.ContentType = rendered.Kind
	resp.PageFG = rendered.PageFG
	resp.PageBG = rendered.PageBG
	if s.GetBrowserPrefs().PageCacheEnabled {
		if len(fetch.Body) <= limits.MaxPageBytes() {
			s.pageCache.Put(fetch.NodeHash, fetch.Path, parsed.Request, fetch.Body, rendered.Kind)
		}
	}

	s.log("info", "page loaded", url)
	s.setLastPage(resp)
	s.recordNetwork(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
		s.app.Event.Emit("node:discovered", s.ListNodes())
	}
	return resp
}

func (s *BrowserService) DownloadFile(rawURL string) (string, error) {
	fetch, err := s.fetchFile(rawURL)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fetch.Body), nil
}

// fetchFile requests a /file/ resource and returns the raw (already
// metadata-stripped) file bytes along with the server-supplied file name,
// if any. Nomad Network nodes respond to /file/ requests with an RNS
// resource transfer carrying a {"name": <filename>} metadata block ahead of
// the file bytes; reticulum-go's Link splits that off before the response
// ever reaches here, so fetch.Body is always the plain file content.
func (s *BrowserService) fetchFile(rawURL string) (nomadnet.FetchResult, error) {
	return s.fetchFileTracked(rawURL, nil)
}

// fetchFileTracked behaves like fetchFile but, when tracker is non-nil,
// also reports byte progress to the download manager and makes the fetch
// cancellable through it.
func (s *BrowserService) fetchFileTracked(rawURL string, tracker *downloadTracker) (nomadnet.FetchResult, error) {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nomadnet.FetchResult{}, errors.New("reticulum not ready")
	}
	parsed, err := nomadnet.ParseURL(rawURL)
	if err != nil {
		return nomadnet.FetchResult{}, err
	}
	if !strings.HasPrefix(parsed.Path, "/file/") {
		parsed.Path = "/file/" + strings.TrimPrefix(parsed.Path, "/")
	}

	// Large files are delivered as multi-packet RNS resource transfers that
	// can legitimately take minutes on slow or high-hop interfaces, so this
	// needs far more headroom than a page fetch; see requestTimeouts in
	// internal/nomadnet/browser.go for the matching inner timeouts.
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	if tracker != nil {
		var trackerCancel context.CancelFunc
		ctx, trackerCancel = context.WithCancel(ctx)
		defer trackerCancel()
		tracker.mgr.bindCancel(tracker.id, trackerCancel)
	}

	hooks := s.fileFetchHooks(rawURL)
	if tracker != nil {
		hooks = mergeFetchHooks(hooks, tracker.fetchHooks())
	}
	fetch := stack.Browser().FetchWithHooks(ctx, parsed.NodeHash, parsed.Path, parsed.Request, hooks)
	if fetch.Error != "" {
		s.log("error", "file fetch failed", fmt.Sprintf("%s: %s", rawURL, fetch.Error))
		return nomadnet.FetchResult{}, errors.New(fetch.Error)
	}
	if len(fetch.Body) == 0 {
		s.log("error", "file fetch failed", fmt.Sprintf("%s: empty file response", rawURL))
		return nomadnet.FetchResult{}, fmt.Errorf("empty file response")
	}
	return fetch, nil
}

// ListActiveDownloads returns the current set of in-flight and recently
// finished mesh file downloads for the downloads panel.
func (s *BrowserService) ListActiveDownloads() []ActiveDownload {
	return s.downloads.list()
}

// CancelDownload aborts an in-flight download started via DownloadToDir.
// Returns false if the download is unknown or already finished.
func (s *BrowserService) CancelDownload(id string) bool {
	return s.downloads.cancel(id)
}

// DismissDownload removes a finished or failed download entry from the
// active downloads list without affecting any file already written to disk.
func (s *BrowserService) DismissDownload(id string) {
	s.downloads.dismiss(id)
}

// fileFetchHooks reports every stage transition and byte-progress update of
// a /file/ fetch to the app's dev log console, so a failed or stalled
// download is visible immediately instead of surfacing only as a single
// opaque error once everything has already timed out.
func (s *BrowserService) fileFetchHooks(rawURL string) *nomadnet.FetchHooks {
	return &nomadnet.FetchHooks{
		OnStage: func(stage, detail string) {
			s.log("debug", "file fetch: "+stage, fmt.Sprintf("%s: %s", rawURL, detail))
		},
		OnProgress: func(p nomadnet.FetchProgress) {
			s.log("debug", "file fetch: progress", fmt.Sprintf("%s: %d/%d bytes", rawURL, p.Received, p.Total))
		},
	}
}

func (s *BrowserService) SaveDownload(rawURL string, destPath string) error {
	fetch, err := s.fetchFile(rawURL)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, fetch.Body, 0o600)
}

func (s *BrowserService) ResolveMicronLink(currentURL, destination, fieldsSpec string, inputs []micron.FieldInput) (string, error) {
	return micron.ResolveNavigation(currentURL, destination, fieldsSpec, inputs)
}

func (s *BrowserService) GetTabs() []TabSnapshot {
	if s.publicMode {
		return []TabSnapshot{}
	}
	return s.store.Tabs()
}

func (s *BrowserService) SaveTabs(tabs []TabSnapshot) []TabSnapshot {
	if s.publicMode {
		return tabs
	}
	result := s.store.SaveTabs(tabs)
	health := s.store.Health()
	if !health.OK {
		s.mu.Lock()
		s.emitStoreHealthLocked(health)
		s.mu.Unlock()
	}
	return result
}

func (s *BrowserService) RenderRaw(path string, raw string) PageResponse {
	rendered := content.Render(path, []byte(raw), "")
	return PageResponse{
		Path:        path,
		ContentType: rendered.Kind,
		HTML:        rendered.HTML,
		Raw:         rendered.Raw,
		PageFG:      rendered.PageFG,
		PageBG:      rendered.PageBG,
	}
}

func (s *BrowserService) RenderRawBase64(path string, rawB64 string) PageResponse {
	body, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		return PageResponse{Path: path, Error: err.Error()}
	}
	rendered := content.Render(path, body, "")
	return PageResponse{
		Path:        path,
		ContentType: rendered.Kind,
		HTML:        rendered.HTML,
		Raw:         rendered.Raw,
		PageFG:      rendered.PageFG,
		PageBG:      rendered.PageBG,
	}
}

func (s *BrowserService) GetLastPage() PageResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastPage
}

func (s *BrowserService) ExportTheme() string {
	theme := s.GetTheme()
	raw, _ := json.MarshalIndent(theme, "", "  ")
	return string(raw)
}

func (s *BrowserService) ImportTheme(jsonData string) (ThemeSettings, error) {
	var theme ThemeSettings
	if err := json.Unmarshal([]byte(jsonData), &theme); err != nil {
		return ThemeSettings{}, err
	}
	return s.SetTheme(theme), nil
}

func (s *BrowserService) GetDevLogs() []DevLogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]DevLogEntry, len(s.devLogs))
	copy(out, s.devLogs)
	return out
}

func (s *BrowserService) GetNetworkLog() []NetworkEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NetworkEntry, len(s.networkLog))
	copy(out, s.networkLog)
	return out
}

func (s *BrowserService) ExportDevLogs() string {
	logs := s.GetDevLogs()
	raw, _ := json.MarshalIndent(logs, "", "  ")
	return string(raw)
}

func (s *BrowserService) ClearDevLogs() {
	s.mu.Lock()
	s.devLogs = nil
	s.mu.Unlock()
}

func (s *BrowserService) HistoryState() HistoryState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	current := ""
	if len(s.history) > 0 && s.histIdx >= 0 && s.histIdx < len(s.history) {
		current = s.history[s.histIdx]
	}
	return HistoryState{
		CanGoBack:    s.histIdx > 0,
		CanGoForward: s.histIdx >= 0 && s.histIdx < len(s.history)-1,
		Current:      current,
	}
}

func (s *BrowserService) GoBack() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.histIdx > 0 {
		s.histIdx--
	}
	if len(s.history) == 0 {
		return ""
	}
	return s.history[s.histIdx]
}

func (s *BrowserService) GoForward() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.histIdx < len(s.history)-1 {
		s.histIdx++
	}
	if len(s.history) == 0 {
		return ""
	}
	return s.history[s.histIdx]
}

func (s *BrowserService) pushHistory(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.histIdx >= 0 && s.histIdx < len(s.history) && s.history[s.histIdx] == url {
		return
	}
	if s.histIdx < len(s.history)-1 {
		s.history = s.history[:s.histIdx+1]
	}
	s.history = append(s.history, url)
	s.histIdx = len(s.history) - 1
}

func (s *BrowserService) setLastPage(page PageResponse) {
	s.mu.Lock()
	s.lastPage = page
	s.mu.Unlock()
}

func (s *BrowserService) recordNetwork(page PageResponse) {
	hops := page.Hops
	if hops < 0 && page.NodeHash != "" {
		hops = s.hopsForNode(page.NodeHash)
	}
	entry := NetworkEntry{
		Time:       time.Now().UnixMilli(),
		URL:        page.URL,
		NodeHash:   page.NodeHash,
		Path:       page.Path,
		DurationMs: page.DurationMs,
		Bytes:      len(page.Raw),
		FromCache:  page.FromCache,
		Hops:       hops,
		Interface:  page.Interface,
		Error:      page.Error,
	}
	s.mu.Lock()
	s.networkLog = append(s.networkLog, entry)
	if len(s.networkLog) > 300 {
		s.networkLog = s.networkLog[len(s.networkLog)-300:]
	}
	s.mu.Unlock()
	if s.app != nil {
		s.app.Event.Emit("network:entry", entry)
	}
}

func (s *BrowserService) RecordPluginNetworkFetch(pluginID, method, rawURL string, statusCode, bytes int, durationMs int64, errMsg string) {
	path := rawURL
	if parsed, err := url.Parse(rawURL); err == nil && parsed != nil {
		if parsed.Host != "" {
			path = parsed.Host + parsed.Path
		}
		if path == "" {
			path = "/"
		}
	}
	entry := NetworkEntry{
		Time:       time.Now().UnixMilli(),
		URL:        rawURL,
		Path:       path,
		DurationMs: durationMs,
		Bytes:      bytes,
		Hops:       -1,
		Interface:  "plugin",
		Source:     "plugin",
		PluginID:   pluginID,
		Method:     method,
		StatusCode: statusCode,
		Error:      errMsg,
	}
	s.mu.Lock()
	s.networkLog = append(s.networkLog, entry)
	if len(s.networkLog) > 300 {
		s.networkLog = s.networkLog[len(s.networkLog)-300:]
	}
	s.mu.Unlock()
	if s.app != nil {
		s.app.Event.Emit("network:entry", entry)
	}
}

func (s *BrowserService) DevLog(level, message, detail string) {
	s.log(level, message, detail)
}

func (s *BrowserService) log(level, message, detail string) {
	s.mu.Lock()
	s.devLogs = append(s.devLogs, DevLogEntry{
		Time:    time.Now().UnixMilli(),
		Level:   level,
		Message: message,
		Detail:  detail,
	})
	if len(s.devLogs) > 500 {
		s.devLogs = s.devLogs[len(s.devLogs)-500:]
	}
	s.mu.Unlock()
	if s.app != nil {
		payload, _ := json.Marshal(DevLogEntry{
			Time:    time.Now().UnixMilli(),
			Level:   level,
			Message: message,
			Detail:  detail,
		})
		s.app.Event.Emit("dev:log", string(payload))
	}
	serverlog.Emit(level, message, detail)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *BrowserService) fetchWithRetry(ctx context.Context, nodeHash, path string, req nomadnet.RequestData) nomadnet.FetchResult {
	s.mu.RLock()
	stack := s.stack
	s.mu.RUnlock()
	if stack == nil {
		return nomadnet.FetchResult{Error: "reticulum not ready"}
	}
	const maxAttempts = 3
	var last nomadnet.FetchResult
	for attempt := range maxAttempts {
		if attempt > 0 {
			wait := time.Duration(attempt) * time.Second
			select {
			case <-ctx.Done():
				last.Error = ctx.Err().Error()
				return last
			case <-time.After(wait):
			}
			if shouldRefreshPathOnRetry(last.Error) {
				stack.Browser().RefreshPathForRetry(nodeHash)
			}
		}
		last = stack.Browser().Fetch(ctx, nodeHash, path, req)
		if last.Error == "" {
			last.Body = s.runAfterFetchHooks(nodeHash, path, req, last.Body)
			return last
		}
		if ctx.Err() != nil {
			last.Error = ctx.Err().Error()
			return last
		}
	}
	return last
}

func shouldRefreshPathOnRetry(errMsg string) bool {
	return nomadnet.ShouldRefreshPath(errMsg)
}

func (s *BrowserService) handlePluginScheme(rawURL string, pushHistory bool) (PageResponse, bool) {
	s.mu.RLock()
	manager := s.plugins
	s.mu.RUnlock()
	if manager == nil {
		return PageResponse{}, false
	}
	scheme, ok := manager.Registry().HandleScheme(rawURL)
	if !ok {
		return PageResponse{}, false
	}
	if scheme.DelegateFrontend {
		resp := PageResponse{
			URL:         scheme.URL,
			ContentType: "plugin-scheme",
			Raw:         scheme.URL,
		}
		if pushHistory {
			s.pushHistory(scheme.URL)
		}
		s.setLastPage(resp)
		if s.app != nil {
			s.app.Event.Emit("plugin:scheme", map[string]string{
				"pluginId": scheme.PluginID,
				"url":      scheme.URL,
				"handler":  scheme.Handler,
			})
			s.app.Event.Emit("page:loaded", resp)
		}
		return resp, true
	}
	resp := PageResponse{
		URL:         scheme.URL,
		Path:        scheme.Path,
		ContentType: scheme.ContentType,
		HTML:        scheme.HTML,
		Raw:         scheme.Raw,
		PageFG:      scheme.PageFG,
		PageBG:      scheme.PageBG,
	}
	if pushHistory {
		s.pushHistory(scheme.URL)
		if scheme.HistoryTitle != "" && !s.publicMode {
			_ = s.store.AddHistory(scheme.URL, scheme.HistoryTitle, "")
		}
	}
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp, true
}

func historyTitle(nodeHash string, handler *nomadnet.AnnounceHandler) string {
	if node, ok := handler.Get(nodeHash); ok && node.Name != "" {
		return node.Name
	}
	return nodeHash
}

func (s *BrowserService) hopsForNode(nodeHash string) int {
	if s == nil || s.stack == nil || nodeHash == "" {
		return -1
	}
	destHash, err := hex.DecodeString(nodeHash)
	if err != nil || len(destHash) != 16 {
		return -1
	}
	tr := s.stack.Transport()
	if tr == nil {
		return announceHops(s.stack.Handler(), nodeHash)
	}
	hops := tr.HopsTo(destHash)
	if hops < transport.PathfinderM {
		return int(hops)
	}
	return announceHops(s.stack.Handler(), nodeHash)
}

func announceHops(handler *nomadnet.AnnounceHandler, nodeHash string) int {
	if handler == nil {
		return -1
	}
	if node, ok := handler.Get(nodeHash); ok {
		return int(node.Hops)
	}
	return -1
}

type MicronWasmFetchResult struct {
	ReleaseTag string `json:"releaseTag"`
	WasmBase64 string `json:"wasmBase64"`
	Sha256Hex  string `json:"sha256Hex"`
}

func (s *BrowserService) FetchMicronParserGoRelease(tag string) (MicronWasmFetchResult, error) {
	result, err := micronwasm.FetchVerifiedRelease(tag)
	if err != nil {
		return MicronWasmFetchResult{}, err
	}
	return MicronWasmFetchResult{
		ReleaseTag: result.ReleaseTag,
		WasmBase64: result.WasmBase64,
		Sha256Hex:  result.Sha256Hex,
	}, nil
}
