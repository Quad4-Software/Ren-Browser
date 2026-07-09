// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it, vi } from "vitest";
import { mount } from "svelte";
import ContentViewer from "./ContentViewer.svelte";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import { LARGE_MICRON_RAW_BYTES } from "$lib/micron/render-page";
import * as renderPage from "$lib/micron/render-page";

const noop = () => {};

describe("ContentViewer micron performance regressions", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
    vi.restoreAllMocks();
  });

  it("uses server html for large micron pages instead of client re-parse", async () => {
    const spy = vi.spyOn(renderPage, "renderClientMicronPage");
    const serverHtml = "<div data-server='1'><span class='Mu-mnt'>S</span></div>";
    const raw = "x".repeat(LARGE_MICRON_RAW_BYTES);

    instance = await mountInBody(ContentViewer, {
      html: serverHtml,
      contentType: "micron",
      loading: false,
      error: "",
      currentURL: "abb3ebcd03cb2388a838e70c001291f9:/page/big.mu",
      raw,
      showSource: false,
      findOpen: false,
      micronEngine: "js",
      micronPreserveLayout: false,
      onNavigate: noop,
      onRetry: noop,
      onReloadFresh: noop,
      onShowSourceChange: noop,
      onFindClose: noop,
    });

    const content = document.querySelector(".content.micron");
    expect(content).not.toBeNull();
    expect(content!.innerHTML).toContain("data-server");
    expect(spy).not.toHaveBeenCalled();
  });

  it("client-renders small micron pages when engine is js", async () => {
    const spy = vi
      .spyOn(renderPage, "renderClientMicronPage")
      .mockReturnValue("<div data-client='1'>ok</div>");

    instance = await mountInBody(ContentViewer, {
      html: "<div data-server='1'>ignored</div>",
      contentType: "micron",
      loading: false,
      error: "",
      currentURL: "abb3ebcd03cb2388a838e70c001291f9:/page/small.mu",
      raw: "`>Hi\nline",
      showSource: false,
      findOpen: false,
      micronEngine: "js",
      micronPreserveLayout: false,
      onNavigate: noop,
      onRetry: noop,
      onReloadFresh: noop,
      onShowSourceChange: noop,
      onFindClose: noop,
    });

    expect(spy).toHaveBeenCalled();
    const content = document.querySelector(".content.micron");
    expect(content).not.toBeNull();
    expect(content!.innerHTML).toContain("data-client");
  });

  it("uses server html when micron engine is go", async () => {
    const spy = vi.spyOn(renderPage, "renderClientMicronPage");
    const serverHtml = "<div data-go='1'><span class='Mu-mnt'>G</span></div>";

    instance = await mountInBody(ContentViewer, {
      html: serverHtml,
      contentType: "micron",
      loading: false,
      error: "",
      currentURL: "abb3ebcd03cb2388a838e70c001291f9:/page/go.mu",
      raw: "`>Hi\nline",
      showSource: false,
      findOpen: false,
      micronEngine: "go",
      onNavigate: noop,
      onRetry: noop,
      onReloadFresh: noop,
      onShowSourceChange: noop,
      onFindClose: noop,
    });

    expect(spy).not.toHaveBeenCalled();
    expect(document.querySelector(".content.micron")!.innerHTML).toContain("data-go");
  });
});
