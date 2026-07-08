// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { cycleReaderRotation, readerFontScaleLabel, stepReaderFontScale } from "./reader-layout";

describe("reader layout", () => {
  it("cycles rotation in quarter turns", () => {
    expect(cycleReaderRotation(0)).toBe(90);
    expect(cycleReaderRotation(90)).toBe(180);
    expect(cycleReaderRotation(180)).toBe(270);
    expect(cycleReaderRotation(270)).toBe(0);
  });

  it("steps font scale within bounds", () => {
    expect(stepReaderFontScale(1, 1)).toBe(1.15);
    expect(stepReaderFontScale(1.5, 1)).toBe(1.5);
    expect(stepReaderFontScale(1.15, -1)).toBe(1);
    expect(stepReaderFontScale(0.85, -1)).toBe(0.85);
  });

  it("formats font scale labels", () => {
    expect(readerFontScaleLabel(1)).toBe("100%");
    expect(readerFontScaleLabel(1.15)).toBe("115%");
  });
});
