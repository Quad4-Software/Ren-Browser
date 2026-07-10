#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

export GOFLAGS="${GOFLAGS:--mod=vendor}"
export SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH:-1700000000}"
export CGO_ENABLED=0

goos="${GOOS:-linux}"
goarch="${GOARCH:-amd64}"
repro_commit="${GIT_COMMIT:-deadbeef}"
repro_time="${BUILD_TIME:-2026-07-10T00:00:00Z}"
out_a="${root}/bin/repro-check-a"
out_b="${root}/bin/repro-check-b"

cleanup() {
  rm -f "${out_a}" "${out_b}"
}
trap cleanup EXIT

build_once() {
  local out="$1"
  rm -rf frontend/dist
  task build:server \
    OUTPUT="${out}" \
    GIT_COMMIT="${repro_commit}" \
    BUILD_TIME="${repro_time}" \
    GOOS="${goos}" \
    GOARCH="${goarch}"
}

build_once "${out_a}"
build_once "${out_b}"

hash_a="$(sha256sum "${out_a}" | awk '{print $1}')"
hash_b="$(sha256sum "${out_b}" | awk '{print $1}')"

if [[ "${hash_a}" != "${hash_b}" ]]; then
  echo "reproducible binary check: server build is not bitwise reproducible" >&2
  echo "  first:  ${hash_a}" >&2
  echo "  second: ${hash_b}" >&2
  exit 1
fi

echo "reproducible binary check: OK (${goos}/${goarch} sha256=${hash_a})"
