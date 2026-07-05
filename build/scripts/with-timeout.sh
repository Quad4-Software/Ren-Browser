#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  printf 'usage: with-timeout.sh DURATION COMMAND...\n' >&2
  exit 2
fi

duration=$1
shift

if command -v timeout >/dev/null 2>&1; then
  exec timeout --foreground "$duration" "$@"
fi

exec "$@"
