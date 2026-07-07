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
  if (url === "settings:") {
    return "settings.json";
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
