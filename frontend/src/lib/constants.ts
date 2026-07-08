// SPDX-License-Identifier: MIT

// Shared values used by the frontend (and build scripts).
// Keep in sync with internal/constants/constants.go.

export const communityDirectoryUrl =
  "https://directory.rns.recipes/api/directory/submitted?search=&type=&status=online";

export const micronParserGoReleaseDownloadBase =
  "https://github.com/Quad4-Software/Micron-Parser-Go/releases/download";
export const micronParserGoReleaseTag = "v1.0.6";
export const micronParserGoWasmFilename = "micron-parser-go.wasm";
export const micronParserGoShasumsFilename = "SHASUMS256.txt";
export const micronParserGoMaxWasmBytes = 14 * 1024 * 1024;
export const micronParserGoGoWasmExecVersion = "go1.26.2";

export const serverAuthSessionCookieName = "renbrowser_session";
export const serverAuthLoginPath = "/api/auth/login";
export const serverAuthLogoutPath = "/api/auth/logout";
export const serverAuthStatusPath = "/api/auth/status";
