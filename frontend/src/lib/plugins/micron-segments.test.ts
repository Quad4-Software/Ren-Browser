// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { isTranslatableText, joinMicronSegments, splitMicronSegments } from "./micron-segments.js";

describe("micron translator segments", () => {
  it("splits text and markup", () => {
    const segments = splitMicronSegments("Hello `>Title` world");
    expect(segments).toEqual([
      { type: "text", value: "Hello " },
      { type: "markup", value: "`>Title`" },
      { type: "text", value: " world" },
    ]);
    expect(joinMicronSegments(segments)).toBe("Hello `>Title` world");
  });

  it("detects translatable text", () => {
    expect(isTranslatableText("   ")).toBe(false);
    expect(isTranslatableText("Hello")).toBe(true);
  });
});
