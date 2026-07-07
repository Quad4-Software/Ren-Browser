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

func TestDisableNestedWebKitSandboxInFlatpak(t *testing.T) {
	t.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "")
	t.Setenv("FLATPAK_ID", "io.quad4.renbrowser")
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
