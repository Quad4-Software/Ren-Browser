#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
vendor_dir="${root}/third_party/reticulum-go"
iface_go="${vendor_dir}/pkg/interfaces/interface.go"
gomod_vendor_iface_go="${root}/vendor/quad4/reticulum-go/pkg/interfaces/interface.go"

bash "${root}/build/scripts/fetch-reticulum-go.sh"

if [[ ! -f "${iface_go}" ]]; then
  echo "patch-reticulum-txbytes: ${iface_go} not found" >&2
  exit 1
fi

patch_txbytes() {
  local target="$1"
  local tmp
  tmp="$(mktemp)"
  awk '
/func \(i \*BaseInterface\) updateBandwidthStats\(bytes uint64\) \{/ { in_fn=1 }
in_fn && /i\.lastTx = time\.Now\(\)/ && !patched {
  print "\ti.TxBytes += bytes"
  patched=1
}
{ print }
' "${target}" > "${tmp}"
  cp "${tmp}" "${target}"
  rm -f "${tmp}"
}

if ! grep -q 'i\.TxBytes += bytes' "${iface_go}"; then
  patch_txbytes "${iface_go}"

  if ! grep -q 'i\.TxBytes += bytes' "${iface_go}"; then
    echo "patch-reticulum-txbytes: failed to patch ${iface_go}" >&2
    exit 1
  fi
fi

# go.mod's replace directive points quad4/reticulum-go at third_party/reticulum-go,
# so `go mod vendor` is what normally copies this fix into vendor/. That step is
# only run by the vendor:go task, not by go:mod:tidy/ci-prep-go, so also patch the
# committed vendor/ copy directly here to keep -mod=vendor builds (the default,
# since vendor/modules.txt is present) from silently reverting to the buggy behavior.
if [[ -f "${gomod_vendor_iface_go}" ]] && ! grep -q 'i\.TxBytes += bytes' "${gomod_vendor_iface_go}"; then
  patch_txbytes "${gomod_vendor_iface_go}"

  if ! grep -q 'i\.TxBytes += bytes' "${gomod_vendor_iface_go}"; then
    echo "patch-reticulum-txbytes: failed to patch ${gomod_vendor_iface_go}" >&2
    exit 1
  fi
fi
