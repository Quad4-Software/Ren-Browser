// SPDX-License-Identifier: MIT
package app

import (
	"renbrowser/internal/content"
	"renbrowser/internal/micron"
	"renbrowser/internal/nomadnet"
	"renbrowser/internal/plugins"
	"renbrowser/internal/plugins/builtin"
	"renbrowser/internal/rns"
)

func (s *BrowserService) SetPluginManager(manager *plugins.Manager) {
	if manager == nil {
		return
	}
	s.mu.Lock()
	s.plugins = manager
	s.mu.Unlock()
	plugins.SanitizeHTML = content.SanitizeHTML
	content.SetRendererRegistry(manager.Registry())
	builtin.RegisterRenderers(manager.Registry())
	micron.GetForceMonospace = func() bool {
		return s.GetBrowserPrefs().MicronPreserveLayout
	}
	builtin.RegisterSchemes(manager.Registry(), builtin.SchemeDeps{
		AboutHTML: func() string {
			return content.RenderAbout(s.aboutContentInfo())
		},
		EditorRaw: func() string {
			return content.DefaultEditorTemplate()
		},
		ConfigRaw: func() string {
			path := s.ConfigPath()
			if path == "" {
				return ""
			}
			text, err := rns.ReadConfigText(path)
			if err != nil {
				return ""
			}
			return text
		},
		LicenseHTML: func() string {
			return content.RenderLicense()
		},
		DocsPage: func(rawURL string) (plugins.SchemeResult, bool) {
			prefs := s.GetBrowserPrefs()
			rendered, ok := content.RenderDocs(content.DocsRenderInput{
				RawURL:    rawURL,
				SavedLang: prefs.DocsLanguage,
				SaveLang: func(lang string) {
					next := s.GetBrowserPrefs()
					next.DocsLanguage = lang
					s.SetBrowserPrefs(next)
				},
			})
			if !ok {
				return plugins.SchemeResult{}, false
			}
			return plugins.SchemeResult{
				URL:          rendered.URL,
				Path:         "/docs",
				ContentType:  "docs",
				HTML:         rendered.HTML,
				Raw:          rendered.Raw,
				HistoryTitle: rendered.HistoryTitle,
			}, true
		},
		SettingsPage: func() (plugins.SchemeResult, bool) {
			return plugins.SchemeResult{
				URL:          "settings:",
				Path:         "/settings",
				ContentType:  "settings",
				HistoryTitle: "Settings",
			}, true
		},
	})
}

func (s *BrowserService) runFetchHooks(nodeHash, path string, req nomadnet.RequestData) (string, nomadnet.RequestData, bool) {
	s.mu.RLock()
	manager := s.plugins
	s.mu.RUnlock()
	if manager == nil {
		return path, req, false
	}
	ctx := plugins.FetchContext{
		NodeHash: nodeHash,
		Path:     path,
		Request:  requestToMap(req),
	}
	updated, cancel, err := manager.Registry().RunBeforeFetch(ctx)
	if err != nil || cancel {
		return path, req, cancel
	}
	if updated.Path != "" {
		path = updated.Path
	}
	return path, req, false
}

func (s *BrowserService) runAfterFetchHooks(nodeHash, path string, req nomadnet.RequestData, body []byte) []byte {
	s.mu.RLock()
	manager := s.plugins
	s.mu.RUnlock()
	if manager == nil {
		return body
	}
	ctx := plugins.FetchContext{
		NodeHash: nodeHash,
		Path:     path,
		Request:  requestToMap(req),
	}
	out, err := manager.Registry().RunAfterFetch(ctx, body)
	if err != nil {
		return body
	}
	return out
}

func requestToMap(req nomadnet.RequestData) map[string]string {
	out := make(map[string]string, len(req.Vars)+len(req.Fields))
	for k, v := range req.Vars {
		out["var_"+k] = v
	}
	for k, v := range req.Fields {
		out["field_"+k] = v
	}
	return out
}
