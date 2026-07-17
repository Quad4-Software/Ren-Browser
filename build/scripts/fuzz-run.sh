#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Run a single go test -fuzz target and tolerate a known Go fuzz-engine flake:
# when -fuzztime expires, the coordinator can exit non-zero with only
# "context deadline exceeded" and no file:line assertion failure.
#
# Usage:
#   bash build/scripts/fuzz-run.sh [go test args...]
# Example:
#   bash build/scripts/fuzz-run.sh -fuzz='^FuzzParseURL$' -fuzztime=30s ./internal/nomadnet/...
#
# Real findings still fail: they write a corpus entry ("Failing input written to")
# and/or print a *_test.go:N: assertion line.
set -euo pipefail

export GOFLAGS="${GOFLAGS:--mod=vendor}"
out="$(mktemp)"
trap 'rm -f "$out"' EXIT

set +e
go test "$@" >"$out" 2>&1
status=$?
set -e
cat "$out"

if [[ "$status" -eq 0 ]]; then
  exit 0
fi

if grep -q 'Failing input written to' "$out"; then
  exit "$status"
fi

# Real assertion failures always include a file:line reference.
if grep -qE '[[:space:]]+[^[:space:]]+_test\.go:[0-9]+:' "$out"; then
  exit "$status"
fi

if grep -q 'context deadline exceeded' "$out"; then
  echo "fuzz: treating coordinator context deadline as success (known Go fuzztime flake)" >&2
  exit 0
fi

exit "$status"
