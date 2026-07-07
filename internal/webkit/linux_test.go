//go:build linux && !android

package webkit

import (
	"os"
	"strings"
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

func TestDisableNestedWebKitSandboxInFlatpak(t *testing.T) {
	t.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "")
	t.Setenv("FLATPAK_ID", "io.quad4.renbrowser")
	disableNestedWebKitSandboxIfNeeded()
	if got := os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS"); got != "" {
		t.Fatalf("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS = %q, want unset in Flatpak", got)
	}
}

func TestDisableNestedWebKitSandboxInAppImage(t *testing.T) {
	t.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "")
	t.Setenv("APPIMAGE", "/tmp/test.AppImage")
	disableNestedWebKitSandboxIfNeeded()
	if got := os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS"); got != "1" {
		t.Fatalf("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS = %q, want 1", got)
	}
}

func TestDisableNestedWebKitSandboxRespectsExistingValue(t *testing.T) {
	t.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "0")
	t.Setenv("FLATPAK_ID", "io.quad4.renbrowser")
	disableNestedWebKitSandboxIfNeeded()
	if got := os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS"); got != "0" {
		t.Fatalf("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS = %q, want 0", got)
	}
}

func TestPreloadHostLibWaylandClientInAppImage(t *testing.T) {
	t.Setenv("APPIMAGE", "/tmp/test.AppImage")
	t.Setenv("LD_PRELOAD", "")
	preloadHostLibWaylandClient()
	got := os.Getenv("LD_PRELOAD")
	if got == "" {
		t.Skip("host libwayland-client.so.0 not found on this system")
	}
	if !strings.Contains(got, "libwayland-client.so") {
		t.Fatalf("LD_PRELOAD = %q, want host libwayland-client", got)
	}
}

func TestPreloadHostLibWaylandClientRespectsExisting(t *testing.T) {
	t.Setenv("APPIMAGE", "/tmp/test.AppImage")
	t.Setenv("LD_PRELOAD", "/usr/lib/libfoo.so")
	preloadHostLibWaylandClient()
	got := os.Getenv("LD_PRELOAD")
	if got == "/usr/lib/libfoo.so" {
		t.Skip("host libwayland-client.so.0 not found on this system")
	}
	if !strings.HasPrefix(got, "/") || !strings.Contains(got, "libwayland-client.so") {
		t.Fatalf("LD_PRELOAD = %q, want host libwayland-client prepended", got)
	}
	if !strings.Contains(got, "/usr/lib/libfoo.so") {
		t.Fatalf("LD_PRELOAD = %q, want to preserve existing preload", got)
	}
}

func TestApplyLinuxDefaultsAppImageWorkaround(t *testing.T) {
	t.Setenv("APPIMAGE", "/tmp/test.AppImage")
	t.Setenv("LD_PRELOAD", "")
	ApplyLinuxDefaults()
	if got := os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER"); got != "1" {
		t.Fatalf("WEBKIT_DISABLE_DMABUF_RENDERER = %q, want 1", got)
	}
}
