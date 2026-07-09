// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { defaultTheme } from "./tokens";

describe("defaultTheme", () => {
  it("uses dark mode and accent token defaults", () => {
    const theme = defaultTheme();
    expect(theme.mode).toBe("dark");
    expect(theme.accent).toBe("#60a5fa");
    expect(theme.fontSize).toBe(14);
    expect(theme.compactToolbar).toBe(false);
    expect(theme.overlaySidebars).toBe(false);
  });
});
