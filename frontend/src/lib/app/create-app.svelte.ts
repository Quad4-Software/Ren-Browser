// SPDX-License-Identifier: MIT
import { SvelteSet } from "svelte/reactivity";
import { Events, System } from "@wailsio/runtime";
import {
  AddFavorite,
  ApplySuggestedCommunityInterfaces,
  CancelDownload,
  CompleteInitialSetup,
  ClearDevLogs,
  ClearBrowsingHistory,
  ClearPageCache,
  ConfigPath,
  DismissDownload,
  ExportDevLogs,
  ExportTheme,
  FetchCommunityInterfaces,
  GetBrowserPrefs,
  GetDownloadDir,
  GetDevLogs,
  GetFavorites,
  GetBrowsingHistory,
  GetInitialSetupState,
  GetKeybinds,
  GetNetworkLog,
  GetStoreHealth,
  GetTabs,
  GetTheme,
  GetWindowChrome,
  GetReticulumConfigText,
  GetPageCacheStats,
  GetRuntimeConfig,
  GetSandboxStatus,
  GetStatus,
  GoBack,
  GoForward,
  HistoryState,
  IdentifyToNode,
  ImportCommunityInterfaces,
  ImportTheme,
  ListActiveDownloads,
  ListDownloads,
  ClearDownloadHistory,
  ListInterfaces,
  ListNodes,
  ListSystemFonts,
  Navigate,
  NavigateFresh,
  OpenDownloadPath,
  OpenURL,
  OpenFreshURL,
  PickDownloadDir,
  PrepareForWake,
  PreviewSuggestedCommunityInterfaces,
  RetryDownload,
  ResetDatabase,
  ResetSettings,
  ResetBrowser,
  RestartReticulum,
  ReloadReticulumConfig,
  RunSelfCheck,
  SaveTabs,
  SaveReticulumConfigText,
  SetBrowserPrefs,
  SetDownloadDir,
  SetEnableTransport,
  SetInterfaceEnabled,
  SetKeybinds,
  SetLogLevel,
  SetNativeTitlebar,
  SetShareInstance,
  SetTheme,
  ShowConfigDir,
  ShowDownloadDir,
  Shutdown,
  SyncMobileChrome,
  TakePendingDeepLink,
} from "../../../bindings/renbrowser/internal/app/browserservice.js";
import type { WindowChrome } from "../../../bindings/renbrowser/internal/app/models.js";
import type { DownloadRow } from "$lib/components/DownloadsMenu.svelte";
import { withProgress, type ActiveDownloadRow } from "$lib/browser/download-progress";
import { exportFilename } from "$lib/brand";
import { isStoreBlockingKind } from "$lib/browser/errors";
import { isCompactViewport } from "$lib/browser/viewport";
import { resolveMobileUI } from "$lib/browser/mobile-layout";
import {
  defaultTheme,
  applyTheme,
  mobileChromeBg,
  mobileChromeUsesLightIcons,
  type ThemeSettings,
} from "$lib/theme/tokens";
import {
  screenshotThemeFromQuery,
  screenshotLayoutFromQuery,
  screenshotSceneFromQuery,
  markScreenshotReady,
  type ScreenshotScene,
} from "$lib/theme/screenshot";
import type {
  CommunityFetchResult,
  SelfCheckResult,
} from "../../../bindings/renbrowser/internal/app/models.js";
import type { CommunityInterface } from "../../../bindings/renbrowser/internal/rns/models.js";
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
import { getContributions as fetchContributions, listPlugins } from "$lib/plugins/api.js";
import { PluginsDir } from "../../../bindings/renbrowser/internal/app/pluginhost.js";
import {
  activateAllPlugins,
  deactivateAllPlugins,
  handlePluginScheme,
} from "$lib/plugins/lifecycle.js";
import { dispatchPluginCommand, matchPluginKeybind } from "$lib/plugins/command-dispatch.js";
import type { ActivePanel, ContributionsSnapshot } from "$lib/plugins/api-types.js";
import { formatBindingError } from "$lib/browser/binding-errors.js";
import {
  downloadPageContent,
  downloadFailureMessage,
  canceledDownloadToast,
  isDownloadCanceledError,
  pageDownloadName,
  type DownloadResult,
} from "$lib/browser/download";
import { blockExternalLinkPointerEvent } from "$lib/browser/navigation-guard";
import { unwrapDeepLink } from "$lib/browser/deeplink";
import {
  canOpenTab,
  MAX_TABS,
  normalizeReticulumURL,
  nodeHomeURL,
  isNodeHomePage,
  orderTabsPinnedFirst,
  pinTabInList,
  reorderTabsInList,
  tabTitleFromURL,
  unpinTabInList,
  type Tab,
  type TabPage,
} from "$lib/browser/url";
import {
  documentURL,
  canonicalDocumentURL,
  isDocumentURL,
  isReadableMeshFileURL,
} from "$lib/documents/types";
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
import type {
  DevLogEntry,
  HistoryEntry,
  InitialSetupStep,
  InterfaceRow,
  NetworkEntry,
  Node,
  PageResponse,
  ReticulumStatusRow,
  SandboxStatus,
  StoreHealth,
  TabSnapshot,
} from "./types";
import { emptyPage, pageFromResponse } from "./page-state";

export type AppController = ReturnType<typeof createApp>;

export function createApp() {
  let activePanel = $state<ActivePanel>("browser");
  let pluginContributions = $state<ContributionsSnapshot>({
    panels: [],
    commands: [],
    devtools: [],
    urlSchemes: [],
  });
  let pluginToast = $state("");
  let pluginToastIsError = $state(false);
  let pluginsDir = $state("");
  let pluginGrantedById = $state<Record<string, string[]>>({});
  let url = $state("");
  let loading = $state(false);
  // Large page bodies are reassigned, never deep-mutated; raw avoids proxy cost.
  let html = $state.raw("");
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
  let reticulumStatus = $state<ReticulumStatusRow>({
    enableTransport: false,
    shareInstance: false,
    connectedToSharedInstance: false,
    sharedInstanceMode: "disabled",
    transportActive: false,
  });
  let configPath = $state("");
  let logLevel = $state(3);
  let systemFonts = $state<string[]>(["system-ui", "sans-serif", "monospace"]);
  let theme = $state<ThemeSettings>(defaultTheme());
  let keybinds = $state<KeybindSettings>(defaultKeybinds());
  let downloadDir = $state("");
  let downloads = $state<DownloadRow[]>([]);
  let activeDownloads = $state<ActiveDownloadRow[]>([]);
  const retryingDownloadIds = new SvelteSet<string>();
  let clearingDownloadHistory = $state(false);
  const activeDownloadViews = $derived(withProgress(activeDownloads));
  let downloadsOpen = $state(false);
  let findOpen = $state(false);
  let pageHighlight = $state("");
  let canGoBack = $state(false);
  let canGoForward = $state(false);
  let lastRaw = $state.raw("");
  let binaryB64 = $state.raw("");
  let pagePath = $state("");
  let fromCache = $state(false);
  let cachedAt = $state(0);
  let showSource = $state(false);
  let openLinksInNewTab = $state(true);
  let nativeTitlebar = $state(false);
  let uiLanguage = $state("");
  let docsLanguage = $state("");
  let initialSetupComplete = $state(false);
  let micronRenderer = $state<MicronRendererPreference>("auto");
  let micronWasmEnabled = $state(true);
  let micronWasmReady = $state(false);
  let micronWasmAvailable = $state(false);
  let micronWasmParserId = $state(BUNDLED_MICRON_WASM_PARSER_ID);
  let micronWasmParserLabel = $state("");
  let identifying = $state(false);
  let identifyConfirmOpen = $state(false);
  let resetDbConfirmOpen = $state(false);
  let resetBrowserConfirmOpen = $state(false);
  let restartReticulumConfirmOpen = $state(false);
  let transportMobileConfirmOpen = $state(false);
  let closeAllConfirmOpen = $state(false);
  let shutdownConfirmOpen = $state(false);
  let clearHistoryConfirmOpen = $state(false);
  let publicMode = $state(false);
  let serverMode = $state(false);
  let storeHealth = $state<StoreHealth>({ ok: true, path: "" });
  let sandboxStatus = $state<SandboxStatus>({
    type: "none",
    enabled: false,
    requested: false,
    supported: false,
    auto: false,
    disabledByEnv: false,
    inFlatpak: false,
    inAppImage: false,
    inContainer: false,
    webkitSandbox: "unavailable",
    onAndroid: false,
  });
  let selfTestResult = $state<SelfCheckResult | null>(null);
  let selfTestRunning = $state(false);
  let meshOnline = $state(true);
  let splitViewOpen = $state(false);
  let splitTabId = $state<string | null>(null);
  let splitRatio = $state(52);
  const desktopChrome = $derived(System.IsDesktop() && !serverMode);
  const layoutOverride = screenshotLayoutFromQuery();
  let compactViewport = $state(typeof window !== "undefined" ? isCompactViewport() : false);
  const mobileUI = $derived(
    resolveMobileUI({
      layoutOverride,
      isMobilePlatform: System.IsMobile(),
      compactViewport,
    }),
  );

  let configText = $state("");
  let configSaving = $state(false);
  let configError = $state("");
  let pageCacheEntries = $state(0);
  let pageCacheMax = $state(128);
  let pageCacheClearing = $state(false);
  let pageCacheEnabled = $state(true);
  let communityItems = $state<CommunityInterface[]>([]);
  let communityLoading = $state(false);
  let communityImporting = $state(false);
  let communityError = $state("");
  let communityFromBundle = $state(false);
  let communityFilter = $state("");
  const communitySelected = new SvelteSet<number>();
  let initialSetupOpen = $state(false);
  let initialSetupStep = $state<InitialSetupStep>("welcome");
  let suggestedItems = $state<CommunityInterface[]>([]);
  let suggestedLoading = $state(false);
  let initialSetupBusy = $state(false);
  let initialSetupError = $state("");
  let discoverySlowMode = $state(false);
  let mobileDevTools = $state(false);
  let tabHoverPreviews = $state(true);
  let micronPreserveLayout = $state(false);
  let mobileTabsOpen = $state(false);
  let settingsSectionsCollapsed = $state<Record<string, boolean>>({});

  const DISCOVERY_POLL_MS = 5000;
  const DISCOVERY_POLL_SLOW_MS = 15000;
  const DISCOVERY_EVENT_DEBOUNCE_MS = 5000;
  const DISCOVERY_EVENT_DEBOUNCE_FAST_MS = 400;
  let statusTimer: ReturnType<typeof setInterval> | undefined;
  let nodeDiscoverTimer: ReturnType<typeof setTimeout> | undefined;
  let appForeground = $state(true);

  let tabs = $state<Tab[]>([{ id: randomId(), title: "", url: "", active: true }]);

  const effectiveMicronEngine = $derived(
    resolveEffectiveMicronEngine(micronRenderer, {
      wasmEnabled: micronWasmEnabled,
      wasmAvailable: micronWasmAvailable,
      wasmReady: micronWasmReady,
      hasServerHtml: html.trim().length > 0,
      rawBytes: lastRaw.length,
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
  const atTabLimit = $derived(tabs.length >= MAX_TABS);
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
    if (panel === "devtools" && !mobileDevTools) {
      return;
    }
    const next = activePanel === panel ? "browser" : panel;
    activePanel = next;
    mobileTabsOpen = false;
    if (next === "settings") {
      void loadPageCacheStats();
    }
  }

  function showPluginToast(message: string, opts: { isError?: boolean } = {}) {
    pluginToast = message;
    pluginToastIsError = !!opts.isError;
    if (opts.isError) {
      console.error("[RenBrowser]", message);
    }
    const duration = opts.isError ? 8000 : 2500;
    setTimeout(() => {
      if (pluginToast === message) {
        pluginToast = "";
      }
    }, duration);
  }

  function pluginHostOpts(pluginId?: string) {
    const granted = pluginId ? (pluginGrantedById[pluginId] ?? []) : [];
    return {
      getCurrentURL: () => url,
      navigate: (next: string) => void browseURL(next),
      showToast: showPluginToast,
      getActivePage: () => ({
        url,
        path: pagePath,
        contentType,
        html,
        raw: lastRaw,
      }),
      updateActivePage: (
        patch: Partial<{
          url: string;
          path: string;
          contentType: string;
          html: string;
          raw: string;
        }>,
      ) => {
        const next = { ...currentPageState() };
        if (patch.html !== undefined) {
          next.html = patch.html;
        }
        if (patch.contentType !== undefined) {
          next.contentType = patch.contentType;
        }
        if (patch.raw !== undefined) {
          next.lastRaw = patch.raw;
        }
        if (patch.path !== undefined) {
          next.path = patch.path;
        }
        applyPageState(next);
        syncActiveTabPage();
        schedulePersistTabs();
      },
      networkFetch: granted.includes("network.fetch"),
      wasmBackend: true,
      onPluginError: (pluginId: string, message: string) => {
        showPluginToast(
          `Extension ${pluginId} failed: ${formatBindingError(message, "Extension failed")}`,
          { isError: true },
        );
      },
    };
  }

  async function bootPlugins() {
    pluginsDir = (await PluginsDir()) ?? "";
    const rows = await listPlugins();
    pluginGrantedById = Object.fromEntries(
      (rows ?? []).map((row) => [row.id, row.grantedPermissions ?? []]),
    );
    const snapshot = await fetchContributions();
    setContributions(snapshot);
    pluginContributions = getContributionsSnapshot();
    await activateAllPlugins(pluginHostOpts());
  }

  async function reloadPlugins() {
    await deactivateAllPlugins();
    await bootPlugins();
  }

  function currentPageState(): TabPage {
    return {
      html,
      contentType,
      error,
      errorKind,
      durationMs,
      lastRaw,
      binaryB64,
      path: pagePath,
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
    binaryB64 = page.binaryB64 ?? "";
    pagePath = page.path ?? "";
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
    if (selected?.url && isDocumentURL(selected.url) && !selected.page?.binaryB64?.trim()) {
      loading = true;
      void openPage(selected.url, false, { tabId: id });
      schedulePersistTabs();
      return;
    }
    if (selected?.page) {
      applyPageState(selected.page);
    } else {
      clearPageState();
    }
    loading = !!selected?.loading;
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
        {
          ...tabs[0],
          id: tabs[0].id,
          title: t("tab.new"),
          url: "",
          active: true,
          page: emptyPage(),
        },
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
    tabs = [{ id: randomId(), title: t("tab.new"), url: "", active: true, page: emptyPage() }];
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

  function requestCloseAllTabs() {
    if (tabs.length <= 1) {
      return;
    }
    closeAllConfirmOpen = true;
  }

  function confirmCloseAllTabs() {
    closeAllConfirmOpen = false;
    closeAllTabs();
    if (mobileTabsOpen && tabs.length <= 1) {
      mobileTabsOpen = false;
    }
  }

  async function loadRuntimeConfig() {
    const config = await GetRuntimeConfig();
    publicMode = !!config.publicMode;
    serverMode = !!config.serverMode;
  }

  function requestClearHistory() {
    if (!history.length) {
      return;
    }
    clearHistoryConfirmOpen = true;
  }

  async function confirmClearHistory() {
    clearHistoryConfirmOpen = false;
    await ClearBrowsingHistory();
    await loadHistory();
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
            loading: false,
          }
        : tab,
    );
    const selected = tabs.find((tab) => tab.id === tabId);
    if (selected?.active) {
      applyPageState(page);
      loading = false;
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
    opts?: { tabId?: string; skipCache?: boolean; highlight?: string },
  ) {
    if (opts && "highlight" in opts) {
      pageHighlight = opts.highlight?.trim() ?? "";
    } else {
      pageHighlight = "";
    }

    syncActiveTabPage();
    const tabId = opts?.tabId ?? tabs.find((tab) => tab.active)?.id;
    if (!tabId) {
      return;
    }

    const normalized = normalizeReticulumURL(nextUrl);
    if (!normalized) {
      return;
    }
    const pageUrl =
      isDocumentURL(normalized) && downloadDir.trim()
        ? canonicalDocumentURL(normalized, downloadDir)
        : normalized;

    if (!isDocumentURL(pageUrl) && isReadableMeshFileURL(normalized)) {
      try {
        handleDownloadResult({ ok: true, pending: true, message: t("downloads.downloading") });
        await downloadPageContent(normalized, "file", "");
        handleDownloadResult({ ok: true, message: t("downloads.fileSaved") });
        await loadDownloads();
        await loadActiveDownloads();
      } catch (err) {
        console.error("[App] mesh document download failed", normalized, err);
        if (isDownloadCanceledError(err)) {
          handleDownloadResult({
            ok: false,
            message: "",
            canceled: true,
            name: pageDownloadName(normalized, "file"),
          });
        } else {
          handleDownloadResult({
            ok: false,
            message: downloadFailureMessage(err, t("downloads.downloadFailed")),
          });
        }
      }
      return;
    }

    const existingTab = tabs.find((tab) => tab.id === tabId);
    const preserveEditorSource =
      normalized === "editor:" &&
      existingTab?.url === "editor:" &&
      (existingTab.page?.lastRaw?.trim() ?? "").length > 0;
    const savedEditorSource = preserveEditorSource ? (existingTab?.page?.lastRaw ?? "") : "";
    const preserveConfigSource =
      normalized === "config:" && existingTab?.url === "config:" && configText.trim().length > 0;
    const savedConfigSource = preserveConfigSource ? configText : "";

    const generation = (existingTab?.navGeneration ?? 0) + 1;
    const isActiveView = tabs.find((tab) => tab.active)?.id === tabId;

    tabs = tabs.map((tab) =>
      tab.id === tabId
        ? {
            ...tab,
            url: pageUrl,
            title: tabTitleFromURL(pageUrl, nodes),
            navGeneration: generation,
            loading: true,
          }
        : tab,
    );

    if (isActiveView) {
      url = pageUrl;
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
      if (normalized === "settings:") {
        contentType = "settings";
      }
    }

    try {
      const skipCache = opts?.skipCache ?? false;
      let page: PageResponse;
      if (skipCache) {
        page = (
          pushHistory ? await NavigateFresh(pageUrl) : await OpenFreshURL(pageUrl)
        ) as PageResponse;
      } else {
        page = (pushHistory ? await Navigate(pageUrl) : await OpenURL(pageUrl)) as PageResponse;
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

      applyPageToTab(tabId, tabPage, page.url?.trim() || pageUrl);
      schedulePersistTabs();
    } catch (err) {
      const current = tabs.find((tab) => tab.id === tabId);
      if (!current || current.navGeneration !== generation) {
        return;
      }
      const failed = {
        ...emptyPage(),
        error: formatBindingError(err, "Request failed"),
        errorKind: "internal",
      };
      applyPageToTab(tabId, failed, normalized);
      schedulePersistTabs();
    } finally {
      tabs = tabs.map((tab) => (tab.id === tabId ? { ...tab, loading: false } : tab));
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

  function browseURL(targetUrl: string, highlight?: string) {
    const highlightValue = highlight === undefined ? "" : highlight.trim();
    if (openLinksInNewTab) {
      if (!canOpenTab(tabs.length)) {
        void openPage(targetUrl, true, { highlight: highlightValue });
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
      void openPage(normalized, true, { tabId: tab.id, highlight: highlightValue });
      schedulePersistTabs();
      return;
    }
    void openPage(targetUrl, true, { highlight: highlightValue });
  }

  function clearPageHighlight() {
    pageHighlight = "";
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

  function pauseBackgroundPolling() {
    if (statusTimer !== undefined) {
      clearInterval(statusTimer);
      statusTimer = undefined;
    }
    if (nodeDiscoverTimer !== undefined) {
      clearTimeout(nodeDiscoverTimer);
      nodeDiscoverTimer = undefined;
    }
  }

  async function resumeForegroundSync() {
    appForeground = true;
    // Drop idle links / soft-stale paths before any reload so post-suspend
    // fetches rediscover instead of hanging on zombie StatusActive caches.
    try {
      await PrepareForWake();
    } catch {
      // Bindings may be unavailable in tests or early boot.
    }
    await Promise.all([loadNodes(), loadInterfaces(), refreshHistoryState()]);
    void refreshNetwork();
    void loadStoreHealth();
    void loadActiveDownloads();

    const active = tabs.find((tab) => tab.active);
    const stuck = tabs.filter((tab) => tab.loading);
    if (stuck.length > 0) {
      tabs = tabs.map((tab) => (tab.loading ? { ...tab, loading: false } : tab));
      if (active?.url && stuck.some((tab) => tab.id === active.id)) {
        void openPage(active.url, false);
      }
    }

    restartStatusTimer();
  }

  function handleAppBackground() {
    appForeground = false;
    pauseBackgroundPolling();
  }

  function discoveryPollInterval(): number {
    return discoverySlowMode ? DISCOVERY_POLL_SLOW_MS : DISCOVERY_POLL_MS;
  }

  function restartStatusTimer() {
    if (statusTimer !== undefined) {
      clearInterval(statusTimer);
    }
    statusTimer = setInterval(() => {
      void loadNodes();
      void loadInterfaces();
    }, discoveryPollInterval());
  }

  function scheduleLoadNodesFromEvent() {
    if (!appForeground) {
      return;
    }
    if (nodeDiscoverTimer !== undefined) {
      return;
    }
    const delay = discoverySlowMode
      ? DISCOVERY_EVENT_DEBOUNCE_MS
      : DISCOVERY_EVENT_DEBOUNCE_FAST_MS;
    nodeDiscoverTimer = setTimeout(() => {
      nodeDiscoverTimer = undefined;
      void loadNodes();
    }, delay);
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
    await loadReticulumStatus();
  }

  async function loadReticulumStatus() {
    try {
      const status = await GetStatus();
      reticulumStatus = {
        enableTransport: Boolean(status?.enableTransport),
        shareInstance: Boolean(status?.shareInstance),
        connectedToSharedInstance: Boolean(status?.connectedToSharedInstance),
        sharedInstanceMode: status?.sharedInstanceMode || "disabled",
        transportActive: Boolean(status?.transportActive),
      };
    } catch {
      // Keep last known status if the binding is unavailable.
    }
  }

  async function syncMobileChromeTheme(current = theme) {
    if (!System.IsMobile()) {
      return;
    }
    await SyncMobileChrome(mobileChromeBg(current), mobileChromeUsesLightIcons(current));
  }

  $effect(() => {
    if (!mobileUI) {
      mobileTabsOpen = false;
      if (splitViewOpen) {
        closeSplitView();
      }
    }
  });

  async function loadConfigText() {
    configError = "";
    try {
      configText = await GetReticulumConfigText();
    } catch (err) {
      configError = formatBindingError(err, "Request failed");
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
      configError = formatBindingError(err, "Request failed");
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
      configError = formatBindingError(err, "Request failed");
    } finally {
      configSaving = false;
    }
  }

  async function loadCommunityInterfaces() {
    communityLoading = true;
    communityError = "";
    try {
      const result = (await FetchCommunityInterfaces()) as CommunityFetchResult;
      communityItems = Array.isArray(result?.items) ? result.items : [];
      communityFromBundle = !!result?.fromBundle;
    } catch (err) {
      communityFromBundle = false;
      communityError = formatBindingError(err, "Request failed");
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
      await loadConfigText();
      await loadCommunityInterfaces();
    } catch (err) {
      communityError = formatBindingError(err, "Request failed");
      throw err;
    } finally {
      communityImporting = false;
    }
  }

  async function checkInitialSetup() {
    if (publicMode) {
      return;
    }
    try {
      const state = await GetInitialSetupState();
      if (state?.needed) {
        initialSetupOpen = true;
        initialSetupStep = "welcome";
        void loadSuggestedPreview();
      }
    } catch {
      // Do not block the shell if setup state cannot be read.
    }
  }

  async function loadSuggestedPreview() {
    suggestedLoading = true;
    initialSetupError = "";
    try {
      const items = await PreviewSuggestedCommunityInterfaces();
      suggestedItems = Array.isArray(items) ? items : [];
    } catch (err) {
      suggestedItems = [];
      initialSetupError = formatBindingError(err, "Request failed");
    } finally {
      suggestedLoading = false;
    }
  }

  function setInitialSetupStep(step: InitialSetupStep) {
    initialSetupStep = step;
    initialSetupError = "";
    if (step === "pick" && communityItems.length === 0) {
      void loadCommunityInterfaces();
    }
    if (step === "config" && !configText.trim()) {
      void loadConfigText();
    }
  }

  async function completeInitialSetup() {
    await CompleteInitialSetup();
    initialSetupComplete = true;
    initialSetupOpen = false;
    initialSetupStep = "welcome";
    initialSetupError = "";
    suggestedItems = [];
    communitySelected.clear();
  }

  async function applySuggestedSetup() {
    initialSetupBusy = true;
    initialSetupError = "";
    try {
      await ApplySuggestedCommunityInterfaces();
      await loadConfigText();
      await loadInterfaces();
      await loadCommunityInterfaces();
      await completeInitialSetup();
    } catch (err) {
      initialSetupError = formatBindingError(err, "Request failed");
    } finally {
      initialSetupBusy = false;
    }
  }

  async function importInitialSetupSelection() {
    if (communitySelected.size === 0) {
      return;
    }
    initialSetupBusy = true;
    initialSetupError = "";
    try {
      await importCommunitySelection();
      await completeInitialSetup();
    } catch (err) {
      initialSetupError = formatBindingError(err, "Request failed");
    } finally {
      initialSetupBusy = false;
    }
  }

  async function saveInitialSetupConfig() {
    initialSetupBusy = true;
    initialSetupError = "";
    configSaving = true;
    configError = "";
    try {
      await SaveReticulumConfigText(configText);
      await loadInterfaces();
      await loadCommunityInterfaces();
      await completeInitialSetup();
    } catch (err) {
      initialSetupError = formatBindingError(err, "Request failed");
    } finally {
      configSaving = false;
      initialSetupBusy = false;
    }
  }

  async function skipInitialSetupAutoOnly() {
    initialSetupBusy = true;
    initialSetupError = "";
    try {
      await completeInitialSetup();
    } catch (err) {
      initialSetupError = formatBindingError(err, "Request failed");
    } finally {
      initialSetupBusy = false;
    }
  }

  async function loadFavorites() {
    favorites = ((await GetFavorites()) ?? []) as string[];
  }

  async function loadHistory() {
    history = (await GetBrowsingHistory(50)) as HistoryEntry[];
  }

  async function loadTheme() {
    const shot = screenshotThemeFromQuery();
    if (shot) {
      theme = { ...theme, mode: shot };
      applyTheme(theme);
      await syncMobileChromeTheme(theme);
      return;
    }
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
      // Drop ready before load so first paint cannot stick on wasm-without-convert.
      micronWasmReady = false;
      micronWasmReady = await ensureMicronWasmReady(true, parserId);
    } else {
      micronWasmReady = false;
    }
  }

  function normalizeSettingsSectionsCollapsed(
    sections: { [_ in string]?: boolean } | null | undefined,
  ): Record<string, boolean> {
    const out: Record<string, boolean> = {};
    for (const [key, value] of Object.entries(sections ?? {})) {
      out[key] = !!value;
    }
    return out;
  }

  async function loadBrowserPrefs() {
    const prefs = await GetBrowserPrefs();
    openLinksInNewTab = !!prefs.openLinksInNewTab;
    nativeTitlebar = !!prefs.nativeTitlebar;
    uiLanguage = prefs.uiLanguage ?? "";
    docsLanguage = prefs.docsLanguage ?? "";
    initialSetupComplete = !!prefs.initialSetupComplete;
    initUILocale(uiLanguage);
    micronWasmEnabled = prefs.micronWasmEnabled ?? true;
    micronWasmAvailable = await isMicronWasmAvailable();
    micronWasmParserId = prefs.micronWasmParserId || BUNDLED_MICRON_WASM_PARSER_ID;
    micronRenderer = normalizeMicronRendererPreference(prefs.micronRenderer);
    discoverySlowMode = !!prefs.discoverySlowMode;
    mobileDevTools = !!prefs.mobileDevTools;
    pageCacheEnabled = prefs.pageCacheEnabled !== false;
    tabHoverPreviews = prefs.tabHoverPreviews !== false;
    micronPreserveLayout = !!prefs.micronPreserveLayout;
    settingsSectionsCollapsed = normalizeSettingsSectionsCollapsed(prefs.settingsSectionsCollapsed);
    if (activePanel === "devtools" && !mobileDevTools) {
      activePanel = "browser";
    }
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
      discoverySlowMode,
      mobileDevTools,
      pageCacheEnabled,
      tabHoverPreviews,
      micronPreserveLayout,
      settingsSectionsCollapsed,
      initialSetupComplete,
    };
  }

  async function persistBrowserPrefs(patch: {
    openLinksInNewTab?: boolean;
    nativeTitlebar?: boolean;
    micronRenderer?: MicronRendererPreference;
    micronWasmEnabled?: boolean;
    micronWasmParserId?: string;
    uiLanguage?: string;
    discoverySlowMode?: boolean;
    mobileDevTools?: boolean;
    pageCacheEnabled?: boolean;
    tabHoverPreviews?: boolean;
    micronPreserveLayout?: boolean;
    settingsSectionsCollapsed?: Record<string, boolean>;
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
    discoverySlowMode = !!prefs.discoverySlowMode;
    mobileDevTools = !!prefs.mobileDevTools;
    pageCacheEnabled = prefs.pageCacheEnabled !== false;
    tabHoverPreviews = prefs.tabHoverPreviews !== false;
    micronPreserveLayout = !!prefs.micronPreserveLayout;
    settingsSectionsCollapsed = normalizeSettingsSectionsCollapsed(prefs.settingsSectionsCollapsed);
    if (activePanel === "devtools" && !mobileDevTools) {
      activePanel = "browser";
    }
    await refreshMicronWasmState(micronWasmParserId);
    return prefs;
  }

  async function saveTabHoverPreviews(value: boolean) {
    tabHoverPreviews = value;
    await persistBrowserPrefs({ tabHoverPreviews: value });
  }

  async function savePageCacheEnabled(value: boolean) {
    if (!value && pageCacheEnabled) {
      await ClearPageCache();
      await loadPageCacheStats();
    }
    pageCacheEnabled = value;
    await persistBrowserPrefs({ pageCacheEnabled: value });
  }

  async function saveMobileDevTools(value: boolean) {
    mobileDevTools = value;
    if (!value && activePanel === "devtools") {
      activePanel = "browser";
    }
    await persistBrowserPrefs({ mobileDevTools: value });
  }

  async function saveMicronPreserveLayout(value: boolean) {
    micronPreserveLayout = value;
    await persistBrowserPrefs({ micronPreserveLayout: value });
  }

  async function saveDiscoverySlowMode(value: boolean) {
    discoverySlowMode = value;
    restartStatusTimer();
    await persistBrowserPrefs({ discoverySlowMode: value });
  }

  async function saveSettingsSectionsCollapsed(sections: Record<string, boolean>) {
    settingsSectionsCollapsed = sections;
    await persistBrowserPrefs({ settingsSectionsCollapsed: sections });
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
      error = formatBindingError(err, "Request failed");
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

  async function loadSandboxStatus() {
    const status = await GetSandboxStatus();
    sandboxStatus = {
      type: status.type ?? "none",
      enabled: !!status.enabled,
      requested: !!status.requested,
      supported: !!status.supported,
      auto: !!status.auto,
      disabledByEnv: !!status.disabledByEnv,
      reason: status.reason,
      inFlatpak: !!status.inFlatpak,
      inAppImage: !!status.inAppImage,
      inContainer: !!status.inContainer,
      containerRuntime: status.containerRuntime,
      webkitSandbox: status.webkitSandbox || "unavailable",
      webkitSandboxNote: status.webkitSandboxNote,
      onAndroid: !!status.onAndroid,
    };
  }

  function handleAndroidBack(): boolean {
    const closeConfirm = (): boolean => {
      if (identifyConfirmOpen) {
        identifyConfirmOpen = false;
        return true;
      }
      if (resetDbConfirmOpen) {
        resetDbConfirmOpen = false;
        return true;
      }
      if (resetBrowserConfirmOpen) {
        resetBrowserConfirmOpen = false;
        return true;
      }
      if (restartReticulumConfirmOpen) {
        restartReticulumConfirmOpen = false;
        return true;
      }
      if (transportMobileConfirmOpen) {
        transportMobileConfirmOpen = false;
        return true;
      }
      if (closeAllConfirmOpen) {
        closeAllConfirmOpen = false;
        return true;
      }
      if (shutdownConfirmOpen) {
        shutdownConfirmOpen = false;
        return true;
      }
      if (clearHistoryConfirmOpen) {
        clearHistoryConfirmOpen = false;
        return true;
      }
      return false;
    };

    if (closeConfirm()) {
      return true;
    }
    if (downloadsOpen) {
      downloadsOpen = false;
      return true;
    }
    if (mobileTabsOpen) {
      mobileTabsOpen = false;
      return true;
    }
    if (findOpen) {
      findOpen = false;
      return true;
    }
    if (activePanel !== "browser") {
      activePanel = "browser";
      return true;
    }
    if (canGoBack) {
      void goBack();
      return true;
    }
    return false;
  }

  function requestResetDatabase() {
    resetDbConfirmOpen = true;
  }

  function requestResetBrowser() {
    resetBrowserConfirmOpen = true;
  }

  function requestRestartReticulum() {
    restartReticulumConfirmOpen = true;
  }

  function requestShutdown() {
    shutdownConfirmOpen = true;
  }

  function confirmShutdown() {
    shutdownConfirmOpen = false;
    void Shutdown();
  }

  async function confirmResetBrowser() {
    resetBrowserConfirmOpen = false;
    await ResetBrowser();
  }

  async function confirmRestartReticulum() {
    restartReticulumConfirmOpen = false;
    try {
      await RestartReticulum();
      await loadInterfaces();
    } catch (err) {
      configError = formatBindingError(err, "Restart failed");
    }
  }

  async function runSelfTest() {
    selfTestRunning = true;
    selfTestResult = null;
    try {
      selfTestResult = await RunSelfCheck();
    } catch (err) {
      console.error("Self check failed:", err);
    } finally {
      selfTestRunning = false;
    }
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
        detail: formatBindingError(err, "Request failed"),
        path: storeHealth.path,
      };
    }
  }

  function tabIdForPageEvent(page: PageResponse): string | undefined {
    const pageUrl = (page.url ?? "").trim();
    if (!pageUrl) {
      return undefined;
    }
    const loadingTab = tabs.find((tab) => tab.url === pageUrl && tab.loading);
    if (loadingTab) {
      return loadingTab.id;
    }
    return tabs.find((tab) => tab.url === pageUrl)?.id;
  }

  function applyAsyncPageLoaded(page: PageResponse) {
    const tabId = tabIdForPageEvent(page);
    if (!tabId) {
      return;
    }
    const target = tabs.find((tab) => tab.id === tabId);
    if (!target?.loading) {
      return;
    }
    const tabPage = pageFromResponse(page);
    applyPageToTab(tabId, tabPage, page.url || target.url || "");
    schedulePersistTabs();
  }

  function applyAsyncPageError(page: PageResponse) {
    const tabId = tabIdForPageEvent(page);
    if (!tabId) {
      return;
    }
    const target = tabs.find((tab) => tab.id === tabId);
    const tabPage = pageFromResponse(page);
    applyPageToTab(tabId, tabPage, page.url || target?.url || "");
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
    discoverySlowMode = !!reset.browserPrefs.discoverySlowMode;
    mobileDevTools = !!reset.browserPrefs.mobileDevTools;
    pageCacheEnabled = reset.browserPrefs.pageCacheEnabled !== false;
    tabHoverPreviews = reset.browserPrefs.tabHoverPreviews !== false;
    micronPreserveLayout = !!reset.browserPrefs.micronPreserveLayout;
    if (activePanel === "devtools" && !mobileDevTools) {
      activePanel = "browser";
    }
    await refreshMicronWasmState(micronWasmParserId);
  }

  async function loadDownloadDir() {
    downloadDir = await GetDownloadDir();
  }

  async function loadDownloads() {
    downloads = ((await ListDownloads()) ?? []) as DownloadRow[];
  }

  async function loadActiveDownloads() {
    activeDownloads = ((await ListActiveDownloads()) ?? []) as ActiveDownloadRow[];
  }

  async function cancelActiveDownload(id: string) {
    const item = activeDownloads.find((entry) => entry.id === id);
    try {
      const ok = await CancelDownload(id);
      if (ok) {
        showPluginToast(canceledDownloadToast(item?.name, t));
        return;
      }
      showPluginToast(t("downloads.cancelFailed"), { isError: true });
    } catch (err) {
      showPluginToast(formatBindingError(err, t("downloads.cancelFailed")), {
        isError: true,
      });
    }
  }

  async function dismissActiveDownload(id: string) {
    await DismissDownload(id);
    activeDownloads = activeDownloads.filter((item) => item.id !== id);
  }

  async function retryActiveDownload(id: string) {
    if (retryingDownloadIds.has(id)) {
      return;
    }
    retryingDownloadIds.add(id);
    try {
      const result = await RetryDownload(id);
      if (result?.ok) {
        await loadActiveDownloads();
        showPluginToast(t("downloads.retryStarted"));
        return;
      }
      showPluginToast(formatBindingError(result?.error, t("downloads.retryFailed")), {
        isError: true,
      });
    } catch (err) {
      showPluginToast(formatBindingError(err, t("downloads.retryFailed")), {
        isError: true,
      });
    } finally {
      retryingDownloadIds.delete(id);
    }
  }

  async function saveDownloadDir(dir: string) {
    downloadDir = await SetDownloadDir(dir);
    await loadDownloads();
  }

  async function pickDownloadDir() {
    downloadDir = await PickDownloadDir();
    await loadDownloads();
  }

  async function clearDownloadHistory() {
    if (clearingDownloadHistory) {
      return;
    }
    clearingDownloadHistory = true;
    try {
      const result = await ClearDownloadHistory();
      if (result?.ok) {
        downloads = [];
        showPluginToast(t("downloads.historyCleared"));
        return;
      }
      showPluginToast(formatBindingError(result?.error, t("downloads.clearHistoryFailed")), {
        isError: true,
      });
    } catch (err) {
      showPluginToast(formatBindingError(err, t("downloads.clearHistoryFailed")), {
        isError: true,
      });
    } finally {
      clearingDownloadHistory = false;
    }
  }

  function handleDownloadResult(result: DownloadResult) {
    if (result.canceled) {
      showPluginToast(canceledDownloadToast(result.name, t));
      return;
    }
    if (result.ok && !result.pending) {
      void loadDownloads();
    }
    if (result.pending) {
      showPluginToast(result.message);
      return;
    }
    showPluginToast(result.message, { isError: !result.ok });
  }

  async function downloadCurrentPage() {
    const tab = tabs.find((item) => item.active);
    if (!tab?.url) {
      showPluginToast(t("downloads.noPageToSave"));
      return;
    }
    syncActiveTabPage();
    const page = currentPageState();
    const payload = page.lastRaw || page.html;
    if (!payload.trim()) {
      showPluginToast(t("downloads.noPageContent"));
      return;
    }
    try {
      await downloadPageContent(tab.url, page.contentType, payload);
      await loadDownloads();
      showPluginToast(t("downloads.saved"));
    } catch (err) {
      if (isDownloadCanceledError(err)) {
        showPluginToast(canceledDownloadToast(pageDownloadName(tab.url, page.contentType), t));
        return;
      }
      showPluginToast(downloadFailureMessage(err, t("downloads.downloadFailed")), {
        isError: true,
      });
    }
  }

  function toggleDownloads() {
    downloadsOpen = !downloadsOpen;
    if (downloadsOpen) {
      void loadDownloads();
      mobileTabsOpen = false;
      if (mobileUI && activePanel !== "browser") {
        activePanel = "browser";
      }
    }
  }

  function openMobileTabs() {
    mobileTabsOpen = true;
    downloadsOpen = false;
    activePanel = "browser";
  }

  function closeMobileTabs() {
    mobileTabsOpen = false;
  }

  function mobileSelectTab(id: string) {
    setActiveTab(id);
    mobileTabsOpen = false;
  }

  function mobileHome() {
    mobileTabsOpen = false;
    downloadsOpen = false;
    activePanel = "browser";
    const home = nodeHomeURL(url);
    if (!home || isNodeHomePage(url)) {
      return;
    }
    void openPage(home);
  }

  async function openDownload(path: string) {
    await OpenDownloadPath(path);
    downloadsOpen = false;
  }

  function readDownload(path: string) {
    downloadsOpen = false;
    void openPage(documentURL(path, downloadDir));
  }

  async function openDownloadFolder() {
    await ShowDownloadDir();
    downloadsOpen = false;
  }

  async function openConfigFolder() {
    configError = "";
    try {
      await ShowConfigDir();
    } catch (err) {
      configError = formatBindingError(err, t("config.openFolderFailed"));
    }
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
      case "search":
        setPanel("search");
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

  async function applyScreenshotScene(scene: ScreenshotScene) {
    try {
      switch (scene) {
        case "about":
          await openPage("about:", false);
          break;
        case "editor":
          await openPage("editor:", false);
          break;
        case "settings":
          activePanel = "settings";
          break;
        case "discovery":
          activePanel = "discovery";
          break;
        case "docs":
          await openPage("docs:?lang=en&page=getting-started", false);
          break;
        case "home":
          break;
      }
      await new Promise((resolve) => setTimeout(resolve, 400));
    } finally {
      markScreenshotReady();
    }
  }

  async function toggleInterface(name: string, enabled: boolean) {
    await SetInterfaceEnabled(name, enabled);
    await loadInterfaces();
  }

  async function applyEnableTransport(enabled: boolean) {
    await SetEnableTransport(enabled);
    await loadReticulumStatus();
  }

  async function toggleTransport(enabled: boolean) {
    if (enabled && mobileUI) {
      transportMobileConfirmOpen = true;
      return;
    }
    await applyEnableTransport(enabled);
  }

  async function confirmEnableTransportMobile() {
    transportMobileConfirmOpen = false;
    await applyEnableTransport(true);
  }

  async function toggleShareInstance(enabled: boolean) {
    await SetShareInstance(enabled);
    await loadReticulumStatus();
    showPluginToast(t("settings.shareInstanceRestartHint"));
  }

  async function exportThemeFile() {
    const json = await ExportTheme();
    const blob = new Blob([json], { type: "application/json" });
    const href = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = href;
    a.download = exportFilename("theme");
    a.click();
    URL.revokeObjectURL(href);
  }

  async function importThemeFile(json: string) {
    theme = (await ImportTheme(json)) as ThemeSettings;
    applyTheme(theme);
    await syncMobileChromeTheme(theme);
  }

  async function clearDevLogsPanel() {
    await ClearDevLogs();
    logs = [];
  }

  async function exportDevLogsFile() {
    const json = await ExportDevLogs();
    const blob = new Blob([json], { type: "application/json" });
    const href = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = href;
    a.download = exportFilename("devlogs");
    a.click();
    URL.revokeObjectURL(href);
  }

  async function setDevLogLevel(level: number) {
    logLevel = await SetLogLevel(level);
  }

  async function addFavoriteUrl(favUrl: string) {
    favorites = (await AddFavorite(favUrl)) as string[];
  }

  function mount() {
    const onResize = () => {
      compactViewport = isCompactViewport();
    };
    onResize();
    window.addEventListener("resize", onResize, { passive: true });

    void loadTheme();
    void loadKeybinds();
    void loadBrowserPrefs();
    void loadWindowChrome();
    void loadDownloadDir();
    void loadDownloads();
    void loadActiveDownloads();
    void ListSystemFonts().then((fonts) => {
      if (Array.isArray(fonts) && fonts.length > 0) {
        systemFonts = fonts as string[];
      }
    });
    const screenshotScene = screenshotSceneFromQuery();
    void loadNodes().then(async () => {
      if (!screenshotScene) {
        const saved = (await GetTabs()) as TabSnapshot[];
        restoreTabs(saved);
        for (const tab of tabs) {
          if (tab.url && isDocumentURL(tab.url) && !tab.page?.binaryB64?.trim()) {
            void openPage(tab.url, false, { tabId: tab.id });
          }
        }
      }
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
    void loadSandboxStatus();
    void loadRuntimeConfig().then(() => checkInitialSetup());
    void bootPlugins();

    Events.On("plugin:loaded", () => {
      void reloadPlugins();
    });
    Events.On("plugin:unloaded", () => {
      void reloadPlugins();
    });
    Events.On("plugin:error", (event) => {
      try {
        const data = JSON.parse(String((event as { data?: string }).data ?? "")) as {
          pluginId?: string;
          message?: string;
        };
        if (data.pluginId && data.message) {
          showPluginToast(
            `Extension ${data.pluginId} disabled: ${formatBindingError(data.message, "Extension failed")}`,
            { isError: true },
          );
        }
      } catch {
        // Ignore malformed plugin error payloads.
      }
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

    let lastDeepLinkURL = "";
    let lastDeepLinkAt = 0;
    const openDeepLinkTarget = (raw: string) => {
      const unwrapped = unwrapDeepLink(String(raw ?? ""));
      const normalized = normalizeReticulumURL(unwrapped || String(raw ?? ""));
      if (!normalized) {
        return;
      }
      const now = Date.now();
      if (normalized === lastDeepLinkURL && now - lastDeepLinkAt < 1500) {
        return;
      }
      lastDeepLinkURL = normalized;
      lastDeepLinkAt = now;
      void openPage(normalized, true);
      void TakePendingDeepLink().catch(() => {});
    };

    Events.On("app:deeplink", (event) => {
      const raw =
        typeof event.data === "string"
          ? event.data
          : String((event as { data?: unknown }).data ?? "");
      openDeepLinkTarget(raw);
    });

    void TakePendingDeepLink()
      .then((pending) => {
        if (pending) {
          openDeepLinkTarget(pending);
        }
      })
      .catch(() => {});

    restartStatusTimer();

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
      }
    });

    Events.On("page:error", (event) => {
      applyAsyncPageError(event.data as PageResponse);
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
      scheduleLoadNodesFromEvent();
    });

    Events.On("downloads:active", (event) => {
      activeDownloads = (event.data ?? []) as ActiveDownloadRow[];
      if (activeDownloads.some((item) => item.status === "completed")) {
        void loadDownloads();
      }
    });

    Events.On("dev:log", (payload: { data: string }) => {
      try {
        const entry = JSON.parse(payload.data) as DevLogEntry;
        logs = [...logs.slice(-499), entry];
      } catch {
        logs = [...logs.slice(-499), { time: Date.now(), level: "info", message: payload.data }];
      }
    });

    Events.On("page:loaded", (event) => {
      applyAsyncPageLoaded(event.data as PageResponse);
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

    const onVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        void resumeForegroundSync();
      } else {
        handleAppBackground();
      }
    };
    document.addEventListener("visibilitychange", onVisibilityChange);

    Events.On("android:ActivityResumed", () => {
      void resumeForegroundSync();
    });
    Events.On("android:ActivityPaused", () => {
      handleAppBackground();
    });
    Events.On("android:ActivityStopped", () => {
      handleAppBackground();
    });

    const androidBackWindow = window as Window & {
      __renHandleAndroidBack?: () => boolean;
    };
    androidBackWindow.__renHandleAndroidBack = handleAndroidBack;

    const onKeyDown = (event: KeyboardEvent) => {
      handleGlobalKeyDown(event);
    };
    const blockExternalLink = (event: Event) => {
      blockExternalLinkPointerEvent(event);
    };
    window.addEventListener("keydown", onKeyDown);
    document.addEventListener("click", blockExternalLink, true);
    document.addEventListener("auxclick", blockExternalLink, true);

    if (screenshotScene) {
      void loadNodes().then(async () => {
        await new Promise((resolve) => setTimeout(resolve, 600));
        await applyScreenshotScene(screenshotScene);
      });
    } else if (screenshotThemeFromQuery()) {
      setTimeout(() => {
        markScreenshotReady();
      }, 1200);
    }

    return () => {
      if (statusTimer !== undefined) {
        clearInterval(statusTimer);
      }
      if (nodeDiscoverTimer !== undefined) {
        clearTimeout(nodeDiscoverTimer);
      }
      window.removeEventListener("keydown", onKeyDown);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      document.removeEventListener("click", blockExternalLink, true);
      document.removeEventListener("auxclick", blockExternalLink, true);
      window.removeEventListener("resize", onResize);
      if (androidBackWindow.__renHandleAndroidBack === handleAndroidBack) {
        delete androidBackWindow.__renHandleAndroidBack;
      }
      if (persistTimer) {
        clearTimeout(persistTimer);
      }
      void persistTabs();
    };
  }

  return {
    get activePanel() {
      return activePanel;
    },
    get pluginContributions() {
      return pluginContributions;
    },
    get pluginToast() {
      return pluginToast;
    },
    get pluginToastIsError() {
      return pluginToastIsError;
    },
    get pluginsDir() {
      return pluginsDir;
    },
    get pluginGrantedById() {
      return pluginGrantedById;
    },
    get url() {
      return url;
    },
    set url(value) {
      url = value;
    },
    get loading() {
      return loading;
    },
    get html() {
      return html;
    },
    get contentType() {
      return contentType;
    },
    get error() {
      return error;
    },
    get errorKind() {
      return errorKind;
    },
    get durationMs() {
      return durationMs;
    },
    get hops() {
      return hops;
    },
    get pageFg() {
      return pageFg;
    },
    get pageBg() {
      return pageBg;
    },
    get nodes() {
      return nodes;
    },
    get logs() {
      return logs;
    },
    get network() {
      return network;
    },
    get favorites() {
      return favorites;
    },
    get history() {
      return history;
    },
    get interfaces() {
      return interfaces;
    },
    get reticulumStatus() {
      return reticulumStatus;
    },
    get configPath() {
      return configPath;
    },
    get logLevel() {
      return logLevel;
    },
    get systemFonts() {
      return systemFonts;
    },
    get theme() {
      return theme;
    },
    set theme(value) {
      theme = value;
    },
    get overlaySidebars() {
      return theme.overlaySidebars;
    },
    get keybinds() {
      return keybinds;
    },
    get downloadDir() {
      return downloadDir;
    },
    set downloadDir(value) {
      downloadDir = value;
    },
    get downloads() {
      return downloads;
    },
    get activeDownloads() {
      return activeDownloads;
    },
    retryingDownloadIds,
    get clearingDownloadHistory() {
      return clearingDownloadHistory;
    },
    get activeDownloadViews() {
      return activeDownloadViews;
    },
    get downloadsOpen() {
      return downloadsOpen;
    },
    set downloadsOpen(value) {
      downloadsOpen = value;
    },
    get findOpen() {
      return findOpen;
    },
    set findOpen(value) {
      findOpen = value;
    },
    get pageHighlight() {
      return pageHighlight;
    },
    get canGoBack() {
      return canGoBack;
    },
    get canGoForward() {
      return canGoForward;
    },
    get lastRaw() {
      return lastRaw;
    },
    get binaryB64() {
      return binaryB64;
    },
    get pagePath() {
      return pagePath;
    },
    get fromCache() {
      return fromCache;
    },
    get cachedAt() {
      return cachedAt;
    },
    get showSource() {
      return showSource;
    },
    get openLinksInNewTab() {
      return openLinksInNewTab;
    },
    get nativeTitlebar() {
      return nativeTitlebar;
    },
    get uiLanguage() {
      return uiLanguage;
    },
    get docsLanguage() {
      return docsLanguage;
    },
    get micronRenderer() {
      return micronRenderer;
    },
    get micronWasmEnabled() {
      return micronWasmEnabled;
    },
    get micronWasmReady() {
      return micronWasmReady;
    },
    get micronWasmAvailable() {
      return micronWasmAvailable;
    },
    get micronWasmParserId() {
      return micronWasmParserId;
    },
    get micronWasmParserLabel() {
      return micronWasmParserLabel;
    },
    get identifying() {
      return identifying;
    },
    get identifyConfirmOpen() {
      return identifyConfirmOpen;
    },
    set identifyConfirmOpen(value) {
      identifyConfirmOpen = value;
    },
    get resetDbConfirmOpen() {
      return resetDbConfirmOpen;
    },
    set resetDbConfirmOpen(value) {
      resetDbConfirmOpen = value;
    },
    get resetBrowserConfirmOpen() {
      return resetBrowserConfirmOpen;
    },
    set resetBrowserConfirmOpen(value) {
      resetBrowserConfirmOpen = value;
    },
    get selfTestResult() {
      return selfTestResult;
    },
    set selfTestResult(value) {
      selfTestResult = value;
    },
    get selfTestRunning() {
      return selfTestRunning;
    },
    set selfTestRunning(value) {
      selfTestRunning = value;
    },
    get restartReticulumConfirmOpen() {
      return restartReticulumConfirmOpen;
    },
    set restartReticulumConfirmOpen(value) {
      restartReticulumConfirmOpen = value;
    },
    get transportMobileConfirmOpen() {
      return transportMobileConfirmOpen;
    },
    set transportMobileConfirmOpen(value) {
      transportMobileConfirmOpen = value;
    },
    get closeAllConfirmOpen() {
      return closeAllConfirmOpen;
    },
    set closeAllConfirmOpen(value) {
      closeAllConfirmOpen = value;
    },
    get shutdownConfirmOpen() {
      return shutdownConfirmOpen;
    },
    set shutdownConfirmOpen(value) {
      shutdownConfirmOpen = value;
    },
    get clearHistoryConfirmOpen() {
      return clearHistoryConfirmOpen;
    },
    set clearHistoryConfirmOpen(value) {
      clearHistoryConfirmOpen = value;
    },
    get publicMode() {
      return publicMode;
    },
    get serverMode() {
      return serverMode;
    },
    get storeHealth() {
      return storeHealth;
    },
    get sandboxStatus() {
      return sandboxStatus;
    },
    get meshOnline() {
      return meshOnline;
    },
    get splitViewOpen() {
      return splitViewOpen;
    },
    get splitTabId() {
      return splitTabId;
    },
    get splitRatio() {
      return splitRatio;
    },
    set splitRatio(value) {
      splitRatio = value;
    },
    get desktopChrome() {
      return desktopChrome;
    },
    get compactViewport() {
      return compactViewport;
    },
    get mobileUI() {
      return mobileUI;
    },
    get configText() {
      return configText;
    },
    set configText(value) {
      configText = value;
    },
    get configSaving() {
      return configSaving;
    },
    get configError() {
      return configError;
    },
    get pageCacheEntries() {
      return pageCacheEntries;
    },
    get pageCacheMax() {
      return pageCacheMax;
    },
    get pageCacheClearing() {
      return pageCacheClearing;
    },
    get pageCacheEnabled() {
      return pageCacheEnabled;
    },
    get communityItems() {
      return communityItems;
    },
    get communityLoading() {
      return communityLoading;
    },
    get communityImporting() {
      return communityImporting;
    },
    get communityError() {
      return communityError;
    },
    get communityFromBundle() {
      return communityFromBundle;
    },
    get communityFilter() {
      return communityFilter;
    },
    set communityFilter(value) {
      communityFilter = value;
    },
    communitySelected,
    get initialSetupOpen() {
      return initialSetupOpen;
    },
    get initialSetupStep() {
      return initialSetupStep;
    },
    get suggestedItems() {
      return suggestedItems;
    },
    get suggestedLoading() {
      return suggestedLoading;
    },
    get initialSetupBusy() {
      return initialSetupBusy;
    },
    get initialSetupError() {
      return initialSetupError;
    },
    get discoverySlowMode() {
      return discoverySlowMode;
    },
    get mobileDevTools() {
      return mobileDevTools;
    },
    get tabHoverPreviews() {
      return tabHoverPreviews;
    },
    get micronPreserveLayout() {
      return micronPreserveLayout;
    },
    get mobileTabsOpen() {
      return mobileTabsOpen;
    },
    get settingsSectionsCollapsed() {
      return settingsSectionsCollapsed;
    },
    get tabs() {
      return tabs;
    },
    get effectiveMicronEngine() {
      return effectiveMicronEngine;
    },
    get micronRendererBadge() {
      return micronRendererBadge;
    },
    get canIdentify() {
      return canIdentify;
    },
    get atTabLimit() {
      return atTabLimit;
    },
    get activeTabId() {
      return activeTabId;
    },
    get splitTab() {
      return splitTab;
    },
    get storeErrorVisible() {
      return storeErrorVisible;
    },
    get activePluginPanel() {
      return activePluginPanel;
    },
    emptyPage,
    setPanel,
    showPluginToast,
    pluginHostOpts,
    openPage,
    browseURL,
    clearPageHighlight,
    goBack,
    goForward,
    setActiveTab,
    closeTab,
    closeOtherTabs,
    closeTabsToRight,
    requestCloseAllTabs,
    confirmCloseAllTabs,
    requestClearHistory,
    confirmClearHistory,
    togglePinTab,
    splitTabView,
    selectSplitTab,
    closeSplitView,
    setSplitTabShowSource,
    downloadTab,
    viewSourceTab,
    newTab,
    reorderTabs,
    reloadTab,
    duplicateTab,
    favoriteTab,
    updateEditorSource,
    setShowSource,
    requestIdentify,
    confirmIdentify,
    requestResetDatabase,
    requestResetBrowser,
    requestRestartReticulum,
    requestShutdown,
    confirmShutdown,
    confirmResetBrowser,
    confirmRestartReticulum,
    confirmResetDatabase,
    runSelfTest,
    resetDefaults,
    saveUILanguage,
    saveTabHoverPreviews,
    savePageCacheEnabled,
    saveMobileDevTools,
    saveDiscoverySlowMode,
    saveSettingsSectionsCollapsed,
    saveOpenLinksInNewTab,
    saveNativeTitlebar,
    saveMicronRenderer,
    saveMicronWasmEnabled,
    saveMicronPreserveLayout,
    saveMicronWasmParser,
    setMicronWasmReady,
    saveTheme,
    saveKeybinds,
    saveDownloadDir,
    pickDownloadDir,
    clearDownloadHistory,
    handleDownloadResult,
    downloadCurrentPage,
    toggleDownloads,
    openMobileTabs,
    closeMobileTabs,
    mobileSelectTab,
    mobileHome,
    openDownload,
    readDownload,
    openDownloadFolder,
    openConfigFolder,
    saveConfigText,
    reloadConfigFromDisk,
    clearPageCache,
    loadCommunityInterfaces,
    toggleCommunitySelection,
    importCommunitySelection,
    setInitialSetupStep,
    loadSuggestedPreview,
    applySuggestedSetup,
    importInitialSetupSelection,
    saveInitialSetupConfig,
    skipInitialSetupAutoOnly,
    reloadPlugins,
    loadStoreHealth,
    toggleInterface,
    toggleTransport,
    confirmEnableTransportMobile,
    toggleShareInstance,
    exportThemeFile,
    importThemeFile,
    clearDevLogsPanel,
    exportDevLogsFile,
    setDevLogLevel,
    addFavoriteUrl,
    cancelActiveDownload,
    dismissActiveDownload,
    retryActiveDownload,
    toggleTheme,
    mount,
  };
}
