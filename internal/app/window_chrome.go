package app

import (
	"errors"
	"strings"
)

type WindowChrome struct {
	NativeTitlebar bool `json:"nativeTitlebar"`
}

func (s *BrowserService) GetWindowChrome() WindowChrome {
	prefs := s.GetBrowserPrefs()
	return WindowChrome{NativeTitlebar: prefs.NativeTitlebar}
}

func (s *BrowserService) SetNativeTitlebar(enabled bool) (BrowserPrefs, error) {
	if s.app == nil {
		return BrowserPrefs{}, errors.New("native titlebar is only available in desktop mode")
	}
	prefs := s.GetBrowserPrefs()
	prefs.NativeTitlebar = enabled
	merged := s.SetBrowserPrefs(prefs)
	window := s.app.Window.Current()
	if window == nil {
		return merged, errors.New("window unavailable")
	}
	window.SetFrameless(!enabled)
	if s.app != nil {
		s.app.Event.Emit("window:chrome", WindowChrome{NativeTitlebar: enabled})
	}
	return merged, nil
}

func nodeHashFromURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "about:" || trimmed == "editor:" {
		return "", errors.New("not a mesh page")
	}
	hash := trimmed
	if idx := strings.Index(trimmed, ":/"); idx >= 0 {
		hash = trimmed[:idx]
	}
	hash = strings.ToLower(strings.TrimSpace(hash))
	if len(hash) != 32 {
		return "", errors.New("invalid node hash")
	}
	return hash, nil
}

func (s *BrowserService) IdentifyToNode(rawURL string) error {
	hash, err := nodeHashFromURL(rawURL)
	if err != nil {
		return err
	}
	if err := s.stack.Identify(hash); err != nil {
		return err
	}
	s.log("info", "identified to node", hash)
	return nil
}
