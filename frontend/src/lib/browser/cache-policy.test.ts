// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { cache } from "$lib/browser/cache-policy";

describe("cache policy", () => {
  it("treats entries older than the default max age as stale", () => {
    const storedAt = Date.now() - cache.defaultMaxAgeMs - 60_000;
    expect(cache.isStale(storedAt)).toBe(true);
  });

  it("keeps fresh entries", () => {
    expect(cache.isStale(Date.now())).toBe(false);
  });
});
