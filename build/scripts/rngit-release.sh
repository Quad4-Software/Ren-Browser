#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
cd "${root}"

RNGIT_REMOTE="${RNGIT_REMOTE:-rns://06a54b505bb67b25ef3f8097e8001edc/public/ren-browser}"
VERSION="${VERSION:-$(grep '^version:' build/brand.yml | sed 's/version: *//' | tr -d '"')}"
DIST_DIR="${DIST_DIR:-${root}/dist}"
LOCAL=false

usage() {
  cat <<EOF
Usage: $(basename "$0") [--local] [--version VERSION] [--dist PATH]

Build Ren Browser release binaries (no AppImage or installers) and publish
them with rngit release create.

Artifacts:
  renbrowser-linux-amd64
  renbrowser-windows-amd64.exe
  renbrowser-server-linux-amd64
  renbrowser-android.apk (signed release)

Requires: Linux host, rngit, Android SDK, and ANDROID_KEYSTORE_* variables.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --local|-L)
      LOCAL=true
      shift
      ;;
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dist)
      DIST_DIR="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "$(basename "$0"): unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

export GOFLAGS="${GOFLAGS:--mod=vendor}"

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "$(basename "$0"): native Linux desktop build requires a Linux host." >&2
  exit 1
fi

if ! command -v rngit >/dev/null 2>&1; then
  echo "$(basename "$0"): rngit not found in PATH." >&2
  exit 1
fi

if [[ -z "${ANDROID_HOME:-}" && -z "${ANDROID_SDK_ROOT:-}" ]]; then
  echo "$(basename "$0"): ANDROID_HOME or ANDROID_SDK_ROOT is required." >&2
  exit 1
fi

for var in ANDROID_KEYSTORE_FILE ANDROID_KEYSTORE_PASSWORD ANDROID_KEY_ALIAS ANDROID_KEY_PASSWORD; do
  if [[ -z "${!var:-}" ]]; then
    echo "$(basename "$0"): ${var} is required for a signed release APK." >&2
    exit 1
  fi
done

if [[ ! -f "${ANDROID_KEYSTORE_FILE}" ]]; then
  echo "$(basename "$0"): keystore not found: ${ANDROID_KEYSTORE_FILE}" >&2
  exit 1
fi

echo "Cleaning bin/"
rm -rf bin
mkdir -p bin

echo "Preparing ${DIST_DIR}"
rm -rf "${DIST_DIR}"
mkdir -p "${DIST_DIR}"

echo "Building Linux amd64 desktop binary..."
task linux:build ARCH=amd64 OUTPUT=bin/renbrowser-linux-amd64
install -m 755 bin/renbrowser-linux-amd64 "${DIST_DIR}/renbrowser-linux-amd64"

echo "Building Windows amd64 binary..."
task windows:build ARCH=amd64
install -m 644 bin/renbrowser.exe "${DIST_DIR}/renbrowser-windows-amd64.exe"
rm -f bin/renbrowser.exe

echo "Building Linux amd64 server binary..."
task build:server GOOS=linux GOARCH=amd64 OUTPUT=bin/renbrowser-server-linux-amd64
install -m 755 bin/renbrowser-server-linux-amd64 "${DIST_DIR}/renbrowser-server-linux-amd64"

echo "Building signed Android release APK..."
task package:android
install -m 644 bin/renbrowser.apk "${DIST_DIR}/renbrowser-android.apk"

echo "Release artifacts:"
ls -la "${DIST_DIR}"

rngit_args=(release "${RNGIT_REMOTE}" create "${VERSION}:${DIST_DIR}")
if [[ "${LOCAL}" == "true" ]]; then
  rngit_args+=(--local)
fi

rngit "${rngit_args[@]}"
echo "rngit release ${VERSION} complete."
