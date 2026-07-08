// SPDX-License-Identifier: MIT
import DOMPurify from "dompurify";
import { stripExternalFromCss } from "$lib/micron/nomad-renderer";

const DOCUMENT_FORBID_TAGS = [
  "script",
  "iframe",
  "object",
  "embed",
  "link",
  "base",
  "meta",
  "form",
  "input",
  "button",
  "textarea",
  "select",
  "option",
  "video",
  "audio",
  "source",
  "track",
  "picture",
];

const DOCUMENT_SANITIZE = {
  USE_PROFILES: { html: true },
  FORBID_TAGS: DOCUMENT_FORBID_TAGS,
  ALLOWED_URI_REGEXP: /^(blob:|data:image\/[a-z0-9+.-]+)/i,
};

type DocumentPurify = {
  sanitize: (dirty: string, config?: object) => string;
  addHook: (entryPoint: string, hook: (node: Element) => void) => void;
};

let documentPurify: DocumentPurify | undefined;
let documentPurifyHooksInstalled = false;

function stripReaderThemeFromCss(css: string): string {
  let s = stripExternalFromCss(css);
  s = s.replace(/\b(?:background(?:-color)?|color)\s*:[^;}"']+;?/gi, "");
  return s;
}

function stripReaderThemeFromInlineStyle(style: string): string {
  const s = stripReaderThemeFromCss(style);
  return s
    .split(";")
    .map((part) => part.trim())
    .filter((part) => part && !/^(?:background(?:-color)?|color)\s*:/i.test(part))
    .join("; ");
}

function getDocumentPurify(): DocumentPurify {
  if (!documentPurify) {
    documentPurify = DOMPurify(window) as unknown as DocumentPurify;
    ensureDocumentPurifyHooks(documentPurify);
  }
  return documentPurify;
}

function ensureDocumentPurifyHooks(purify: DocumentPurify): void {
  if (documentPurifyHooksInstalled) {
    return;
  }
  documentPurifyHooksInstalled = true;

  purify.addHook("uponSanitizeElement", (node: Element) => {
    if (node.nodeName === "STYLE" && node.textContent) {
      node.textContent = stripReaderThemeFromCss(node.textContent);
    }
  });

  purify.addHook("afterSanitizeAttributes", (node: Element) => {
    if (node.nodeName === "A" && node.hasAttribute("href")) {
      node.removeAttribute("href");
    }
    if (node.nodeName === "IMG" && node.hasAttribute("src")) {
      const src = node.getAttribute("src") ?? "";
      if (!/^(blob:|data:image\/[a-z0-9+.-]+)/i.test(src)) {
        node.removeAttribute("src");
      }
    }
    const attrs = node.attributes;
    for (let i = attrs.length - 1; i >= 0; i--) {
      const name = attrs[i]?.name;
      if (name && name.toLowerCase().startsWith("on")) {
        node.removeAttribute(name);
      }
    }
    if (node.hasAttribute("bgcolor")) {
      node.removeAttribute("bgcolor");
    }
    if (node.hasAttribute("style")) {
      const style = node.getAttribute("style");
      if (style) {
        const cleaned = stripReaderThemeFromInlineStyle(style);
        if (cleaned) {
          node.setAttribute("style", cleaned);
        } else {
          node.removeAttribute("style");
        }
      }
    }
  });
}

function stripScriptTags(html: string): string {
  return html.replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, "");
}

export function sanitizeDocumentHtml(html: string): string {
  return String(getDocumentPurify().sanitize(stripScriptTags(html), DOCUMENT_SANITIZE));
}
