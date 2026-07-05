// SPDX-License-Identifier: MIT
import {
  DownloadToDir,
  SaveTextToDownloadDir,
} from "../../../bindings/renbrowser/internal/app/browserservice.js";

export function downloadText(filename: string, text: string, mime = "text/plain") {
  const blob = new Blob([text], { type: mime });
  const href = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = href;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(href);
}

function meshBarePath(url: string): string {
  const path = url.includes(":/") ? (url.split(":/")[1] ?? "") : url;
  const bare = path.split(/[?`]/)[0] ?? path;
  return bare.startsWith("/") ? bare : `/${bare}`;
}

export function pageDownloadName(url: string, contentType: string): string {
  if (url === "editor:") {
    return "editor.mu";
  }
  if (url === "about:") {
    return "about.html";
  }
  if (url === "license:") {
    return "LICENSE";
  }
  if (url === "config:") {
    return "reticulum.conf";
  }
  const barePath = meshBarePath(url);
  const leaf = barePath.split("/").filter(Boolean).at(-1);
  if (leaf) {
    return leaf;
  }
  if (contentType === "micron" || contentType === "editor") {
    return "page.mu";
  }
  if (contentType === "html") {
    return "page.html";
  }
  if (contentType === "markdown") {
    return "page.md";
  }
  return "page.txt";
}

export function isFileURL(url: string): boolean {
  if (!url || url === "about:" || url === "license:" || url === "editor:" || url === "config:") {
    return false;
  }
  return meshBarePath(url).startsWith("/file/");
}

export async function downloadMeshFile(url: string): Promise<string> {
  return await DownloadToDir(url);
}

export async function savePageToDownloadDir(filename: string, text: string): Promise<string> {
  return await SaveTextToDownloadDir(filename, text);
}

export async function downloadPageContent(
  url: string,
  contentType: string,
  text: string,
): Promise<string> {
  if (isFileURL(url)) {
    return await downloadMeshFile(url);
  }
  return await savePageToDownloadDir(pageDownloadName(url, contentType), text);
}
