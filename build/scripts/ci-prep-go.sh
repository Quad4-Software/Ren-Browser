#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"

bash "${root}/build/scripts/patch-reticulum-txbytes.sh"

dist="${root}/frontend/dist"
mkdir -p "${dist}"
if [[ ! -f "${dist}/index.html" ]]; then
  printf '%s\n' '<!DOCTYPE html><html><head></head><body></body></html>' > "${dist}/index.html"
fi
