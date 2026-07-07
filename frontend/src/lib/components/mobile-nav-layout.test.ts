import { afterEach, describe, expect, it } from "vitest";
import { mount, flushSync, tick } from "svelte";
import MobileNav from "./MobileNav.svelte";
import { cleanupMount } from "$lib/test/svelte-mount";
import { parsePx } from "$lib/browser/layout-regression";

const noop = () => {};

describe("MobileNav layout regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    document.body.innerHTML = "";
    instance = null;
  });

  it("shows when rendered inside a mobile shell and truncates labels", async () => {
    const shell = document.createElement("div");
    shell.className = "app-shell mobile-ui";
    document.body.appendChild(shell);

    flushSync(() => {
      instance = mount(MobileNav, {
        target: shell,
        props: {
          activePanel: "browser",
          pluginPanels: [],
          mobileDevTools: false,
          downloadsOpen: false,
          activeDownloadCount: 0,
          onPanel: noop,
          onToggleDownloads: noop,
        },
      });
    });
    await tick();

    const nav = shell.querySelector(".mobile-nav") as HTMLElement | null;
    expect(nav).not.toBeNull();
    expect(getComputedStyle(nav!).display).toBe("grid");
    expect(parsePx(getComputedStyle(nav!).width)).toBeGreaterThan(0);

    const label = shell.querySelector(".mobile-nav button span") as HTMLElement | null;
    expect(label).not.toBeNull();
    expect(getComputedStyle(label!).textOverflow).toBe("ellipsis");
  });
});
