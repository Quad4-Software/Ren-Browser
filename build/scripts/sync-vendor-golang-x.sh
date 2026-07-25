#!/usr/bin/env bash
# Sync vendored golang.org/x/{crypto,net,sync,text} with go.mod after go mod tidy.
# Full `go mod vendor` can fail on Wails webview2 resolution; this keeps -mod=vendor builds consistent.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

modules="${root}/vendor/modules.txt"

mod_ver() {
  local mod=$1
  GOFLAGS=-mod=mod go list -m -f '{{.Version}}' "${mod}"
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
        mkdir -p "${dst}/$(dirname "${rel}")"
        cp -a "${src}/${rel}" "${dst}/${rel}"
      elif [[ -f "${src}/${rel}" ]]; then
        mkdir -p "${dst}/$(dirname "${rel}")"
        cp -a "${src}/${rel}" "${dst}/${rel}"
      fi
    fi
  done < "${modules}"
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
