// SPDX-License-Identifier: MIT

export type DocumentKind = "pdf" | "epub";

export function isDocumentContentType(contentType: string): contentType is DocumentKind {
  return contentType === "pdf" || contentType === "epub";
}

export function isReadableDocumentName(name: string): boolean {
  return /\.(pdf|epub)$/i.test(name.trim());
}

export function documentURL(path: string): string {
  return `document:?path=${encodeURIComponent(path)}`;
}

export function isDocumentURL(url: string): boolean {
  return url.trim().toLowerCase().startsWith("document:");
}

export function isReadableMeshFileURL(url: string): boolean {
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
