#!/usr/bin/env bash
# Run go mod tidy while keeping -mod=vendor builds consistent.
# tidy -e can drop require lines when the module proxy streams fail, leaving
# vendor/modules.txt marking packages as explicit while go.mod no longer does.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

snap_dir="$(mktemp -d "${TMPDIR:-/tmp}/go-mod-tidy.XXXXXX")"
created_stub=0
dist="${root}/frontend/dist"

cleanup() {
  # Do not leave a stub dist behind: Task treats dist/index.html as a generate
  # and would skip vite on the next build:frontend (breaks reproducible checks).
  if [[ "${created_stub}" -eq 1 ]]; then
    rm -f "${dist}/index.html"
    rmdir "${dist}" 2>/dev/null || true
  fi
  rm -rf "${snap_dir}"
}
trap cleanup EXIT

cp go.mod "${snap_dir}/go.mod"
cp go.sum "${snap_dir}/go.sum"
cp vendor/modules.txt "${snap_dir}/modules.txt"

# go list loads //go:embed all:frontend/dist from main_*.go. Create a temporary
# stub only for the vendor consistency check, then remove it in cleanup.
mkdir -p "${dist}"
if [[ ! -f "${dist}/index.html" ]]; then
  printf '%s\n' '<!DOCTYPE html><html><head></head><body></body></html>' > "${dist}/index.html"
  created_stub=1
fi

bash "${root}/build/scripts/fetch-reticulum-go.sh"
GOFLAGS=-mod=mod go mod tidy -e
bash "${root}/build/scripts/ensure-go-mod-vendor-deps.sh"
bash "${root}/build/scripts/sync-vendor-golang-x.sh"

vendor_check_err="${snap_dir}/vendor-check.err"
if ! GOFLAGS=-mod=vendor go list ./... >/dev/null 2>"${vendor_check_err}"; then
  if ! grep -Eqi 'inconsistent vendoring|vendor/modules\.txt|explicitly required|marked as explicit' "${vendor_check_err}"; then
    echo "go-mod-tidy-ci: go list failed (not treating as vendor inconsistency):" >&2
    cat "${vendor_check_err}" >&2 || true
    exit 1
  fi
  echo "go-mod-tidy-ci: tidy left -mod=vendor inconsistent; restoring go.mod/go.sum/modules.txt" >&2
  cat "${vendor_check_err}" >&2 || true
  cp "${snap_dir}/go.mod" go.mod
  cp "${snap_dir}/go.sum" go.sum
  cp "${snap_dir}/modules.txt" vendor/modules.txt
  GOFLAGS=-mod=vendor go list ./... >/dev/null
fi
