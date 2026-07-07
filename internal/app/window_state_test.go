// SPDX-License-Identifier: MIT
package app

import (
	"testing"
)

func TestSaveAndLoadWindowState(t *testing.T) {
	svc := newTestBrowserService(t)
	state := WindowState{X: 120, Y: 80, Width: 1440, Height: 900, Maximized: true}
	if err := svc.saveWindowState(state); err != nil {
		t.Fatalf("saveWindowState: %v", err)
	}
	loaded, err := svc.loadWindowState()
	if err != nil {
		t.Fatalf("loadWindowState: %v", err)
	}
	if loaded != state {
		t.Fatalf("loaded=%#v want=%#v", loaded, state)
	}
}

func TestSaveWindowStateClampsMinimumSize(t *testing.T) {
	svc := newTestBrowserService(t)
	if err := svc.saveWindowState(WindowState{Width: 100, Height: 200}); err != nil {
		t.Fatalf("saveWindowState: %v", err)
	}
	loaded, err := svc.loadWindowState()
	if err != nil {
		t.Fatalf("loadWindowState: %v", err)
	}
	if loaded.Width != 360 || loaded.Height != 480 {
		t.Fatalf("loaded=%#v", loaded)
	}
}
