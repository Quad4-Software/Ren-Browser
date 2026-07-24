// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  expandHexColor,
  micronPageColors,
  micronShellStyle,
  normalizeReticulumURL,
} from "$lib/browser/url";
import { wrapPreviewSrcdoc } from "$lib/browser/preview-srcdoc";
import { isAllowedNavigationURL, isBlockedNavigationURL } from "$lib/browser/navigation-guard";
import { pageDownloadName } from "$lib/browser/download";
import {
  ALLOWED_NAV_CORPUS,
  assertCssColorTokenSafe,
  assertNoStyleBreakout,
  assertSafeDownloadFilename,
  BLOCKED_NAV_CORPUS,
  CSS_BREAKOUT_COLORS,
  DOWNLOAD_NAME_ATTACKS,
  mulberry32,
  NON_HEX_HEADER_COLORS,
  randomHexish,
} from "$lib/test/oracles/security-oracles";

describe("oracle: micron page colors must be CSS-safe", () => {
  it("rejects non-hex colors that micron parsers accept by length alone", () => {
    for (const dirty of NON_HEX_HEADER_COLORS) {
      const colors = micronPageColors(dirty, dirty);
      assertCssColorTokenSafe(colors.fg, `micronPageColors.fg(${dirty})`);
      assertCssColorTokenSafe(colors.bg, `micronPageColors.bg(${dirty})`);
      expect(colors.fg).toBe("#ffffff");
      expect(colors.bg).toBe("#000000");
    }
  });

  it("expandHexColor must not echo attacker-controlled non-hex", () => {
    for (const dirty of CSS_BREAKOUT_COLORS) {
      const expanded = expandHexColor(dirty);
      expect(expanded === "" || /^[0-9a-fA-F]{6}$/.test(expanded)).toBe(true);
    }
  });

  it("micronShellStyle never embeds raw quotes or style closers", () => {
    for (const dirty of CSS_BREAKOUT_COLORS) {
      const style = micronShellStyle("micron", dirty, dirty);
      expect(style).not.toMatch(/<\/style/i);
      expect(style).not.toContain('"');
      expect(style).not.toContain("'");
      expect(style).toMatch(/^background:#[0-9a-fA-F]{6};color:#[0-9a-fA-F]{6}$/);
    }
  });
});

describe("oracle: wrapPreviewSrcdoc resists color / markup breakout", () => {
  it("keeps adversarial colors inside the style block", () => {
    for (const dirty of CSS_BREAKOUT_COLORS) {
      const doc = wrapPreviewSrcdoc("<p>hi</p>", { fg: dirty, bg: dirty });
      assertNoStyleBreakout(doc, `wrapPreviewSrcdoc(${dirty})`);
      expect(doc.toLowerCase()).not.toContain("<script");
    }
  });

  it("does not pass through full documents with executable markup unchecked", () => {
    const poisoned = `<!DOCTYPE html><html><body><script>window.__xss=1</script><p>ok</p></body></html>`;
    const out = wrapPreviewSrcdoc(poisoned);
    expect(out.toLowerCase()).not.toContain("<script");
  });

  it("fuzzes hexish header colors into srcdoc without breakout", () => {
    const rand = mulberry32(0xc0ffee);
    for (let i = 0; i < 200; i++) {
      const fg = randomHexish(rand, rand() > 0.5 ? 3 : 6);
      const bg = randomHexish(rand, rand() > 0.5 ? 3 : 6);
      const colors = micronPageColors(fg, bg);
      const doc = wrapPreviewSrcdoc("<div>preview</div>", colors);
      assertNoStyleBreakout(doc, `fuzz#${i} fg=${fg} bg=${bg}`);
      assertCssColorTokenSafe(colors.fg, `fuzz.fg#${i}`);
      assertCssColorTokenSafe(colors.bg, `fuzz.bg#${i}`);
    }
  });
});

describe("oracle: navigation allow/block vs normalize consistency", () => {
  it("blocked external URLs normalize to empty and stay blocked", () => {
    for (const url of BLOCKED_NAV_CORPUS) {
      expect(isBlockedNavigationURL(url), url).toBe(true);
      expect(isAllowedNavigationURL(url), url).toBe(false);
      expect(normalizeReticulumURL(url), url).toBe("");
    }
  });

  it("known allowed mesh/special URLs are not blocked", () => {
    for (const url of ALLOWED_NAV_CORPUS) {
      expect(isBlockedNavigationURL(url), url).toBe(false);
      expect(isAllowedNavigationURL(url), url).toBe(true);
      expect(normalizeReticulumURL(url).length, url).toBeGreaterThan(0);
    }
  });

  it("scheme-like smuggling cannot sneak past the guard", () => {
    const smugglers = [
      " https://evil.example",
      "\thttps://evil.example",
      "https://evil.example\n",
      "rns:https://evil.example",
      "http://abb3ebcd03cb2388a838e70c001291f9:/page/index.mu",
      "file:abb3ebcd03cb2388a838e70c001291f9:/page/index.mu",
      "javascript:abb3ebcd03cb2388a838e70c001291f9:/page/index.mu",
      "data:text/html,abb3ebcd03cb2388a838e70c001291f9:/page/index.mu",
    ];
    for (const url of smugglers) {
      expect(isAllowedNavigationURL(url) && normalizeReticulumURL(url) !== "", url).toBe(false);
    }
  });

  it("fuzz: random scheme prefixes never become allowed absolute web URLs", () => {
    const rand = mulberry32(42);
    const schemes = [
      "http",
      "https",
      "ftp",
      "file",
      "javascript",
      "data",
      "blob",
      "ws",
      "wss",
      "mailto",
      "tel",
      "vbscript",
    ];
    for (let i = 0; i < 100; i++) {
      const scheme = schemes[Math.floor(rand() * schemes.length)]!;
      const url = `${scheme}://${rand().toString(36)}.example/${i}`;
      expect(isBlockedNavigationURL(url), url).toBe(true);
      expect(normalizeReticulumURL(url), url).toBe("");
    }
  });
});

describe("oracle: download filenames cannot traverse", () => {
  it("rejects path traversal and drive letters from a= query attacks", () => {
    for (const attack of DOWNLOAD_NAME_ATTACKS) {
      const name = pageDownloadName(attack.url, attack.contentType);
      assertSafeDownloadFilename(name, attack.url);
    }
  });
});
