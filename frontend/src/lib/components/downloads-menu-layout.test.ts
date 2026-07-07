import { afterEach, describe, expect, it } from "vitest";
import { mount } from "svelte";
import DownloadsMenu from "./DownloadsMenu.svelte";
import { expectsEllipsis, expectsNoHorizontalScroll } from "$lib/browser/layout-regression";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";

const noop = () => {};

describe("DownloadsMenu layout regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    document.body.style.width = "";
    instance = null;
  });

  it("truncates long download names in the sheet layout", async () => {
    const longName = "very-long-download-name-that-should-not-expand-the-menu".repeat(2) + ".zip";
    document.body.style.width = "360px";
    instance = await mountInBody(DownloadsMenu, {
      open: true,
      downloads: [
        {
          name: longName,
          path: `/tmp/${longName}`,
          size: 1024,
          modifiedAt: 1_700_000_000,
        },
      ],
      downloadDir: "/home/user/Downloads",
      variant: "sheet",
      onDownloadPage: noop,
      onOpenFile: noop,
      onOpenFolder: noop,
      onClose: noop,
    });

    const menu = document.querySelector(".menu.sheet");
    expect(menu).not.toBeNull();
    expectsNoHorizontalScroll(menu!);

    const name = document.querySelector(".file-row .name");
    expect(name).not.toBeNull();
    expectsEllipsis(name!, { requireMinWidthZero: true });
    expect(name!.textContent).toContain(".zip");
  });

  it("truncates active download names in the header row", async () => {
    instance = await mountInBody(DownloadsMenu, {
      open: true,
      active: [
        {
          id: "dl-1",
          url: "mesh:/file/big.bin",
          name: "x".repeat(96) + ".bin",
          status: "downloading",
          received: 12,
          total: 100,
          startedAt: 1,
          updatedAt: 2,
          speedBps: 4,
          etaSeconds: 10,
        },
      ],
      downloads: [],
      downloadDir: "/home/user/Downloads",
      variant: "sheet",
      onDownloadPage: noop,
      onOpenFile: noop,
      onOpenFolder: noop,
      onClose: noop,
    });

    const name = document.querySelector(".active-head .name");
    expect(name).not.toBeNull();
    expectsEllipsis(name!, { requireMinWidthZero: true });
  });
});
