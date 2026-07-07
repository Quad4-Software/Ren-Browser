//go:build linux && !android

package webkit

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ApplyLinuxDefaults sets WebKitGTK environment defaults on Linux.
// Existing user values are never overwritten.
func ApplyLinuxDefaults() {
	if os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER") == "" {
		_ = os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
	applyAppImageGraphicsWorkaround()
	disableNestedWebKitSandboxIfNeeded()
}

func applyAppImageGraphicsWorkaround() {
	if !runningInAppImage() {
		return
	}
	preloadHostLibWaylandClient()
}

func preloadHostLibWaylandClient() {
	lib := hostLibWaylandClient()
	if lib == "" {
		return
	}
	existing := os.Getenv("LD_PRELOAD")
	if strings.Contains(existing, lib) {
		return
	}
	if existing == "" {
		_ = os.Setenv("LD_PRELOAD", lib)
		return
	}
	_ = os.Setenv("LD_PRELOAD", lib+":"+existing)
}

func hostLibWaylandClient() string {
	for _, dir := range hostLibDirs() {
		for _, name := range []string{"libwayland-client.so.0", "libwayland-client.so"} {
			path := filepath.Join(dir, name)
			if st, err := os.Stat(path); err == nil && !st.IsDir() {
				if abs, err := filepath.Abs(path); err == nil {
					return abs
				}
				return path
			}
		}
	}
	return ""
}

func hostLibDirs() []string {
	switch runtime.GOARCH {
	case "amd64":
		return []string{"/usr/lib/x86_64-linux-gnu", "/usr/lib64", "/usr/lib"}
	case "arm64":
		return []string{"/usr/lib/aarch64-linux-gnu", "/usr/lib64", "/usr/lib"}
	default:
		return []string{"/usr/lib64", "/usr/lib"}
	}
}

func disableNestedWebKitSandboxIfNeeded() {
	if os.Getenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS") != "" {
		return
	}
	// AppImages run outside Flatpak's portal-based WebKit launcher and cannot
	// mount transient AppImage paths into bwrap. Flatpak must keep WebKit's
	// sandbox enabled so WebKitNetworkProcess/WebKitWebProcess spawn via portal.
	if runningInAppImage() {
		_ = os.Setenv("WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS", "1")
	}
}

func runningInAppImage() bool {
	return strings.TrimSpace(os.Getenv("APPIMAGE")) != ""
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
