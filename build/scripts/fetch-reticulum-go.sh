#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
vendor_dir="${root}/third_party/reticulum-go"
mod_go="${vendor_dir}/go.mod"
iface_go="${vendor_dir}/pkg/interfaces/interface.go"
version_file="${root}/build/scripts/reticulum-go-version"
repo="${RETICULUM_GO_REPO:-https://git.quad4.io/reticulum/Reticulum-Go.git}"

if [[ ! -f "${version_file}" ]]; then
  echo "fetch-reticulum-go: missing ${version_file}" >&2
  exit 1
fi

ref="$(tr -d '[:space:]' < "${version_file}")"
if [[ -z "${ref}" ]]; then
  echo "fetch-reticulum-go: empty ref in ${version_file}" >&2
  exit 1
fi

if [[ -f "${mod_go}" && -f "${iface_go}" && -d "${vendor_dir}/.git" ]]; then
  head="$(git -C "${vendor_dir}" rev-parse HEAD 2>/dev/null || true)"
  want=""
  if [[ "${ref}" =~ ^[0-9a-fA-F]{40}$ ]]; then
    want="$(printf '%s' "${ref}" | tr 'A-F' 'a-f')"
  else
    want="$(git -C "${vendor_dir}" rev-parse -q --verify "${ref}^{commit}" 2>/dev/null || true)"
  fi
  if [[ -n "${head}" && "${head}" == "${want}" ]]; then
    exit 0
  fi
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
