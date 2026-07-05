import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { defaultKeybinds } from "./keybinds";
import {
  clearLocalBrowserData,
  nextHistoryID,
  readLocalBrowserData,
  writeLocalBrowserData,
  type LocalBrowserData,
  type LocalHistoryEntry,
} from "./local-store";
import { defaultTheme } from "../theme/tokens";

const profile = "test-profile";
const storage = new Map<string, string>();

beforeEach(() => {
  storage.clear();
  vi.stubGlobal("localStorage", {
    getItem: (key: string) => storage.get(key) ?? null,
    setItem: (key: string, value: string) => {
      storage.set(key, value);
    },
    removeItem: (key: string) => {
      storage.delete(key);
    },
  });
});

afterEach(() => {
  clearLocalBrowserData(profile);
  vi.unstubAllGlobals();
});

function sampleData(): LocalBrowserData {
  return {
    tabs: [],
    favorites: [],
    history: [],
    theme: defaultTheme(),
    keybinds: defaultKeybinds(),
    browserPrefs: {
      openLinksInNewTab: true,
      openLinksInNewWindow: false,
      nativeTitlebar: true,
    },
  };
}

describe("nextHistoryID", () => {
  it("starts at 1 for an empty list", () => {
    expect(nextHistoryID([])).toBe(1);
  });

  it("increments from the highest existing id", () => {
    const entries: LocalHistoryEntry[] = [
      { id: 3, url: "a", title: "A", nodeHash: "abc", visitedAt: 1 },
      { id: 7, url: "b", title: "B", nodeHash: "def", visitedAt: 2 },
    ];
    expect(nextHistoryID(entries)).toBe(8);
  });
});

describe("local browser storage", () => {
  it("round-trips profile data", () => {
    const data = sampleData();
    data.favorites = ["abb3ebcd03cb2388a838e70c001291f9:/page/index.mu"];
    writeLocalBrowserData(profile, data);
    const loaded = readLocalBrowserData(profile);
    expect(loaded?.favorites).toEqual(data.favorites);
  });

  it("returns null for missing profiles", () => {
    expect(readLocalBrowserData("missing-profile")).toBeNull();
  });
});
