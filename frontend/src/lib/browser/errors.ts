// SPDX-License-Identifier: MIT
import { displayName } from "$lib/brand";
import { translate } from "$lib/i18n/catalog";

export type PageErrorKind =
  | "connection_failed"
  | "connection_lost"
  | "not_found"
  | "internal"
  | "storage_full"
  | "database_corrupt"
  | "unknown";

export type StoreErrorKind = "storage_full" | "database_corrupt";

export type ErrorPageContent = {
  title: string;
  description: string;
  showRetry: boolean;
  showResetDatabase: boolean;
  tone: "danger" | "warning";
};

const PAGE_ERROR_KEYS: Record<
  PageErrorKind,
  {
    title: string;
    description: string;
    showRetry: boolean;
    showResetDatabase: boolean;
    tone: "danger" | "warning";
  }
> = {
  connection_failed: {
    title: "errors.connectionFailedTitle",
    description: "errors.connectionFailedDescription",
    showRetry: true,
    showResetDatabase: false,
    tone: "warning",
  },
  connection_lost: {
    title: "errors.connectionLostTitle",
    description: "errors.connectionLostDescription",
    showRetry: true,
    showResetDatabase: false,
    tone: "warning",
  },
  not_found: {
    title: "errors.notFoundTitle",
    description: "errors.notFoundDescription",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
  internal: {
    title: "errors.internalTitle",
    description: "",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
  storage_full: {
    title: "errors.storageFullTitle",
    description: "errors.storageFullDescription",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
  database_corrupt: {
    title: "errors.databaseCorruptTitle",
    description: "errors.databaseCorruptDescription",
    showRetry: false,
    showResetDatabase: true,
    tone: "danger",
  },
  unknown: {
    title: "errors.unknownTitle",
    description: "errors.unknownDescription",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
};

export function normalizePageErrorKind(kind: string | undefined, error: string): PageErrorKind {
  if (kind && kind in PAGE_ERROR_KEYS) {
    return kind as PageErrorKind;
  }
  const msg = error.toLowerCase();
  if (
    msg.includes("reticulum not ready") ||
    msg.includes("no path") ||
    msg.includes("link establish")
  ) {
    return "connection_failed";
  }
  if (msg.includes("context canceled") || msg.includes("connection lost")) {
    return "connection_lost";
  }
  if (msg.includes("empty response") || msg.includes("not found") || msg.includes("404")) {
    return "not_found";
  }
  if (msg.includes("500") || msg.includes("internal")) {
    return "internal";
  }
  return error ? "internal" : "unknown";
}

export function pageErrorContent(kind: PageErrorKind, detail: string): ErrorPageContent {
  const base = PAGE_ERROR_KEYS[kind];
  const title = translate(base.title);
  let description = base.description ? translate(base.description, { app: displayName }) : "";
  if (kind === "internal" && detail.trim()) {
    description = detail.trim();
  }
  if (kind === "unknown" && detail.trim()) {
    description = detail.trim();
  }
  return {
    title,
    description,
    showRetry: base.showRetry,
    showResetDatabase: base.showResetDatabase,
    tone: base.tone,
  };
}

export function isStoreBlockingKind(kind: string | undefined): kind is StoreErrorKind {
  return kind === "storage_full" || kind === "database_corrupt";
}
