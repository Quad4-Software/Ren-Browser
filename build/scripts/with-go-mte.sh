#!/usr/bin/env bash
# Run go from the mobile MTE toolchain (builds it on first use).
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
eval "$(bash "${root}/build/scripts/ensure-go-mte.sh" --print-env)"
exec go "$@"
