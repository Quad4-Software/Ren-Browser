#!/bin/sh
# Verify renbrowser.rsm signature and byte-level file hashes.
#
# Env:
#   RNS_REQUIRED_SIGNER  identity hash (default: e46112d44649266d71fe2193e00a4710)
#   RNS_RSM_PATH         path to .rsm (default: renbrowser.rsm)
#   RNS_INVENTORY_OUT    if set, write extracted inventory here (for end-of-job recheck)
#
# Usage:
#   sh build/scripts/verify-tree-rsm.sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

SIGNER="${RNS_REQUIRED_SIGNER:-e46112d44649266d71fe2193e00a4710}"
RSM_PATH="${RNS_RSM_PATH:-$ROOT/renbrowser.rsm}"
HEADER="# renbrowser tree manifest v1"

if [ ! -f "$RSM_PATH" ]; then
	echo "verify-tree-rsm.sh: missing $RSM_PATH" >&2
	exit 1
fi

INV="$(mktemp "${TMPDIR:-/tmp}/tree-inv-verify.XXXXXX")"
RAW="$(mktemp "${TMPDIR:-/tmp}/tree-rsm-raw.XXXXXX")"
trap 'rm -f "$INV" "$RAW"' EXIT INT

if command -v rnid >/dev/null 2>&1; then
	if ! rnid -i "$SIGNER" -V "$RSM_PATH" >"$RAW" 2>/dev/null; then
		echo "verify-tree-rsm.sh: RSM signature verification failed" >&2
		exit 1
	fi
	awk -v h="$HEADER" 'BEGIN{p=0} $0==h{p=1} p{print}' "$RAW" >"$INV"
elif command -v reticulum-go >/dev/null 2>&1; then
	if ! reticulum-go id -i "$SIGNER" -V "$RSM_PATH" -extract >"$INV" 2>/dev/null; then
		# Fallback for older reticulum-go without -extract
		if ! reticulum-go id -i "$SIGNER" -V "$RSM_PATH" >"$RAW" 2>/dev/null; then
			echo "verify-tree-rsm.sh: RSM signature verification failed" >&2
			exit 1
		fi
		awk -v h="$HEADER" 'BEGIN{p=0} $0==h{p=1} p{print}' "$RAW" >"$INV"
	fi
else
	echo "verify-tree-rsm.sh: need rnid or reticulum-go on PATH" >&2
	exit 1
fi

if [ ! -s "$INV" ]; then
	echo "verify-tree-rsm.sh: could not extract inventory from RSM" >&2
	exit 1
fi

if [ -n "${RNS_INVENTORY_OUT:-}" ]; then
	cp "$INV" "$RNS_INVENTORY_OUT"
fi

sh "$ROOT/build/scripts/tree-manifest.sh" verify-tracked "$INV"
echo "verify-tree-rsm.sh: OK (signer $SIGNER)"
