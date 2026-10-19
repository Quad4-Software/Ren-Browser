// SPDX-License-Identifier: MIT

// Package constants holds shared values used by the Go backend.
// Keep in sync with frontend/src/lib/constants.ts.
package constants

const (
	MicronParserGoReleaseDownloadBase = "https://github.com/Quad4-Software/Micron-Parser-Go/releases/download"
	MicronParserGoReleaseTag          = "v1.3.0"
	MicronParserGoWasmFilename        = "micron-parser-go.wasm"
	MicronParserGoShasumsFilename     = "SHASUMS256.txt"
	MicronParserGoMaxWasmBytes        = 14 * 1024 * 1024
	MicronParserGoGoWasmExecVersion   = "go1.27.1"

	ServerAuthSessionCookieName = "renbrowser_session"
	ServerAuthLoginPath         = "/api/auth/login"
	ServerAuthLogoutPath        = "/api/auth/logout"
	ServerAuthStatusPath        = "/api/auth/status"
)
