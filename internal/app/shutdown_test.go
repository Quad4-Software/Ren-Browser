// SPDX-License-Identifier: MIT
package app_test

import (
	"testing"
)

func TestServiceShutdownClosesStore(t *testing.T) {
	svc := newTestService(t)
	if err := svc.ServiceShutdown(); err != nil {
		t.Fatalf("ServiceShutdown: %v", err)
	}
	if err := svc.ServiceShutdown(); err != nil {
		t.Fatalf("second ServiceShutdown: %v", err)
	}
}

func TestNavigateDuringShutdown(t *testing.T) {
	svc := newTestService(t)
	if err := svc.ServiceShutdown(); err != nil {
		t.Fatal(err)
	}
	resp := svc.Navigate("editor")
	if resp.Error == "" {
		t.Fatal("expected shutdown error")
	}
}
