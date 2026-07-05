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

const PAGE_ERROR_COPY: Record<PageErrorKind, ErrorPageContent> = {
  connection_failed: {
    title: "Connection to server failed",
    description:
      "Could not reach this node. Check that Reticulum interfaces are online and the destination is available on the mesh.",
    showRetry: true,
    showResetDatabase: false,
    tone: "warning",
  },
  connection_lost: {
    title: "Connection lost",
    description: "The link to the server was interrupted before the page finished loading.",
    showRetry: true,
    showResetDatabase: false,
    tone: "warning",
  },
  not_found: {
    title: "404 — Page not found",
    description: "The server does not have a page at this address.",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
  internal: {
    title: "500 — Internal server error",
    description: "",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
  storage_full: {
    title: "Storage unavailable",
    description:
      "Ren Browser cannot write to disk. Free up space or fix permissions for your profile folder.",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
  database_corrupt: {
    title: "Database corrupted",
    description:
      "Your local profile data could not be read. Resetting the database removes saved tabs, history, favorites, and settings.",
    showRetry: false,
    showResetDatabase: true,
    tone: "danger",
  },
  unknown: {
    title: "Could not open page",
    description: "Something went wrong while loading this page.",
    showRetry: true,
    showResetDatabase: false,
    tone: "danger",
  },
};

export function normalizePageErrorKind(kind: string | undefined, error: string): PageErrorKind {
  if (kind && kind in PAGE_ERROR_COPY) {
    return kind as PageErrorKind;
  }
  const msg = error.toLowerCase();
  if (msg.includes("reticulum not ready") || msg.includes("no path") || msg.includes("link establish")) {
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
  const base = PAGE_ERROR_COPY[kind];
  if (kind === "internal" && detail.trim()) {
    return {
      ...base,
      description: detail.trim(),
    };
  }
  if (kind === "unknown" && detail.trim()) {
    return {
      ...base,
      description: detail.trim(),
    };
  }
  return base;
}

export function isStoreBlockingKind(kind: string | undefined): kind is StoreErrorKind {
  return kind === "storage_full" || kind === "database_corrupt";
}
