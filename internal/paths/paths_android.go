//go:build android

// SPDX-License-Identifier: MIT

package paths

import (
	"strings"

	_ "unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:linkname androidBridgeString github.com/wailsapp/wails/v3/pkg/application.androidBridgeString
func androidBridgeString(method string) (string, bool)

func InitAndroid() {
	if root := application.Mobile.StoragePath(); root != "" {
		SetDataRoot(root)
	}
}

func UserDownloadDir() string {
	s, ok := androidBridgeString("getDownloadPath")
	if ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	return "/storage/emulated/0/Download"
}
