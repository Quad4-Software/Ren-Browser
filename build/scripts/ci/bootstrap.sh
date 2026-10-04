#!/usr/bin/env bash
# Verify or install CI toolchain without GitHub Actions or other marketplace steps.
#
# Usage:
#   bash build/scripts/ci/bootstrap.sh verify
#   bash build/scripts/ci/bootstrap.sh install
#
# Env:
#   CI_TOOLCHAIN_ROOT  install prefix (default: repo/.cache/ci-toolchain)
#   CI_BOOTSTRAP       verify | install (default from first argument)
set -euo pipefail

. "$(dirname "$0")/env.sh"
. "$ROOT/build/scripts/ci/versions.env"

mode="${1:-${CI_BOOTSTRAP:-verify}}"

toolchain="${CI_TOOLCHAIN_ROOT:-$ROOT/.cache/ci-toolchain}"
bindir="$toolchain/bin"
gopath="$toolchain/gopath"

export PATH="$bindir:$PATH"
export GOPATH="$gopath"
export GOCACHE="${GOCACHE:-$toolchain/gocache}"
export GOTOOLCHAIN=local

go_want="${GO_VERSION}"
node_want="${NODE_VERSION}"
pnpm_want="${PNPM_VERSION}"
task_want="${TASK_VERSION}"

semver_ge() {
	# True when $1 >= $2 (major.minor.patch numeric compare)
	awk -v a="$1" -v b="$2" 'BEGIN {
		split(a, A, "."); split(b, B, ".")
		for (i = 1; i <= 3; i++) {
			ai = A[i] + 0; bi = B[i] + 0
			if (ai > bi) exit 0
			if (ai < bi) exit 1
		}
		exit 0
	}'
}

have_go() {
	command -v go >/dev/null 2>&1 || return 1
	local v
	v="$(go env GOVERSION 2>/dev/null | sed 's/^go//')"
	semver_ge "$v" "$go_want"
}

have_node() {
	command -v node >/dev/null 2>&1 || return 1
	local v
	v="$(node -p "process.versions.node" 2>/dev/null)"
	semver_ge "$v" "$node_want"
}

have_pnpm() {
	command -v pnpm >/dev/null 2>&1 || return 1
	local v
	v="$(pnpm --version 2>/dev/null | tr -d 'v')"
	semver_ge "$v" "$pnpm_want"
}

have_task() {
	command -v task >/dev/null 2>&1 || return 1
	local v
	v="$(task --version 2>/dev/null | awk '{print $NF}' | tr -d 'v')"
	semver_ge "$v" "$task_want"
}

verify_all() {
	local ok=1
	have_go || { echo "bootstrap: need Go >= $go_want on PATH" >&2; ok=0; }
	have_node || { echo "bootstrap: need Node >= $node_want on PATH" >&2; ok=0; }
	have_pnpm || { echo "bootstrap: need pnpm >= $pnpm_want on PATH" >&2; ok=0; }
	have_task || { echo "bootstrap: need Task >= $task_want on PATH" >&2; ok=0; }
	[ "$ok" -eq 1 ]
}

linux_arch() {
	local m
	m="$(uname -m)"
	case "$m" in
	x86_64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*) echo "bootstrap: unsupported cpu: $m" >&2; exit 1 ;;
	esac
}

fetch() {
	local url="$1" dest="$2"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$dest"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$dest" "$url"
	else
		echo "bootstrap: need curl or wget" >&2
		exit 1
	fi
}

install_go() {
	if have_go; then
		return 0
	fi
	local arch go_tgz
	arch="$(linux_arch)"
	go_tgz="go${go_want}.linux-${arch}.tar.gz"
	mkdir -p "$toolchain"
	tmp="$(mktemp "${TMPDIR:-/tmp}/go-bootstrap.XXXXXX")"
	fetch "https://go.dev/dl/${go_tgz}" "$tmp"
	rm -rf "$toolchain/go"
	tar -C "$toolchain" -xzf "$tmp"
	rm -f "$tmp"
	ln -sf "$toolchain/go/bin/go" "$bindir/go"
	ln -sf "$toolchain/go/bin/gofmt" "$bindir/gofmt"
	export PATH="$bindir:$PATH"
}

node_linux_arch() {
	local arch
	arch="$(linux_arch)"
	case "$arch" in
	amd64) echo x64 ;;
	arm64) echo arm64 ;;
	*) echo "$arch" ;;
	esac
}

install_node() {
	if have_node && have_pnpm; then
		return 0
	fi
	local arch node_tgz ver_major
	arch="$(node_linux_arch)"
	ver_major="${node_want%%.*}"
	node_tgz="node-v${node_want}-linux-${arch}.tar.xz"
	mkdir -p "$toolchain"
	tmp="$(mktemp "${TMPDIR:-/tmp}/node-bootstrap.XXXXXX")"
	fetch "https://nodejs.org/dist/v${node_want}/${node_tgz}" "$tmp"
	rm -rf "$toolchain/node"
	mkdir -p "$toolchain/node"
	tar -C "$toolchain/node" --strip-components=1 -xJf "$tmp"
	rm -f "$tmp"
	ln -sf "$toolchain/node/bin/node" "$bindir/node"
	ln -sf "$toolchain/node/bin/npm" "$bindir/npm"
	ln -sf "$toolchain/node/bin/npx" "$bindir/npx"
	export PATH="$bindir:$PATH"
	if ! have_pnpm; then
		npm config set prefix "$toolchain/node"
		npm install -g "pnpm@${pnpm_want}"
		ln -sf "$toolchain/node/bin/pnpm" "$bindir/pnpm" 2>/dev/null || true
	fi
}

install_task() {
	if have_task; then
		return 0
	fi
	install_go
	mkdir -p "$gopath/bin"
	# Task is fetched from the module proxy, not vendor/
	GOFLAGS= GOBIN="$bindir" go install "github.com/go-task/task/v3/cmd/task@v${task_want}"
	export PATH="$bindir:$PATH"
}

install_all() {
	mkdir -p "$bindir" "$gopath" "$GOCACHE"
	install_go
	install_node
	install_task
	verify_all
	echo "bootstrap: toolchain ready under $toolchain"
}

case "$mode" in
verify)
	verify_all
	echo "bootstrap: verify OK"
	;;
install)
	install_all
	;;
*)
	echo "bootstrap: usage: bootstrap.sh verify|install" >&2
	exit 2
	;;
esac
