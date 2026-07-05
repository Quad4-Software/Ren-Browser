import type { ThemeSettings } from "$lib/theme/tokens";
import type { KeybindSettings } from "$lib/browser/keybinds";

export type LocalTabSnapshot = {
  id: string;
  title: string;
  url: string;
  active: boolean;
  pinned?: boolean;
  html?: string;
  contentType?: string;
  error?: string;
  durationMs?: number;
  lastRaw?: string;
  pageFg?: string;
  pageBg?: string;
};

export type LocalHistoryEntry = {
  id: number;
  url: string;
  title: string;
  nodeHash: string;
  visitedAt: number;
};

export type LocalBrowserPrefs = {
  openLinksInNewTab: boolean;
  openLinksInNewWindow: boolean;
  nativeTitlebar: boolean;
};

export type LocalBrowserData = {
  tabs: LocalTabSnapshot[];
  favorites: string[];
  history: LocalHistoryEntry[];
  theme: ThemeSettings;
  keybinds: KeybindSettings;
  browserPrefs: LocalBrowserPrefs;
};

const dataVersion = 1;

function storageKey(profile: string): string {
  const safe = profile.trim() || "default";
  return `renbrowser:${safe}:v${dataVersion}`;
}

export function readLocalBrowserData(profile: string): LocalBrowserData | null {
  try {
    const raw = localStorage.getItem(storageKey(profile));
    if (!raw) {
      return null;
    }
    return JSON.parse(raw) as LocalBrowserData;
  } catch {
    return null;
  }
}

export function writeLocalBrowserData(profile: string, data: LocalBrowserData): void {
  localStorage.setItem(storageKey(profile), JSON.stringify(data));
}

export function clearLocalBrowserData(profile: string): void {
  localStorage.removeItem(storageKey(profile));
}

export function nextHistoryID(entries: LocalHistoryEntry[]): number {
  if (!entries.length) {
    return 1;
  }
  return Math.max(...entries.map((entry) => entry.id)) + 1;
}
