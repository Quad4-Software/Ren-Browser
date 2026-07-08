// SPDX-License-Identifier: MIT
import JSZip from "jszip";
import { sanitizeDocumentHtml } from "$lib/documents/sanitize-html";

export type EpubChapter = {
  id: string;
  title: string;
  html: string;
};

export type EpubBook = {
  title: string;
  chapters: EpubChapter[];
  blobUrls: string[];
};

function textContent(el: Element | null | undefined): string {
  return el?.textContent?.trim() ?? "";
}

function elementsByLocalName(root: ParentNode, localName: string): Element[] {
  const want = localName.toLowerCase();
  const nodes =
    root instanceof Document || root instanceof Element ? root.getElementsByTagName("*") : [];
  return [...nodes].filter((el) => el.localName.toLowerCase() === want);
}

function firstByLocalName(root: ParentNode, localName: string): Element | null {
  return elementsByLocalName(root, localName)[0] ?? null;
}

function readContainerRootPath(containerXml: string, container: Document): string {
  const rootfile = firstByLocalName(container, "rootfile");
  const fromDom = rootfile?.getAttribute("full-path")?.trim();
  if (fromDom) {
    return fromDom;
  }
  const match = containerXml.match(/full-path\s*=\s*["']([^"']+)["']/i);
  return match?.[1]?.trim() ?? "";
}

function childElementsByLocalName(parent: Element, localName: string): Element[] {
  const want = localName.toLowerCase();
  return [...parent.children].filter((child) => child.localName.toLowerCase() === want);
}

function resolveZipPath(base: string, href: string): string {
  if (href.startsWith("/")) {
    return href.slice(1);
  }
  const baseDir = base.includes("/") ? base.slice(0, base.lastIndexOf("/") + 1) : "";
  const parts = (baseDir + href).split("/");
  const out: string[] = [];
  for (const part of parts) {
    if (!part || part === ".") {
      continue;
    }
    if (part === "..") {
      out.pop();
      continue;
    }
    out.push(part);
  }
  return out.join("/");
}

function normalizeEpubPath(path: string): string {
  const bare = path.split("#")[0] ?? path;
  return bare.replace(/\\/g, "/").replace(/^\/+/, "").trim();
}

function parseNcxToc(ncxXml: string): Map<string, string> {
  const titles = new Map<string, string>();
  const doc = new DOMParser().parseFromString(ncxXml, "application/xml");
  for (const point of elementsByLocalName(doc, "navPoint")) {
    const label = textContent(firstByLocalName(point, "text"));
    const content = firstByLocalName(point, "content");
    const src = content?.getAttribute("src") ?? "";
    const path = normalizeEpubPath(src);
    if (label && path) {
      titles.set(path, label);
    }
  }
  if (titles.size === 0) {
    const pattern =
      /<navLabel>\s*<(?:\w+:)?text>([^<]+)<\/(?:\w+:)?text>[\s\S]*?<content[^>]+src="([^"]+)"/gi;
    for (const match of ncxXml.matchAll(pattern)) {
      const label = match[1]?.trim() ?? "";
      const path = normalizeEpubPath(match[2] ?? "");
      if (label && path) {
        titles.set(path, label);
      }
    }
  }
  return titles;
}

function parseNavToc(navHtml: string, opfDir: string): Map<string, string> {
  const titles = new Map<string, string>();
  const doc = new DOMParser().parseFromString(navHtml, "application/xhtml+xml");
  const navRoots = [...elementsByLocalName(doc, "nav")];
  const tocNav =
    navRoots.find((nav) => {
      const type = nav.getAttribute("epub:type") ?? nav.getAttribute("type") ?? "";
      return type.toLowerCase().split(/\s+/).includes("toc");
    }) ?? navRoots[0];
  if (!tocNav) {
    return titles;
  }
  for (const link of elementsByLocalName(tocNav, "a")) {
    const href = link.getAttribute("href");
    const label = textContent(link);
    if (!href || !label) {
      continue;
    }
    titles.set(normalizeEpubPath(resolveZipPath(opfDir, href)), label);
  }
  return titles;
}

function lookupTocTitle(titles: Map<string, string>, chapterPath: string, href: string): string {
  const candidates = new Set([
    normalizeEpubPath(chapterPath),
    normalizeEpubPath(href),
    normalizeEpubPath(resolveZipPath("", href)),
  ]);
  const leaf = href.split("/").pop() ?? href;
  candidates.add(normalizeEpubPath(leaf));
  for (const key of candidates) {
    const hit = titles.get(key);
    if (hit) {
      return hit;
    }
  }
  for (const [key, label] of titles) {
    const keyLeaf = key.split("/").pop() ?? key;
    if (keyLeaf === leaf || key.endsWith(`/${leaf}`)) {
      return label;
    }
  }
  return "";
}

function headingFromBody(doc: Document): string {
  const body = doc.body;
  if (!body) {
    return "";
  }
  for (const tag of ["h1", "h2", "h3", "h4", "h5", "h6"]) {
    const text = textContent(body.querySelector(tag));
    if (text) {
      return text;
    }
  }
  return "";
}

function titleFromFilename(href: string): string {
  const base =
    href
      .split("/")
      .pop()
      ?.replace(/\.[^.]+$/i, "") ?? "";
  const numbered = base.match(/(?:^|_)c(\d{1,3})(?:_|$)/i);
  if (numbered) {
    return `Chapter ${Number(numbered[1])}`;
  }
  const chapterWord = base.match(/chapter[\s._-]*(\d+)/i);
  if (chapterWord) {
    return `Chapter ${Number(chapterWord[1])}`;
  }
  return base.replace(/[_-]+/g, " ").replace(/\s+/g, " ").trim();
}

async function loadTocTitles(
  zip: JSZip,
  opfDir: string,
  manifest: Map<string, { href: string; mediaType: string }>,
  manifestEl: Element,
  spineEl: Element,
): Promise<Map<string, string>> {
  const titles = new Map<string, string>();

  for (const item of childElementsByLocalName(manifestEl, "item")) {
    const properties = item.getAttribute("properties") ?? "";
    if (!properties.split(/\s+/).includes("nav")) {
      continue;
    }
    const href = item.getAttribute("href");
    if (!href) {
      continue;
    }
    const navHtml = await readZipText(zip, resolveZipPath(opfDir, href));
    for (const [key, label] of parseNavToc(navHtml, opfDir)) {
      titles.set(key, label);
    }
  }

  const tocId = spineEl.getAttribute("toc");
  if (tocId) {
    const entry = manifest.get(tocId);
    if (entry) {
      const ncxXml = await readZipText(zip, resolveZipPath(opfDir, entry.href));
      for (const [key, label] of parseNcxToc(ncxXml)) {
        titles.set(key, label);
      }
    }
  }

  if (titles.size === 0) {
    for (const entry of manifest.values()) {
      if (!entry.mediaType.includes("dtbncx") && !entry.href.toLowerCase().endsWith(".ncx")) {
        continue;
      }
      const ncxXml = await readZipText(zip, resolveZipPath(opfDir, entry.href));
      for (const [key, label] of parseNcxToc(ncxXml)) {
        titles.set(key, label);
      }
    }
  }

  return titles;
}

function chapterTitle(
  doc: Document,
  chapterPath: string,
  href: string,
  tocTitles: Map<string, string>,
  index: number,
): string {
  return (
    lookupTocTitle(tocTitles, chapterPath, href) ||
    headingFromBody(doc) ||
    titleFromFilename(href) ||
    `Chapter ${index + 1}`
  );
}

async function readZipText(zip: JSZip, path: string): Promise<string> {
  const file = zip.file(path);
  if (!file) {
    throw new Error(`missing epub file: ${path}`);
  }
  return await file.async("text");
}

async function readZipBytes(zip: JSZip, path: string): Promise<Uint8Array> {
  const file = zip.file(path);
  if (!file) {
    throw new Error(`missing epub asset: ${path}`);
  }
  return await file.async("uint8array");
}

function mimeForPath(path: string): string {
  const lower = path.toLowerCase();
  if (lower.endsWith(".png")) {
    return "image/png";
  }
  if (lower.endsWith(".jpg") || lower.endsWith(".jpeg")) {
    return "image/jpeg";
  }
  if (lower.endsWith(".gif")) {
    return "image/gif";
  }
  if (lower.endsWith(".svg")) {
    return "image/svg+xml";
  }
  if (lower.endsWith(".webp")) {
    return "image/webp";
  }
  return "application/octet-stream";
}

async function resolveChapterImages(
  html: string,
  chapterDir: string,
  zip: JSZip,
  blobCache: Map<string, string>,
  blobUrls: string[],
): Promise<string> {
  const doc = new DOMParser().parseFromString(html, "text/html");
  const images = doc.querySelectorAll("img[src]");
  for (const img of images) {
    const src = img.getAttribute("src");
    if (!src || src.startsWith("blob:") || src.startsWith("data:")) {
      continue;
    }
    if (/^[a-z][a-z0-9+.-]*:/i.test(src.trim())) {
      img.removeAttribute("src");
      continue;
    }
    const assetPath = resolveZipPath(chapterDir, src);
    let blobUrl = blobCache.get(assetPath);
    if (!blobUrl) {
      const bytes = await readZipBytes(zip, assetPath);
      blobUrl = URL.createObjectURL(
        new Blob([new Uint8Array(bytes)], { type: mimeForPath(assetPath) }),
      );
      blobCache.set(assetPath, blobUrl);
      blobUrls.push(blobUrl);
    }
    img.setAttribute("src", blobUrl);
  }
  return sanitizeDocumentHtml(doc.body?.innerHTML ?? html);
}

export async function parseEpub(data: Uint8Array): Promise<EpubBook> {
  const zip = await JSZip.loadAsync(data);
  const containerXml = await readZipText(zip, "META-INF/container.xml");
  const container = new DOMParser().parseFromString(containerXml, "application/xml");
  const opfPath = readContainerRootPath(containerXml, container);
  if (!opfPath) {
    throw new Error("epub rootfile not found");
  }

  const opfXml = await readZipText(zip, opfPath);
  const opf = new DOMParser().parseFromString(opfXml, "application/xml");
  const packageEl = opf.documentElement;
  if (!packageEl) {
    throw new Error("epub package root not found");
  }
  const opfDir = opfPath.includes("/") ? opfPath.slice(0, opfPath.lastIndexOf("/") + 1) : "";

  const manifest = new Map<string, { href: string; mediaType: string }>();
  const manifestEl = childElementsByLocalName(packageEl, "manifest")[0];
  if (!manifestEl) {
    throw new Error("epub manifest not found");
  }
  for (const item of childElementsByLocalName(manifestEl, "item")) {
    const id = item.getAttribute("id");
    const href = item.getAttribute("href");
    if (!id || !href) {
      continue;
    }
    manifest.set(id, {
      href,
      mediaType: item.getAttribute("media-type") ?? "",
    });
  }

  const spineEl = childElementsByLocalName(packageEl, "spine")[0];
  if (!spineEl) {
    throw new Error("epub spine not found");
  }
  const tocTitles = await loadTocTitles(zip, opfDir, manifest, manifestEl, spineEl);
  const spineIds: string[] = [];
  for (const itemref of childElementsByLocalName(spineEl, "itemref")) {
    const idref = itemref.getAttribute("idref");
    if (idref) {
      spineIds.push(idref);
    }
  }

  const metadataEl = childElementsByLocalName(packageEl, "metadata")[0];
  const title =
    (metadataEl ? textContent(childElementsByLocalName(metadataEl, "title")[0]) : "") || "EPUB";

  const blobCache = new Map<string, string>();
  const blobUrls: string[] = [];
  const chapters: EpubChapter[] = [];
  for (const id of spineIds) {
    const entry = manifest.get(id);
    if (!entry) {
      continue;
    }
    const chapterPath = resolveZipPath(opfDir, entry.href);
    const chapterDir = chapterPath.includes("/")
      ? chapterPath.slice(0, chapterPath.lastIndexOf("/") + 1)
      : "";
    const rawHtml = await readZipText(zip, chapterPath);
    const doc = new DOMParser().parseFromString(rawHtml, "application/xhtml+xml");
    const heading = chapterTitle(doc, chapterPath, entry.href, tocTitles, chapters.length);
    const bodyHtml = doc.body?.innerHTML ?? rawHtml;
    chapters.push({
      id,
      title: heading,
      html: await resolveChapterImages(bodyHtml, chapterDir, zip, blobCache, blobUrls),
    });
  }

  if (chapters.length === 0) {
    throw new Error("epub has no readable chapters");
  }

  return { title, chapters, blobUrls };
}

export function revokeEpubBlobUrls(book: EpubBook): void {
  const seen = new Set<string>();
  for (const url of book.blobUrls) {
    if (!seen.has(url)) {
      seen.add(url);
      URL.revokeObjectURL(url);
    }
  }
}
