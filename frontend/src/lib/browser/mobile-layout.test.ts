import { describe, expect, it } from "vitest";
import { resolveMobileUI } from "./mobile-layout";
import { isCompactViewport, MOBILE_LAYOUT_MAX_WIDTH } from "./viewport";

describe("resolveMobileUI", () => {
  it("honours screenshot layout overrides", () => {
    expect(
      resolveMobileUI({
        layoutOverride: "mobile",
        isMobilePlatform: false,
        compactViewport: false,
      }),
    ).toBe(true);
    expect(
      resolveMobileUI({
        layoutOverride: "desktop",
        isMobilePlatform: true,
        compactViewport: true,
      }),
    ).toBe(false);
  });

  it("uses platform or compact viewport when not overridden", () => {
    expect(
      resolveMobileUI({
        layoutOverride: null,
        isMobilePlatform: true,
        compactViewport: false,
      }),
    ).toBe(true);
    expect(
      resolveMobileUI({
        layoutOverride: null,
        isMobilePlatform: false,
        compactViewport: true,
      }),
    ).toBe(true);
    expect(
      resolveMobileUI({
        layoutOverride: null,
        isMobilePlatform: false,
        compactViewport: false,
      }),
    ).toBe(false);
  });
});

describe("mobile layout breakpoint", () => {
  it("aligns compact viewport with the mobile UI width gate", () => {
    expect(isCompactViewport(MOBILE_LAYOUT_MAX_WIDTH)).toBe(true);
    expect(isCompactViewport(MOBILE_LAYOUT_MAX_WIDTH + 1)).toBe(false);
  });
});
