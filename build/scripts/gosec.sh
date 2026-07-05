#!/usr/bin/env bash
set -euo pipefail

export GOFLAGS="${GOFLAGS:--mod=vendor}"
GOFLAGS=-mod=mod go run -exec "env GOFLAGS=-mod=vendor" github.com/securego/gosec/v2/cmd/gosec@latest \
  -exclude-generated \
  -fmt=text \
  -exclude-dir=node_modules \
  -exclude-dir=frontend \
  -exclude-dir=build/android \
  -exclude-dir=build/ios \
  -exclude-dir=vendor \
  -exclude-dir=third_party \
  "$@"
