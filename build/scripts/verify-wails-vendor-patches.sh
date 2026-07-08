#!/usr/bin/env bash
# Fail fast when vendored Wails patches are missing or syntactically broken.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
app="${root}/vendor/github.com/wailsapp/wails/v3/pkg/application"

fail() {
  echo "verify-wails-vendor-patches: $*" >&2
  exit 1
}

if [ ! -d "${app}" ]; then
  exit 0
fi

grep -q 'Force shutdown' "${app}/application_server.go" || fail 'missing Force shutdown patch in application_server.go'
grep -q 'func (b \*WebSocketBroadcaster) closeAll()' "${app}/websocket_server.go" || fail 'missing closeAll patch in websocket_server.go'
grep -q 'window_apply_frameless' "${app}/linux_cgo.c" || fail 'missing linux frameless patch in linux_cgo.c'

if grep -q 'shutdownDone <- _ = h.server.Shutdown' "${app}/application_server.go"; then
  fail 'broken application_server.go shutdown patch'
fi

cd "${root}"
GOFLAGS=-mod=vendor go build -tags server,production -trimpath -o /dev/null .
echo "verify-wails-vendor-patches: ok"
