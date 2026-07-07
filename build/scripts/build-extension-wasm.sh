#!/usr/bin/env bash
# Build micron-translator extension WASM for release assets.
# Exits 0 when TinyGo or the build is unavailable (CI may skip upload).
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
out_dir="${1:-${root}/dist/release}"
ext="${root}/extensions/micron-translator"
mkdir -p "$out_dir"

install_tinygo() {
  if command -v tinygo >/dev/null 2>&1; then
    tinygo version
    return 0
  fi
  if [ "$(uname -s)" != "Linux" ] || [ "$(uname -m)" != "x86_64" ]; then
    echo "build-extension-wasm: TinyGo not installed and auto-install is Linux amd64 only" >&2
    return 1
  fi
  local ver="${TINYGO_VERSION:-0.41.0}"
  local deb="tinygo_${ver}_amd64.deb"
  local url="https://github.com/tinygo-org/tinygo/releases/download/v${ver}/${deb}"
  echo "build-extension-wasm: installing TinyGo ${ver}..."
  curl -fsSL -o "/tmp/${deb}" "$url"
  sudo dpkg -i "/tmp/${deb}" >/dev/null 2>&1 || sudo apt-get install -f -y >/dev/null
  command -v tinygo >/dev/null 2>&1
}

if ! install_tinygo; then
  echo "build-extension-wasm: skipping (TinyGo unavailable)" >&2
  exit 0
fi

build_ok=0
if bash "${ext}/build-wasm.sh"; then
  build_ok=1
else
  echo "build-extension-wasm: build-wasm.sh failed; staging any wasm outputs" >&2
fi

staged=0
if [ -f "${ext}/renbrowser.micron-translator.wasm" ]; then
  cp "${ext}/renbrowser.micron-translator.wasm" "${out_dir}/renbrowser-micron-translator.wasm"
  staged=1
fi
if [ -f "${ext}/translator.wasm" ]; then
  cp "${ext}/translator.wasm" "${out_dir}/renbrowser-micron-translator-backend.wasm"
  staged=1
fi

if [ "$staged" -eq 0 ]; then
  echo "build-extension-wasm: skipping (no wasm outputs)" >&2
  exit 0
fi

if [ "$build_ok" -eq 0 ]; then
  echo "build-extension-wasm: staged partial wasm outputs" >&2
fi

echo "build-extension-wasm: staged extension wasm in ${out_dir}"
