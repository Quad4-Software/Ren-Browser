#!/bin/sh
# Fail if tracked file bytes changed vs a saved inventory, or unexpected
# untracked files appeared (GitHub runner mutation check).
#
# Usage:
#   verify-workspace-clean.sh <inventory-file>
#
# Env:
#   RNS_CLEAN_ALLOW   space-separated path prefixes always ignored (optional)
#   RNS_CLEAN_SOFT    if set to 1, inventory hash mismatches warn but do not
#                     fail the job. Unexpected untracked files still fail.
#                     Start-of-job verify-tree-rsm.sh stays hard either way.
set -eu

ROOT="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

INV="${1:?inventory file}"
if [ ! -f "$INV" ]; then
	echo "verify-workspace-clean.sh: missing inventory: $INV" >&2
	exit 1
fi

SOFT=0
case "${RNS_CLEAN_SOFT:-}" in
1 | true | TRUE | yes | YES)
	SOFT=1
	;;
esac

if [ "$SOFT" -eq 1 ]; then
	if ! sh "$ROOT/build/scripts/tree-manifest.sh" verify "$INV"; then
		echo "verify-workspace-clean.sh: WARNING inventory mismatch (soft mode)" >&2
	fi
else
	sh "$ROOT/build/scripts/tree-manifest.sh" verify "$INV"
fi

# Default ephemeral prefixes created by CI / local builds
ALLOW="bin/ bench/ sbom/ reports/ coverage/ dist/ node_modules/ frontend/node_modules/ frontend/dist/ frontend/.svelte-kit/ frontend/playwright-report/ frontend/test-results/ .docker/ .flatpak-builder/ build/linux/ build/flatpak/ build/tools/ build/packaging/ .task/ __pycache__/ .cache/"
ALLOW="$ALLOW ${RNS_CLEAN_ALLOW:-}"

is_allowed() {
	p="$1"
	for a in $ALLOW; do
		case "$p" in
		"$a" | "$a"*)
			return 0
			;;
		esac
	done
	case "$p" in
	vendor | vendor/* | */vendor | */vendor/*)
		return 0
		;;
	*.log | *.tmp | *.swp | *.exe | *.wasm | *.apk | *.AppImage)
		return 0
		;;
	esac
	return 1
}

fail=0
tmp="$(mktemp "${TMPDIR:-/tmp}/ws-clean.XXXXXX")"
trap 'rm -f "$tmp"' EXIT INT
git status --porcelain -u --ignored=no >"$tmp" 2>/dev/null || git status --porcelain -u >"$tmp"
while IFS= read -r line; do
	[ -z "$line" ] && continue
	xy="$(printf '%s\n' "$line" | cut -c1-2)"
	path="$(printf '%s\n' "$line" | sed 's/^.. //;s/.* -> //')"
	case "$xy" in
	"??")
		if is_allowed "$path"; then
			continue
		fi
		echo "verify-workspace-clean.sh: unexpected untracked: $path" >&2
		fail=1
		;;
	*)
		if [ "$path" = "renbrowser.rsm" ]; then
			continue
		fi
		if is_allowed "$path"; then
			continue
		fi
		if [ "$SOFT" -eq 1 ]; then
			echo "verify-workspace-clean.sh: WARNING unexpected change: $line" >&2
			continue
		fi
		echo "verify-workspace-clean.sh: unexpected change: $line" >&2
		fail=1
		;;
	esac
done <"$tmp"

if [ "$fail" -ne 0 ]; then
	echo "verify-workspace-clean.sh: workspace not clean" >&2
	exit 1
fi
echo "verify-workspace-clean.sh: OK"
