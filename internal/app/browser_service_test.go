// SPDX-License-Identifier: MIT
package app_test

import (
	"testing"
)

func TestHistoryNavigation(t *testing.T) {
	svc := newTestService(t)

	svc.Navigate("editor")
	svc.Navigate("config")

	state := svc.HistoryState()
	if !state.CanGoBack {
		t.Fatal("expected can go back")
	}
	if state.CanGoForward {
		t.Fatal("unexpected can go forward")
	}

	back := svc.GoBack()
	if back == "" {
		t.Fatal("expected back url")
	}
	state = svc.HistoryState()
	if !state.CanGoForward {
		t.Fatal("expected can go forward after back")
	}
}

func TestDeleteInterface(t *testing.T) {
	svc := newTestService(t)

	err := svc.DeleteInterface("non_existent")
	if err == nil {
		t.Fatal("expected error for non-existent interface")
	}
}
