#!/usr/bin/env bash
# Sync vendored golang.org/x/{crypto,net,sync,text} with go.mod after go mod tidy.
# Full `go mod vendor` can fail on Wails webview2 resolution; this keeps -mod=vendor builds consistent.
#
# File selection mirrors cmd/go's go mod vendor:
# - skip *_test.go and testdata/
# - skip .go files tagged //go:build ignore (or +build ignore)
# - copy module metadata (LICENSE, PATENTS, …) into the vendored module root
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

modules="${root}/vendor/modules.txt"

# Metadata name prefixes copied by go mod vendor (cmd/go/internal/modcmd/vendor.go).
meta_prefixes=(
  AUTHORS
  CONTRIBUTORS
  COPYLEFT
  COPYING
  COPYRIGHT
  LEGAL
  LICENSE
  NOTICE
  PATENTS
)

mod_ver() {
  local mod=$1
  GOFLAGS=-mod=mod go list -m -f '{{.Version}}' "${mod}"
}

# True when a .go file is tagged ignore and must not be vendored.
is_ignore_tagged_go() {
  local f=$1
  # Match leading build constraints only (same region go/parser considers).
  awk '
    /^[[:space:]]*$/ { next }
    /^\/\// {
      if ($0 ~ /^\/\/go:build[[:space:]]+ignore([[:space:]]|$)/) { found=1; exit }
      if ($0 ~ /^\/\/[[:space:]]*\+build[[:space:]]+ignore([[:space:]]|$)/) { found=1; exit }
      next
    }
    { exit }
    END { exit found ? 0 : 1 }
  ' "${f}"
}

is_metadata_name() {
  local name=$1 prefix
  for prefix in "${meta_prefixes[@]}"; do
    if [[ "${name}" == "${prefix}"* ]]; then
      return 0
    fi
  done
  return 1
}

copy_metadata() {
  local src=$1 dst=$2
  local f base
  shopt -s nullglob
  for f in "${src}"/*; do
    [[ -f "${f}" ]] || continue
    base="$(basename "${f}")"
    if is_metadata_name "${base}"; then
      mkdir -p "${dst}"
      cp -a "${f}" "${dst}/${base}"
    fi
  done
  shopt -u nullglob
}

# Copy one package directory the way go mod vendor does (no tests, no ignore-tagged .go).
copy_package_dir() {
  local src_pkg=$1 dst_pkg=$2
  local f base
  mkdir -p "${dst_pkg}"
  shopt -s nullglob
  for f in "${src_pkg}"/*; do
    base="$(basename "${f}")"
    # Package dirs only; nested packages are listed separately in modules.txt.
    # testdata/ and other subdirs are not part of the package for vendoring.
    if [[ -d "${f}" ]]; then
      continue
    fi
    if [[ ! -f "${f}" ]]; then
      continue
    fi
    if [[ "${base}" == *_test.go ]]; then
      continue
    fi
    if [[ "${base}" == go.mod || "${base}" == go.sum ]]; then
      continue
    fi
    if [[ "${base}" == *.go ]] && is_ignore_tagged_go "${f}"; then
      continue
    fi
    cp -a "${f}" "${dst_pkg}/${base}"
  done
  shopt -u nullglob
}

copy_listed_packages() {
  local modpath=$1 ver=$2
  local modcache src dst rel
  modcache="$(go env GOMODCACHE)"
  src="${modcache}/${modpath}@${ver}"
  dst="${root}/vendor/${modpath}"

  if [[ ! -d "${src}" ]]; then
    GOFLAGS=-mod=mod go mod download "${modpath}@${ver}"
  fi

  chmod -R u+w "${dst}" 2>/dev/null || true
  rm -rf "${dst}"
  mkdir -p "${dst}"

  local in_block=0
  while IFS= read -r line; do
    if [[ "${line}" == "# ${modpath} "* ]]; then
      in_block=1
      continue
    fi
    if [[ "${in_block}" -eq 1 && "${line}" == \#* && "${line}" != \#\#* ]]; then
      break
    fi
    if [[ "${in_block}" -eq 1 && "${line}" == "${modpath}/"* ]]; then
      rel="${line#${modpath}/}"
      if [[ -d "${src}/${rel}" ]]; then
        copy_package_dir "${src}/${rel}" "${dst}/${rel}"
      elif [[ -f "${src}/${rel}" ]]; then
        mkdir -p "${dst}/$(dirname "${rel}")"
        cp -a "${src}/${rel}" "${dst}/${rel}"
      fi
    fi
  done < "${modules}"

  copy_metadata "${src}" "${dst}"
  # Module cache entries are often mode 0444; keep the working tree writable.
  chmod -R u+w "${dst}" 2>/dev/null || true
}

crypto_ver="$(mod_ver golang.org/x/crypto)"
net_ver="$(mod_ver golang.org/x/net)"
sync_ver="$(mod_ver golang.org/x/sync)"
text_ver="$(mod_ver golang.org/x/text)"

modules_tmp="$(mktemp)"
sed \
  -e "s|^# golang.org/x/crypto v.*|# golang.org/x/crypto ${crypto_ver}|" \
  -e "s|^# golang.org/x/net v.*|# golang.org/x/net ${net_ver}|" \
  -e "s|^# golang.org/x/sync v.*|# golang.org/x/sync ${sync_ver}|" \
  -e "s|^# golang.org/x/text v.*|# golang.org/x/text ${text_ver}|" \
  "${modules}" > "${modules_tmp}"
mv "${modules_tmp}" "${modules}"

copy_listed_packages golang.org/x/crypto "${crypto_ver}"
copy_listed_packages golang.org/x/net "${net_ver}"
copy_listed_packages golang.org/x/sync "${sync_ver}"
copy_listed_packages golang.org/x/text "${text_ver}"
