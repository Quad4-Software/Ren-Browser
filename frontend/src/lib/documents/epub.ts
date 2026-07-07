// SPDX-License-Identifier: MIT
import DOMPurify from "dompurify";
import JSZip from "jszip";

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

const EPUB_SANITIZE = {
  USE_PROFILES: { html: true },
  FORBID_TAGS: ["script", "iframe", "object", "embed", "link", "base", "meta"],
  FORBID_ATTR: ["onerror", "onload", "onclick", "onmouseover"],
  ALLOWED_URI_REGEXP: /^(blob:|data:image\/[a-z0-9+.-]+)/i,
};

function textContent(el: Element | null | undefined): string {
  return el?.textContent?.trim() ?? "";
}

function elementsByLocalName(root: ParentNode, localName: string): Element[] {
  const want = localName.toLowerCase();
  const nodes =
    root instanceof Document || root instanceof Element
      ? root.getElementsByTagName("*")
      : [];
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
  return String(DOMPurify.sanitize(doc.body?.innerHTML ?? html, EPUB_SANITIZE));
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
    const heading =
      textContent(doc.querySelector("h1")) ||
      textContent(doc.querySelector("title")) ||
      `Chapter ${chapters.length + 1}`;
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
