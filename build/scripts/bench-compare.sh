#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Compare current Go benchmark output against a baseline artifact.
# Fails when a matching benchmark regresses by more than BENCH_REGRESSION_PCT (default 25%).
set -euo pipefail

BASELINE="${1:-}"
CURRENT="${2:-bench/go.txt}"
THRESHOLD="${BENCH_REGRESSION_PCT:-25}"

if [[ -z "$BASELINE" || ! -f "$BASELINE" ]]; then
  echo "No baseline benchmark file; skipping regression gate."
  exit 0
fi

if [[ ! -f "$CURRENT" ]]; then
  echo "Missing current benchmark file: $CURRENT" >&2
  exit 1
fi

python3 - "$BASELINE" "$CURRENT" "$THRESHOLD" <<'PY'
import re
import sys

baseline_path, current_path, threshold_s = sys.argv[1:4]
threshold = float(threshold_s)

def parse(path: str) -> dict[str, float]:
    out: dict[str, float] = {}
    pat = re.compile(r'^(Benchmark\S+)\s+\d+\s+([\d\.]+)\s+ns/op')
    with open(path, encoding='utf-8', errors='replace') as f:
        for line in f:
            m = pat.match(line.strip())
            if m:
                out[m.group(1)] = float(m.group(2))
    return out

base = parse(baseline_path)
cur = parse(current_path)
if not base or not cur:
    print("Benchmark parse produced no entries; skipping gate.")
    sys.exit(0)

regressions = []
for name, base_ns in base.items():
    if name not in cur or base_ns <= 0:
        continue
    cur_ns = cur[name]
    pct = ((cur_ns - base_ns) / base_ns) * 100.0
    if pct > threshold:
        regressions.append((name, base_ns, cur_ns, pct))

if regressions:
    print(f"Benchmark regressions over {threshold:.0f}%:")
    for name, base_ns, cur_ns, pct in sorted(regressions, key=lambda x: -x[3]):
        print(f"  {name}: {base_ns:.0f} -> {cur_ns:.0f} ns/op ({pct:+.1f}%)")
    sys.exit(1)

print(f"Benchmark regression gate passed (threshold {threshold:.0f}%).")
PY
