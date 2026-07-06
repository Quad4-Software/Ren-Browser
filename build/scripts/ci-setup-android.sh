#!/usr/bin/env bash
set -euo pipefail

root="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$HOME/Android/Sdk}}"
export ANDROID_HOME="${root}"
export ANDROID_SDK_ROOT="${root}"
mkdir -p "${root}"

if ! command -v java >/dev/null 2>&1; then
  echo "ci-setup-android: java is required (install JDK 21 before this script)" >&2
  exit 1
fi

sdkmanager_bin=""
if command -v sdkmanager >/dev/null 2>&1; then
  sdkmanager_bin="$(command -v sdkmanager)"
else
  tools_dir="${root}/cmdline-tools/latest/bin"
  if [ -x "${tools_dir}/sdkmanager" ]; then
    sdkmanager_bin="${tools_dir}/sdkmanager"
  else
    tmp="$(mktemp -d)"
    trap 'rm -rf "${tmp}"' EXIT
    zip="${tmp}/cmdline-tools.zip"
    curl -fsSL -o "${zip}" "https://dl.google.com/android/repository/commandlinetools-linux-13114758_latest.zip"
    mkdir -p "${root}/cmdline-tools"
    unzip -q "${zip}" -d "${tmp}/extract"
    rm -rf "${root}/cmdline-tools/latest"
    mv "${tmp}/extract/cmdline-tools" "${root}/cmdline-tools/latest"
    sdkmanager_bin="${root}/cmdline-tools/latest/bin/sdkmanager"
  fi
fi

export PATH="$(dirname "${sdkmanager_bin}"):${root}/platform-tools:${PATH}"
yes | "${sdkmanager_bin}" --licenses >/dev/null
"${sdkmanager_bin}" \
  "platform-tools" \
  "platforms;android-34" \
  "build-tools;34.0.0" \
  "ndk;26.3.11579264"

ndk="$(ls -d "${root}/ndk/"* 2>/dev/null | sort -V | tail -1 || true)"
if [ -z "${ndk}" ] || [ ! -d "${ndk}" ]; then
  echo "ci-setup-android: NDK not found under ${root}/ndk" >&2
  exit 1
fi

if [ -n "${GITHUB_ENV:-}" ]; then
  {
    echo "ANDROID_HOME=${root}"
    echo "ANDROID_SDK_ROOT=${root}"
    echo "ANDROID_NDK_HOME=${ndk}"
    echo "PATH=${root}/platform-tools:${root}/cmdline-tools/latest/bin:${PATH}"
  } >> "${GITHUB_ENV}"
fi

echo "Android SDK: ${root}"
echo "Android NDK: ${ndk}"
