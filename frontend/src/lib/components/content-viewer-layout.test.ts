import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import ContentViewer from "./ContentViewer.svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";

const noop = () => {};

describe("ContentViewer layout regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("wraps the cache banner without forcing horizontal overflow", async () => {
    instance = await mountInBody(ContentViewer, {
      html: "<p>Ren Browser</p>",
      contentType: "html",
      loading: false,
      error: "",
      currentURL: "mesh:/page",
      raw: "<p>Ren Browser</p>",
      fromCache: true,
      cachedAt: Date.now(),
      showSource: false,
      findOpen: false,
      micronEngine: "js",
      onNavigate: noop,
      onRetry: noop,
      onReloadFresh: noop,
      onShowSourceChange: noop,
      onFindClose: noop,
    });

    const banner = document.querySelector(".cache-banner");
    expect(banner).not.toBeNull();
    expect(getComputedStyle(banner!).flexWrap).toBe("wrap");

    const text = document.querySelector(".cache-text");
    expect(text).not.toBeNull();
    expect(getComputedStyle(text!).overflowWrap).toBe("anywhere");
  });
});
