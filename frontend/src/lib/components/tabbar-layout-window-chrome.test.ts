// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it, vi } from "vitest";
import { mount } from "svelte";
import TabBar from "./TabBar.svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import * as windowActions from "$lib/browser/window-actions";

const baseTab = {
  id: "tab-1",
  title: "Example",
  url: "deadbeef:/page/index.mu",
  active: true,
  loading: false,
  pinned: false,
};

const noop = () => {};

describe("TabBar titlebar double-click", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
    vi.restoreAllMocks();
  });

  it("toggles window size when the drag strip is double-clicked", async () => {
    const toggle = vi
      .spyOn(windowActions, "handleTitlebarDoubleClick")
      .mockResolvedValue({ ok: true, action: "maximized" });

    instance = await mountInBody(TabBar, {
      tabs: [baseTab],
      nativeTitlebar: false,
      mobileUI: false,
      tabHoverPreviews: false,
      splitTabId: null,
      splitViewOpen: false,
      onSelect: noop,
      onClose: noop,
      onNew: noop,
      onReorder: noop,
      onReload: noop,
      onDuplicate: noop,
      onFavorite: noop,
      onViewSource: noop,
      onDownload: noop,
      onSplit: noop,
      onCloseSplit: noop,
      onCloseOthers: noop,
      onCloseRight: noop,
      onCloseAll: noop,
      onTogglePin: noop,
    });

    const strip = document.querySelector(".drag-strip");
    expect(strip).not.toBeNull();
    strip!.dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));

    await vi.waitFor(() => {
      expect(toggle).toHaveBeenCalledTimes(1);
    });
  });

  it("reports window action errors", async () => {
    const onWindowChromeError = vi.fn();
    vi.spyOn(windowActions, "handleTitlebarDoubleClick").mockResolvedValue({
      ok: false,
      error: "Maximise failed",
    });

    instance = await mountInBody(TabBar, {
      tabs: [baseTab],
      nativeTitlebar: false,
      mobileUI: false,
      tabHoverPreviews: false,
      splitTabId: null,
      splitViewOpen: false,
      onSelect: noop,
      onClose: noop,
      onNew: noop,
      onReorder: noop,
      onReload: noop,
      onDuplicate: noop,
      onFavorite: noop,
      onViewSource: noop,
      onDownload: noop,
      onSplit: noop,
      onCloseSplit: noop,
      onCloseOthers: noop,
      onCloseRight: noop,
      onCloseAll: noop,
      onTogglePin: noop,
      onWindowChromeError,
    });

    document
      .querySelector(".drag-strip")!
      .dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));

    await vi.waitFor(() => {
      expect(onWindowChromeError).toHaveBeenCalledWith("Maximise failed");
    });
  });

  it("does not handle double-click with the native titlebar", async () => {
    const toggle = vi
      .spyOn(windowActions, "handleTitlebarDoubleClick")
      .mockResolvedValue({ ok: true, action: "maximized" });

    instance = await mountInBody(TabBar, {
      tabs: [baseTab],
      nativeTitlebar: true,
      mobileUI: false,
      tabHoverPreviews: false,
      splitTabId: null,
      splitViewOpen: false,
      onSelect: noop,
      onClose: noop,
      onNew: noop,
      onReorder: noop,
      onReload: noop,
      onDuplicate: noop,
      onFavorite: noop,
      onViewSource: noop,
      onDownload: noop,
      onSplit: noop,
      onCloseSplit: noop,
      onCloseOthers: noop,
      onCloseRight: noop,
      onCloseAll: noop,
      onTogglePin: noop,
    });

    expect(document.querySelector(".drag-strip")).not.toBeNull();
    document
      .querySelector(".drag-strip")!
      .dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));

    await new Promise((resolve) => setTimeout(resolve, 0));
    expect(toggle).not.toHaveBeenCalled();
  });
});
