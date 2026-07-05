package app_test

import (
	"testing"

	"renbrowser/internal/app"
	"renbrowser/internal/rns"
)

func TestHistoryNavigation(t *testing.T) {
	stack, err := rns.NewStack("")
	if err != nil {
		t.Fatal(err)
	}
	svc, err := app.NewBrowserService(stack, nil)
	if err != nil {
		t.Fatal(err)
	}

	svc.Navigate("abb3ebcd03cb2388a838e70c001291f9:/page/a.mu")
	svc.Navigate("abb3ebcd03cb2388a838e70c001291f9:/page/b.mu")

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
