import { afterEach, describe, expect, it } from "vitest";
import { mount, tick, type ComponentProps } from "svelte";
import MobileTabsPage from "./MobileTabsPage.svelte";
import type { Tab } from "$lib/browser/url";
import {
  expectsEllipsis,
  expectsNoHorizontalScroll,
  expectsVerticalScroll,
  parsePx,
} from "$lib/browser/layout-regression";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";

const noop = () => {};

function tab(id: string, title: string, url = `mesh:${id}`): Tab {
  return { id, title, url, active: id === "a" };
}

describe("MobileTabsPage layout regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  async function mountTabs(overrides: Partial<ComponentProps<typeof MobileTabsPage>> = {}) {
    instance = await mountInBody(MobileTabsPage, {
      tabs: [tab("a", "Alpha"), tab("b", "Beta"), tab("c", "Gamma")],
      activeTabId: "a",
      atTabLimit: false,
      onSelect: noop,
      onClose: noop,
      onCloseAll: noop,
      onNew: noop,
      onDismiss: noop,
      ...overrides,
    });
  }

  it("scrolls vertically and keeps tab cards at the fixed mobile width", async () => {
    await mountTabs();

    const grid = document.querySelector(".tabs-grid");
    expect(grid).not.toBeNull();
    expectsVerticalScroll(grid!);
    expectsNoHorizontalScroll(grid!);

    const cards = document.querySelectorAll(".tab-card");
    expect(cards.length).toBe(3);
    for (const card of cards) {
      expect(parsePx(getComputedStyle(card).width)).toBeGreaterThanOrEqual(160);
    }
  });

  it("truncates long tab titles instead of widening cards", async () => {
    const longTitle = "a".repeat(80);
    await mountTabs({
      tabs: [tab("a", longTitle)],
      activeTabId: "a",
    });

    const title = document.querySelector(".card-title");
    expect(title).not.toBeNull();
    expectsEllipsis(title!);

    const card = document.querySelector(".tab-card");
    expect(card).not.toBeNull();
    expect(parsePx(getComputedStyle(card!).width)).toBeLessThan(220);
  });
});
