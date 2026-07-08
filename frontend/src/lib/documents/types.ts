// SPDX-License-Identifier: MIT

export type DocumentKind = "pdf" | "epub";

export function isDocumentContentType(contentType: string): contentType is DocumentKind {
  return contentType === "pdf" || contentType === "epub";
}

export function isReadableDocumentName(name: string): boolean {
  return /\.(pdf|epub)$/i.test(name.trim());
}

function joinPaths(base: string, relative: string): string {
  const trimmedBase = base.replace(/\/+$/, "");
  const trimmedRelative = relative.replace(/^\/+/, "");
  if (!trimmedBase) {
    return `/${trimmedRelative}`;
  }
  return `${trimmedBase}/${trimmedRelative}`;
}

function relativeDocumentPath(downloadDir: string, absolutePath: string): string | null {
  const base = downloadDir.replace(/\/+$/, "");
  const file = absolutePath.replace(/\\/g, "/");
  const baseNorm = base.replace(/\\/g, "/");
  if (file === baseNorm) {
    return "";
  }
  const prefix = `${baseNorm}/`;
  if (!file.startsWith(prefix)) {
    return null;
  }
  return file.slice(prefix.length);
}

export function documentURL(absolutePath: string, downloadDir = ""): string {
  const path = absolutePath.trim();
  if (downloadDir.trim()) {
    const rel = relativeDocumentPath(downloadDir, path);
    if (rel !== null) {
      return `document:/${rel}`;
    }
  }
  return `document:?path=${encodeURIComponent(path)}`;
}

export function isDocumentURL(url: string): boolean {
  return url.trim().toLowerCase().startsWith("document:");
}

export function parseDocumentPathFromURL(url: string): string | null {
  const trimmed = url.trim();
  if (!isDocumentURL(trimmed)) {
    return null;
  }
  if (trimmed.includes("?")) {
    try {
      const path = new URL(trimmed).searchParams.get("path")?.trim();
      return path || null;
    } catch {
      return null;
    }
  }
  const rest = trimmed.slice("document:".length).replace(/^\/+/, "");
  return rest || null;
}

export function resolveDocumentAbsolutePath(url: string, downloadDir: string): string | null {
  const trimmed = url.trim();
  if (!isDocumentURL(trimmed)) {
    return null;
  }
  if (trimmed.includes("?")) {
    return parseDocumentPathFromURL(trimmed);
  }
  const rel = parseDocumentPathFromURL(trimmed);
  if (!rel || !downloadDir.trim()) {
    return null;
  }
  return joinPaths(downloadDir, rel);
}

export function canonicalDocumentURL(url: string, downloadDir: string): string {
  const absolute = resolveDocumentAbsolutePath(url, downloadDir);
  if (!absolute) {
    return url;
  }
  return documentURL(absolute, downloadDir);
}

export function isReadableMeshFileURL(url: string): boolean {
  if (isDocumentURL(url)) {
    return false;
  }
  const path = meshLeafPath(url);
  return isReadableDocumentName(path);
}

function meshLeafPath(url: string): string {
  const path = url.includes(":/") ? (url.split(":/")[1] ?? "") : url;
  const bare = path.split(/[?`]/)[0] ?? path;
  const leaf = bare.split("/").filter(Boolean).at(-1) ?? "";
  return leaf;
}

export function decodeBase64ToUint8Array(b64: string): Uint8Array {
  const binary = atob(b64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}
