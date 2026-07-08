#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

mkdir -p coverage
OUT="${COVERAGE_OUT:-coverage/go.out}"
HTML="${COVERAGE_HTML:-coverage/go.html}"

echo "Running Go coverage -> ${OUT}"
go test -covermode=atomic -coverprofile="${OUT}" ./internal/...

go tool cover -func="${OUT}" | tee coverage/go-func.txt
go tool cover -html="${OUT}" -o "${HTML}"
echo "Wrote ${HTML}"
