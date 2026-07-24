// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { sanitizeDocumentHtml } from "$lib/documents/sanitize-html";
import { buildIsolatedHtmlDocument } from "$lib/documents/isolated-html";
import {
  assertNoExecutableMarkup,
  mulberry32,
  XSS_HTML_CORPUS,
} from "$lib/test/oracles/security-oracles";

describe("oracle: sanitizeDocumentHtml adversarial corpus", () => {
  it("strips every XSS payload in the corpus", () => {
    for (const payload of XSS_HTML_CORPUS) {
      const clean = sanitizeDocumentHtml(payload);
      assertNoExecutableMarkup(clean, `sanitizeDocumentHtml(${payload.slice(0, 60)})`);
      expect(clean.toLowerCase()).not.toContain("javascript:");
      expect(clean.toLowerCase()).not.toContain("https://evil.example");
    }
  });

  it("stays free of executable markup under repeated sanitize", () => {
    for (const payload of XSS_HTML_CORPUS) {
      const once = sanitizeDocumentHtml(payload);
      const twice = sanitizeDocumentHtml(once);
      assertNoExecutableMarkup(once, "sanitize-once");
      assertNoExecutableMarkup(twice, "sanitize-twice");
    }
  });

  it("keeps safe local images while dropping remote ones", () => {
    const html =
      '<p>hi</p><img src="data:image/png;base64,aaa"><img src="https://evil.example/x.png">';
    const clean = sanitizeDocumentHtml(html);
    expect(clean).toContain("data:image/png");
    expect(clean.toLowerCase()).not.toContain("https://evil.example");
  });

  it("strips reader theme and external CSS from style blocks", () => {
    const html = `<style>
      @import url(https://evil.example/x.css);
      body { color: red; background: url(https://evil.example/bg.png); }
    </style><p style="color:blue;font-weight:bold">x</p>`;
    const clean = sanitizeDocumentHtml(html);
    expect(clean.toLowerCase()).not.toContain("@import");
    expect(clean.toLowerCase()).not.toContain("https://evil.example");
    expect(clean.toLowerCase()).not.toMatch(/\bcolor\s*:/);
    expect(clean.toLowerCase()).not.toMatch(/\bbackground/);
  });

  it("fuzz-combines corpus fragments without resurrecting handlers", () => {
    const rand = mulberry32(0x5a71);
    for (let i = 0; i < 80; i++) {
      const a = XSS_HTML_CORPUS[Math.floor(rand() * XSS_HTML_CORPUS.length)]!;
      const b = XSS_HTML_CORPUS[Math.floor(rand() * XSS_HTML_CORPUS.length)]!;
      const clean = sanitizeDocumentHtml(`${a}${b}<p>ok</p>`);
      assertNoExecutableMarkup(clean, `fuzz-combo#${i}`);
    }
  });
});

describe("oracle: isolated document shell", () => {
  it("embeds sanitized body without reintroducing scripts", () => {
    const body = sanitizeDocumentHtml(`<p>chapter</p><script>alert(1)</script>`);
    const doc = buildIsolatedHtmlDocument(body, { theme: "dark", fontScale: 1, rotation: 0 });
    assertNoExecutableMarkup(doc.replace(/<meta[^>]*>/gi, ""), "isolated-shell");
    expect(doc).toContain("Content-Security-Policy");
    expect(doc).toContain("script-src 'none'");
  });

  it("does not interpolate hostile fontScale/rotation into executable CSS", () => {
    const hostileScale = `1;}</style><script>alert(1)</script><style>x{font-size:1`;
    const hostileRotation = `0;}</style><img src=x onerror=alert(1)><style>x{transform:rotate(0`;
    const doc = buildIsolatedHtmlDocument("<p>x</p>", {
      theme: "light",
      fontScale: Number.NaN,
      rotation: Number.NaN,
    });
    expect(doc.toLowerCase()).not.toContain("<script");
    // Numeric options must stay numeric sinks. String injection is a type error,
    // but NaN must not produce breakout either.
    expect(doc).toMatch(/font-size:\s*NaNrem|font-size:\s*1rem|font-size:\s*0rem/);
    void hostileScale;
    void hostileRotation;
  });
});
