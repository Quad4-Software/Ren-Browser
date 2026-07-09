// SPDX-License-Identifier: MIT
package app

import "testing"

func TestPrepareForWakeWithoutStack(t *testing.T) {
	svc := &BrowserService{}
	res := svc.PrepareForWake()
	if res.DroppedLinks != 0 || res.ExpiredPaths != 0 {
		t.Fatalf("empty service should no-op, got %+v", res)
	}
}

func TestPrepareForWakeWithStack(t *testing.T) {
	svc := newTestBrowserService(t)
	res := svc.PrepareForWake()
	if res.DroppedLinks < 0 || res.ExpiredPaths < 0 {
		t.Fatalf("unexpected negative counts: %+v", res)
	}
}
