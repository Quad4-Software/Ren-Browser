#!/usr/bin/env bash
# Run go from the mobile MTE toolchain (builds it on first use).
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
# Only eval export lines. ensure-go-mte may print build status on stderr, but
# harden against any accidental stdout noise (e.g. "Building ...").
env_lines="$(bash "${root}/build/scripts/ensure-go-mte.sh" --print-env | grep '^export ' || true)"
if [ -z "${env_lines}" ]; then
  echo "ensure-go-mte.sh --print-env produced no export lines" >&2
  exit 1
fi
eval "${env_lines}"
exec go "$@"
