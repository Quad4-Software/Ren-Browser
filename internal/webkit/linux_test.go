//go:build linux && !android

package webkit

import (
	"os"
	"testing"
)

func TestApplyLinuxDefaults(t *testing.T) {
	t.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "")
	ApplyLinuxDefaults()
	if got := os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER"); got != "1" {
		t.Fatalf("WEBKIT_DISABLE_DMABUF_RENDERER = %q, want 1", got)
	}
}

func TestApplyLinuxDefaultsRespectsExistingValue(t *testing.T) {
	t.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "0")
	ApplyLinuxDefaults()
	if got := os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER"); got != "0" {
		t.Fatalf("WEBKIT_DISABLE_DMABUF_RENDERER = %q, want 0", got)
	}
}
