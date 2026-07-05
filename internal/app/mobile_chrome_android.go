//go:build android

package app

import (
	"encoding/json"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func (s *BrowserService) SyncMobileChrome(chromeBg string, lightStatusIcons bool) {
	style := "light"
	if !lightStatusIcons {
		style = "dark"
	}
	payload, err := json.Marshal(map[string]any{
		"style":         style,
		"statusBar":     chromeBg,
		"navigationBar": chromeBg,
		"webView":       chromeBg,
	})
	if err != nil {
		return
	}
	application.Android.SetStatusBar(string(payload))
}
