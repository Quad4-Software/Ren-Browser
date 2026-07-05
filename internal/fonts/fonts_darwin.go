//go:build darwin
// SPDX-License-Identifier: MIT

package fonts

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
