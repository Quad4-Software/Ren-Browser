<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onMount } from "svelte";
  import { SvelteSet } from "svelte/reactivity";
  import { Events, System } from "@wailsio/runtime";
  import {
    AddFavorite,
    ClearDevLogs,
    ClearPageCache,
    ConfigPath,
    ExportDevLogs,
    ExportTheme,
    FetchCommunityInterfaces,
    GetBrowserPrefs,
    GetDownloadDir,
    GetDevLogs,
    GetFavorites,
    GetBrowsingHistory,
    GetKeybinds,
    GetNetworkLog,
    GetStoreHealth,
    GetTabs,
    GetTheme,
    GetWindowChrome,
    GetReticulumConfigText,
    GetPageCacheStats,
    GoBack,
    GoForward,
    HistoryState,
    IdentifyToNode,
    ImportCommunityInterfaces,
    ImportTheme,
    ListDownloads,
    ListInterfaces,
    ListNodes,
    ListSystemFonts,
    Navigate,
    NavigateFresh,
    OpenDownloadPath,
    OpenURL,
    OpenFreshURL,
    PickDownloadDir,
    ResetSettings,
    ResetDatabase,
    ReloadReticulumConfig,
    SaveTabs,
    SaveReticulumConfigText,
    SetBrowserPrefs,
    SetDownloadDir,
    SetInterfaceEnabled,
    SetKeybinds,
    SetLogLevel,
    SetNativeTitlebar,
    SetTheme,
    ShowDownloadDir,
    SyncMobileChrome,
  } from "../bindings/renbrowser/internal/app/browserservice.js";
  import type { WindowChrome } from "../bindings/renbrowser/internal/app/models.js";
  import type { DownloadRow } from "$lib/components/DownloadsMenu.svelte";
  import { exportFilename } from "$lib/brand";
  import BrowserChrome from "$lib/components/BrowserChrome.svelte";
  import TabBar from "$lib/components/TabBar.svelte";
  import ContentViewer from "$lib/components/ContentViewer.svelte";
  import MicronEditor from "$lib/components/MicronEditor.svelte";
  import ReticulumConfigEditor from "$lib/components/ReticulumConfigEditor.svelte";
  import DiscoveryPanel from "$lib/components/DiscoveryPanel.svelte";
  import HistoryPanel from "$lib/components/HistoryPanel.svelte";
  import DevToolsPanel from "$lib/components/DevToolsPanel.svelte";
  import SettingsPanel from "$lib/components/SettingsPanel.svelte";
  import SplitPane from "$lib/components/SplitPane.svelte";
  import SplitTabPicker from "$lib/components/SplitTabPicker.svelte";
  import MobileNav from "$lib/components/MobileNav.svelte";
  import DownloadsMenu from "$lib/components/DownloadsMenu.svelte";
  import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
  import AppStoreError from "$lib/components/AppStoreError.svelte";
  import { isStoreBlockingKind } from "$lib/browser/errors";
  import {
    defaultTheme,
    applyTheme,
    mobileChromeBg,
    mobileChromeUsesLightIcons,
    type ThemeSettings,
  } from "$lib/theme/tokens";
  import type { CommunityInterface } from "$lib/components/CommunityInterfaces.svelte";
  import {
    defaultKeybinds,
    matchKeybind,
    isKeybindRecording,
    mergeKeybinds,
    type KeybindAction,
    type KeybindSettings,
  } from "$lib/browser/keybinds";
  import {
    getContributionsSnapshot,
    setContributions,
    findPanel,
    parsePanelKey,
  } from "$lib/plugins/registry.js";
  import { getContributions as fetchContributions } from "$lib/plugins/api.js";
  import { PluginsDir } from "../bindings/renbrowser/internal/app/pluginhost.js";
  import {
    activateAllPlugins,
    deactivateAllPlugins,
    handlePluginScheme,
  } from "$lib/plugins/lifecycle.js";
  import { dispatchPluginCommand, matchPluginKeybind } from "$lib/plugins/command-dispatch.js";
  import type { ActivePanel, ContributionsSnapshot } from "$lib/plugins/api-types.js";
  import PluginPanelHost from "$lib/components/PluginPanelHost.svelte";
  import { downloadPageContent } from "$lib/browser/download";
  import {
    canOpenTab,
    normalizeReticulumURL,
    orderTabsPinnedFirst,
    pinTabInList,
    reorderTabsInList,
    tabTitleFromURL,
    unpinTabInList,
    type Tab,
    type TabPage,
  } from "$lib/browser/url";
  import {
    BUNDLED_MICRON_WASM_PARSER_ID,
    ensureMicronWasmReady,
    micronRendererBadgeLabel,
    normalizeMicronRendererPreference,
    resolveEffectiveMicronEngine,
    resolveMicronWasmParserLabel,
    shouldPreloadMicronWasm,
    type MicronRendererPreference,
  } from "$lib/micron/render-page";
  import { isMicronWasmAvailable } from "$lib/micron/wasm-loader";
  import { randomId } from "$lib/browser/id";
  import { initUILocale, setUILocale, t, detectOSLocale } from "$lib/i18n/i18n.svelte";

  type Node = {
    hash: string;
    name: string;
    hops: number;
    lastSeen: number;
  };

  type PageResponse = {
    url: string;
    nodeHash: string;
    path: string;
    contentType: string;
    html: string;
    raw: string;
    pageFg?: string;
    pageBg?: string;
    durationMs: number;
    fromCache?: boolean;
    cachedAt?: number;
    hops?: number;
    error?: string;
    errorKind?: string;
  };

  type DevLogEntry = {
    time: number;
    level: string;
    message: string;
    detail?: string;
  };

  type NetworkEntry = {
    time: number;
    url: string;
    nodeHash: string;
    path: string;
    durationMs: number;
    bytes: number;
    fromCache: boolean;
    hops: number;
    error?: string;
  };

  type InterfaceRow = {
    name: string;
    type: string;
    enabled: boolean;
    online: boolean;
    txBytes: number;
    rxBytes: number;
  };

  type TabSnapshot = {
    id: string;
    title: string;
    url: string;
    active: boolean;
    pinned?: boolean;
    html?: string;
    contentType?: string;
    error?: string;
    errorKind?: string;
    durationMs?: number;
    lastRaw?: string;
    pageFg?: string;
    pageBg?: string;
  };

  type StoreHealth = {
    ok: boolean;
    kind?: string;
    detail?: string;
    path: string;
  };

  type HistoryEntry = {
    id: number;
    url: string;
    title: string;
    nodeHash: string;
    visitedAt: number;
  };

  let activePanel = $state<ActivePanel>("browser");
  let pluginContributions = $state<ContributionsSnapshot>({
    panels: [],
    commands: [],
    devtools: [],
    urlSchemes: [],
  });
  let pluginToast = $state("");
  let pluginsDir = $state("");
  let url = $state("");
  let loading = $state(false);
  let html = $state("");
  let contentType = $state("");
  let error = $state("");
  let errorKind = $state("");
  let durationMs = $state(0);
  let hops = $state(-1);
  let pageFg = $state("");
  let pageBg = $state("");
  let nodes = $state<Node[]>([]);
  let logs = $state<DevLogEntry[]>([]);
  let network = $state<NetworkEntry[]>([]);
  let favorites = $state<string[]>([]);
  let history = $state<HistoryEntry[]>([]);
  let interfaces = $state<InterfaceRow[]>([]);
  let configPath = $state("");
  let logLevel = $state(3);
  let systemFonts = $state<string[]>(["system-ui", "sans-serif", "monospace"]);
  let theme = $state<ThemeSettings>(defaultTheme());
  let keybinds = $state<KeybindSettings>(defaultKeybinds());
  let downloadDir = $state("");
  let downloads = $state<DownloadRow[]>([]);
  let downloadsOpen = $state(false);
  let findOpen = $state(false);
  let canGoBack = $state(false);
  let canGoForward = $state(false);
  let lastRaw = $state("");
  let fromCache = $state(false);
  let cachedAt = $state(0);
  let showSource = $state(false);
  let openLinksInNewTab = $state(true);
  let nativeTitlebar = $state(false);
  let uiLanguage = $state("");
  let docsLanguage = $state("");
  let micronRenderer = $state<MicronRendererPreference>("auto");
  let micronWasmEnabled = $state(true);
  let micronWasmReady = $state(false);
  let micronWasmAvailable = $state(false);
  let micronWasmParserId = $state(BUNDLED_MICRON_WASM_PARSER_ID);
  let micronWasmParserLabel = $state("");
  let identifying = $state(false);
  let identifyConfirmOpen = $state(false);
  let resetDbConfirmOpen = $state(false);
  let storeHealth = $state<StoreHealth>({ ok: true, path: "" });
  let meshOnline = $state(true);
  let splitViewOpen = $state(false);
  let splitTabId = $state<string | null>(null);
  let splitRatio = $state(52);
  const desktopChrome = System.IsDesktop();
  const mobileUI = !desktopChrome;

  let configText = $state("");
  let configSaving = $state(false);
  let configError = $state("");
  let pageCacheEntries = $state(0);
  let pageCacheMax = $state(128);
  let pageCacheClearing = $state(false);
  let communityItems = $state<CommunityInterface[]>([]);
  let communityLoading = $state(false);
  let communityImporting = $state(false);
  let communityError = $state("");
  let communityFilter = $state("");
  let communitySelected = new SvelteSet<number>();

  let tabs = $state<Tab[]>([{ id: randomId(), title: "", url: "", active: true }]);

  const effectiveMicronEngine = $derived(
    resolveEffectiveMicronEngine(micronRenderer, {
      wasmEnabled: micronWasmEnabled,
      wasmAvailable: micronWasmAvailable,
      wasmReady: micronWasmReady,
      hasServerHtml: html.trim().length > 0,
    }),
  );

  const micronRendererBadge = $derived(
    contentType === "micron" && !showSource
      ? micronRendererBadgeLabel(micronRenderer, effectiveMicronEngine, micronWasmParserLabel)
      : "",
  );

  const canIdentify = $derived(
    /^[a-f0-9]{32}:/i.test(url.trim()) || /^[a-f0-9]{32}$/i.test(url.trim()),
  );
  const activeTabId = $derived(tabs.find((tab) => tab.active)?.id ?? "");
  const splitTab = $derived(
    splitTabId && splitTabId !== activeTabId ? tabs.find((tab) => tab.id === splitTabId) : null,
  );
  const storeErrorVisible = $derived(!storeHealth.ok && isStoreBlockingKind(storeHealth.kind));
  const activePluginPanel = $derived.by(() => {
    const parsed = parsePanelKey(activePanel);
    if (!parsed) {
      return null;
    }
    return findPanel(activePanel);
  });

  let persistTimer: ReturnType<typeof setTimeout> | undefined;

  function setPanel(panel: ActivePanel) {
    const next = activePanel === panel ? "browser" : panel;
    activePanel = next;
    if (next === "settings") {
      void loadPageCacheStats();
    }
  }

  function showPluginToast(message: string) {
    pluginToast = message;
    setTimeout(() => {
      if (pluginToast === message) {
        pluginToast = "";
      }
    }, 2500);
  }

  async function bootPlugins() {
    pluginsDir = (await PluginsDir()) ?? "";
    const snapshot = await fetchContributions();
    setContributions(snapshot);
    pluginContributions = getContributionsSnapshot();
    await activateAllPlugins({
      getCurrentURL: () => url,
      navigate: (next) => void browseURL(next),
      showToast: showPluginToast,
    });
  }

  async function reloadPlugins() {
    await deactivateAllPlugins();
    await bootPlugins();
  }

  function emptyPage(): TabPage {
    return {
      html: "",
      contentType: "",
      error: "",
      errorKind: "",
      durationMs: 0,
      lastRaw: "",
      pageFg: "",
      pageBg: "",
      fromCache: false,
      cachedAt: 0,
      hops: -1,
      showSource: false,
    };
  }

  function pageFromResponse(page: PageResponse): TabPage {
    return {
      html: page.html ?? "",
      contentType: page.contentType ?? "",
      error: page.error ?? "",
      errorKind: page.errorKind ?? "",
      durationMs: page.durationMs ?? 0,
      lastRaw: page.raw ?? "",
      pageFg: page.pageFg ?? "",
      pageBg: page.pageBg ?? "",
      fromCache: page.fromCache ?? false,
      cachedAt: page.cachedAt ?? 0,
      hops: page.hops ?? -1,
      showSource: false,
    };
  }

  function currentPageState(): TabPage {
    return {
      html,
      contentType,
      error,
      errorKind,
      durationMs,
      lastRaw,
      pageFg,
      pageBg,
      fromCache,
      cachedAt,
      hops,
      showSource,
    };
  }

  function applyPageState(page: TabPage) {
    html = page.html ?? "";
    contentType = page.contentType ?? "";
    error = page.error ?? "";
    errorKind = page.errorKind ?? "";
    durationMs = page.durationMs ?? 0;
    lastRaw = page.lastRaw ?? "";
    pageFg = page.pageFg ?? "";
    pageBg = page.pageBg ?? "";
    fromCache = page.fromCache ?? false;
    cachedAt = page.cachedAt ?? 0;
    hops = page.hops ?? -1;
    showSource = page.showSource ?? false;
  }

  function clearPageState() {
    applyPageState(emptyPage());
  }

  function syncActiveTabPage() {
    const page = currentPageState();
    tabs = tabs.map((tab) =>
      tab.active
        ? {
            ...tab,
            url,
            title: tabTitleFromURL(url, nodes),
            page,
          }
        : tab,
    );
  }

  function schedulePersistTabs() {
    if (persistTimer) {
      clearTimeout(persistTimer);
    }
    persistTimer = setTimeout(() => {
      void persistTabs();
    }, 250);
  }

  async function persistTabs() {
    syncActiveTabPage();
    const payload: TabSnapshot[] = tabs.map((tab) => ({
      id: tab.id,
      title: tab.title,
      url: tab.url,
      active: tab.active,
      pinned: tab.pinned,
      html: tab.page?.html,
      contentType: tab.page?.contentType,
      error: tab.page?.error,
      errorKind: tab.page?.errorKind,
      durationMs: tab.page?.durationMs,
      lastRaw: tab.page?.lastRaw,
      pageFg: tab.page?.pageFg,
      pageBg: tab.page?.pageBg,
    }));
    await SaveTabs(payload);
  }

  function restoreTabs(saved: TabSnapshot[]) {
    if (!saved.length) {
      return;
    }
    tabs = orderTabsPinnedFirst(
      saved.map((tab) => ({
        id: tab.id || randomId(),
        title: tab.title || tabTitleFromURL(tab.url, nodes),
        url: tab.url ?? "",
        active: tab.active,
        pinned: tab.pinned,
        page: {
          html: tab.html ?? "",
          contentType: tab.contentType ?? "",
          error: tab.error ?? "",
          errorKind: tab.errorKind ?? "",
          durationMs: tab.durationMs ?? 0,
          lastRaw: tab.lastRaw ?? "",
          pageFg: tab.pageFg ?? "",
          pageBg: tab.pageBg ?? "",
        },
      })),
    );
    if (!tabs.some((tab) => tab.active)) {
      tabs = tabs.map((tab, index) => ({ ...tab, active: index === 0 }));
    }
    const selected = tabs.find((tab) => tab.active) ?? tabs[0];
    url = selected.url;
    if (selected.page) {
      applyPageState(selected.page);
    } else {
      clearPageState();
    }
  }

  function refreshTabTitles() {
    tabs = tabs.map((tab) => ({
      ...tab,
      title: tab.url ? tabTitleFromURL(tab.url, nodes) : t("tab.new"),
    }));
  }

  function reconcileSplitView() {
    if (tabs.length < 2) {
      closeSplitView();
      return;
    }
    if (splitTabId && splitTabId === activeTabId) {
      splitTabId = null;
    }
  }

  function setActiveTab(id: string) {
    syncActiveTabPage();
    tabs = tabs.map((tab) => ({ ...tab, active: tab.id === id }));
    const selected = tabs.find((tab) => tab.id === id);
    url = selected?.url ?? "";
    if (selected?.page) {
      applyPageState(selected.page);
    } else {
      clearPageState();
    }
    loading = false;
    if (splitViewOpen && splitTabId === id) {
      splitTabId = null;
    }
    schedulePersistTabs();
  }

  function closeTab(id: string) {
    syncActiveTabPage();
    const closing = tabs.find((tab) => tab.id === id);
    if (closing?.pinned) {
      return;
    }
    if (splitTabId === id) {
      splitTabId = null;
    }
    if (tabs.length === 1) {
      tabs = [
        { ...tabs[0], id: tabs[0].id, title: t("tab.new"), url: "", active: true, page: emptyPage() },
      ];
      url = "";
      clearPageState();
      closeSplitView();
      schedulePersistTabs();
      return;
    }
    const idx = tabs.findIndex((tab) => tab.id === id);
    const closingActive = tabs[idx]?.active;
    tabs = tabs.filter((tab) => tab.id !== id);
    if (closingActive) {
      const next = tabs[Math.max(0, idx - 1)];
      setActiveTab(next.id);
      reconcileSplitView();
      return;
    }
    schedulePersistTabs();
    reconcileSplitView();
  }

  function resetToSingleTab() {
    tabs = [
      { id: randomId(), title: t("tab.new"), url: "", active: true, page: emptyPage() },
    ];
    url = "";
    closeSplitView();
    clearPageState();
    schedulePersistTabs();
  }

  function closeOtherTabs(keepId: string) {
    syncActiveTabPage();
    const keep = tabs.find((tab) => tab.id === keepId);
    if (!keep) {
      return;
    }
    if (splitTabId && splitTabId !== keepId && !tabs.find((tab) => tab.id === splitTabId)?.pinned) {
      splitTabId = null;
    }
    tabs = tabs
      .filter((tab) => tab.pinned || tab.id === keepId)
      .map((tab) => ({ ...tab, active: tab.id === keepId }));
    url = keep.url;
    applyPageState(keep.page ?? emptyPage());
    schedulePersistTabs();
    reconcileSplitView();
  }

  function closeTabsToRight(tabId: string) {
    syncActiveTabPage();
    const idx = tabs.findIndex((tab) => tab.id === tabId);
    if (idx < 0) {
      return;
    }
    const removed = tabs
      .slice(idx + 1)
      .filter((tab) => !tab.pinned)
      .map((tab) => tab.id);
    if (splitTabId && removed.includes(splitTabId)) {
      splitTabId = null;
    }
    const next = tabs.filter((tab, index) => index <= idx || tab.pinned);
    if (!next.some((tab) => tab.active)) {
      next[next.length - 1] = { ...next[next.length - 1], active: true };
      url = next[next.length - 1].url;
      applyPageState(next[next.length - 1].page ?? emptyPage());
    }
    tabs = next;
    schedulePersistTabs();
    reconcileSplitView();
  }

  function closeAllTabs() {
    syncActiveTabPage();
    const pinned = tabs.filter((tab) => tab.pinned);
    if (!pinned.length) {
      resetToSingleTab();
      return;
    }
    if (splitTabId && !pinned.some((tab) => tab.id === splitTabId)) {
      splitTabId = null;
    }
    const activePinned = pinned.find((tab) => tab.active) ?? pinned[0];
    tabs = pinned.map((tab) => ({ ...tab, active: tab.id === activePinned.id }));
    url = activePinned.url;
    applyPageState(activePinned.page ?? emptyPage());
    schedulePersistTabs();
    reconcileSplitView();
  }

  function togglePinTab(id: string) {
    syncActiveTabPage();
    const tab = tabs.find((item) => item.id === id);
    if (!tab) {
      return;
    }
    tabs = tab.pinned ? unpinTabInList(tabs, id) : pinTabInList(tabs, id);
    schedulePersistTabs();
  }

  function splitTabView(tabId: string) {
    if (tabs.length < 2) {
      return;
    }
    const tab = tabs.find((item) => item.id === tabId);
    if (!tab) {
      return;
    }
    splitViewOpen = true;
    activePanel = "browser";
    if (tabId !== activeTabId) {
      splitTabId = tabId;
      return;
    }
    const others = tabs.filter((item) => item.id !== activeTabId);
    splitTabId = others.length === 1 ? others[0].id : null;
  }

  function selectSplitTab(tabId: string) {
    if (!splitViewOpen || tabId === activeTabId) {
      return;
    }
    splitTabId = tabId;
  }

  function closeSplitView() {
    splitViewOpen = false;
    splitTabId = null;
  }

  function setSplitTabShowSource(tabId: string, value: boolean) {
    tabs = tabs.map((tab) =>
      tab.id === tabId
        ? {
            ...tab,
            page: {
              ...(tab.page ?? emptyPage()),
              showSource: value,
            },
          }
        : tab,
    );
    schedulePersistTabs();
  }

  async function downloadTab(id: string) {
    const tab = tabs.find((item) => item.id === id);
    if (!tab?.url) {
      return;
    }
    const page = tab.active ? currentPageState() : (tab.page ?? emptyPage());
    const payload = page.lastRaw || page.html;
    await downloadPageContent(tab.url, page.contentType, payload);
    await loadDownloads();
  }

  function viewSourceTab(id: string) {
    if (!tabs.find((tab) => tab.id === id)?.active) {
      setActiveTab(id);
    }
    showSource = true;
    activePanel = "browser";
    syncActiveTabPage();
    schedulePersistTabs();
  }

  function newTab() {
    if (!canOpenTab(tabs.length)) {
      return;
    }
    syncActiveTabPage();
    tabs = tabs.map((tab) => ({ ...tab, active: false }));
    const tab: Tab = {
      id: randomId(),
      title: t("tab.new"),
      url: "",
      active: true,
      page: emptyPage(),
    };
    tabs = [...tabs, tab];
    url = "";
    clearPageState();
    activePanel = "browser";
    schedulePersistTabs();
  }

  function reorderTabs(fromId: string, toId: string) {
    tabs = reorderTabsInList(tabs, fromId, toId);
    schedulePersistTabs();
  }

  function reloadTab(id: string) {
    const tab = tabs.find((t) => t.id === id);
    if (!tab?.url) {
      return;
    }
    if (!tab.active) {
      setActiveTab(id);
    }
    void openPage(tab.url);
  }

  function duplicateTab(id: string) {
    if (!canOpenTab(tabs.length)) {
      return;
    }
    syncActiveTabPage();
    const tab = tabs.find((t) => t.id === id);
    if (!tab) {
      return;
    }
    const dup: Tab = {
      id: randomId(),
      title: tab.title,
      url: tab.url,
      active: true,
      page: tab.page ? { ...tab.page } : emptyPage(),
    };
    tabs = [...tabs.map((t) => ({ ...t, active: false })), dup];
    url = dup.url;
    applyPageState(dup.page ?? emptyPage());
    schedulePersistTabs();
  }

  async function favoriteTab(id: string) {
    const tab = tabs.find((t) => t.id === id);
    if (!tab?.url) {
      return;
    }
    favorites = (await AddFavorite(tab.url)) as string[];
  }

  function applyPageToTab(tabId: string, page: TabPage, pageUrl: string) {
    tabs = tabs.map((tab) =>
      tab.id === tabId
        ? {
            ...tab,
            url: pageUrl,
            title: tabTitleFromURL(pageUrl, nodes),
            page,
          }
        : tab,
    );
    if (tabs.find((tab) => tab.id === tabId)?.active) {
      applyPageState(page);
    }
  }

  function updateEditorSource(source: string) {
    lastRaw = source;
    syncActiveTabPage();
    schedulePersistTabs();
  }

  async function openPage(
    nextUrl: string,
    pushHistory = true,
    opts?: { tabId?: string; skipCache?: boolean },
  ) {
    syncActiveTabPage();
    const tabId = opts?.tabId ?? tabs.find((tab) => tab.active)?.id;
    if (!tabId) {
      return;
    }

    const normalized = normalizeReticulumURL(nextUrl);
    if (!normalized) {
      return;
    }

    const existingTab = tabs.find((tab) => tab.id === tabId);
    const preserveEditorSource =
      normalized === "editor:" &&
      existingTab?.url === "editor:" &&
      (existingTab.page?.lastRaw?.trim() ?? "").length > 0;
    const savedEditorSource = preserveEditorSource ? (existingTab?.page?.lastRaw ?? "") : "";
    const preserveConfigSource =
      normalized === "config:" &&
      existingTab?.url === "config:" &&
      configText.trim().length > 0;
    const savedConfigSource = preserveConfigSource ? configText : "";

    const generation = (existingTab?.navGeneration ?? 0) + 1;
    const isActiveView = tabs.find((tab) => tab.active)?.id === tabId;

    tabs = tabs.map((tab) =>
      tab.id === tabId
        ? {
            ...tab,
            url: normalized,
            title: tabTitleFromURL(normalized, nodes),
            navGeneration: generation,
          }
        : tab,
    );

    if (isActiveView) {
      url = normalized;
      loading = true;
      error = "";
      errorKind = "";
      showSource = false;
      activePanel = "browser";
      if (normalized === "editor:") {
        contentType = "editor";
      }
      if (normalized === "config:") {
        contentType = "config";
      }
    }

    try {
      const skipCache = opts?.skipCache ?? false;
      let page: PageResponse;
      if (skipCache) {
        page = (
          pushHistory ? await NavigateFresh(normalized) : await OpenFreshURL(normalized)
        ) as PageResponse;
      } else {
        page = (
          pushHistory ? await Navigate(normalized) : await OpenURL(normalized)
        ) as PageResponse;
      }

      const current = tabs.find((tab) => tab.id === tabId);
      if (!current || current.navGeneration !== generation) {
        return;
      }

      let tabPage = pageFromResponse(page);
      if (preserveEditorSource && savedEditorSource) {
        tabPage = { ...tabPage, lastRaw: savedEditorSource };
      }
      if (normalized === "config:") {
        configText =
          preserveConfigSource && savedConfigSource ? savedConfigSource : (page.raw ?? "");
        configError = "";
        tabPage = { ...tabPage, lastRaw: configText };
      }

      applyPageToTab(tabId, tabPage, normalized);
      schedulePersistTabs();
    } catch (err) {
      const current = tabs.find((tab) => tab.id === tabId);
      if (!current || current.navGeneration !== generation) {
        return;
      }
      const failed = {
        ...emptyPage(),
        error: err instanceof Error ? err.message : String(err),
        errorKind: "internal",
      };
      applyPageToTab(tabId, failed, normalized);
      schedulePersistTabs();
    } finally {
      if (tabs.find((tab) => tab.id === tabId)?.active) {
        loading = false;
      }
      await refreshHistoryState();
      await refreshNetwork();
      if (pushHistory) {
        await loadHistory();
      }
    }
  }

  function browseURL(targetUrl: string) {
    if (openLinksInNewTab) {
      if (!canOpenTab(tabs.length)) {
        void openPage(targetUrl);
        return;
      }
      syncActiveTabPage();
      const normalized = normalizeReticulumURL(targetUrl);
      if (!normalized) {
        return;
      }
      const tab: Tab = {
        id: randomId(),
        title: tabTitleFromURL(normalized, nodes),
        url: normalized,
        active: true,
        page: emptyPage(),
        navGeneration: 0,
      };
      tabs = [...tabs.map((t) => ({ ...t, active: false })), tab];
      url = normalized;
      clearPageState();
      activePanel = "browser";
      void openPage(normalized, true, { tabId: tab.id });
      schedulePersistTabs();
      return;
    }
    void openPage(targetUrl);
  }

  async function refreshHistoryState() {
    const state = await HistoryState();
    canGoBack = state.canGoBack;
    canGoForward = state.canGoForward;
  }

  async function goBack() {
    const previous = await GoBack();
    if (previous) {
      await openPage(previous, false);
    } else {
      await refreshHistoryState();
    }
  }

  async function goForward() {
    const next = await GoForward();
    if (next) {
      await openPage(next, false);
    } else {
      await refreshHistoryState();
    }
  }

  async function loadNodes() {
    nodes = (await ListNodes()) as Node[];
    refreshTabTitles();
  }

  async function loadLogs() {
    logs = (await GetDevLogs()) as DevLogEntry[];
  }

  async function refreshNetwork() {
    network = (await GetNetworkLog()) as NetworkEntry[];
  }

  async function loadInterfaces() {
    interfaces = (await ListInterfaces()) as InterfaceRow[];
    configPath = await ConfigPath();
  }

  async function syncMobileChromeTheme(current = theme) {
    if (!mobileUI) {
      return;
    }
    await SyncMobileChrome(mobileChromeBg(current), mobileChromeUsesLightIcons(current));
  }

  async function loadConfigText() {
    configError = "";
    try {
      configText = await GetReticulumConfigText();
    } catch (err) {
      configError = err instanceof Error ? err.message : String(err);
    }
  }

  async function reloadConfigFromDisk() {
    configSaving = true;
    configError = "";
    try {
      configText = await ReloadReticulumConfig();
      await loadInterfaces();
      await loadCommunityInterfaces();
    } catch (err) {
      configError = err instanceof Error ? err.message : String(err);
    } finally {
      configSaving = false;
    }
  }

  async function loadPageCacheStats() {
    try {
      const stats = await GetPageCacheStats();
      pageCacheEntries = stats.entries ?? 0;
      pageCacheMax = stats.max ?? 128;
    } catch {
      pageCacheEntries = 0;
      pageCacheMax = 128;
    }
  }

  async function clearPageCache() {
    pageCacheClearing = true;
    try {
      await ClearPageCache();
      await loadPageCacheStats();
    } finally {
      pageCacheClearing = false;
    }
  }

  async function saveConfigText() {
    configSaving = true;
    configError = "";
    try {
      await SaveReticulumConfigText(configText);
      await loadInterfaces();
      await loadCommunityInterfaces();
    } catch (err) {
      configError = err instanceof Error ? err.message : String(err);
    } finally {
      configSaving = false;
    }
  }

  async function loadCommunityInterfaces() {
    communityLoading = true;
    communityError = "";
    try {
      communityItems = (await FetchCommunityInterfaces()) as CommunityInterface[];
    } catch (err) {
      communityError = err instanceof Error ? err.message : String(err);
    } finally {
      communityLoading = false;
    }
  }

  function toggleCommunitySelection(id: number) {
    if (communitySelected.has(id)) {
      communitySelected.delete(id);
    } else {
      communitySelected.add(id);
    }
  }

  async function importCommunitySelection() {
    const configs = communityItems
      .filter((item) => communitySelected.has(item.id))
      .map((item) => item.config);
    if (configs.length === 0) {
      return;
    }
    communityImporting = true;
    communityError = "";
    try {
      await ImportCommunityInterfaces(configs);
      communitySelected.clear();
      await loadInterfaces();
      await loadCommunityInterfaces();
    } catch (err) {
      communityError = err instanceof Error ? err.message : String(err);
    } finally {
      communityImporting = false;
    }
  }

  async function loadFavorites() {
    favorites = ((await GetFavorites()) ?? []) as string[];
  }

  async function loadHistory() {
    history = (await GetBrowsingHistory(50)) as HistoryEntry[];
  }

  async function loadTheme() {
    theme = (await GetTheme()) as ThemeSettings;
    applyTheme(theme);
    await syncMobileChromeTheme(theme);
  }

  async function loadKeybinds() {
    const saved = (await GetKeybinds()) as KeybindSettings;
    keybinds = mergeKeybinds(saved);
  }

  function setShowSource(value: boolean) {
    showSource = value;
    syncActiveTabPage();
    schedulePersistTabs();
  }

  async function refreshMicronWasmState(parserId = micronWasmParserId) {
    micronWasmAvailable = await isMicronWasmAvailable();
    micronWasmParserLabel = await resolveMicronWasmParserLabel(parserId);
    if (shouldPreloadMicronWasm(micronRenderer, micronWasmEnabled) && micronWasmAvailable) {
      micronWasmReady = await ensureMicronWasmReady(true, parserId);
    } else {
      micronWasmReady = false;
    }
  }

  async function loadBrowserPrefs() {
    const prefs = await GetBrowserPrefs();
    openLinksInNewTab = !!prefs.openLinksInNewTab;
    nativeTitlebar = !!prefs.nativeTitlebar;
    uiLanguage = prefs.uiLanguage ?? "";
    docsLanguage = prefs.docsLanguage ?? "";
    initUILocale(uiLanguage);
    micronWasmEnabled = prefs.micronWasmEnabled ?? true;
    micronWasmAvailable = await isMicronWasmAvailable();
    micronWasmParserId = prefs.micronWasmParserId || BUNDLED_MICRON_WASM_PARSER_ID;
    micronRenderer = normalizeMicronRendererPreference(prefs.micronRenderer);
    await refreshMicronWasmState(micronWasmParserId);
  }

  function currentBrowserPrefsPayload() {
    return {
      openLinksInNewTab,
      openLinksInNewWindow: false,
      nativeTitlebar,
      micronRenderer,
      micronWasmEnabled,
      micronWasmParserId,
      uiLanguage,
      docsLanguage,
    };
  }

  async function persistBrowserPrefs(patch: {
    openLinksInNewTab?: boolean;
    nativeTitlebar?: boolean;
    micronRenderer?: MicronRendererPreference;
    micronWasmEnabled?: boolean;
    micronWasmParserId?: string;
    uiLanguage?: string;
  }) {
    const prefs = await SetBrowserPrefs({
      ...currentBrowserPrefsPayload(),
      ...patch,
    });
    openLinksInNewTab = !!prefs.openLinksInNewTab;
    nativeTitlebar = !!prefs.nativeTitlebar;
    micronWasmEnabled = prefs.micronWasmEnabled ?? true;
    micronWasmAvailable = await isMicronWasmAvailable();
    micronWasmParserId = prefs.micronWasmParserId || BUNDLED_MICRON_WASM_PARSER_ID;
    micronRenderer = normalizeMicronRendererPreference(prefs.micronRenderer);
    await refreshMicronWasmState(micronWasmParserId);
    return prefs;
  }

  async function saveOpenLinksInNewTab(value: boolean) {
    await persistBrowserPrefs({ openLinksInNewTab: value });
  }

  async function saveUILanguage(value: string) {
    uiLanguage = value;
    setUILocale(value.trim() ? value : detectOSLocale());
    await persistBrowserPrefs({ uiLanguage: value });
  }

  async function saveNativeTitlebar(value: boolean) {
    try {
      const prefs = await SetNativeTitlebar(value);
      openLinksInNewTab = !!prefs.openLinksInNewTab;
      nativeTitlebar = !!prefs.nativeTitlebar;
      micronWasmEnabled = prefs.micronWasmEnabled ?? true;
      micronWasmParserId = prefs.micronWasmParserId || BUNDLED_MICRON_WASM_PARSER_ID;
      micronWasmAvailable = await isMicronWasmAvailable();
      micronRenderer = normalizeMicronRendererPreference(prefs.micronRenderer);
      await refreshMicronWasmState(micronWasmParserId);
    } catch {
      nativeTitlebar = value;
    }
  }

  async function saveMicronRenderer(value: MicronRendererPreference) {
    await persistBrowserPrefs({ micronRenderer: value });
  }

  async function saveMicronWasmEnabled(value: boolean) {
    await persistBrowserPrefs({ micronWasmEnabled: value });
  }

  async function saveMicronWasmParser(parserId: string) {
    micronWasmParserId = parserId;
    await persistBrowserPrefs({ micronWasmParserId: parserId });
  }

  function setMicronWasmReady(ready: boolean) {
    micronWasmReady = ready;
  }

  async function loadWindowChrome() {
    if (!desktopChrome) {
      return;
    }
    const chrome = (await GetWindowChrome()) as { nativeTitlebar?: boolean };
    nativeTitlebar = !!chrome.nativeTitlebar;
  }

  function requestIdentify() {
    if (!canIdentify || identifying) {
      return;
    }
    identifyConfirmOpen = true;
  }

  async function confirmIdentify() {
    identifyConfirmOpen = false;
    if (!canIdentify || identifying) {
      return;
    }
    identifying = true;
    try {
      await IdentifyToNode(url);
      await openPage(url, false, { skipCache: true });
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      errorKind = "internal";
    } finally {
      identifying = false;
    }
  }

  async function loadStoreHealth() {
    const health = (await GetStoreHealth()) as StoreHealth;
    storeHealth = {
      ok: !!health.ok,
      kind: health.kind,
      detail: health.detail,
      path: health.path ?? "",
    };
  }

  function requestResetDatabase() {
    resetDbConfirmOpen = true;
  }

  async function confirmResetDatabase() {
    resetDbConfirmOpen = false;
    try {
      await ResetDatabase();
      await loadStoreHealth();
      await loadFavorites();
      await loadHistory();
      tabs = [{ id: randomId(), title: t("tab.new"), url: "", active: true }];
      clearPageState();
      url = "";
    } catch (err) {
      storeHealth = {
        ok: false,
        kind: "database_corrupt",
        detail: err instanceof Error ? err.message : String(err),
        path: storeHealth.path,
      };
    }
  }

  function applyAsyncPageError(page: PageResponse) {
    const tabId = tabs.find((tab) => tab.active)?.id;
    if (!tabId || page.url !== url) {
      return;
    }
    const tabPage = pageFromResponse(page);
    applyPageToTab(tabId, tabPage, page.url || url);
    schedulePersistTabs();
  }

  async function resetDefaults() {
    if (!confirm(t("settings.resetConfirm"))) {
      return;
    }
    const reset = await ResetSettings();
    theme = reset.theme as ThemeSettings;
    applyTheme(theme);
    keybinds = mergeKeybinds(reset.keybinds);
    openLinksInNewTab = !!reset.browserPrefs.openLinksInNewTab;
    nativeTitlebar = !!reset.browserPrefs.nativeTitlebar;
    uiLanguage = reset.browserPrefs.uiLanguage ?? "";
    docsLanguage = reset.browserPrefs.docsLanguage ?? "";
    initUILocale(uiLanguage);
    micronWasmEnabled = reset.browserPrefs.micronWasmEnabled ?? true;
    micronWasmParserId = reset.browserPrefs.micronWasmParserId || BUNDLED_MICRON_WASM_PARSER_ID;
    micronWasmAvailable = await isMicronWasmAvailable();
    micronRenderer = normalizeMicronRendererPreference(reset.browserPrefs.micronRenderer);
    await refreshMicronWasmState(micronWasmParserId);
  }

  async function loadDownloadDir() {
    downloadDir = await GetDownloadDir();
  }

  async function loadDownloads() {
    downloads = ((await ListDownloads()) ?? []) as DownloadRow[];
  }

  async function saveDownloadDir(dir: string) {
    downloadDir = await SetDownloadDir(dir);
    await loadDownloads();
  }

  async function pickDownloadDir() {
    downloadDir = await PickDownloadDir();
    await loadDownloads();
  }

  async function downloadCurrentPage() {
    const payload = lastRaw || html;
    await downloadPageContent(url, contentType, payload);
    await loadDownloads();
  }

  function toggleDownloads() {
    downloadsOpen = !downloadsOpen;
    if (downloadsOpen) {
      void loadDownloads();
      if (mobileUI && activePanel !== "browser") {
        activePanel = "browser";
      }
    }
  }

  async function openDownload(path: string) {
    await OpenDownloadPath(path);
    downloadsOpen = false;
  }

  async function openDownloadFolder() {
    await ShowDownloadDir();
    downloadsOpen = false;
  }

  async function saveKeybinds(next: KeybindSettings) {
    keybinds = mergeKeybinds((await SetKeybinds(next)) as KeybindSettings);
  }

  function isEditableTarget(target: EventTarget | null): boolean {
    if (!(target instanceof HTMLElement)) {
      return false;
    }
    const tag = target.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") {
      return true;
    }
    return target.isContentEditable;
  }

  function runKeybindAction(action: KeybindAction) {
    switch (action) {
      case "focusUrl":
        document.querySelector<HTMLInputElement>(".url-input")?.focus();
        break;
      case "reload":
        if (url) {
          void openPage(url);
        }
        break;
      case "devtools":
        setPanel("devtools");
        break;
      case "findInPage":
        findOpen = true;
        activePanel = "browser";
        break;
      case "discovery":
        setPanel("discovery");
        break;
      case "settings":
        setPanel("settings");
        break;
      case "newTab":
        newTab();
        break;
      case "closeTab": {
        const tab = tabs.find((t) => t.active);
        if (tab && !tab.pinned) {
          closeTab(tab.id);
        }
        break;
      }
    }
  }

  function handleGlobalKeyDown(event: KeyboardEvent) {
    if (isKeybindRecording()) {
      return;
    }
    if (findOpen && event.key === "Escape") {
      return;
    }
    const pluginHit = matchPluginKeybind(event);
    if (pluginHit) {
      event.preventDefault();
      void dispatchPluginCommand(pluginHit.pluginId, pluginHit.commandId);
      return;
    }
    const editing = isEditableTarget(event.target);
    for (const [action, chord] of Object.entries(keybinds.bindings) as [KeybindAction, string][]) {
      if (!matchKeybind(event, chord)) {
        continue;
      }
      if (editing && action !== "findInPage" && action !== "devtools") {
        continue;
      }
      event.preventDefault();
      runKeybindAction(action);
      return;
    }
  }

  async function saveTheme(next: ThemeSettings) {
    theme = (await SetTheme(next)) as ThemeSettings;
    applyTheme(theme);
    await syncMobileChromeTheme(theme);
  }

  function toggleTheme() {
    const nextMode = theme.mode === "dark" ? "light" : "dark";
    void saveTheme({ ...theme, mode: nextMode });
  }

  onMount(() => {
    void loadTheme();
    void loadKeybinds();
    void loadBrowserPrefs();
    void loadWindowChrome();
    void loadDownloadDir();
    void loadDownloads();
    void ListSystemFonts().then((fonts) => {
      if (Array.isArray(fonts) && fonts.length > 0) {
        systemFonts = fonts as string[];
      }
    });
    void loadNodes().then(async () => {
      const saved = (await GetTabs()) as TabSnapshot[];
      restoreTabs(saved);
    });
    void loadLogs();
    void loadFavorites();
    void loadHistory();
    void loadInterfaces();
    void loadConfigText();
    void loadPageCacheStats();
    void loadCommunityInterfaces();
    void refreshNetwork();
    void loadStoreHealth();
    void bootPlugins();

    Events.On("plugin:loaded", () => {
      void reloadPlugins();
    });
    Events.On("plugin:unloaded", () => {
      void reloadPlugins();
    });
    Events.On("plugin:scheme", async (event) => {
      const data = event.data as { pluginId?: string; url?: string; handler?: string };
      if (!data?.pluginId || !data.url) {
        return;
      }
      const result = await handlePluginScheme(data.pluginId, data.handler ?? "", data.url);
      if (result?.html) {
        html = result.html;
        contentType = result.contentType || "html";
        error = "";
        url = data.url;
      }
    });

    const statusTimer = setInterval(() => {
      void loadNodes();
      void loadInterfaces();
    }, 5000);

    Events.On("rns:status", (event) => {
      const status = typeof event.data === "string" ? event.data : "";
      const wasOnline = meshOnline;
      meshOnline = status === "online";
      void loadInterfaces();
      if (wasOnline && !meshOnline && loading) {
        const lost = {
          ...emptyPage(),
          error: t("errors.connectionLostEvent"),
          errorKind: "connection_lost",
        };
        const tabId = tabs.find((tab) => tab.active)?.id;
        if (tabId) {
          applyPageToTab(tabId, lost, url);
          schedulePersistTabs();
        }
        loading = false;
      }
    });

    Events.On("page:error", (event) => {
      applyAsyncPageError(event.data as PageResponse);
      loading = false;
    });

    Events.On("store:health", (event) => {
      const health = event.data as StoreHealth;
      storeHealth = {
        ok: !!health.ok,
        kind: health.kind,
        detail: health.detail,
        path: health.path ?? "",
      };
    });

    Events.On("node:discovered", () => {
      void loadNodes();
    });

    Events.On("dev:log", (payload: { data: string }) => {
      try {
        const entry = JSON.parse(payload.data) as DevLogEntry;
        logs = [...logs.slice(-499), entry];
      } catch {
        logs = [...logs.slice(-499), { time: Date.now(), level: "info", message: payload.data }];
      }
    });

    Events.On("page:loaded", () => {
      void refreshHistoryState();
      void refreshNetwork();
      void loadHistory();
    });

    Events.On("network:entry", () => {
      void refreshNetwork();
    });

    Events.On("window:chrome", (event) => {
      const chrome = event.data as WindowChrome;
      nativeTitlebar = !!chrome.nativeTitlebar;
    });

    const onKeyDown = (event: KeyboardEvent) => {
      handleGlobalKeyDown(event);
    };
    window.addEventListener("keydown", onKeyDown);

    return () => {
      clearInterval(statusTimer);
      window.removeEventListener("keydown", onKeyDown);
      if (persistTimer) {
        clearTimeout(persistTimer);
      }
      void persistTabs();
    };
  });
</script>

<div class="app-shell">
  {#if pluginToast}
    <div class="plugin-toast" role="status">{pluginToast}</div>
  {/if}
  <TabBar
    {tabs}
    {nativeTitlebar}
    {mobileUI}
    {splitViewOpen}
    {splitTabId}
    onSelect={setActiveTab}
    onClose={closeTab}
    onNew={newTab}
    onReorder={reorderTabs}
    onReload={reloadTab}
    onDuplicate={duplicateTab}
    onFavorite={favoriteTab}
    onViewSource={viewSourceTab}
    onDownload={downloadTab}
    onSplit={splitTabView}
    onCloseSplit={closeSplitView}
    onCloseOthers={closeOtherTabs}
    onCloseRight={closeTabsToRight}
    onCloseAll={closeAllTabs}
    onTogglePin={togglePinTab}
  />

  <BrowserChrome
    bind:url
    {canGoBack}
    {canGoForward}
    {activePanel}
    pluginPanels={pluginContributions.panels}
    themeMode={theme.mode === "light" ? "light" : "dark"}
    {downloadsOpen}
    {downloads}
    {downloadDir}
    {canIdentify}
    {identifying}
    onNavigate={openPage}
    onBack={goBack}
    onForward={goForward}
    onReload={() => openPage(url)}
    onDownloadPage={downloadCurrentPage}
    onToggleDownloads={toggleDownloads}
    onCloseDownloads={() => (downloadsOpen = false)}
    onOpenDownload={openDownload}
    onOpenDownloadFolder={openDownloadFolder}
    onIdentify={requestIdentify}
    onPanel={setPanel}
    onToggleTheme={toggleTheme}
  />

  <main class="workspace" class:split={activePanel !== "browser"}>
    <section class="page-pane">
      {#snippet primaryPane()}
        {#if contentType === "editor"}
          {#if loading}
            <div class="editor-loading">{t("editor.loadingMicron")}</div>
          {:else}
            <MicronEditor
              source={lastRaw}
              currentURL={url}
              onSourceChange={updateEditorSource}
              onNavigate={openPage}
            />
          {/if}
        {:else if contentType === "config"}
          {#if loading}
            <div class="editor-loading">{t("editor.loadingConfig")}</div>
          {:else}
            <section class="config-page">
              <ReticulumConfigEditor
                bind:configText
                {configPath}
                saving={configSaving}
                error={configError}
                onChange={(text) => {
                  configText = text;
                }}
                onSave={() => void saveConfigText()}
                onReload={() => void reloadConfigFromDisk()}
              />
            </section>
          {/if}
        {:else}
          <ContentViewer
            {html}
            {contentType}
            {loading}
            {error}
            {errorKind}
            {pageFg}
            {pageBg}
            raw={lastRaw}
            {fromCache}
            {cachedAt}
            {showSource}
            currentURL={url}
            {findOpen}
            micronEngine={effectiveMicronEngine}
            onFindClose={() => (findOpen = false)}
            onNavigate={openPage}
            onRetry={() => openPage(url)}
            onReloadFresh={() => openPage(url, true, { skipCache: true })}
            onShowSourceChange={setShowSource}
          />
        {/if}
      {/snippet}

      {#if splitViewOpen}
        {#snippet secondaryPane()}
          {#if splitTab}
            {@const splitPage = splitTab.page ?? emptyPage()}
            <ContentViewer
              html={splitPage.html}
              contentType={splitPage.contentType}
              loading={false}
              error={splitPage.error}
              errorKind={splitPage.errorKind}
              pageFg={splitPage.pageFg}
              pageBg={splitPage.pageBg}
              raw={splitPage.lastRaw}
              fromCache={splitPage.fromCache}
              cachedAt={splitPage.cachedAt ?? 0}
              showSource={splitPage.showSource ?? false}
              currentURL={splitTab.url}
              findOpen={false}
              micronEngine={effectiveMicronEngine}
              onFindClose={() => {}}
              onNavigate={(target) => void openPage(target, true, { tabId: splitTab.id })}
              onRetry={() => void openPage(splitTab.url, false, { tabId: splitTab.id })}
              onReloadFresh={() =>
                void openPage(splitTab.url, true, { tabId: splitTab.id, skipCache: true })}
              onShowSourceChange={(value) => setSplitTabShowSource(splitTab.id, value)}
            />
          {:else}
            <SplitTabPicker
              {tabs}
              {activeTabId}
              onSelect={selectSplitTab}
              onClose={closeSplitView}
            />
          {/if}
        {/snippet}
        <SplitPane
          ratio={splitRatio}
          onRatioChange={(value) => (splitRatio = value)}
          primary={primaryPane}
          secondary={secondaryPane}
        />
      {:else}
        {@render primaryPane()}
      {/if}
    </section>

    {#if activePanel === "discovery"}
      <aside class="side-pane">
        <DiscoveryPanel
          {nodes}
          {favorites}
          onOpen={browseURL}
          onFavorite={async (favUrl) => {
            favorites = (await AddFavorite(favUrl)) as string[];
          }}
        />
      </aside>
    {:else if activePanel === "history"}
      <aside class="side-pane">
        <HistoryPanel {history} onOpen={browseURL} />
      </aside>
    {:else if activePanel === "devtools"}
      <aside class="side-pane">
        <DevToolsPanel
          {logs}
          {network}
          raw={lastRaw}
          {logLevel}
          {contentType}
          {durationMs}
          {hops}
          {fromCache}
          {cachedAt}
          {micronRendererBadge}
          pluginTabs={pluginContributions.devtools}
          onClear={() => {
            void ClearDevLogs();
            logs = [];
          }}
          onExport={async () => {
            const json = await ExportDevLogs();
            const blob = new Blob([json], { type: "application/json" });
            const href = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = href;
            a.download = exportFilename("devlogs");
            a.click();
            URL.revokeObjectURL(href);
          }}
          onLogLevel={(level) => {
            void SetLogLevel(level).then((v) => {
              logLevel = v;
            });
          }}
        />
      </aside>
    {:else if activePanel === "settings"}
      <aside class="side-pane">
        <SettingsPanel
          bind:theme
          {systemFonts}
          {keybinds}
          {interfaces}
          {configPath}
          {pluginsDir}
          bind:downloadDir
          {uiLanguage}
          onChangeUILanguage={saveUILanguage}
          {openLinksInNewTab}
          {nativeTitlebar}
          {micronRenderer}
          {micronWasmEnabled}
          {micronWasmParserId}
          {desktopChrome}
          {mobileUI}
          bind:configText
          {configSaving}
          {configError}
          {communityItems}
          {communityLoading}
          {communityImporting}
          {communityError}
          bind:communityFilter
          {communitySelected}
          onChange={saveTheme}
          onChangeKeybinds={saveKeybinds}
          onChangeDownloadDir={saveDownloadDir}
          onPickDownloadDir={pickDownloadDir}
          onChangeOpenLinksInNewTab={saveOpenLinksInNewTab}
          onChangeNativeTitlebar={saveNativeTitlebar}
          onChangeMicronRenderer={saveMicronRenderer}
          onChangeMicronWasmEnabled={saveMicronWasmEnabled}
          onChangeMicronWasmParser={saveMicronWasmParser}
          onMicronWasmReadyChange={setMicronWasmReady}
          onResetDefaults={resetDefaults}
          onToggleInterface={async (name, enabled) => {
            await SetInterfaceEnabled(name, enabled);
            await loadInterfaces();
          }}
          onExportTheme={async () => {
            const json = await ExportTheme();
            const blob = new Blob([json], { type: "application/json" });
            const href = URL.createObjectURL(blob);
            const a = document.createElement("a");
            a.href = href;
            a.download = exportFilename("theme");
            a.click();
            URL.revokeObjectURL(href);
          }}
          onImportTheme={async (json) => {
            theme = (await ImportTheme(json)) as ThemeSettings;
            applyTheme(theme);
            await syncMobileChromeTheme(theme);
          }}
          onConfigChange={(text) => {
            configText = text;
          }}
          onConfigSave={() => void saveConfigText()}
          onConfigReload={() => void reloadConfigFromDisk()}
          onClearPageCache={() => void clearPageCache()}
          {pageCacheEntries}
          {pageCacheMax}
          {pageCacheClearing}
          onCommunityRefresh={() => void loadCommunityInterfaces()}
          onCommunityFilter={(value) => {
            communityFilter = value;
          }}
          onCommunityToggle={toggleCommunitySelection}
          onCommunityImport={() => void importCommunitySelection()}
          onPluginsChanged={() => void reloadPlugins()}
        />
      </aside>
    {:else if activePluginPanel}
      <aside class="side-pane">
        <PluginPanelHost
          pluginId={activePluginPanel.pluginId}
          panelId={activePluginPanel.id}
          title={activePluginPanel.title}
          entry={activePluginPanel.entry}
          getCurrentURL={() => url}
          navigate={(next) => void browseURL(next)}
          showToast={showPluginToast}
        />
      </aside>
    {/if}
  </main>

  <MobileNav
    {activePanel}
    pluginPanels={pluginContributions.panels}
    {downloadsOpen}
    onPanel={setPanel}
    onToggleDownloads={toggleDownloads}
  />

  {#if mobileUI}
    <div class="mobile-downloads">
      <DownloadsMenu
        open={downloadsOpen}
        {downloads}
        {downloadDir}
        onDownloadPage={downloadCurrentPage}
        onOpenFile={openDownload}
        onOpenFolder={openDownloadFolder}
        onClose={() => (downloadsOpen = false)}
      />
    </div>
  {/if}

  <ConfirmDialog
    open={identifyConfirmOpen}
    title={t("dialog.identifyTitle")}
    message={t("dialog.identifyMessage")}
    confirmLabel={t("common.identify")}
    onConfirm={confirmIdentify}
    onCancel={() => (identifyConfirmOpen = false)}
  />

  <ConfirmDialog
    open={resetDbConfirmOpen}
    title={t("dialog.resetDbTitle")}
    message={t("dialog.resetDbMessage")}
    confirmLabel={t("dialog.resetDbConfirm")}
    onConfirm={confirmResetDatabase}
    onCancel={() => (resetDbConfirmOpen = false)}
  />

  {#if storeErrorVisible}
    <AppStoreError
      kind={storeHealth.kind ?? "database_corrupt"}
      detail={storeHealth.detail ?? ""}
      path={storeHealth.path}
      onResetDatabase={requestResetDatabase}
      onRetry={() => void loadStoreHealth()}
    />
  {/if}
</div>

<style>
  .app-shell {
    height: 100vh;
    display: grid;
    grid-template-rows: auto auto 1fr auto;
    background: var(--ren-surface-bg);
  }

  .workspace {
    min-height: 0;
    display: grid;
    grid-template-columns: 1fr;
  }

  .workspace.split {
    grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.8fr);
  }

  .page-pane,
  .side-pane {
    min-height: 0;
    min-width: 0;
    border-top: 1px solid var(--ren-border);
  }

  .side-pane {
    border-left: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
  }

  @media (max-width: 900px) {
    .workspace.split {
      grid-template-columns: 1fr;
      grid-template-rows: 1fr auto;
    }

    .side-pane {
      max-height: 45vh;
      border-left: none;
    }
  }

  .editor-loading {
    height: 100%;
    display: grid;
    place-items: center;
    color: var(--ren-muted);
  }

  .config-page {
    height: 100%;
    overflow: auto;
    padding: 1rem;
    background: var(--ren-content-bg);
  }

  .mobile-downloads :global(.menu) {
    position: fixed;
    left: 0.75rem;
    right: 0.75rem;
    bottom: calc(3.6rem + env(safe-area-inset-bottom));
    max-height: min(55vh, 28rem);
    z-index: 40;
  }

  .mobile-downloads :global(.backdrop) {
    z-index: 35;
  }

  .plugin-toast {
    position: fixed;
    left: 50%;
    bottom: 1.25rem;
    transform: translateX(-50%);
    z-index: 60;
    padding: 0.55rem 0.9rem;
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    border: 1px solid var(--ren-border);
    box-shadow: var(--ren-shadow);
    font-size: 0.85rem;
  }
</style>
