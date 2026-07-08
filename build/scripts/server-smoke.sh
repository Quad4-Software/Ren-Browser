#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Smoke-test a built renbrowser-server binary: start, hit /, stop.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN="${1:-$ROOT/bin/renbrowser-server}"
HOST="${REN_BROWSER_SMOKE_HOST:-127.0.0.1}"
PORT="${REN_BROWSER_SMOKE_PORT:-19285}"
BASE="http://${HOST}:${PORT}/"

if [[ ! -x "$BIN" ]]; then
  echo "server binary not found or not executable: $BIN" >&2
  exit 1
fi

PROFILE="$(mktemp -d "${TMPDIR:-/tmp}/renbrowser-smoke.XXXXXX")"
cleanup() {
  if [[ -n "${PID:-}" ]] && kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null || true
    wait "$PID" 2>/dev/null || true
  fi
  rm -rf "$PROFILE"
}
trap cleanup EXIT

HOME="$PROFILE" REN_BROWSER_PUBLIC_MODE=1 \
  "$BIN" --host "$HOST" --port "$PORT" >/tmp/renbrowser-smoke.log 2>&1 &
PID=$!

ok=0
for _ in $(seq 1 120); do
  if curl -fsS "$BASE" >/dev/null 2>&1; then
    ok=1
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "server exited early; log:" >&2
    cat /tmp/renbrowser-smoke.log >&2 || true
    exit 1
  fi
  sleep 0.25
done

if [[ "$ok" -ne 1 ]]; then
  echo "server did not become ready at $BASE" >&2
  cat /tmp/renbrowser-smoke.log >&2 || true
  exit 1
fi

BODY="$(curl -fsS "$BASE")"
if [[ -z "$BODY" ]]; then
  echo "empty response from $BASE" >&2
  exit 1
fi

echo "server smoke ok: $BASE"
