//go:build android

// SPDX-License-Identifier: MIT

package paths

import (
	"strings"

	_ "unsafe"
)

//go:linkname androidBridgeString github.com/wailsapp/wails/v3/pkg/application.androidBridgeString
func androidBridgeString(method string) (string, bool)

func InitAndroid() {
	root, ok := androidBridgeString("getStoragePath")
	if !ok {
		return
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return
	}
	legacy, _ := androidBridgeString("getLegacyStoragePath")
	migrateAndroidStorage(strings.TrimSpace(legacy), root)
	SetDataRoot(root)
}

func UserDownloadDir() string {
	s, ok := androidBridgeString("getDownloadPath")
	if ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	return "/storage/emulated/0/Download"
}
