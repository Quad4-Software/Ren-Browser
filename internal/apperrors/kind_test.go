// SPDX-License-Identifier: MIT
package apperrors

import "testing"

func TestClassifyFetchConnectionFailed(t *testing.T) {
	kind, _ := ClassifyFetch("no path to node: path discovery timed out", nil)
	if kind != KindConnectionFailed {
		t.Fatalf("kind=%q want connection_failed", kind)
	}
}

func TestClassifyFetchNotFound(t *testing.T) {
	kind, _ := ClassifyFetch("empty response", nil)
	if kind != KindNotFound {
		t.Fatalf("kind=%q want not_found", kind)
	}
}

func TestClassifyFetchConnectionLost(t *testing.T) {
	kind, _ := ClassifyFetch("context canceled", nil)
	if kind != KindConnectionLost {
		t.Fatalf("kind=%q want connection_lost", kind)
	}
}

func TestClassifyFetchInternal(t *testing.T) {
	kind, detail := ClassifyFetch("unexpected server fault", nil)
	if kind != KindInternal {
		t.Fatalf("kind=%q want internal", kind)
	}
	if detail != "unexpected server fault" {
		t.Fatalf("detail=%q", detail)
	}
}
