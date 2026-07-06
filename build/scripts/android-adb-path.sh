#!/usr/bin/env bash
set +e
if command -v adb >/dev/null 2>&1; then
  command -v adb
  exit 0
fi
home="${HOME:-/home/runner}"
root="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
if [ -z "${root}" ]; then
  if [ -d "${home}/Library/Android/sdk" ]; then
    root="${home}/Library/Android/sdk"
  else
    root="${home}/Android/Sdk"
  fi
fi
echo "${root}/platform-tools/adb"
exit 0
