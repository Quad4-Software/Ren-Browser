#!/usr/bin/env bash
set -euo pipefail

mode="${1:-check}"
root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

mapfile -t files < <(
  find . -name '*.go' \
    -not -path './vendor/*' \
    -not -path './frontend/*' \
    -not -path './node_modules/*' \
    -not -path './third_party/*'
)

if ((${#files[@]} == 0)); then
  exit 0
fi

case "${mode}" in
  write)
    gofmt -w "${files[@]}"
    ;;
  check)
    unformatted="$(gofmt -l "${files[@]}")"
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
