#!/bin/sh
# Sign a byte-level tree inventory into renbrowser.rsm (rnid-compatible).
#
# Requires a private identity file (64-byte .rid / rngit client_identity).
# Never commit the identity file.
#
# Env:
#   RNS_ID_PATH   path to identity (required unless -i given)
#   RNS_RSM_PATH  output path (default: renbrowser.rsm in repo root)
#
# Usage:
#   RNS_ID_PATH=~/.local/share/reticulum-go/reticulum-go-release.rid sh build/scripts/sign-tree-rsm.sh
set -eu

ROOT="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

ID_PATH="${RNS_ID_PATH:-}"
RSM_PATH="${RNS_RSM_PATH:-$ROOT/renbrowser.rsm}"

while [ "$#" -gt 0 ]; do
	case "$1" in
	-i)
		ID_PATH="${2:?}"
		shift 2
		;;
	-o)
		RSM_PATH="${2:?}"
		shift 2
		;;
	*)
		echo "sign-tree-rsm.sh: unknown arg: $1" >&2
		exit 2
		;;
	esac
done

if [ -z "$ID_PATH" ]; then
	echo "sign-tree-rsm.sh: set RNS_ID_PATH or pass -i /path/to.rid" >&2
	exit 1
fi
if [ ! -f "$ID_PATH" ]; then
	echo "sign-tree-rsm.sh: identity not found: $ID_PATH" >&2
	exit 1
fi

run_signer() {
	# Prefer rnid. older reticulum-go builds lack -S @file and -extract.
	if command -v rnid >/dev/null 2>&1; then
		rnid -i "$ID_PATH" -S -r "$1" -w "$RSM_PATH" -f
	elif command -v reticulum-go >/dev/null 2>&1; then
		reticulum-go id -i "$ID_PATH" -S "@$1" -w "$RSM_PATH" -f
	else
		echo "sign-tree-rsm.sh: need rnid or reticulum-go on PATH" >&2
		return 1
	fi
}

INV="$(mktemp "${TMPDIR:-/tmp}/tree-inv.XXXXXX")"
trap 'rm -f "$INV"' EXIT INT

sh "$ROOT/build/scripts/tree-manifest.sh" generate >"$INV"
run_signer "$INV"
echo "sign-tree-rsm.sh: wrote $RSM_PATH"
