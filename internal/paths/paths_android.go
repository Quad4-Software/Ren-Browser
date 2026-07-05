//go:build android

package paths

import "github.com/wailsapp/wails/v3/pkg/application"

func InitAndroid() {
	if root := application.Mobile.StoragePath(); root != "" {
		SetDataRoot(root)
	}
}
