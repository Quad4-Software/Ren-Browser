//go:build android
// SPDX-License-Identifier: MIT

package fonts

func detectPlatformFonts() []string {
	return []string{
		"Roboto",
		"Noto Sans",
		"Droid Sans",
		"Droid Sans Mono",
	}
}
