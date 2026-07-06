//go:build !android

// SPDX-License-Identifier: MIT

package app

import (
	"os/exec"
	"runtime"
)

func openPathDesktop(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start() // #nosec G204 -- path validated by validateDownloadPath before OpenDownloadPath
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start() // #nosec G204 -- path validated by validateDownloadPath before OpenDownloadPath
	default:
		return exec.Command("xdg-open", path).Start() // #nosec G204 -- path validated by validateDownloadPath before OpenDownloadPath
	}
}
