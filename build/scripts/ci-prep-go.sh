#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"

reticulum_dir="${root}/third_party/reticulum-go"
if [[ ! -f "${reticulum_dir}/go.mod" || ! -f "${reticulum_dir}/pkg/interfaces/interface.go" ]]; then
  bash "${root}/build/scripts/fetch-reticulum-go.sh"
elif [[ -d "${reticulum_dir}/.git" || "${RETICULUM_GO_FETCH:-}" == "1" ]]; then
  bash "${root}/build/scripts/fetch-reticulum-go.sh"
elif [[ "${GOFLAGS:-}" == *"-mod=mod"* || "${GOFLAGS:-}" == *"-mod=readonly"* ]]; then
  bash "${root}/build/scripts/fetch-reticulum-go.sh"
else
  echo "ci-prep-go: using vendored Reticulum-Go copy"
fi

dist="${root}/frontend/dist"
mkdir -p "${dist}"
if [[ ! -f "${dist}/index.html" ]]; then
  printf '%s\n' '<!DOCTYPE html><html><head></head><body></body></html>' > "${dist}/index.html"
fi
