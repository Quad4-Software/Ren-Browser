// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { emptyPage } from "$lib/app/page-state";
import type { Tab } from "$lib/browser/url";
import {
  TAB_PREVIEW_HTML_MAX,
  compactInactiveTabPages,
  compactTabPage,
  hotTabIds,
  tabPageNeedsReload,
  tabSnapshotForPersist,
} from "./tab-page-memory";

function tab(partial: Partial<Tab> & Pick<Tab, "id">): Tab {
  return {
    title: "t",
    url: "nomad://abc/page/index.mu",
    active: false,
    page: {
      ...emptyPage(),
      html: "<p>full page body</p>",
      lastRaw: ">>full",
      contentType: "micron",
    },
    ...partial,
  };
}

describe("tab page memory", () => {
  it("keeps full bodies only for hot tabs", () => {
    const tabs = [
      tab({ id: "a", active: true, page: { ...emptyPage(), html: "A".repeat(8000), lastRaw: "raw-a" } }),
      tab({ id: "b", page: { ...emptyPage(), html: "B".repeat(8000), lastRaw: "raw-b", binaryB64: "qq" } }),
    ];
    const compacted = compactInactiveTabPages(tabs, null);
    expect(compacted[0].page?.lastRaw).toBe("raw-a");
    expect(compacted[0].page?.html.length).toBe(8000);
    expect(compacted[1].page?.lastRaw).toBe("");
    expect(compacted[1].page?.binaryB64).toBe("");
    expect(compacted[1].page?.html.length).toBe(TAB_PREVIEW_HTML_MAX);
  });

  it("treats split and editor tabs as hot", () => {
    const tabs = [
      tab({ id: "a", active: true }),
      tab({ id: "split" }),
      tab({ id: "ed", url: "editor:", page: { ...emptyPage(), lastRaw: "draft", contentType: "editor" } }),
    ];
    const hot = hotTabIds(tabs, "split");
    expect([...hot].sort()).toEqual(["a", "ed", "split"]);
  });

  it("omits bodies from persist payloads for cold tabs", () => {
    const cold = tab({ id: "b", page: { ...emptyPage(), html: "<p>x</p>", lastRaw: "x" } });
    const payload = tabSnapshotForPersist(cold, false);
    expect(payload.html).toBe("");
    expect(payload.lastRaw).toBe("");
    expect(payload.url).toBe(cold.url);
  });

  it("reloads compacted mesh tabs and skips error-only tabs", () => {
    const compacted = compactTabPage({ ...emptyPage(), html: "H".repeat(8000), lastRaw: "raw" }, false);
    expect(
      tabPageNeedsReload(tab({ id: "b", page: compacted })),
    ).toBe(true);
    expect(
      tabPageNeedsReload(
        tab({
          id: "err",
          page: { ...emptyPage(), error: "offline", errorKind: "timeout" },
        }),
      ),
    ).toBe(false);
    expect(
      tabPageNeedsReload(
        tab({
          id: "ed",
          url: "editor:",
          page: { ...emptyPage(), contentType: "editor", lastRaw: "draft" },
        }),
      ),
    ).toBe(false);
  });
});
