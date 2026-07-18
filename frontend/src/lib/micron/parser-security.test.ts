// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import MicronParser from "./parser";
import { LARGE_MICRON_RAW_BYTES, resolveEffectiveMicronEngine } from "./render-page";

describe("micron parser security surfaces", () => {
  it("prefers Go HTML for large pages to avoid main-thread re-parse", () => {
    const engine = resolveEffectiveMicronEngine("auto", {
      wasmEnabled: true,
      wasmAvailable: true,
      wasmReady: true,
      hasServerHtml: true,
      rawBytes: LARGE_MICRON_RAW_BYTES,
    });
    expect(engine).toBe("go");
  });

  it("detects box-drawing lines as needing per-char cells", () => {
    expect(MicronParser.lineNeedsPerCharCells("ABC")).toBe(false);
    expect(MicronParser.lineNeedsPerCharCells("┌─┐")).toBe(true);
  });

  it("strips fixed overlay styles that could cover the UI", () => {
    const dirty = '<span style="position: fixed; top: 0; z-index: 9999; color: red">x</span>';
    const clean = MicronParser.stripOverlayStyles(dirty);
    expect(clean.toLowerCase()).not.toContain("position: fixed");
    expect(clean.toLowerCase()).not.toContain("z-index");
    expect(clean.toLowerCase()).not.toContain("top:");
  });

  it("documents Latin lines do not need per-char ForceMonospace cells", () => {
    // Go RenderDark always emits one Mu-mnt span per rune (~29x HTML).
    // JS can group Latin runs when lineNeedsPerCharCells is false.
    const latin = "A".repeat(4096);
    expect(MicronParser.lineNeedsPerCharCells(latin)).toBe(false);
  });
});
