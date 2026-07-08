// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it, vi } from "vitest";
import { mount } from "svelte";
import { SvelteSet } from "svelte/reactivity";
import { cleanupMount, mountInBody } from "$lib/test/svelte-mount";
import InitialSetupModal from "./InitialSetupModal.svelte";
import type { AppController } from "$lib/app/create-app.svelte";
import type { CommunityInterface } from "../../../bindings/renbrowser/internal/rns/models.js";

describe("InitialSetupModal component", () => {
  let instance: ReturnType<typeof mount> | null = null;

  afterEach(() => {
    cleanupMount(instance);
    instance = null;
  });

  const createMockApp = (overrides = {}): Partial<AppController> => {
    const selected = new SvelteSet<number>();
    return {
      initialSetupOpen: true,
      initialSetupStep: "welcome",
      initialSetupError: "",
      initialSetupBusy: false,
      suggestedItems: [] as CommunityInterface[],
      suggestedLoading: false,
      communityItems: [] as CommunityInterface[],
      communityLoading: false,
      communityError: "",
      communityFromBundle: false,
      communityFilter: "",
      communitySelected: selected,
      configText: "[[Auto Discovery]]\n  enabled = yes\n",
      configPath: "/mock/reticulum.conf",
      configSaving: false,
      configError: "",
      setInitialSetupStep: vi.fn(),
      skipInitialSetupAutoOnly: vi.fn(),
      loadSuggestedPreview: vi.fn(),
      applySuggestedSetup: vi.fn(),
      toggleCommunitySelection: vi.fn(),
      importInitialSetupSelection: vi.fn(),
      saveInitialSetupConfig: vi.fn(),
      reloadConfigFromDisk: vi.fn(),
      openConfigFolder: vi.fn(),
      loadCommunityInterfaces: vi.fn(),
      ...overrides,
    };
  };

  it("does not render when app.initialSetupOpen is false", async () => {
    const app = createMockApp({ initialSetupOpen: false }) as AppController;
    instance = await mountInBody(InitialSetupModal, { app });

    const dialog = document.querySelector(".dialog");
    expect(dialog).toBeNull();
  });

  it("renders welcome screen choices", async () => {
    const app = createMockApp() as AppController;
    instance = await mountInBody(InitialSetupModal, { app });

    const title = document.getElementById("initial-setup-title");
    expect(title?.textContent).toBe("Welcome to Ren Browser");

    const choices = document.querySelectorAll(".choice");
    expect(choices.length).toBe(4);
  });

  it("triggers step navigation on choosing a route", async () => {
    const app = createMockApp() as AppController;
    instance = await mountInBody(InitialSetupModal, { app });

    const choices = document.querySelectorAll(".choice");
    // Click suggested interfaces
    (choices[0] as HTMLButtonElement).click();
    expect(app.setInitialSetupStep).toHaveBeenCalledWith("suggested");
  });

  it("renders suggested screen layout with preview items", async () => {
    const items: CommunityInterface[] = [
      {
        id: 1,
        name: "Test Community Hub",
        type: "TCPClientInterface",
        typeName: "TCP Client",
        network: "MichMesh",
        host: "rns.example.com",
        port: 7822,
        status: "online",
        config: "",
        installed: false,
      },
    ];
    const app = createMockApp({
      initialSetupStep: "suggested",
      suggestedItems: items,
    }) as AppController;

    instance = await mountInBody(InitialSetupModal, { app });

    const stepHint = document.querySelector(".step-hint");
    expect(stepHint?.textContent).toContain("hubs from the public directory");

    const name = document.querySelector(".suggested-list .name");
    expect(name?.textContent).toBe("Test Community Hub");

    const meta = document.querySelector(".suggested-list .meta");
    expect(meta?.textContent).toBe("TCP Client · MichMesh");
  });

  it("renders custom config editor layout", async () => {
    const app = createMockApp({
      initialSetupStep: "config",
    }) as AppController;

    instance = await mountInBody(InitialSetupModal, { app });

    const textarea = document.querySelector(".editor") as HTMLTextAreaElement;
    expect(textarea).not.toBeNull();
    expect(textarea.value).toBe(app.configText);
  });
});
