#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
vendor_dir="${root}/third_party/reticulum-go"
mod_go="${vendor_dir}/go.mod"
iface_go="${vendor_dir}/pkg/interfaces/interface.go"
version_file="${root}/build/scripts/reticulum-go-version"
repo="${RETICULUM_GO_REPO:-https://github.com/Quad4-Software/Reticulum-Go.git}"

if [[ -f "${mod_go}" && -f "${iface_go}" ]]; then
  exit 0
fi

if [[ ! -f "${version_file}" ]]; then
  echo "fetch-reticulum-go: missing ${version_file}" >&2
  exit 1
fi

ref="$(tr -d '[:space:]' < "${version_file}")"
if [[ -z "${ref}" ]]; then
  echo "fetch-reticulum-go: empty ref in ${version_file}" >&2
  exit 1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "fetch-reticulum-go: git is required" >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

git -C "${tmpdir}" init -q
git -C "${tmpdir}" remote add origin "${repo}"
if ! git -C "${tmpdir}" fetch --depth 1 origin "${ref}"; then
  echo "fetch-reticulum-go: failed to fetch ${ref} from ${repo}" >&2
  exit 1
fi
git -C "${tmpdir}" checkout -q FETCH_HEAD

mkdir -p "${root}/third_party"
rm -rf "${vendor_dir}"
mkdir -p "${vendor_dir}"
if command -v rsync >/dev/null 2>&1; then
  rsync -a --delete "${tmpdir}/" "${vendor_dir}/"
else
  cp -a "${tmpdir}/." "${vendor_dir}/"
  rm -rf "${vendor_dir}/.git"
fi

if [[ ! -f "${mod_go}" ]]; then
  echo "fetch-reticulum-go: ${mod_go} not found after checkout ${ref}" >&2
  exit 1
fi

if [[ ! -f "${iface_go}" ]]; then
  echo "fetch-reticulum-go: ${iface_go} not found after checkout ${ref}" >&2
  exit 1
fi
