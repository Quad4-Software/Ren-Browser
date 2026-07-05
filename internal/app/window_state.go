// SPDX-License-Identifier: MIT
package app

import (
	"encoding/json"
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

const windowStateKey = "windowState"

type WindowState struct {
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	Maximized bool `json:"maximized"`
}

func defaultWindowState() WindowState {
	return WindowState{Width: 1280, Height: 800}
}

func (s *BrowserService) loadWindowState() (WindowState, error) {
	raw, err := s.store.GetSetting(windowStateKey)
	if err != nil || raw == "" {
		return defaultWindowState(), err
	}
	var state WindowState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return defaultWindowState(), err
	}
	if state.Width < 360 {
		state.Width = 1280
	}
	if state.Height < 480 {
		state.Height = 800
	}
	return state, nil
}

func (s *BrowserService) GetWindowState() WindowState {
	state, err := s.loadWindowState()
	if err != nil {
		return defaultWindowState()
	}
	return state
}

func (s *BrowserService) saveWindowState(state WindowState) error {
	if state.Width < 360 {
		state.Width = 360
	}
	if state.Height < 480 {
		state.Height = 480
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.store.SetSetting(windowStateKey, string(raw))
}

func (s *BrowserService) SaveWindowState(state WindowState) WindowState {
	_ = s.saveWindowState(state)
	return state
}

func (s *BrowserService) ResetWindowState() WindowState {
	_ = s.store.SetSetting(windowStateKey, "")
	return defaultWindowState()
}

func (s *BrowserService) CaptureWindowState() (WindowState, error) {
	if s.app == nil {
		return WindowState{}, errors.New("window state is only available in desktop mode")
	}
	window := s.app.Window.Current()
	if window == nil {
		return WindowState{}, errors.New("window unavailable")
	}
	width, height := window.Size()
	x, y := window.Position()
	state := WindowState{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Maximized: window.IsMaximised(),
	}
	_ = s.saveWindowState(state)
	return state, nil
}

func (s *BrowserService) ToggleFullscreen() error {
	if s.app == nil {
		return errors.New("fullscreen is only available in desktop mode")
	}
	window := s.app.Window.Current()
	if window == nil {
		return errors.New("window unavailable")
	}
	if window.IsFullscreen() {
		window.UnFullscreen()
	} else {
		window.Fullscreen()
	}
	return nil
}

func (s *BrowserService) InitialWindowOptions(frameless bool, reset bool) application.WebviewWindowOptions {
	opts := application.WebviewWindowOptions{
		Title:            "Ren Browser",
		Width:            1280,
		Height:           800,
		MinWidth:         360,
		MinHeight:        480,
		Frameless:        frameless,
		BackgroundColour: application.NewRGB(9, 9, 11),
		URL:              "/",
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
	if reset {
		return opts
	}
	state := s.GetWindowState()
	if state.Width > 0 && state.Height > 0 {
		opts.Width = state.Width
		opts.Height = state.Height
	}
	if state.X != 0 || state.Y != 0 {
		opts.X = state.X
		opts.Y = state.Y
		opts.InitialPosition = application.WindowXY
	}
	return opts
}

func (s *BrowserService) AttachWindowPersistence(window application.Window) {
	if window == nil || s.resetWindow {
		return
	}
	window.OnWindowEvent(events.Common.WindowDidMove, func(*application.WindowEvent) {
		_, _ = s.CaptureWindowState()
	})
	window.OnWindowEvent(events.Common.WindowDidResize, func(*application.WindowEvent) {
		_, _ = s.CaptureWindowState()
	})
}
