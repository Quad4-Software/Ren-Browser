// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  formatDocsURL,
  parseDocsURL,
  renderDocsMarkdown,
  renderDocsPage,
  rewriteDocsHref,
  sanitizeDocsPage,
} from "./docs-render";

describe("docs URL helpers", () => {
  it("formats and parses docs urls", () => {
    expect(formatDocsURL("en", "faq")).toBe("docs:?lang=en&page=faq");
    expect(parseDocsURL("docs:?lang=en&page=faq")).toEqual({ lang: "en", page: "faq" });
  });

  it("rejects unsafe page names", () => {
    expect(sanitizeDocsPage("../secrets")).toBe("");
    expect(sanitizeDocsPage("faq.md")).toBe("faq");
  });

  it("rewrites internal markdown links", () => {
    expect(rewriteDocsHref("getting-started.md", "en", "")).toBe(
      "docs:?lang=en&page=getting-started",
    );
    expect(rewriteDocsHref("#section", "en", "faq")).toBe("docs:?lang=en&page=faq#section");
  });

  it("drops external markdown links", () => {
    expect(rewriteDocsHref("https://example.com", "en", "")).toBe("");
  });
});

describe("renderDocsMarkdown", () => {
  it("renders headings and lists", () => {
    const html = renderDocsMarkdown("# Title\n\n- one\n- two", "en", "");
    expect(html).toContain("<h1>Title</h1>");
    expect(html).toContain("<ul>");
    expect(html).toContain("<li>one</li>");
  });

  it("rewrites relative doc links for navigation", () => {
    const html = renderDocsMarkdown("[Install](installation.md)", "en", "faq");
    expect(html).toContain("docs:?lang=en");
    expect(html).toContain("page=installation");
    expect(html).toContain("Install");
  });

  it("renders external references without navigable links", () => {
    const html = renderDocsMarkdown("[Reticulum](https://reticulum.network/)", "en", "");
    expect(html).not.toMatch(/<a[^>]+href="https:\/\/reticulum\.network\/"/);
    expect(html).toContain('class="docs-external-ref"');
    expect(html).toContain("Reticulum");
  });
});

describe("renderDocsPage", () => {
  it("wraps markdown with docs chrome", () => {
    const html = renderDocsPage("# Docs\n\nHello.", "docs:?lang=en");
    expect(html).toContain('class="docs-page"');
    expect(html).toContain('class="docs-body"');
    expect(html).toContain("<h1>Docs</h1>");
    expect(html).toContain("<p>Hello.</p>");
    expect(html).toContain('href="docs:?lang=ru"');
  });
});
