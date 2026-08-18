import { afterEach, describe, expect, it } from "vitest";
import { mount, flushSync, tick } from "svelte";
import MobileUrlBar from "./MobileUrlBar.svelte";
import { cleanupMount } from "$lib/test/svelte-mount";
import { MOBILE_INPUT_FONT_PX } from "$lib/browser/mobile-keyboard";

const noop = () => {};

describe("MobileUrlBar layout regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    document.body.innerHTML = "";
    instance = null;
  });

  it("uses 16px on the URL field so WebView does not zoom on focus", async () => {
    flushSync(() => {
      instance = mount(MobileUrlBar, {
        target: document.body,
        props: {
          url: "rn://example",
          tabCount: 1,
          canIdentify: false,
          identifying: false,
          atTabLimit: false,
          onNavigate: noop,
          onHome: noop,
          onNewTab: noop,
          onOpenTabs: noop,
          onIdentify: noop,
        },
      });
    });
    await tick();

    const input = document.querySelector(".url-input") as HTMLInputElement | null;
    expect(input).not.toBeNull();
    expect(getComputedStyle(input!).fontSize).toBe(`${MOBILE_INPUT_FONT_PX}px`);
  });
});
