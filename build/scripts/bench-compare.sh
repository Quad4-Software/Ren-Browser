#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Compare current Go benchmark output against a baseline artifact.
# Fails when a matching benchmark regresses by more than BENCH_REGRESSION_PCT (default 35%).
set -euo pipefail

BASELINE="${1:-}"
CURRENT="${2:-bench/go.txt}"
THRESHOLD="${BENCH_REGRESSION_PCT:-35}"
# Ignore tiny absolute deltas so sub-microsecond benches do not flake CI.
MIN_ABS_NS="${BENCH_REGRESSION_MIN_NS:-1500}"

if [[ -z "$BASELINE" || ! -f "$BASELINE" ]]; then
  echo "No baseline benchmark file; skipping regression gate."
  exit 0
fi

if [[ ! -f "$CURRENT" ]]; then
  echo "Missing current benchmark file: $CURRENT" >&2
  exit 1
fi

python3 - "$BASELINE" "$CURRENT" "$THRESHOLD" "$MIN_ABS_NS" <<'PY'
import re
import statistics
import sys

baseline_path, current_path, threshold_s, min_abs_s = sys.argv[1:5]
threshold = float(threshold_s)
min_abs_ns = float(min_abs_s)

# Strip trailing -N (GOMAXPROCS) so baselines compare across runner shapes.
name_pat = re.compile(r'^(Benchmark\S+?)(?:-\d+)?$')
line_pat = re.compile(r'^(Benchmark\S+)\s+\d+\s+([\d\.]+)\s+ns/op')

def canon(name: str) -> str:
    m = name_pat.match(name)
    return m.group(1) if m else name

def parse(path: str) -> dict[str, list[float]]:
    out: dict[str, list[float]] = {}
    with open(path, encoding='utf-8', errors='replace') as f:
        for line in f:
            m = line_pat.match(line.strip())
            if not m:
                continue
            key = canon(m.group(1))
            out.setdefault(key, []).append(float(m.group(2)))
    return out

def median_ns(samples: list[float]) -> float:
    return float(statistics.median(samples))

base = parse(baseline_path)
cur = parse(current_path)
if not base or not cur:
    print("Benchmark parse produced no entries; skipping gate.")
    sys.exit(0)

regressions = []
for name, base_samples in base.items():
    if name not in cur or not base_samples:
        continue
    cur_samples = cur[name]
    if not cur_samples:
        continue
    base_ns = median_ns(base_samples)
    cur_ns = median_ns(cur_samples)
    if base_ns <= 0:
        continue
    delta = cur_ns - base_ns
    pct = (delta / base_ns) * 100.0
    if pct > threshold and delta > min_abs_ns:
        regressions.append((name, base_ns, cur_ns, pct, delta))

if regressions:
    print(f"Benchmark regressions over {threshold:.0f}% and {min_abs_ns:.0f} ns (median):")
    for name, base_ns, cur_ns, pct, delta in sorted(regressions, key=lambda x: -x[3]):
        print(f"  {name}: {base_ns:.0f} -> {cur_ns:.0f} ns/op ({pct:+.1f}%, +{delta:.0f} ns)")
    sys.exit(1)

print(
    f"Benchmark regression gate passed "
    f"(median, threshold {threshold:.0f}%, min delta {min_abs_ns:.0f} ns)."
)
PY
