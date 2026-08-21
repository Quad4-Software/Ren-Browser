#!/usr/bin/env bash
# Run go mod tidy while keeping -mod=vendor builds consistent.
# tidy -e can drop require lines when the module proxy streams fail, leaving
# vendor/modules.txt marking packages as explicit while go.mod no longer does.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

snap_dir="$(mktemp -d "${TMPDIR:-/tmp}/go-mod-tidy.XXXXXX")"
cleanup() { rm -rf "${snap_dir}"; }
trap cleanup EXIT

cp go.mod "${snap_dir}/go.mod"
cp go.sum "${snap_dir}/go.sum"
cp vendor/modules.txt "${snap_dir}/modules.txt"

bash "${root}/build/scripts/fetch-reticulum-go.sh"
GOFLAGS=-mod=mod go mod tidy -e
bash "${root}/build/scripts/ensure-go-mod-vendor-deps.sh"
bash "${root}/build/scripts/sync-vendor-golang-x.sh"

if ! GOFLAGS=-mod=vendor go list ./... >/dev/null 2>"${snap_dir}/vendor-check.err"; then
  echo "go-mod-tidy-ci: tidy left -mod=vendor inconsistent; restoring go.mod/go.sum/modules.txt" >&2
  if [[ -s "${snap_dir}/vendor-check.err" ]]; then
    cat "${snap_dir}/vendor-check.err" >&2 || true
  fi
  cp "${snap_dir}/go.mod" go.mod
  cp "${snap_dir}/go.sum" go.sum
  cp "${snap_dir}/modules.txt" vendor/modules.txt
  GOFLAGS=-mod=vendor go list ./... >/dev/null
fi
