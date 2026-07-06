#!/usr/bin/env bash
# Build the headless server for Windows 7/8/8.1 using go-legacy-win7.
# https://github.com/thongtech/go-legacy-win7
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$root"

go_version="$(awk '/^go / { print $2; exit }' go.mod)"
legacy_tag="v${go_version}-1"
legacy_archive="go-legacy-win7-${go_version}-1.linux_amd64.tar.gz"
legacy_url="https://github.com/thongtech/go-legacy-win7/releases/download/${legacy_tag}/${legacy_archive}"
toolchain_dir="${TOOLCHAIN_DIR:-${root}/build/tools/go-legacy-win7}"
output="${OUTPUT:-${root}/bin/renbrowser-server-windows7-amd64.exe}"
version="$(awk '/^version:/ { gsub(/"/, "", $2); print $2; exit }' build/brand.yml)"
git_commit="$(git rev-parse --short HEAD 2>/dev/null || echo dev)"

if [ ! -f frontend/dist/index.html ]; then
  echo "frontend/dist is missing; build the frontend or set REN_BROWSER_FRONTEND_PREBUILT=1 and download the artifact first." >&2
  exit 1
fi

if [ ! -x "${toolchain_dir}/bin/go" ]; then
  mkdir -p "$(dirname "${toolchain_dir}")"
  tmp="$(mktemp -d)"
  trap 'rm -rf "${tmp}"' EXIT
  echo "Downloading ${legacy_url}"
  curl -fsSL "${legacy_url}" -o "${tmp}/toolchain.tar.gz"
  tar -C "${tmp}" -xzf "${tmp}/toolchain.tar.gz"
  rm -rf "${toolchain_dir}"
  mv "${tmp}/go-legacy-win7" "${toolchain_dir}"
fi

bash build/scripts/patch-wails-vendor.sh
bash build/scripts/ci-prep-go.sh

export GOROOT="${toolchain_dir}"
export PATH="${GOROOT}/bin:${PATH}"
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64
export GOFLAGS=-mod=vendor

"${GOROOT}/bin/go" version
"${GOROOT}/bin/go" build \
  -tags server,production \
  -trimpath \
  -buildvcs=false \
  -ldflags="-w -s -X renbrowser/internal/buildinfo.Version=${version} -X renbrowser/internal/buildinfo.Commit=${git_commit}" \
  -o "${output}"

echo "Built ${output}"
