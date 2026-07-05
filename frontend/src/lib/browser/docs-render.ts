// SPDX-License-Identifier: MIT
import { marked } from "marked";

const SUPPORTED_LANGS = ["en", "ru", "es", "de"] as const;
const PAGE_NAME_RE = /^[a-z0-9-]+$/;

marked.use({
  gfm: true,
  breaks: false,
  async: false,
});

function normalizeMarkdownInput(md: string): string {
  let s = String(md ?? "").replace(/\r\n/g, "\n");
  s = s.replace(/^(\s{0,3})(#{1,6})([^\s#])/gm, "$1$2 $3");
  return s;
}

export function sanitizeDocsPage(page: string): string {
  const trimmed = page.trim().toLowerCase().replace(/\.md$/i, "");
  if (!trimmed || trimmed === "readme") {
    return "";
  }
  return PAGE_NAME_RE.test(trimmed) ? trimmed : "";
}

export function parseDocsURL(url: string): { lang: string; page: string } {
  const idx = url.indexOf("?");
  if (idx < 0) {
    return { lang: "", page: "" };
  }
  const params = new URLSearchParams(url.slice(idx + 1));
  return {
    lang: (params.get("lang") ?? "").trim().toLowerCase(),
    page: sanitizeDocsPage(params.get("page") ?? ""),
  };
}

export function formatDocsURL(lang: string, page = ""): string {
  if (!lang) {
    return "docs:";
  }
  const params = new URLSearchParams({ lang });
  if (page) {
    params.set("page", page);
  }
  return `docs:?${params.toString()}`;
}

export function rewriteDocsHref(href: string, lang: string, currentPage: string): string {
  const trimmed = href.trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.startsWith("#")) {
    return formatDocsURL(lang, currentPage) + trimmed;
  }
  if (/^https?:\/\//i.test(trimmed)) {
    return "";
  }
  const path = trimmed.replace(/^\.\//, "");
  if (path.includes("://") || path.startsWith("..")) {
    return formatDocsURL(lang, "");
  }
  const base = path.split("/").pop() ?? path;
  if (base === "README.md" || base === "." || base === "/") {
    return formatDocsURL(lang, "");
  }
  const page = sanitizeDocsPage(base.replace(/\.md$/i, ""));
  return formatDocsURL(lang, page);
}

function docsLangLabel(lang: string): string {
  switch (lang) {
    case "en":
      return "English";
    case "ru":
      return "Russian";
    case "es":
      return "Spanish";
    case "de":
      return "German";
    default:
      return lang.toUpperCase();
  }
}

function docsLangSwitcher(current: string): string {
  const links = SUPPORTED_LANGS.filter((lang) => lang !== current).map(
    (lang) => `<a href="${formatDocsURL(lang)}">${docsLangLabel(lang)}</a>`,
  );
  if (links.length === 0) {
    return "";
  }
  return `<span class="docs-lang-switch"> · ${links.join(" · ")}</span>`;
}

function isAllowedDocsHref(href: string | null): boolean {
  if (!href) {
    return false;
  }
  const h = href.trim().toLowerCase();
  if (h.startsWith("#")) {
    return true;
  }
  return h.startsWith("docs:") || h.startsWith("docs?");
}

function parseMarkdownHtml(markdown: string, lang: string, currentPage: string): string {
  const renderer = new marked.Renderer();
  renderer.link = ({ href, title, text }) => {
    const raw = (href ?? "").trim();
    const titleAttr = title ? ` title="${title.replace(/"/g, "&quot;")}"` : "";
    if (/^https?:\/\//i.test(raw)) {
      const escapedHref = raw.replace(/&/g, "&amp;").replace(/"/g, "&quot;");
      return `<span class="docs-external-ref"${titleAttr} data-href="${escapedHref}">${text}</span>`;
    }
    const target = rewriteDocsHref(raw, lang, currentPage);
    if (!isAllowedDocsHref(target)) {
      return text;
    }
    const escapedHref = target.replace(/&/g, "&amp;").replace(/"/g, "&quot;");
    return `<a href="${escapedHref}"${titleAttr}>${text}</a>`;
  };

  const parsed = marked.parse(normalizeMarkdownInput(markdown), { renderer, async: false });
  return typeof parsed === "string" ? parsed : "";
}

export function renderDocsMarkdown(markdown: string, lang: string, currentPage: string): string {
  return parseMarkdownHtml(markdown, lang, currentPage);
}

export function renderDocsPage(markdown: string, currentURL: string): string {
  const { lang, page } = parseDocsURL(currentURL);
  const body = parseMarkdownHtml(markdown, lang, page);
  const nav =
    `<nav class="docs-nav"><a href="${formatDocsURL(lang)}">Index</a>` +
    docsLangSwitcher(lang) +
    `</nav>`;
  return `<article class="docs-page">${nav}<div class="docs-body">${body}</div></article>`;
}
