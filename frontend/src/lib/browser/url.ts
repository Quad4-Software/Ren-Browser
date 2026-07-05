// SPDX-License-Identifier: MIT
import { translate } from "$lib/i18n/catalog";
export type TabPage = {
  html: string;
  contentType: string;
  error: string;
  errorKind?: string;
  durationMs: number;
  lastRaw: string;
  pageFg?: string;
  pageBg?: string;
  fromCache?: boolean;
  cachedAt?: number;
  hops?: number;
  showSource?: boolean;
};

export type Tab = {
  id: string;
  title: string;
  url: string;
  active: boolean;
  pinned?: boolean;
  page?: TabPage;
  navGeneration?: number;
};

export const MAX_TABS = 32;

export const TAB_WIDTH_MAX = 220;
export const TAB_WIDTH_MIN = 76;
export const PINNED_TAB_WIDTH = 44;
export const TAB_GAP_PX = 4;

export function isTabPinned(tab: Tab): boolean {
  return !!tab.pinned;
}

export function orderTabsPinnedFirst(tabs: Tab[]): Tab[] {
  const pinned = tabs.filter((tab) => tab.pinned);
  const unpinned = tabs.filter((tab) => !tab.pinned);
  return [...pinned, ...unpinned];
}

export function pinTabInList(tabs: Tab[], id: string): Tab[] {
  const tab = tabs.find((item) => item.id === id);
  if (!tab || tab.pinned) {
    return tabs;
  }
  return orderTabsPinnedFirst(
    tabs.map((item) => (item.id === id ? { ...item, pinned: true } : item)),
  );
}

export function unpinTabInList(tabs: Tab[], id: string): Tab[] {
  const tab = tabs.find((item) => item.id === id);
  if (!tab?.pinned) {
    return tabs;
  }
  const pinned = tabs.filter((item) => item.pinned && item.id !== id);
  const unpinned = tabs.filter((item) => !item.pinned);
  return [...pinned, { ...tab, pinned: false }, ...unpinned];
}

export function reorderTabsInList(tabs: Tab[], fromId: string, toId: string): Tab[] {
  const fromIdx = tabs.findIndex((tab) => tab.id === fromId);
  const toIdx = tabs.findIndex((tab) => tab.id === toId);
  if (fromIdx < 0 || toIdx < 0) {
    return tabs;
  }
  const from = tabs[fromIdx];
  const to = tabs[toIdx];
  if (!!from.pinned !== !!to.pinned) {
    return tabs;
  }
  const next = [...tabs];
  const [moved] = next.splice(fromIdx, 1);
  next.splice(toIdx, 0, moved);
  return next;
}

export function unpinnedWidthForStrip(stripWidth: number, tabs: Tab[]): number {
  const pinnedCount = tabs.filter((tab) => tab.pinned).length;
  const unpinnedCount = tabs.length - pinnedCount;
  if (unpinnedCount <= 0) {
    return TAB_WIDTH_MAX;
  }
  const pinnedTotal = pinnedCount * PINNED_TAB_WIDTH;
  const gaps = Math.max(0, tabs.length - 1) * TAB_GAP_PX;
  return tabWidthForCount(stripWidth - pinnedTotal - gaps, unpinnedCount);
}

export function tabWidthForTab(stripWidth: number, tabs: Tab[], tab: Tab): number {
  if (tab.pinned) {
    return PINNED_TAB_WIDTH;
  }
  return unpinnedWidthForStrip(stripWidth, tabs);
}
export const TAB_NEW_BUTTON_GAP_PX = 6;

export function tabsAreaWidth(stripWidth: number, newTabButtonWidth: number): number {
  const gap = newTabButtonWidth > 0 ? TAB_NEW_BUTTON_GAP_PX : 0;
  return Math.max(0, stripWidth - newTabButtonWidth - gap);
}

export function tabWidthForCount(stripWidth: number, count: number): number {
  if (count <= 0 || stripWidth <= 0) {
    return TAB_WIDTH_MAX;
  }
  const totalGap = Math.max(0, count - 1) * TAB_GAP_PX;
  const perTab = (stripWidth - totalGap) / count;
  return Math.min(TAB_WIDTH_MAX, Math.max(52, perTab));
}

export function canOpenTab(count: number): boolean {
  return count < MAX_TABS;
}

export type DiscoveredNode = {
  hash: string;
  name: string;
};

export function normalizeReticulumURL(input: string): string {
  const trimmed = input.trim();
  if (!trimmed) {
    return "";
  }
  const lower = trimmed.toLowerCase();
  if (lower === "about" || lower === "about:") {
    return "about:";
  }
  if (lower === "license" || lower === "license:") {
    return "license:";
  }
  if (lower === "editor" || lower === "editor:") {
    return "editor:";
  }
  if (lower === "config" || lower === "config:") {
    return "config:";
  }
  if (lower === "settings" || lower === "settings:") {
    return "settings:";
  }
  if (lower === "docs" || lower === "docs:") {
    return "docs:";
  }
  if (lower.startsWith("docs?")) {
    return `docs:?${trimmed.slice(trimmed.indexOf("?") + 1)}`;
  }
  if (lower.startsWith("docs:?")) {
    return trimmed;
  }
  if (trimmed.includes(":/")) {
    return trimmed;
  }
  if (/^[a-f0-9]{32}$/i.test(trimmed)) {
    return `${trimmed.toLowerCase()}:/page/index.mu`;
  }
  if (trimmed.startsWith("/page/")) {
    return trimmed;
  }
  return trimmed;
}

export function tabTitleFromURL(url: string, nodes: DiscoveredNode[] = []): string {
  if (!url) {
    return translate("tab.new");
  }
  if (url === "about:") {
    return translate("tab.about");
  }
  if (url === "license:") {
    return translate("tab.license");
  }
  if (url === "editor:") {
    return translate("tab.micronEditor");
  }
  if (url === "config:") {
    return translate("tab.reticulumConfig");
  }
  if (url === "settings:") {
    return translate("tab.settings");
  }
  if (url.startsWith("docs")) {
    const query = url.includes("?") ? url.slice(url.indexOf("?") + 1) : "";
    const page = new URLSearchParams(query).get("page");
    if (page) {
      const title = page
        .split("-")
        .map((part) => (part ? part[0].toUpperCase() + part.slice(1) : part))
        .join(" ");
      return title.length <= 40 ? title : `${title.slice(0, 39)}…`;
    }
    return translate("tab.documentation");
  }
  const hash = url.split(":/")[0]?.toLowerCase();
  const node = nodes.find((n) => n.hash.toLowerCase() === hash);
  if (node?.name) {
    const name = node.name;
    if (name.length <= 40) {
      return name;
    }
    return `${name.slice(0, 39)}…`;
  }
  const path = url.split(":/").at(1) ?? url;
  const leaf = path.split("/").filter(Boolean).at(-1);
  const raw = leaf || translate("tab.nomadNet");
  if (raw.length <= 40) {
    return raw;
  }
  return `${raw.slice(0, 39)}…`;
}

export function expandHexColor(color: string): string {
  const c = color.trim().replace(/^#/, "");
  if (c.length === 3) {
    return c
      .split("")
      .map((ch) => ch + ch)
      .join("");
  }
  return c;
}

export function micronShellStyle(contentType: string, pageFg?: string, pageBg?: string): string {
  if (contentType !== "micron") {
    return "";
  }
  const bg = pageBg ? `#${expandHexColor(pageBg)}` : "#000000";
  const fg = pageFg ? `#${expandHexColor(pageFg)}` : "#ffffff";
  return `background:${bg};color:${fg}`;
}
