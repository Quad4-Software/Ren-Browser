// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  isolateNomadLinksInHtml,
  sanitizeNomadHtmlDocument,
  sanitizeNomadHtmlFragment,
  stripExternalFromCss,
} from "$lib/micron/nomad-renderer";
import MicronParser, { parseMicronHeaderTagColors } from "$lib/micron/parser";
import {
  assertNoExecutableMarkup,
  mulberry32,
  NON_HEX_HEADER_COLORS,
  XSS_HTML_CORPUS,
} from "$lib/test/oracles/security-oracles";

const NODE = "abb3ebcd03cb2388a838e70c001291f9";

describe("oracle: nomad HTML sanitize", () => {
  it("strips XSS corpus from fragments and documents", () => {
    for (const payload of XSS_HTML_CORPUS) {
      const frag = sanitizeNomadHtmlFragment(payload);
      const doc = sanitizeNomadHtmlDocument(`<!DOCTYPE html><html><body>${payload}</body></html>`);
      assertNoExecutableMarkup(frag, `nomad-fragment(${payload.slice(0, 40)})`);
      assertNoExecutableMarkup(doc, `nomad-document(${payload.slice(0, 40)})`);
    }
  });

  it("blocks external http(s) urls in CSS", () => {
    const css = `
      @import url("https://evil.example/a.css");
      .x { background: url(https://evil.example/b.png); color: expression(alert(1)); }
      .y { behavior: url(https://evil.example/x.htc); -moz-binding: url(https://evil.example/x.xml#x); }
    `;
    const cleaned = stripExternalFromCss(css).toLowerCase();
    expect(cleaned).not.toContain("https://evil.example");
    expect(cleaned).not.toContain("@import");
    expect(cleaned).not.toContain("expression(");
    expect(cleaned).not.toContain("-moz-binding");
  });

  it("isolates links so external hrefs cannot navigate", () => {
    const html = `
      <a href="https://evil.example">ext</a>
      <a href="/page/index.mu">local</a>
      <a href="${NODE}:/page/x.mu">abs</a>
      <a href="javascript:alert(1)">js</a>
    `;
    const isolated = isolateNomadLinksInHtml(html, NODE);
    expect(isolated.toLowerCase()).not.toContain('href="https://');
    expect(isolated.toLowerCase()).not.toContain("javascript:");
    expect(isolated).toContain(`data-nomad-url="${NODE}:/page/index.mu"`);
    expect(isolated).toContain(`data-nomad-url="${NODE}:/page/x.mu"`);
  });
});

describe("oracle: micron header colors are hex-only (parser contract)", () => {
  it("documents upstream length-only check accepts non-hex 3/6 char tokens", () => {
    // This is the known weak contract in micron-parser. RenBrowser sinks must
    // not trust these values. The test fails loudly if upstream starts validating
    // hex (then we can delete the sink hardening cautiously).
    const accepted: string[] = [];
    for (const dirty of NON_HEX_HEADER_COLORS) {
      const source = `#!fg=${dirty}\n#!bg=${dirty}\nHello`;
      const colors = parseMicronHeaderTagColors(source, true);
      if (colors.fg === dirty || colors.bg === dirty) {
        accepted.push(dirty);
      }
    }
    expect(accepted.length).toBeGreaterThan(0);
  });

  it("stripOverlayStyles removes fixed overlays that could cover chrome", () => {
    const attacks = [
      "position:fixed;top:0;left:0;right:0;bottom:0;z-index:2147483647",
      "position: FIXED; inset: 0; z-index: 99999",
      "position:fixed;z-index:9;width:100vw;height:100vh",
    ];
    for (const style of attacks) {
      const dirty = `<span style="${style}">cover</span>`;
      const clean = MicronParser.stripOverlayStyles(dirty).toLowerCase();
      expect(clean).not.toContain("position:fixed");
      expect(clean).not.toContain("position: fixed");
      expect(clean).not.toContain("z-index");
    }
  });

  it("fuzz micron markup through stripOverlayStyles without throwing", () => {
    const rand = mulberry32(99);
    for (let i = 0; i < 50; i++) {
      const styleBits = ["position:fixed", "z-index:9999", "top:0", "color:red", "background:blue"];
      const picked = styleBits.filter(() => rand() > 0.4).join(";");
      const html = `<div style="${picked}">${"x".repeat(Math.floor(rand() * 20))}</div>`;
      expect(() => MicronParser.stripOverlayStyles(html)).not.toThrow();
    }
  });
});
