#!/usr/bin/env bash
set -euo pipefail

mode="${1:-check}"
root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

unformatted="$(find . -name '*.go' \
  -not -path './vendor/*' \
  -not -path './frontend/*' \
  -not -path './node_modules/*' \
  -not -path './third_party/*' \
  -print0 | xargs -0 gofmt -l)"

case "${mode}" in
  write)
    find . -name '*.go' \
      -not -path './vendor/*' \
      -not -path './frontend/*' \
      -not -path './node_modules/*' \
      -not -path './third_party/*' \
      -print0 | xargs -0 gofmt -w
    ;;
  check)
    if [[ -n "${unformatted}" ]]; then
      echo "Unformatted Go files:"
      echo "${unformatted}"
      exit 1
    fi
    ;;
  *)
    echo "gofmt-project.sh: unknown mode ${mode} (use check or write)" >&2
    exit 2
    ;;
esac
