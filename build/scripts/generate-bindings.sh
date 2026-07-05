#!/usr/bin/env bash
set -euo pipefail

args=(wails3 generate bindings -clean=true -ts -i)
if [[ -n "${BUILD_FLAGS:-}" ]]; then
  args+=(-f "${BUILD_FLAGS}")
fi
if [[ "${OBFUSCATED:-}" == "true" ]]; then
  args+=(-obfuscated)
fi
exec "${args[@]}"
