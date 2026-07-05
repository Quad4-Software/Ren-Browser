#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
vendor_dir="${root}/third_party/reticulum-go"
iface_go="${vendor_dir}/pkg/interfaces/interface.go"

bash "${root}/build/scripts/fetch-reticulum-go.sh"

if [[ ! -f "${iface_go}" ]]; then
  echo "patch-reticulum-txbytes: ${iface_go} not found" >&2
  exit 1
fi

if grep -q 'i\.TxBytes += bytes' "${iface_go}"; then
  exit 0
fi

tmp="$(mktemp)"
awk '
/func \(i \*BaseInterface\) updateBandwidthStats\(bytes uint64\) \{/ { in_fn=1 }
in_fn && /i\.lastTx = time\.Now\(\)/ && !patched {
  print "\ti.TxBytes += bytes"
  patched=1
}
{ print }
' "${iface_go}" > "${tmp}"
cp "${tmp}" "${iface_go}"
rm -f "${tmp}"

if ! grep -q 'i\.TxBytes += bytes' "${iface_go}"; then
  echo "patch-reticulum-txbytes: failed to patch ${iface_go}" >&2
  exit 1
fi
