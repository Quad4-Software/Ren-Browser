#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")" && pwd)"
wasm_src="${root}/wasm"
out="${root}/translator.wasm"

if ! command -v tinygo >/dev/null 2>&1; then
  echo "build-wasm: tinygo not found; run build/scripts/build-extension-wasm.sh or install TinyGo" >&2
  if [[ -f "$out" ]]; then
    exit 0
  fi
  echo "build-wasm: translator.wasm missing and tinygo unavailable" >&2
  exit 127
fi

tinygo build -tags wasm -o "$out" -target=wasm-unknown -scheduler=none -opt=2 "${wasm_src}"
echo "build-wasm: wrote ${out}"

bundle_out="${root}/renbrowser.micron-translator.wasm"
if command -v go >/dev/null 2>&1; then
  (cd "${root}/bundle" && go run . -root ".." -wasm "../translator.wasm" -out "../renbrowser.micron-translator.wasm")
else
  echo "build-wasm: go not found; skipping bundled ${bundle_out}" >&2
fi
