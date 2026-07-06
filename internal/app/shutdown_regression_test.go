// SPDX-License-Identifier: MIT
package app_test

import (
	"context"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"

	"renbrowser/internal/store"
)

func TestServiceShutdownClosesStoreForReuseChecks(t *testing.T) {
	svc := newTestService(t)
	if err := svc.ServiceShutdown(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatal(err)
	}
	health := svc.GetStoreHealth()
	if health.OK {
		t.Fatalf("health=%#v; store should be unavailable after shutdown", health)
	}
}

func TestSaveTabsAfterShutdownDoesNotPanic(t *testing.T) {
	svc := newTestService(t)
	if err := svc.ServiceShutdown(context.Background(), application.ServiceOptions{}); err != nil {
		t.Fatal(err)
	}
	_ = svc.SaveTabs([]store.TabSnapshot{{
		ID:    "tab-1",
		Title: "Test",
		URL:   "editor",
	}})
}
