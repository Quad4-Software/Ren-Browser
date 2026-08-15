#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

# Ignore a stale vendor/ while resolving modules so go mod tidy bumps
# (e.g. inlined WebView2 loader DLLs) can be synced without chicken-and-egg failures.
export GOFLAGS="${GOFLAGS:--mod=mod}"

bash "${root}/build/scripts/fetch-reticulum-go.sh"

gomod="$(go env GOMODCACHE)"
src="${gomod}/github.com/wailsapp/go-webview2@v1.0.23/webviewloader"
wails_ver="$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)"
wails_wv2="${gomod}/github.com/wailsapp/wails/v3@${wails_ver}/internal/webview2/webviewloader"

go mod download github.com/wailsapp/go-webview2@v1.0.23

copy_webview2_dlls() {
  local dest=$1
  chmod -R u+w "${dest}" "$(dirname "${dest}")" 2>/dev/null || true
  mkdir -p "${dest}"
  chmod -R u+w "${dest}" 2>/dev/null || true
  for arch in arm64 x64 x86; do
    mkdir -p "${dest}/${arch}"
    cp -f "${src}/${arch}/WebView2Loader.dll" "${dest}/${arch}/" 2>/dev/null || true
  done
}

copy_webview2_dlls "${wails_wv2}"

if wv2_ver="$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/webview2 2>/dev/null)" && [ -n "${wv2_ver}" ]; then
  copy_webview2_dlls "${gomod}/github.com/wailsapp/wails/webview2@${wv2_ver}/webviewloader"
fi

go mod vendor
bash "${root}/build/scripts/patch-wails-vendor.sh"
bash "${root}/build/scripts/verify-wails-vendor-patches.sh"
