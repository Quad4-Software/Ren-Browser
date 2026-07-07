import { describe, expect, it } from "vitest";
import { previewScaleForBox, wrapPreviewSrcdoc } from "./preview-srcdoc";

describe("preview srcdoc", () => {
  it("scales iframe to fill preview width", () => {
    expect(previewScaleForBox(280)).toBeCloseTo(280 / 1280, 5);
    expect(previewScaleForBox(0)).toBe(0.22);
  });

  it("wraps fragment html in a fixed-width shell", () => {
    const out = wrapPreviewSrcdoc("<p>Hello</p>", { fg: "#fff", bg: "#111" });
    expect(out).toContain("<!DOCTYPE html>");
    expect(out).toContain("width:1280px");
    expect(out).toContain("<p>Hello</p>");
    expect(out).toContain("background:#111");
  });

  it("defaults wrapped shell to micron dark colors", () => {
    const out = wrapPreviewSrcdoc("<p>Hello</p>");
    expect(out).toContain("background:#000000");
    expect(out).toContain("color:#ffffff");
  });

  it("leaves full documents unchanged", () => {
    const doc = "<!DOCTYPE html><html><body>ok</body></html>";
    expect(wrapPreviewSrcdoc(doc)).toBe(doc);
  });
});
