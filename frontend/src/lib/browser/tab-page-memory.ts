// SPDX-License-Identifier: MIT
import type { TabSnapshot } from "$lib/app/types";
import { emptyPage } from "$lib/app/page-state";
import type { Tab, TabPage } from "$lib/browser/url";

export const TAB_PREVIEW_HTML_MAX = 4096;

export function isEditorTab(
  tab: Pick<Tab, "url"> & { page?: Pick<TabPage, "contentType"> },
): boolean {
  const url = tab.url.trim().toLowerCase();
  return url === "editor" || url === "editor:" || tab.page?.contentType === "editor";
}

export function hotTabIds(tabs: Tab[], splitTabId: string | null): Set<string> {
  const ids = new Set<string>();
  for (const tab of tabs) {
    if (tab.active || isEditorTab(tab)) {
      ids.add(tab.id);
    }
  }
  if (splitTabId) {
    ids.add(splitTabId);
  }
  return ids;
}

export function compactTabPage(page: TabPage | undefined, keepFull: boolean): TabPage {
  const base = page ?? emptyPage();
  if (keepFull) {
    return base;
  }
  const html = base.html ?? "";
  return {
    ...base,
    html: html.length > TAB_PREVIEW_HTML_MAX ? html.slice(0, TAB_PREVIEW_HTML_MAX) : html,
    lastRaw: "",
    binaryB64: "",
  };
}

export function compactInactiveTabPages(tabs: Tab[], splitTabId: string | null): Tab[] {
  const hot = hotTabIds(tabs, splitTabId);
  let changed = false;
  const next = tabs.map((tab) => {
    const keepFull = hot.has(tab.id);
    const page = compactTabPage(tab.page, keepFull);
    if (page === tab.page) {
      return tab;
    }
    if (
      (tab.page?.html ?? "") === page.html &&
      (tab.page?.lastRaw ?? "") === page.lastRaw &&
      (tab.page?.binaryB64 ?? "") === page.binaryB64
    ) {
      return tab;
    }
    changed = true;
    return { ...tab, page };
  });
  return changed ? next : tabs;
}

export function tabPageNeedsReload(tab: Tab): boolean {
  const url = tab.url.trim();
  if (!url) {
    return false;
  }
  const page = tab.page;
  if (
    page?.error?.trim() &&
    !(page.html?.trim() || page.lastRaw?.trim() || page.binaryB64?.trim())
  ) {
    return false;
  }
  if (isEditorTab(tab) && (page?.lastRaw?.trim() ?? "").length > 0) {
    return false;
  }
  if ((page?.binaryB64?.trim() ?? "").length > 0) {
    return false;
  }
  if ((page?.lastRaw?.trim() ?? "").length > 0) {
    return false;
  }
  const html = page?.html ?? "";
  if (html.length > TAB_PREVIEW_HTML_MAX) {
    return false;
  }
  return true;
}

export function tabSnapshotForPersist(tab: Tab, keepFull: boolean): TabSnapshot {
  return {
    id: tab.id,
    title: tab.title,
    url: tab.url,
    active: tab.active,
    pinned: tab.pinned,
    contentType: tab.page?.contentType,
    error: tab.page?.error,
    errorKind: tab.page?.errorKind,
    durationMs: tab.page?.durationMs,
    pageFg: tab.page?.pageFg,
    pageBg: tab.page?.pageBg,
    html: keepFull ? tab.page?.html : "",
    lastRaw: keepFull ? tab.page?.lastRaw : "",
  };
}
