#!/usr/bin/env bash
# Portable CI setup (no GitHub Actions). Installs nothing system-wide except
# optional Linux packages when CI_INSTALL_LINUX_DEPS=1.
set -euo pipefail

. "$(dirname "$0")/env.sh"

ensure_rnid() {
	if command -v rnid >/dev/null 2>&1 || command -v reticulum-go >/dev/null 2>&1; then
		return 0
	fi
	require_cmd python3
	local venv="${CI_RNS_VENV:-${TMPDIR:-/tmp}/renbrowser-rns-venv}"
	python3 -m venv "$venv"
	# shellcheck disable=SC1091
	. "$venv/bin/activate"
	python -m pip install -q 'rns>=1.3.8'
	export PATH="$venv/bin:$PATH"
}

if ! ci_truthy "${CI_SKIP_TREE_VERIFY:-}"; then
	ensure_rnid
	export RNS_INVENTORY_OUT="${RNS_INVENTORY_OUT:-${TMPDIR:-/tmp}/renbrowser-ci-inventory.$$}"
	sh "$ROOT/build/scripts/verify-tree-rsm.sh"
fi

bash "$ROOT/build/scripts/ci-prep-go.sh"

if ci_truthy "${CI_INSTALL_LINUX_DEPS:-}" && ci_on_linux; then
	bash "$ROOT/build/scripts/ci-install-linux-deps.sh"
fi

if ci_truthy "${CI_INSTALL_FRONTEND_DEPS:-}"; then
	require_cmd pnpm
	if command -v task >/dev/null 2>&1; then
		task common:install:frontend:deps
	else
		pnpm install --dir frontend
	fi
fi

echo "ci setup: OK"
