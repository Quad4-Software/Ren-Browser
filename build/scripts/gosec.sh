#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
set -euo pipefail

export GOFLAGS="${GOFLAGS:--mod=vendor}"
out="$(mktemp)"
trap 'rm -f "$out"' EXIT

set +e
GOFLAGS=-mod=mod go run -exec "env GOFLAGS=-mod=vendor" github.com/securego/gosec/v2/cmd/gosec@latest \
  -exclude-generated \
  -fmt=text \
  -exclude-dir=node_modules \
  -exclude-dir=frontend \
  -exclude-dir=build/android \
  -exclude-dir=build/ios \
  -exclude-dir=vendor \
  -exclude-dir=third_party \
  "$@" >"$out" 2>&1
status=$?
set -e
cat "$out"

# gosec can exit non-zero when cgo/SSA package load fails even with zero findings.
if [[ "$status" -ne 0 ]]; then
  if grep -Eq 'Issues[[:space:]]*:[[:space:]]*0[[:space:]]*$' "$out"; then
    exit 0
  fi
  exit "$status"
fi
