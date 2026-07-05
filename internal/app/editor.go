package app

import (
	"strings"

	"renbrowser/internal/content"
)

func isEditorURL(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "editor", "editor:":
		return true
	default:
		return false
	}
}

func (s *BrowserService) editorPage(pushHistory bool) PageResponse {
	raw := content.DefaultEditorTemplate
	resp := PageResponse{
		URL:         "editor:",
		Path:        "/page/editor.mu",
		ContentType: "editor",
		Raw:         raw,
	}
	if pushHistory {
		s.pushHistory("editor:")
		_ = s.store.AddHistory("editor:", "Micron Editor", "")
	}
	s.setLastPage(resp)
	if s.app != nil {
		s.app.Event.Emit("page:loaded", resp)
	}
	return resp
}
