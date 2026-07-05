// SPDX-License-Identifier: MIT
package plugins

import (
	"sort"
	"strings"
	"sync"
)

type SchemeResult struct {
	URL              string
	Path             string
	ContentType      string
	HTML             string
	Raw              string
	PageFG           string
	PageBG           string
	HistoryTitle     string
	DelegateFrontend bool
	PluginID         string
	Handler          string
}

type SchemeHandler func(rawURL string) (SchemeResult, bool)

type FetchContext struct {
	NodeHash string
	Path     string
	Request  map[string]string
}

type FetchHookResult struct {
	Cancel bool
	Path   string
}

type BeforeFetchHook func(ctx FetchContext) (FetchHookResult, error)
type AfterFetchHook func(ctx FetchContext, body []byte) ([]byte, error)

type Renderer interface {
	ID() string
	Priority() int
	PluginID() string
	Match(path string, body []byte, detected string) bool
	Render(path string, body []byte, nodeHash string) (Rendered, error)
}

type Rendered struct {
	Kind    string
	HTML    string
	Raw     string
	PageFG  string
	PageBG  string
	IsError bool
}

type ContributionsView struct {
	Renderers  []RendererContribView  `json:"renderers"`
	URLSchemes []URLSchemeContribView `json:"urlSchemes"`
	Panels     []PanelContribView     `json:"panels"`
	Commands   []CommandContribView   `json:"commands"`
	Themes     []ThemeContribView     `json:"themes"`
	Settings   []SettingsContribView  `json:"settings"`
	DevTools   []DevToolsContribView  `json:"devtools"`
}

type RendererContribView struct {
	PluginID string `json:"pluginId"`
	RendererContrib
}

type URLSchemeContribView struct {
	PluginID string `json:"pluginId"`
	URLSchemeContrib
}

type PanelContribView struct {
	PluginID string `json:"pluginId"`
	PanelContrib
}

type CommandContribView struct {
	PluginID string `json:"pluginId"`
	CommandContrib
}

type ThemeContribView struct {
	PluginID string `json:"pluginId"`
	ThemeContrib
}

type SettingsContribView struct {
	PluginID string `json:"pluginId"`
	SettingsContrib
}

type DevToolsContribView struct {
	PluginID string `json:"pluginId"`
	DevToolsContrib
}

type Registry struct {
	mu          sync.RWMutex
	renderers   []Renderer
	schemes     map[string]schemeEntry
	beforeFetch []fetchHookEntry
	afterFetch  []afterFetchEntry
	panels      []PanelContribView
	commands    []CommandContribView
	themes      []ThemeContribView
	settings    []SettingsContribView
	devtools    []DevToolsContribView
	manifests   map[string]Manifest
}

type schemeEntry struct {
	pluginID string
	handler  SchemeHandler
	contrib  URLSchemeContrib
}

type fetchHookEntry struct {
	pluginID string
	hook     BeforeFetchHook
}

type afterFetchEntry struct {
	pluginID string
	hook     AfterFetchHook
}

func NewRegistry() *Registry {
	return &Registry{
		schemes:   make(map[string]schemeEntry),
		manifests: make(map[string]Manifest),
	}
}

func (r *Registry) RegisterRenderer(renderer Renderer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.renderers = append(r.renderers, renderer)
	sort.SliceStable(r.renderers, func(i, j int) bool {
		return r.renderers[i].Priority() > r.renderers[j].Priority()
	})
}

func (r *Registry) UnregisterPlugin(pluginID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	next := r.renderers[:0]
	for _, renderer := range r.renderers {
		if renderer.PluginID() != pluginID {
			next = append(next, renderer)
		}
	}
	r.renderers = next
	for scheme, entry := range r.schemes {
		if entry.pluginID == pluginID {
			delete(r.schemes, scheme)
		}
	}
	r.beforeFetch = filterFetchHooks(r.beforeFetch, pluginID)
	r.afterFetch = filterAfterHooks(r.afterFetch, pluginID)
	r.panels = filterPanels(r.panels, pluginID)
	r.commands = filterCommands(r.commands, pluginID)
	r.themes = filterThemes(r.themes, pluginID)
	r.settings = filterSettings(r.settings, pluginID)
	r.devtools = filterDevtools(r.devtools, pluginID)
	delete(r.manifests, pluginID)
}

func filterFetchHooks(in []fetchHookEntry, pluginID string) []fetchHookEntry {
	out := in[:0]
	for _, h := range in {
		if h.pluginID != pluginID {
			out = append(out, h)
		}
	}
	return out
}

func filterAfterHooks(in []afterFetchEntry, pluginID string) []afterFetchEntry {
	out := in[:0]
	for _, h := range in {
		if h.pluginID != pluginID {
			out = append(out, h)
		}
	}
	return out
}

func filterPanels(in []PanelContribView, pluginID string) []PanelContribView {
	out := in[:0]
	for _, p := range in {
		if p.PluginID != pluginID {
			out = append(out, p)
		}
	}
	return out
}

func filterCommands(in []CommandContribView, pluginID string) []CommandContribView {
	out := in[:0]
	for _, c := range in {
		if c.PluginID != pluginID {
			out = append(out, c)
		}
	}
	return out
}

func filterThemes(in []ThemeContribView, pluginID string) []ThemeContribView {
	out := in[:0]
	for _, t := range in {
		if t.PluginID != pluginID {
			out = append(out, t)
		}
	}
	return out
}

func filterSettings(in []SettingsContribView, pluginID string) []SettingsContribView {
	out := in[:0]
	for _, s := range in {
		if s.PluginID != pluginID {
			out = append(out, s)
		}
	}
	return out
}

func filterDevtools(in []DevToolsContribView, pluginID string) []DevToolsContribView {
	out := in[:0]
	for _, d := range in {
		if d.PluginID != pluginID {
			out = append(out, d)
		}
	}
	return out
}

func (r *Registry) RegisterManifest(m Manifest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.manifests[m.ID] = m
}

func (r *Registry) RegisterScheme(pluginID string, contrib URLSchemeContrib, handler SchemeHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	scheme := normalizeScheme(contrib.Scheme)
	r.schemes[scheme] = schemeEntry{pluginID: pluginID, handler: handler, contrib: contrib}
}

func (r *Registry) RegisterContributions(pluginID string, c Contributions) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range c.Panels {
		r.panels = append(r.panels, PanelContribView{PluginID: pluginID, PanelContrib: p})
	}
	for _, cmd := range c.Commands {
		r.commands = append(r.commands, CommandContribView{PluginID: pluginID, CommandContrib: cmd})
	}
	for _, th := range c.Themes {
		r.themes = append(r.themes, ThemeContribView{PluginID: pluginID, ThemeContrib: th})
	}
	for _, st := range c.Settings {
		r.settings = append(r.settings, SettingsContribView{PluginID: pluginID, SettingsContrib: st})
	}
	for _, dt := range c.DevTools {
		r.devtools = append(r.devtools, DevToolsContribView{PluginID: pluginID, DevToolsContrib: dt})
	}
}

func (r *Registry) RegisterBeforeFetch(pluginID string, hook BeforeFetchHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.beforeFetch = append(r.beforeFetch, fetchHookEntry{pluginID: pluginID, hook: hook})
}

func (r *Registry) RegisterAfterFetch(pluginID string, hook AfterFetchHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.afterFetch = append(r.afterFetch, afterFetchEntry{pluginID: pluginID, hook: hook})
}

func (r *Registry) BestRenderer(path string, body []byte, detected string) (Renderer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, renderer := range r.renderers {
		if renderer.Match(path, body, detected) {
			return renderer, true
		}
	}
	return nil, false
}

func (r *Registry) Render(path string, body []byte, nodeHash string, detected string) (Rendered, bool) {
	renderer, ok := r.BestRenderer(path, body, detected)
	if !ok {
		return Rendered{}, false
	}
	out, err := renderer.Render(path, body, nodeHash)
	if err != nil {
		return Rendered{Kind: detected, IsError: true, HTML: "", Raw: string(body)}, true
	}
	return out, true
}

func (r *Registry) HandleScheme(rawURL string) (SchemeResult, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	scheme := extractScheme(rawURL)
	if scheme == "" {
		return SchemeResult{}, false
	}
	entry, ok := r.schemes[scheme]
	if !ok {
		return SchemeResult{}, false
	}
	if entry.handler != nil {
		return entry.handler(rawURL)
	}
	return SchemeResult{
		DelegateFrontend: true,
		PluginID:         entry.pluginID,
		Handler:          entry.contrib.Handler,
		URL:              rawURL,
	}, true
}

func (r *Registry) RunBeforeFetch(ctx FetchContext) (FetchContext, bool, error) {
	r.mu.RLock()
	hooks := append([]fetchHookEntry(nil), r.beforeFetch...)
	r.mu.RUnlock()
	for _, entry := range hooks {
		res, err := entry.hook(ctx)
		if err != nil {
			return ctx, false, err
		}
		if res.Cancel {
			return ctx, true, nil
		}
		if res.Path != "" {
			ctx.Path = res.Path
		}
	}
	return ctx, false, nil
}

func (r *Registry) RunAfterFetch(ctx FetchContext, body []byte) ([]byte, error) {
	r.mu.RLock()
	hooks := append([]afterFetchEntry(nil), r.afterFetch...)
	r.mu.RUnlock()
	current := body
	for _, entry := range hooks {
		next, err := entry.hook(ctx, current)
		if err != nil {
			return current, err
		}
		current = next
	}
	return current, nil
}

func (r *Registry) Contributions() ContributionsView {
	r.mu.RLock()
	defer r.mu.RUnlock()
	view := ContributionsView{
		Panels:   append([]PanelContribView(nil), r.panels...),
		Commands: append([]CommandContribView(nil), r.commands...),
		Themes:   append([]ThemeContribView(nil), r.themes...),
		Settings: append([]SettingsContribView(nil), r.settings...),
		DevTools: append([]DevToolsContribView(nil), r.devtools...),
	}
	for _, renderer := range r.renderers {
		if renderer.PluginID() == "" {
			continue
		}
		if m, ok := r.manifests[renderer.PluginID()]; ok {
			view.Renderers = append(view.Renderers, contribRenderers(m)...)
		}
	}
	for scheme, entry := range r.schemes {
		if entry.pluginID == "" {
			continue
		}
		c := entry.contrib
		if c.Scheme == "" {
			c.Scheme = scheme
		}
		view.URLSchemes = append(view.URLSchemes, URLSchemeContribView{
			PluginID:         entry.pluginID,
			URLSchemeContrib: c,
		})
	}
	sort.Slice(view.URLSchemes, func(i, j int) bool {
		return view.URLSchemes[i].Scheme < view.URLSchemes[j].Scheme
	})
	return view
}

func contribRenderers(m Manifest) []RendererContribView {
	out := make([]RendererContribView, 0, len(m.Contributes.Renderers))
	for _, rc := range m.Contributes.Renderers {
		out = append(out, RendererContribView{PluginID: m.ID, RendererContrib: rc})
	}
	return out
}

func normalizeScheme(scheme string) string {
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	scheme = strings.TrimSuffix(scheme, ":")
	return scheme
}

func extractScheme(rawURL string) string {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		return ""
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "docs?") {
		return "docs"
	}
	if idx := strings.IndexByte(lower, ':'); idx > 0 {
		return lower[:idx]
	}
	switch lower {
	case "about", "editor", "license", "hello", "docs":
		return lower
	default:
		return ""
	}
}
