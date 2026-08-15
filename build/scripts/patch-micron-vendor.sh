#!/usr/bin/env bash
# Re-apply Ren Browser micron-parser-go patches after `go mod vendor`.
set -euo pipefail

root="$(cd "$(dirname "$0")/../.." && pwd)"
line_go="${root}/vendor/micron-parser-go/micron/line.go"

if [ ! -f "${line_go}" ]; then
  exit 0
fi

if ! grep -q 's.Depth > 16' "${line_go}"; then
  sed -i '/s\.Depth = i$/a\
\t\t\t\tif s.Depth > 16 {\
\t\t\t\t\ts.Depth = 16\
\t\t\t\t}
' "${line_go}"
fi

if grep -q 'ind := max((s.Depth-1)\*2, 0)' "${line_go}"; then
  sed -i '/func sectionIndentStyleEm(s \*State) float64 {/,/return float64(ind) \* 0.6/c\
func sectionIndentStyleEm(s *State) float64 {\
\tdepth := s.Depth\
\tif depth > 16 {\
\t\tdepth = 16\
\t}\
\tind := max((depth-1)*2, 0)\
\tif ind <= 0 {\
\t\treturn 0\
\t}\
\treturn float64(ind) * 0.6
' "${line_go}"
fi
