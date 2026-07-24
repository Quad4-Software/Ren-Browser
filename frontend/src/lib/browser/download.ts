// SPDX-License-Identifier: MIT
import {
  DownloadToDir,
  SaveTextToDownloadDir,
} from "../../../bindings/renbrowser/internal/app/browserservice.js";
import { errorText, formatBindingError, unwrapBindingErrorMessage } from "./binding-errors.js";

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

export function sanitizeDownloadFilename(name: string, fallback = "page.txt"): string {
  let leaf = name.trim().replace(/\\/g, "/");
  const parts = leaf.split("/").filter(Boolean);
  leaf = parts.at(-1) ?? "";
  leaf = leaf.replace(/\0/g, "");
  leaf = leaf.replace(/[<>:"|?*]/g, "_");
  if (leaf === "." || leaf === ".." || !leaf) {
    return fallback;
  }
  if (/^[a-zA-Z]:/.test(leaf)) {
    leaf = leaf.slice(2);
  }
  leaf = leaf.replace(/:/g, "_");
  return leaf || fallback;
}

export function pageDownloadName(url: string, contentType: string): string {
  const typeFallback =
    contentType === "micron" || contentType === "editor"
      ? "page.mu"
      : contentType === "html"
        ? "page.html"
        : contentType === "markdown"
          ? "page.md"
          : "page.txt";

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
  if (url === "settings:") {
    return "settings.json";
  }

  // Check for 'a' parameter in query string
  try {
    const u = new URL(url.includes("://") ? url : `mesh://${url}`);
    // Prioritize splitting by & or | if a= is present
    const q = u.search;
    if (q.includes("a=")) {
      for (const part of q.slice(1).split(/[&|]/)) {
        if (part.startsWith("a=")) {
          return sanitizeDownloadFilename(decodeURIComponent(part.slice(2)), typeFallback);
        }
      }
    }
    const a = u.searchParams.get("a");
    if (a) {
      return sanitizeDownloadFilename(a, typeFallback);
    }
  } catch {
    const q = url.indexOf("?");
    if (q >= 0) {
      const query = url.slice(q + 1);
      for (const part of query.split(/[&|]/)) {
        if (part.startsWith("a=")) {
          try {
            return sanitizeDownloadFilename(decodeURIComponent(part.slice(2)), typeFallback);
          } catch {
            return sanitizeDownloadFilename(part.slice(2), typeFallback);
          }
        }
      }
    }
  }

  const barePath = meshBarePath(url);
  const leaf = barePath.split("/").filter(Boolean).at(-1);
  if (leaf && leaf !== "artifact") {
    return sanitizeDownloadFilename(leaf, typeFallback);
  }
  return typeFallback;
}

export function isFileURL(url: string): boolean {
  if (
    !url ||
    url === "about:" ||
    url === "license:" ||
    url === "editor:" ||
    url === "config:" ||
    url === "settings:"
  ) {
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

export type DownloadResult = {
  ok: boolean;
  message: string;
  pending?: boolean;
  canceled?: boolean;
  name?: string;
};

const canceledDownloadLabelMax = 48;

export function truncateDownloadLabel(name: string, max = canceledDownloadLabelMax): string {
  const trimmed = name.trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.length <= max) {
    return trimmed;
  }
  return `${trimmed.slice(0, max - 1)}\u2026`;
}

export function canceledDownloadToast(
  name: string | undefined,
  translate: (key: string, params?: Record<string, string>) => string,
): string {
  const label = truncateDownloadLabel(name ?? "");
  if (!label) {
    return translate("downloads.canceled");
  }
  return translate("downloads.canceledNamed", { name: label });
}

function isCanceledMessage(text: string): boolean {
  const lower = text.toLowerCase();
  return lower.includes("context canceled") || lower === "download canceled";
}

export function isDownloadCanceledError(err: unknown): boolean {
  const message = unwrapBindingErrorMessage(errorText(err)).toLowerCase();
  return isCanceledMessage(message);
}

export function downloadFailureMessage(err: unknown, fallback: string): string {
  if (isDownloadCanceledError(err)) {
    return "";
  }
  return formatBindingError(err, fallback);
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
