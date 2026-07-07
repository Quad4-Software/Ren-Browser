//go:build linux && !android

package webkit

import "os"

// ApplyLinuxDefaults sets WebKitGTK environment defaults on Linux.
// Existing user values are never overwritten.
func ApplyLinuxDefaults() {
	if os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER") == "" {
		_ = os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1")
	}
}

func init() {
	ApplyLinuxDefaults()
}
