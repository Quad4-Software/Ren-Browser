// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  MAX_TABS,
  TAB_NEW_BUTTON_GAP_PX,
  canOpenTab,
  expandHexColor,
  micronShellStyle,
  normalizeReticulumURL,
  isNodeHomePage,
  nodeHomeURL,
  orderTabsPinnedFirst,
  pinTabInList,
  reorderTabsInList,
  tabTitleFromURL,
  tabsAreaWidth,
  tabWidthForCount,
  tabWidthForTab,
  unpinTabInList,
  type Tab,
} from "./url";

function sampleTab(id: string, pinned = false): Tab {
  return { id, title: id, url: `${id}:/page/index.mu`, active: false, pinned };
}

describe("normalizeReticulumURL", () => {
  it("expands bare node hash to index page", () => {
    expect(normalizeReticulumURL("ABB3EBCD03CB2388A838E70C001291F9")).toBe(
      "abb3ebcd03cb2388a838e70c001291f9:/page/index.mu",
    );
  });

  it("keeps mesh urls intact", () => {
    const url = "abb3ebcd03cb2388a838e70c001291f9:/page/about.mu";
    expect(normalizeReticulumURL(url)).toBe(url);
  });

  it("normalizes about, license, editor, config, and docs urls", () => {
    expect(normalizeReticulumURL("about")).toBe("about:");
    expect(normalizeReticulumURL("license")).toBe("license:");
    expect(normalizeReticulumURL("editor")).toBe("editor:");
    expect(normalizeReticulumURL("config")).toBe("config:");
    expect(normalizeReticulumURL("settings")).toBe("settings:");
    expect(normalizeReticulumURL("docs")).toBe("docs:");
    expect(normalizeReticulumURL("docs?lang=en")).toBe("docs:?lang=en");
    expect(normalizeReticulumURL("docs:?lang=en&page=faq")).toBe("docs:?lang=en&page=faq");
  });

  it("preserves query and backtick field suffixes", () => {
    const url = "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice&action=go";
    expect(normalizeReticulumURL(url)).toBe(url);
    const backtick = "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu`user=alice|action=go";
    expect(normalizeReticulumURL(backtick)).toBe(backtick);
  });

  it("normalizes document urls", () => {
    expect(normalizeReticulumURL("document:book.epub")).toBe("document:/book.epub");
    expect(normalizeReticulumURL("document:/book.epub")).toBe("document:/book.epub");
  });

  it("rejects external internet urls", () => {
    expect(normalizeReticulumURL("https://example.com")).toBe("");
    expect(normalizeReticulumURL("http://example.com/path")).toBe("");
    expect(normalizeReticulumURL("//cdn.example.com")).toBe("");
  });

  it("unwraps renbrowser deeplinks", () => {
    expect(normalizeReticulumURL("renbrowser:about")).toBe("about:");
    expect(normalizeReticulumURL("renbrowser://open?url=license%3A")).toBe("license:");
  });
});

describe("nodeHomeURL", () => {
  const hash = "abb3ebcd03cb2388a838e70c001291f9";
  const home = `${hash}:/page/index.mu`;

  it("builds index.mu for the current node", () => {
    expect(nodeHomeURL(`${hash}:/page/about.mu`)).toBe(home);
    expect(nodeHomeURL(`rns://${hash}/page/guide.mu`)).toBe(home);
  });

  it("returns empty when the url is not a mesh node page", () => {
    expect(nodeHomeURL("about:")).toBe("");
    expect(nodeHomeURL("")).toBe("");
  });
});

describe("isNodeHomePage", () => {
  const hash = "abb3ebcd03cb2388a838e70c001291f9";

  it("matches index.mu for the node", () => {
    expect(isNodeHomePage(`${hash}:/page/index.mu`)).toBe(true);
    expect(isNodeHomePage(`${hash}:/page/index.mu?x=1`)).toBe(true);
  });

  it("rejects other mesh paths", () => {
    expect(isNodeHomePage(`${hash}:/page/about.mu`)).toBe(false);
    expect(isNodeHomePage(`${hash}:/page/guide/index.mu`)).toBe(false);
    expect(isNodeHomePage("about:")).toBe(false);
  });
});

describe("tabTitleFromURL", () => {
  it("uses node name when known", () => {
    expect(
      tabTitleFromURL("deadbeefdeadbeefdeadbeefdeadbeef:/page/guide/index.mu", [
        { hash: "deadbeefdeadbeefdeadbeefdeadbeef", name: "Mesh Node" },
      ]),
    ).toBe("Mesh Node");
  });

  it("uses page leaf name when node unknown", () => {
    expect(tabTitleFromURL("deadbeef:/page/guide/index.mu")).toBe("index.mu");
  });

  it("labels special pages", () => {
    expect(tabTitleFromURL("editor:")).toBe("Micron Editor");
    expect(tabTitleFromURL("config:")).toBe("Reticulum Config");
    expect(tabTitleFromURL("settings:")).toBe("Settings");
    expect(tabTitleFromURL("about:")).toBe("About");
    expect(tabTitleFromURL("license:")).toBe("License");
    expect(tabTitleFromURL("docs:")).toBe("Documentation");
    expect(tabTitleFromURL("docs:?lang=en&page=faq")).toBe("Faq");
  });

  it("truncates very long leaf names", () => {
    const long = "a".repeat(50);
    expect(tabTitleFromURL(`deadbeef:/page/${long}.mu`)).toBe(`${"a".repeat(39)}…`);
  });
});

describe("tabWidthForCount", () => {
  it("compresses tabs to fit the strip", () => {
    expect(tabWidthForCount(600, 5)).toBe(116.8);
    expect(tabWidthForCount(600, 10)).toBe(56.4);
  });

  it("caps width when there is room", () => {
    expect(tabWidthForCount(1200, 2)).toBe(220);
  });

  it("uses compact width before layout is measured", () => {
    expect(tabWidthForCount(0, 8)).toBe(52);
  });

  it("accounts for the new-tab button width via tabsAreaWidth", () => {
    const stripWidth = 600;
    const newTabWidth = 28;
    const area = tabsAreaWidth(stripWidth, newTabWidth);
    expect(area).toBe(stripWidth - newTabWidth - TAB_NEW_BUTTON_GAP_PX);
    expect(tabWidthForCount(area, 5)).toBeLessThan(tabWidthForCount(stripWidth, 5));
  });
});

describe("tabsAreaWidth", () => {
  it("reserves space for the new-tab button and gap", () => {
    expect(tabsAreaWidth(600, 28)).toBe(600 - 28 - TAB_NEW_BUTTON_GAP_PX);
  });

  it("uses the full strip when the button is not measured yet", () => {
    expect(tabsAreaWidth(600, 0)).toBe(600);
  });

  it("never returns negative width", () => {
    expect(tabsAreaWidth(20, 28)).toBe(0);
  });
});

describe("canOpenTab", () => {
  it("allows tabs below the limit", () => {
    expect(canOpenTab(0)).toBe(true);
    expect(canOpenTab(MAX_TABS - 1)).toBe(true);
  });

  it("blocks at the limit", () => {
    expect(canOpenTab(MAX_TABS)).toBe(false);
  });
});

describe("expandHexColor", () => {
  it("expands 3-digit shorthand", () => {
    expect(expandHexColor("#f0a")).toBe("ff00aa");
    expect(expandHexColor("f0a")).toBe("ff00aa");
  });

  it("passes through 6-digit values", () => {
    expect(expandHexColor("ff00aa")).toBe("ff00aa");
  });

  it("rejects non-hex tokens that are only length-valid", () => {
    expect(expandHexColor("zzz")).toBe("");
    expect(expandHexColor('000";x')).toBe("");
    expect(expandHexColor("red")).toBe("");
  });
});

describe("micronShellStyle", () => {
  it("returns empty for non-micron content", () => {
    expect(micronShellStyle("html", "fff", "000")).toBe("");
  });

  it("builds fg/bg style for micron pages", () => {
    expect(micronShellStyle("micron", "f0a", "123")).toBe("background:#112233;color:#ff00aa");
  });

  it("uses defaults when colors are omitted", () => {
    expect(micronShellStyle("micron")).toBe("background:#000000;color:#ffffff");
  });

  it("falls back to defaults for hostile color injection", () => {
    expect(micronShellStyle("micron", 'fff" onload="x', "000;}</style><script>")).toBe(
      "background:#000000;color:#ffffff",
    );
  });
});

describe("tab pinning", () => {
  it("keeps pinned tabs on the left", () => {
    const tabs = [sampleTab("a"), sampleTab("b", true), sampleTab("c")];
    expect(orderTabsPinnedFirst(tabs).map((item) => item.id)).toEqual(["b", "a", "c"]);
  });

  it("pins a tab and moves it left", () => {
    const tabs = [sampleTab("a"), sampleTab("b"), sampleTab("c", true)];
    expect(pinTabInList(tabs, "b").map((item) => [item.id, item.pinned])).toEqual([
      ["b", true],
      ["c", true],
      ["a", false],
    ]);
  });

  it("unpins a tab after the pinned group", () => {
    const tabs = [sampleTab("a", true), sampleTab("b", true), sampleTab("c")];
    expect(unpinTabInList(tabs, "b").map((item) => [item.id, item.pinned])).toEqual([
      ["a", true],
      ["b", false],
      ["c", false],
    ]);
  });

  it("blocks reorder across pinned boundaries", () => {
    const tabs = [sampleTab("a", true), sampleTab("b"), sampleTab("c")];
    expect(reorderTabsInList(tabs, "b", "a").map((item) => item.id)).toEqual(["a", "b", "c"]);
    expect(reorderTabsInList(tabs, "a", "b").map((item) => item.id)).toEqual(["a", "b", "c"]);
    expect(reorderTabsInList(tabs, "b", "c").map((item) => item.id)).toEqual(["a", "c", "b"]);
  });

  it("uses a fixed width for pinned tabs", () => {
    const tabs = [sampleTab("a", true), sampleTab("b"), sampleTab("c")];
    expect(tabWidthForTab(600, tabs, tabs[0])).toBe(44);
    expect(tabWidthForTab(600, tabs, tabs[1])).toBeGreaterThan(44);
  });
});
