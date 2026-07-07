import { describe, expect, it } from "vitest";
import { isCompactViewport, MOBILE_LAYOUT_MAX_WIDTH } from "./viewport";

describe("isCompactViewport", () => {
  it("matches the mobile layout breakpoint", () => {
    expect(isCompactViewport(MOBILE_LAYOUT_MAX_WIDTH)).toBe(true);
    expect(isCompactViewport(MOBILE_LAYOUT_MAX_WIDTH + 1)).toBe(false);
  });
});
