import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import ExtensionsPanel from "./ExtensionsPanel.svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";

describe("ExtensionsPanel layout regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("prevents long plugin directory paths from overflowing on mobile", async () => {
    instance = await mountInBody(ExtensionsPanel, {
      pluginsDir:
        "/home/user1/.config/renbrowser/extremely-long-custom-plugins-directory-path-that-would-normally-overflow-the-screen-on-mobile-devices",
      showTitle: true,
    });

    const hint = document.querySelector(".extensions .hint");
    expect(hint).not.toBeNull();
    const style = getComputedStyle(hint!);
    expect(style.wordBreak).toBe("break-all");
    expect(style.overflowWrap).toBe("anywhere");
  });
});
