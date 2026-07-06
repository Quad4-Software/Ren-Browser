#!/usr/bin/env bash
# Re-apply Ren Browser patches to vendored Wails after `go mod vendor`.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
patch_dir="${root}/build/patches/wails"
vendor="${root}/vendor/github.com/wailsapp/wails/v3"
extract="${vendor}/pkg/updater/extract.go"

if [ ! -d "${vendor}" ]; then
  exit 0
fi

install -D -m 0644 "${patch_dir}/internal/operatingsystem/os_bsd.go" "${vendor}/internal/operatingsystem/os_bsd.go"
install -D -m 0644 "${patch_dir}/internal/assetserver/assetserver_bsd.go" "${vendor}/internal/assetserver/assetserver_bsd.go"
install -D -m 0644 "${patch_dir}/internal/fileexplorer/fileexplorer_bsd.go" "${vendor}/internal/fileexplorer/fileexplorer_bsd.go"

if [ -f "${extract}" ]; then
  sed -i \
    -e 's/exceeds %d bytes", maxArchiveTotalSize)/exceeds %d bytes", int64(maxArchiveTotalSize))/g' \
    "${extract}"
fi
