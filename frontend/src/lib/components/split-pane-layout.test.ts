// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import SplitPaneHarness from "$lib/test/SplitPaneHarness.svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";

describe("SplitPane layout", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  it("sizes the primary pane from the ratio custom property", async () => {
    instance = await mountInBody(SplitPaneHarness, { ratio: 40 });

    const root = document.querySelector(".split-root") as HTMLElement | null;
    expect(root).not.toBeNull();
    expect(root!.style.getPropertyValue("--split-ratio").trim()).toBe("40%");
    expect(getComputedStyle(root!).display).toBe("flex");

    const divider = document.querySelector(".divider") as HTMLElement | null;
    expect(divider).not.toBeNull();
    expect(getComputedStyle(divider!).cursor).toBe("col-resize");
  });
});
