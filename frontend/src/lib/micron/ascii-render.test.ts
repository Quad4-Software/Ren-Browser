// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { renderClientMicronPage } from "./render-page";

function plainText(html: string): string {
  return html.replace(/<[^>]+>/g, "");
}

describe("micron ASCII / force-monospace regressions", () => {
  const art = ["+-----+", "| ASCII |", "+-----+"].join("\n");
  const url = "abb3ebcd03cb2388a838e70c001291f9:/page/art.mu";

  it("always emits monospace cells for ASCII art", () => {
    const html = renderClientMicronPage(url, art, "js");
    expect(html).toMatch(/Mu-mnt/);
    expect(plainText(html)).toContain("ASCII");
  });

  it("keeps force-monospace even when preserveLayout is false", () => {
    const html = renderClientMicronPage(url, art, "js", { preserveLayout: false });
    expect(html).toMatch(/Mu-mnt/);
    expect(plainText(html)).toContain("ASCII");
  });

  it("keeps column characters present after render", () => {
    const html = renderClientMicronPage(url, "|===|", "js");
    const plain = plainText(html);
    expect(plain).toContain("|");
    expect(plain).toContain("=");
  });
});
