//go:build linux && !android

package webkit

import "os"

// ApplyLinuxDefaults sets WebKitGTK environment defaults on Linux.
// Existing user values are never overwritten.
func ApplyLinuxDefaults() {
	if os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER") == "" {
		_ = os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
	disableNestedWebKitSandboxIfNeeded()
}

func disableNestedWebKitSandboxIfNeeded() {
	if os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS") != "" {
		return
	}
	if runningInFlatpak() {
		_ = os.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "1")
	}
}

func runningInFlatpak() bool {
	if os.Getenv("FLATPAK_ID") != "" {
		return true
	}
	_, err := os.Stat("/.flatpak-info")
	return err == nil
}

func init() {
	ApplyLinuxDefaults()
}
