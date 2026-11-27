//go:build darwin

// SPDX-License-Identifier: MIT

package fonts

import "os"

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func detectPlatformFonts() []string {
	home := userHome()
	dirs := []string{
		"/System/Library/Fonts",
		"/Library/Fonts",
	}
	if home != "" {
		dirs = append(dirs, home+"/Library/Fonts")
	}
	return scanFontDirs(dirs)
}
