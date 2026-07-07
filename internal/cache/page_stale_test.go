// SPDX-License-Identifier: MIT
package cache

import (
	"testing"
	"time"
)

func TestEntryIsStale(t *testing.T) {
	entry := Entry{StoredAt: time.Now().Add(-2 * time.Hour)}
	if !entry.IsStale(time.Hour) {
		t.Fatal("expected entry older than max age to be stale")
	}
	if entry.IsStale(0) {
		t.Fatal("zero max age should never be stale")
	}
	fresh := Entry{StoredAt: time.Now()}
	if fresh.IsStale(time.Hour) {
		t.Fatal("fresh entry should not be stale")
	}
}

func TestDefaultPageCacheMaxAge(t *testing.T) {
	if DefaultPageCacheMaxAge <= 0 {
		t.Fatalf("default max age = %v", DefaultPageCacheMaxAge)
	}
}
