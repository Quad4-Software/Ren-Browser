// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { resolveLinkURL } from "./micron-links";

describe("resolveLinkURL", () => {
  it("resolves relative micron page links", () => {
    expect(resolveLinkURL("abc:/page/index.mu", "/page/about.mu")).toBe("abc:/page/about.mu");
  });

  it("prefixes bare page names", () => {
    expect(resolveLinkURL("abc:/page/index.mu", "guide.mu")).toBe("abc:/page/guide.mu");
  });
});
