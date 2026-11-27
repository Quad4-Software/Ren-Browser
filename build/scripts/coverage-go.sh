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

TOTAL="$(awk '/^total:/ { gsub(/%/, "", $NF); print $NF }' coverage/go-func.txt | tail -n 1)"
MINIMUM="${GO_COVERAGE_MIN:-50.0}"
if [ -z "${TOTAL}" ]; then
	echo "coverage gate: unable to parse total coverage" >&2
	exit 1
fi
if awk -v total="${TOTAL}" -v minimum="${MINIMUM}" 'BEGIN { exit !(total < minimum) }'; then
	echo "coverage gate: total Go coverage ${TOTAL}% is below ${MINIMUM}%" >&2
	exit 1
fi
echo "coverage gate: total Go coverage ${TOTAL}% meets ${MINIMUM}%"
echo "Wrote ${HTML}"
