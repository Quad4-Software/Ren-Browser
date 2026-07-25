#!/usr/bin/env bash
# go mod tidy -e can drop modules that only appear in vendored Wails / Windows paths
# when build/android fails to load. Re-pin anything still listed in vendor/modules.txt.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

modules="${root}/vendor/modules.txt"
if [[ ! -f "${modules}" ]]; then
  exit 0
fi

declare -a pins=()
while IFS= read -r line; do
  if [[ "${line}" =~ ^#\ ([^[:space:]]+)\ ([^[:space:]]+)$ ]] && [[ "${line}" != \#\#* ]]; then
    mod="${BASH_REMATCH[1]}"
    ver="${BASH_REMATCH[2]}"
    case "${mod}" in
      github.com/jchv/go-winloader|github.com/wailsapp/wails/webview2)
        pins+=("${mod}@${ver}")
        ;;
    esac
  fi
done < "${modules}"

if ((${#pins[@]} == 0)); then
  exit 0
fi

GOFLAGS=-mod=mod go get "${pins[@]}"
