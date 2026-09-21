// SPDX-License-Identifier: MIT
import { translate } from "$lib/i18n/catalog";
import { parseDocumentPathFromURL } from "$lib/documents/types";
import { isBlockedNavigationURL } from "./navigation-guard";
import { unwrapDeepLink } from "./deeplink";

export type TabPage = {
  html: string;
  contentType: string;
  error: string;
  errorKind?: string;
  durationMs: number;
  lastRaw: string;
  binaryB64?: string;
  path?: string;
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
  groupId?: string;
  page?: TabPage;
  navGeneration?: number;
  loading?: boolean;
};

export type TabLayout = "top" | "left";

export type TabGroup = {
  id: string;
  name: string;
  color: string;
  collapsed: boolean;
};

export const TAB_GROUP_COLORS = [
  "#71717a",
  "#60a5fa",
  "#34d399",
  "#fbbf24",
  "#f87171",
  "#a78bfa",
] as const;

export function tabGroupColor(groups: TabGroup[], groupId: string | undefined): string {
  if (!groupId) {
    return "";
  }
  return groups.find((group) => group.id === groupId)?.color ?? "";
}

export function setTabGroupInList(tabs: Tab[], id: string, groupId: string | undefined): Tab[] {
  return tabs.map((tab) => {
    if (tab.id !== id) {
      return tab;
    }
    if (!groupId) {
      const { groupId: _drop, ...rest } = tab;
      return rest;
    }
    return { ...tab, groupId };
  });
}

export type TabListItem =
  { type: "tab"; tab: Tab } | { type: "group"; group: TabGroup; tabs: Tab[] };

// Assigns a tab to a group and moves it directly after the last current
// member so grouped tabs stay contiguous in the list.
export function assignTabToGroupInList(
  tabs: Tab[],
  tabId: string,
  groupId: string | undefined,
): Tab[] {
  const tab = tabs.find((item) => item.id === tabId);
  if (!tab || tab.pinned || tab.groupId === groupId) {
    return tabs;
  }
  const next = setTabGroupInList(tabs, tabId, groupId);
  if (!groupId) {
    return next;
  }
  const members = next.filter((item) => item.groupId === groupId && item.id !== tabId);
  if (members.length === 0) {
    return next;
  }
  const moved = next.find((item) => item.id === tabId);
  if (!moved) {
    return next;
  }
  const without = next.filter((item) => item.id !== tabId);
  const lastMemberIdx = without.findIndex((item) => item.id === members[members.length - 1].id);
  without.splice(lastMemberIdx + 1, 0, moved);
  return without;
}

// Drops groups with no member tabs and clears dangling groupId references.
// Returns null when nothing changed so callers can skip state updates.
export function pruneTabGroups(
  tabs: Tab[],
  groups: TabGroup[],
): { tabs: Tab[]; groups: TabGroup[] } | null {
  const groupIds = new Set(groups.map((group) => group.id));
  const memberCounts = new Map<string, number>();
  let tabsChanged = false;
  const nextTabs = tabs.map((tab) => {
    if (!tab.groupId || groupIds.has(tab.groupId)) {
      if (tab.groupId) {
        memberCounts.set(tab.groupId, (memberCounts.get(tab.groupId) ?? 0) + 1);
      }
      return tab;
    }
    tabsChanged = true;
    const { groupId: _drop, ...rest } = tab;
    return rest;
  });
  const nextGroups = groups.filter((group) => (memberCounts.get(group.id) ?? 0) > 0);
  if (!tabsChanged && nextGroups.length === groups.length) {
    return null;
  }
  return { tabs: nextTabs, groups: nextGroups };
}

// Groups render at the position of their first member tab. Pinned tabs
// always lead and are never grouped.
export function groupTabsForDisplay(tabs: Tab[], groups: TabGroup[]): TabListItem[] {
  const groupById = new Map(groups.map((group) => [group.id, group]));
  const items: TabListItem[] = [];
  const seen = new Set<string>();
  for (const tab of tabs) {
    const group = tab.groupId ? groupById.get(tab.groupId) : undefined;
    if (!group) {
      items.push({ type: "tab", tab });
      continue;
    }
    if (seen.has(group.id)) {
      continue;
    }
    seen.add(group.id);
    items.push({
      type: "group",
      group,
      tabs: tabs.filter((member) => member.groupId === group.id),
    });
  }
  return items;
}

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
    tabs.map((item) => {
      if (item.id !== id) {
        return item;
      }
      const { groupId: _drop, ...rest } = item;
      return { ...rest, pinned: true };
    }),
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
    return TAB_WIDTH_MIN;
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
    return 52;
  }
  const totalGap = Math.max(0, count - 1) * TAB_GAP_PX;
  const perTab = (stripWidth - totalGap) / count;
  return Math.min(TAB_WIDTH_MAX, Math.max(52, perTab));
}

export function canOpenTab(count: number): boolean {
  return count < MAX_TABS;
}

const NODE_HASH_RE = /^[a-f0-9]{32}$/i;

export function nodeHashFromMeshURL(url: string): string {
  const trimmed = url.trim();
  if (!trimmed) {
    return "";
  }
  if (trimmed.toLowerCase().startsWith("rns://")) {
    try {
      const parsed = new URL(trimmed);
      const hash = parsed.hostname.toLowerCase();
      return NODE_HASH_RE.test(hash) ? hash : "";
    } catch {
      return "";
    }
  }
  if (!trimmed.includes(":/")) {
    return "";
  }
  const hash = trimmed.split(":/")[0]?.trim().toLowerCase() ?? "";
  return NODE_HASH_RE.test(hash) ? hash : "";
}

export function nodeHomeURL(url: string): string {
  const hash = nodeHashFromMeshURL(url);
  return hash ? `${hash}:/page/index.mu` : "";
}

export function isNodeHomePage(url: string): boolean {
  const hash = nodeHashFromMeshURL(url);
  if (!hash) {
    return false;
  }
  const rest = url.includes(":/") ? (url.split(":/")[1] ?? "") : "";
  const path = rest.split(/[?`]/)[0] ?? "";
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return normalized === "/page/index.mu";
}

export type DiscoveredNode = {
  hash: string;
  name: string;
};

export function normalizeReticulumURL(input: string): string {
  const unwrapped = unwrapDeepLink(input);
  const trimmed = (unwrapped || input).trim();
  if (!trimmed) {
    return "";
  }
  if (isBlockedNavigationURL(trimmed)) {
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
  if (lower.startsWith("document:")) {
    if (trimmed.includes("?")) {
      return trimmed;
    }
    let rest = trimmed.slice("document:".length);
    if (!rest.startsWith("/")) {
      rest = `/${rest}`;
    }
    return `document:${rest}`;
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
  if (url.startsWith("document:")) {
    const path = parseDocumentPathFromURL(url) ?? "";
    const leaf = path.split(/[/\\]/).filter(Boolean).at(-1) ?? "";
    if (leaf) {
      return leaf.length <= 40 ? leaf : `${leaf.slice(0, 39)}…`;
    }
    return translate("tab.document");
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

const MICRON_HEX_COLOR_RE = /^[0-9a-fA-F]{3}$|^[0-9a-fA-F]{6}$/;

export function isMicronHexColor(color: string): boolean {
  return MICRON_HEX_COLOR_RE.test(color.trim().replace(/^#/, ""));
}

export function expandHexColor(color: string): string {
  const c = color.trim().replace(/^#/, "");
  if (!MICRON_HEX_COLOR_RE.test(c)) {
    return "";
  }
  if (c.length === 3) {
    return c
      .split("")
      .map((ch) => ch + ch)
      .join("");
  }
  return c;
}

export function micronPageColors(pageFg?: string, pageBg?: string): { fg: string; bg: string } {
  const bg = pageBg?.trim() ? expandHexColor(pageBg) : "";
  const fg = pageFg?.trim() ? expandHexColor(pageFg) : "";
  return {
    bg: bg ? `#${bg}` : "#000000",
    fg: fg ? `#${fg}` : "#ffffff",
  };
}

export function micronShellStyle(contentType: string, pageFg?: string, pageBg?: string): string {
  if (contentType !== "micron") {
    return "";
  }
  const { fg, bg } = micronPageColors(pageFg, pageBg);
  return `background:${bg};color:${fg}`;
}
