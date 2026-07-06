#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
browserservice="${root}/frontend/bindings/renbrowser/internal/app/browserservice.ts"
pluginhost="${root}/frontend/bindings/renbrowser/internal/app/pluginhost.ts"

if ! command -v wails3 >/dev/null 2>&1; then
  if [[ -f "$browserservice" && -f "$pluginhost" ]]; then
    echo "generate-bindings: wails3 not found; using committed bindings"
    exit 0
  fi
  echo "generate-bindings: wails3 not found and bindings are missing" >&2
  exit 127
fi

args=(wails3 generate bindings -clean=true -ts -i)
if [[ -n "${BUILD_FLAGS:-}" ]]; then
  args+=(-f "${BUILD_FLAGS}")
fi
if [[ "${OBFUSCATED:-}" == "true" ]]; then
  args+=(-obfuscated)
fi
exec "${args[@]}"
