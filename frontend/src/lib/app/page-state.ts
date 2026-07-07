// SPDX-License-Identifier: MIT
import type { PageResponse, TabPage } from "./types";

export function emptyPage(): TabPage {
  return {
    html: "",
    contentType: "",
    error: "",
    errorKind: "",
    durationMs: 0,
    lastRaw: "",
    binaryB64: "",
    path: "",
    pageFg: "",
    pageBg: "",
    fromCache: false,
    cachedAt: 0,
    hops: -1,
    showSource: false,
  };
}

export function pageFromResponse(page: PageResponse): TabPage {
  return {
    html: page.html ?? "",
    contentType: page.contentType ?? "",
    error: page.error ?? "",
    errorKind: page.errorKind ?? "",
    durationMs: page.durationMs ?? 0,
    lastRaw: page.raw ?? "",
    binaryB64: page.binaryB64 ?? "",
    path: page.path ?? "",
    pageFg: page.pageFg ?? "",
    pageBg: page.pageBg ?? "",
    fromCache: page.fromCache ?? false,
    cachedAt: page.cachedAt ?? 0,
    hops: page.hops ?? -1,
    showSource: false,
  };
}
