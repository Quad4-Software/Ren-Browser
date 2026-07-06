#!/usr/bin/env bash
# Resolves the Android SDK root for Task vars. Always exits 0 so unrelated tasks
# (e.g. linux:build) are not failed when the SDK is absent.
set +e
home="${HOME:-/home/runner}"
for d in "${ANDROID_HOME:-}" "${ANDROID_SDK_ROOT:-}" "${home}/Library/Android/sdk" "${home}/Android/Sdk"; do
  if [ -n "${d}" ] && [ -d "${d}/ndk" ]; then
    echo "${d}"
    exit 0
  fi
done
if [ -n "${ANDROID_HOME:-}" ]; then
  echo "${ANDROID_HOME}"
elif [ -n "${ANDROID_SDK_ROOT:-}" ]; then
  echo "${ANDROID_SDK_ROOT}"
else
  echo "${home}/Android/Sdk"
fi
exit 0
