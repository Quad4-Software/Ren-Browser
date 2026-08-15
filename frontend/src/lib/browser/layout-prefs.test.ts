// SPDX-License-Identifier: MIT
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SIDEBAR_DEFAULT_PX, SIDEBAR_MIN_PX } from "./pane-resize";
import { readSidebarWidth, writeSidebarWidth } from "./layout-prefs";

const storage = new Map<string, string>();

beforeEach(() => {
  storage.clear();
  vi.stubGlobal("localStorage", {
    getItem: (key: string) => storage.get(key) ?? null,
    setItem: (key: string, value: string) => {
      storage.set(key, value);
    },
    removeItem: (key: string) => {
      storage.delete(key);
    },
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("layout prefs", () => {
  it("returns the default width when nothing is stored", () => {
    expect(readSidebarWidth()).toBe(SIDEBAR_DEFAULT_PX);
  });

  it("round-trips a clamped sidebar width", () => {
    writeSidebarWidth(410);
    expect(readSidebarWidth()).toBe(410);
    writeSidebarWidth(80);
    expect(readSidebarWidth()).toBe(SIDEBAR_MIN_PX);
  });
});
