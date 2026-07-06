#!/usr/bin/env bash
# Exit 0 when vite build should be skipped (prebuilt dist present); non-zero to run it.
set -euo pipefail
root="$(cd "$(dirname "$0")/../.." && pwd)"
if [[ "${REN_BROWSER_FRONTEND_PREBUILT:-0}" == "1" ]] && [[ -f "${root}/frontend/dist/index.html" ]]; then
  exit 0
fi
exit 1
