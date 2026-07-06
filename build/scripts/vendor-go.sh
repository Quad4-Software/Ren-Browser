#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

bash "${root}/build/scripts/fetch-reticulum-go.sh"

gomod="$(go env GOMODCACHE)"
wv2_ver="$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/webview2)"
wv2="${gomod}/github.com/wailsapp/wails/webview2@${wv2_ver}/webviewloader"
src="${gomod}/github.com/wailsapp/go-webview2@v1.0.23/webviewloader"

go mod download github.com/wailsapp/go-webview2@v1.0.23
chmod -R u+w "$(dirname "${wv2}")" 2>/dev/null || true
for arch in arm64 x64 x86; do
  mkdir -p "${wv2}/${arch}"
  cp "${src}/${arch}/WebView2Loader.dll" "${wv2}/${arch}/"
done

go mod vendor
bash "${root}/build/scripts/patch-wails-vendor.sh"
