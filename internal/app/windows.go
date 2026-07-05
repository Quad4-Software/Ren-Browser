package app

import (
	"errors"
	"net/url"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func (s *BrowserService) OpenNewWindow(rawURL string) error {
	if s.app == nil {
		return errors.New("new window is only available in desktop mode")
	}
	prefs := s.GetBrowserPrefs()
	frameless := !prefs.NativeTitlebar

	target := "/?window=secondary"
	if rawURL != "" {
		target = "/?window=secondary&open=" + url.QueryEscape(rawURL)
	}

	opts := application.WebviewWindowOptions{
		Title:            "Ren Browser",
		Width:            1280,
		Height:           800,
		MinWidth:         360,
		MinHeight:        480,
		Frameless:        frameless,
		BackgroundColour: application.NewRGB(9, 9, 11),
		URL:              target,
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBar{
				Hide:            frameless,
				FullSizeContent: frameless,
			},
		},
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: frameless,
		},
	}

	s.app.Window.NewWithOptions(opts).Show()
	return nil
}
