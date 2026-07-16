#!/usr/bin/env bash
# Build or reuse a Go tip toolchain with Android MTE fixes (upstream CLs
# 749062 / 751020) plus a local iOS parity patch for findnull / IndexByte.
#
# Used only for Android and iOS app builds. Desktop and server builds keep
# the system Go from go.mod.
#
# Env:
#   GO_MTE_DIR   install root (default: build/tools/go-mte)
#   GO_MTE_REF   go git ref to build (default: pinned tip commit)
#   GO_MTE_FORCE=1  rebuild even when stamp matches
#   GO_MTE_BUILD_DIR  scratch dir for make.bash (default: cache outside module)
#
# Usage:
#   bash build/scripts/ensure-go-mte.sh
#   eval "$(bash build/scripts/ensure-go-mte.sh --print-env)"
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
# Tip commit after both Android MTE CLs. Override with GO_MTE_REF if needed.
default_ref="1ad6b283d410fb05be4ba63aaacb25b560d23839"
ref="${GO_MTE_REF:-${default_ref}}"
toolchain_dir="${GO_MTE_DIR:-${root}/build/tools/go-mte}"
patch_file="${root}/build/patches/go/mte-ios.patch"
stamp_file="${toolchain_dir}/.renbrowser-mte-stamp"
print_env=0

for arg in "$@"; do
  case "$arg" in
    --print-env) print_env=1 ;;
    --force) GO_MTE_FORCE=1 ;;
    -h|--help)
      sed -n '2,20p' "$0"
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      exit 2
      ;;
  esac
done

patch_hash="$(sha256sum "${patch_file}" | awk '{print $1}')"
desired_stamp="${ref}:${patch_hash}"

have_toolchain() {
  [ -x "${toolchain_dir}/bin/go" ] || return 1
  [ -f "${stamp_file}" ] || return 1
  [ "$(cat "${stamp_file}")" = "${desired_stamp}" ] || return 1
  # Tip must still carry the upstream Android MTE findnull change.
  grep -q 'goos.IsAndroid' "${toolchain_dir}/src/runtime/string.go" 2>/dev/null || return 1
  grep -q 'goos.IsIos' "${toolchain_dir}/src/runtime/string.go" 2>/dev/null || return 1
  # Stale installs that still contain Go's test/ tree poison go test ./...
  [ ! -d "${toolchain_dir}/test" ] || return 1
  return 0
}

print_exports() {
  echo "export GOROOT=\"${toolchain_dir}\""
  echo "export PATH=\"${toolchain_dir}/bin:\${PATH}\""
  echo "export GOTOOLCHAIN=local"
}

if have_toolchain && [ "${GO_MTE_FORCE:-0}" != "1" ]; then
  if [ "${print_env}" = "1" ]; then
    print_exports
  else
    echo "Using MTE Go toolchain at ${toolchain_dir} (${desired_stamp})"
    "${toolchain_dir}/bin/go" version
  fi
  exit 0
fi

if ! command -v go >/dev/null 2>&1; then
  echo "bootstrap go not found on PATH (needed to build tip)" >&2
  exit 1
fi
if ! command -v git >/dev/null 2>&1; then
  echo "git is required to fetch the Go toolchain sources" >&2
  exit 1
fi
if [ ! -f "${patch_file}" ]; then
  echo "missing patch: ${patch_file}" >&2
  exit 1
fi

bootstrap="$(go env GOROOT)"
echo "Building MTE Go toolchain"
echo "  ref:       ${ref}"
echo "  install:   ${toolchain_dir}"
echo "  bootstrap: ${bootstrap} ($(go version))"

mkdir -p "$(dirname "${toolchain_dir}")"
# Scratch must stay outside the renbrowser module. A failed build left under
# build/tools/go-mte-build previously made go test ./... and go mod tidy fail
# on Go's relative-import test trees.
cache_root="${XDG_CACHE_HOME:-${HOME}/.cache}/renbrowser"
workdir="${GO_MTE_BUILD_DIR:-${cache_root}/go-mte-build}"
rm -rf "${workdir}"
mkdir -p "${workdir}"
tmp="${workdir}/src"
mkdir -p "${tmp}"
cleanup() { rm -rf "${workdir}"; }
trap cleanup EXIT

# Keep Go's compile scratch on the same volume as the workdir.
export GOTMPDIR="${workdir}/gotmp"
mkdir -p "${GOTMPDIR}"
export TMPDIR="${GOTMPDIR}"

# Shallow fetch of the pinned commit.
git init -q "${tmp}/go"
git -C "${tmp}/go" remote add origin https://github.com/golang/go.git
git -C "${tmp}/go" fetch -q --depth=1 origin "${ref}"
git -C "${tmp}/go" checkout -q FETCH_HEAD

if ! grep -q '4096\*(1-goos.IsAndroid) + 16\*goos.IsAndroid' "${tmp}/go/src/runtime/string.go"; then
  echo "pinned Go ref ${ref} is missing upstream Android MTE findnull fix" >&2
  echo "bump GO_MTE_REF to a tip commit after CL 749062 / 751020" >&2
  exit 1
fi

# Reset any previous patch markers then apply iOS parity.
git -C "${tmp}/go" apply --check "${patch_file}"
git -C "${tmp}/go" apply "${patch_file}"

export GOROOT_BOOTSTRAP="${bootstrap}"
# Avoid tip trying to download another toolchain while bootstrapping itself.
export GOTOOLCHAIN=local
(
  cd "${tmp}/go/src"
  ./make.bash
)

rm -rf "${toolchain_dir}"
mv "${tmp}/go" "${toolchain_dir}"
# Drop the temp trap target so we do not delete the installed tree.
trap - EXIT
rm -rf "${workdir}"
# History is not needed at build time and bloats CI caches.
rm -rf "${toolchain_dir}/.git"
# Go's top-level test/ tree is not inside module std. If left under
# build/tools/go-mte it is picked up by the parent module's go test ./...
# and go mod tidy (relative imports, mixed packages). Strip it and other
# non-runtime trees from the installed GOROOT.
rm -rf \
  "${toolchain_dir}/test" \
  "${toolchain_dir}/api" \
  "${toolchain_dir}/doc" \
  "${toolchain_dir}/misc" \
  "${toolchain_dir}/.github"

echo "${desired_stamp}" > "${stamp_file}"

echo "Built $("${toolchain_dir}/bin/go" version)"
if [ "${print_env}" = "1" ]; then
  print_exports
fi
