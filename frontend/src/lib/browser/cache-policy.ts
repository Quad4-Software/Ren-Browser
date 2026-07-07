// SPDX-License-Identifier: MIT

export const defaultMaxAgeMs = 24 * 60 * 60 * 1000;

export function isStale(cachedAtMs: number, now = Date.now(), maxAgeMs = defaultMaxAgeMs): boolean {
  if (!cachedAtMs || maxAgeMs <= 0) {
    return false;
  }
  return now - cachedAtMs > maxAgeMs;
}

export const cache = {
  defaultMaxAgeMs,
  isStale,
};
