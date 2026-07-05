//go:build android

package fonts

func detectPlatformFonts() []string {
	return []string{
		"Roboto",
		"Noto Sans",
		"Droid Sans",
		"Droid Sans Mono",
	}
}
