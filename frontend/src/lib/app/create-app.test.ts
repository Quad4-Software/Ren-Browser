// SPDX-License-Identifier: MIT
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effect_root } from "svelte/internal/client";
import {
  createBrowserserviceMocks,
  createPluginhostMocks,
  createRuntimeMock,
} from "$lib/test/mock-bindings";
import type { AppController } from "./create-app.svelte";

const browserMocks = createBrowserserviceMocks();
const pluginMocks = createPluginhostMocks();
const runtime = createRuntimeMock();

vi.mock("@wailsio/runtime", () => ({
  Events: runtime.Events,
  System: runtime.System,
}));

vi.mock("../../../bindings/renbrowser/internal/app/browserservice.js", () => browserMocks);

vi.mock("../../../bindings/renbrowser/internal/app/pluginhost.js", () => pluginMocks);

vi.mock("$lib/plugins/api.js", () => ({
  getContributions: vi.fn(async () => ({
    panels: [],
    commands: [],
    devtools: [],
    urlSchemes: [],
  })),
  listPlugins: vi.fn(async () => []),
}));

vi.mock("$lib/plugins/lifecycle.js", () => ({
  activateAllPlugins: vi.fn(async () => {}),
  deactivateAllPlugins: vi.fn(async () => {}),
  handlePluginScheme: vi.fn(async () => false),
}));

vi.mock("$lib/browser/download", async () => {
  const actual =
    await vi.importActual<typeof import("$lib/browser/download")>("$lib/browser/download");
  return {
    ...actual,
    downloadPageContent: vi.fn(async () => ({ ok: true, message: "saved", pending: false })),
  };
});

vi.mock("$lib/micron/wasm-loader", async (importOriginal) => {
  const actual = await importOriginal<typeof import("$lib/micron/wasm-loader")>();
  return {
    ...actual,
    isMicronWasmAvailable: vi.fn(async () => false),
    isMicronWasmBundled: vi.fn(() => false),
    preloadNomadMicronWasm: vi.fn(async () => false),
  };
});

vi.mock("$lib/micron/render-page", async (importOriginal) => {
  const actual = await importOriginal<typeof import("$lib/micron/render-page")>();
  return {
    ...actual,
    ensureMicronWasmReady: vi.fn(async () => false),
    resolveMicronWasmParserLabel: vi.fn(async () => ""),
  };
});

async function withApp(run: (app: AppController) => void | Promise<void>): Promise<void> {
  const { createApp } = await import("./create-app.svelte");
  let app!: AppController;
  const dispose = effect_root(() => {
    app = createApp();
  });
  try {
    await run(app);
  } finally {
    dispose();
  }
}

describe("createApp controller", () => {
  beforeEach(() => {
    Object.assign(browserMocks, createBrowserserviceMocks());
    Object.assign(pluginMocks, createPluginhostMocks());
    runtime.Events.On.mockClear();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("starts with a single blank active tab", async () => {
    await withApp((app) => {
      expect(app.tabs).toHaveLength(1);
      expect(app.tabs[0].active).toBe(true);
      expect(app.tabs[0].url).toBe("");
      expect(app.atTabLimit).toBe(false);
      expect(app.activePanel).toBe("browser");
    });
  });

  it("opens and closes tabs", async () => {
    await withApp((app) => {
      const firstId = app.tabs[0].id;
      app.newTab();
      expect(app.tabs).toHaveLength(2);
      expect(app.tabs.filter((t) => t.active)).toHaveLength(1);
      expect(app.activeTabId).not.toBe(firstId);
      app.closeTab(app.activeTabId);
      expect(app.tabs).toHaveLength(1);
      expect(app.tabs[0].id).toBe(firstId);
    });
  });

  it("resets the last tab instead of removing it", async () => {
    await withApp((app) => {
      const id = app.tabs[0].id;
      app.url = "about:";
      app.closeTab(id);
      expect(app.tabs).toHaveLength(1);
      expect(app.tabs[0].url).toBe("");
      expect(app.url).toBe("");
    });
  });

  it("pins tabs and refuses to close pinned tabs", async () => {
    await withApp((app) => {
      app.newTab();
      const pinnedId = app.tabs[0].id;
      app.togglePinTab(pinnedId);
      expect(app.tabs.find((t) => t.id === pinnedId)?.pinned).toBe(true);
      app.closeTab(pinnedId);
      expect(app.tabs.some((t) => t.id === pinnedId)).toBe(true);
    });
  });

  it("navigates via openPage and updates tab url", async () => {
    browserMocks.Navigate.mockResolvedValue({
      url: "about:",
      html: "<article class='about-page'>About</article>",
      contentType: "text/html",
      raw: "",
      binaryB64: "",
      path: "",
      error: "",
      errorKind: "",
      durationMs: 2,
      pageFg: "",
      pageBg: "",
      fromCache: false,
      cachedAt: 0,
      hops: -1,
    });

    await withApp(async (app) => {
      await app.openPage("about:");
      expect(browserMocks.Navigate).toHaveBeenCalled();
      expect(app.url).toBe("about:");
      expect(app.html).toContain("about-page");
      expect(app.loading).toBe(false);
    });
  });

  it("toggles downloads menu and loads history", async () => {
    browserMocks.ListDownloads.mockResolvedValue([
      { name: "page.mu", path: "/tmp/page.mu", size: 12, modifiedAt: 1 },
    ]);
    await withApp(async (app) => {
      expect(app.downloadsOpen).toBe(false);
      app.toggleDownloads();
      expect(app.downloadsOpen).toBe(true);
      await vi.waitFor(() => {
        expect(browserMocks.ListDownloads).toHaveBeenCalled();
      });
    });
  });

  it("persists tabs after debounce", async () => {
    vi.useFakeTimers();
    await withApp(async (app) => {
      app.newTab();
      expect(browserMocks.SaveTabs).not.toHaveBeenCalled();
      await vi.advanceTimersByTimeAsync(300);
      expect(browserMocks.SaveTabs).toHaveBeenCalled();
    });
  });

  it("registers deeplink handler on mount", async () => {
    browserMocks.TakePendingDeepLink.mockResolvedValue("");
    await withApp(async (app) => {
      const cleanup = app.mount();
      await vi.waitFor(() => {
        expect(runtime.Events.On).toHaveBeenCalledWith("app:deeplink", expect.any(Function));
        expect(browserMocks.TakePendingDeepLink).toHaveBeenCalled();
      });
      cleanup();
    });
  });
});
