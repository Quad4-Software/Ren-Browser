// SPDX-License-Identifier: MIT
package app_test

import (
	"testing"

	"renbrowser/internal/store"
)

func TestServiceShutdownClosesStoreForReuseChecks(t *testing.T) {
	svc := newTestService(t)
	if err := svc.ServiceShutdown(); err != nil {
		t.Fatal(err)
	}
	health := svc.GetStoreHealth()
	if health.OK {
		t.Fatalf("health=%#v; store should be unavailable after shutdown", health)
	}
}

func TestSaveTabsAfterShutdownDoesNotPanic(t *testing.T) {
	svc := newTestService(t)
	if err := svc.ServiceShutdown(); err != nil {
		t.Fatal(err)
	}
	_ = svc.SaveTabs([]store.TabSnapshot{{
		ID:    "tab-1",
		Title: "Test",
		URL:   "editor",
	}})
}
