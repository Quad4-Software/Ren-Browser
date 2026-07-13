// SPDX-License-Identifier: MIT
import { vi } from "vitest";

export type BindingMock = ReturnType<typeof vi.fn>;

const emptyPage = {
  url: "",
  html: "",
  contentType: "text/html",
  raw: "",
  binaryB64: "",
  path: "",
  error: "",
  errorKind: "",
  durationMs: 1,
  pageFg: "",
  pageBg: "",
  fromCache: false,
  cachedAt: 0,
  hops: -1,
};

const defaultTheme = {
  mode: "dark",
  accent: "#3b82f6",
  fontFamily: "system-ui",
  fontSize: 14,
  customTokens: {},
  compactToolbar: false,
  overlaySidebars: false,
};

/** Default resolved values for BrowserService bindings used by createApp. */
export function browserserviceDefaults(): Record<string, unknown> {
  return {
    AddFavorite: undefined,
    ApplySuggestedCommunityInterfaces: undefined,
    AttachStack: undefined,
    CancelDownload: undefined,
    CaptureWindowState: undefined,
    ClearBrowsingHistory: undefined,
    ClearDevLogs: undefined,
    ClearDownloadHistory: undefined,
    ClearPageCache: undefined,
    CompleteInitialSetup: undefined,
    ConfigPath: "/tmp/reticulum.conf",
    CreateIdentity: undefined,
    DeleteIdentity: undefined,
    DevLog: undefined,
    DismissDownload: undefined,
    DownloadFile: undefined,
    DownloadToDir: undefined,
    ExportDevLogs: "",
    ExportIdentity: undefined,
    ExportProfile: undefined,
    ExportTheme: "",
    FetchCommunityInterfaces: { items: [], fromBundle: false, error: "" },
    FetchMicronParserGoRelease: undefined,
    GetAboutInfo: { version: "0.0.0", commit: "test" },
    GetApkShareInfo: undefined,
    GetApkShareSession: undefined,
    GetBrowserPrefs: {
      openLinksInNewTab: true,
      openLinksInNewWindow: false,
      nativeTitlebar: false,
      uiLanguage: "en",
      docsLanguage: "en",
      micronRenderer: "auto",
      micronWasmEnabled: true,
      micronPreserveLayout: false,
      micronWasmParserId: "",
      pageCacheEnabled: true,
      tabHoverPreviews: true,
      mobileDevTools: false,
      discoverySlowMode: false,
      initialSetupComplete: true,
      settingsSectionsCollapsed: null,
    },
    GetBrowsingHistory: [],
    GetDevLogs: [],
    GetDownloadDir: "/tmp/downloads",
    GetFavorites: [],
    GetInitialSetupState: { needed: false, suggestedCount: 0 },
    GetKeybinds: { bindings: null },
    GetLastPage: emptyPage,
    GetNetworkLog: [],
    GetPageCacheStats: { entries: 0, max: 128 },
    GetRecent: [],
    GetReticulumConfigText: "",
    GetRuntimeConfig: {
      publicMode: false,
      serverMode: true,
      profile: "default",
      profilePath: "/tmp/profile",
    },
    GetSandboxStatus: {
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
    },
    GetStatus: {
      online: true,
      nodeCount: 0,
      interfaceCount: 0,
      interfacesOnline: 0,
      configPath: "/tmp/reticulum.conf",
      enableTransport: false,
      shareInstance: false,
      connectedToSharedInstance: false,
      sharedInstanceMode: "disabled",
      transportActive: false,
    },
    GetStoreHealth: { ok: true, path: "/tmp/renbrowser.db" },
    GetTabs: [],
    GetTheme: defaultTheme,
    GetWindowChrome: { nativeTitlebar: false },
    GetWindowState: undefined,
    GoBack: emptyPage,
    GoForward: emptyPage,
    HandleDeepLink: undefined,
    HistoryState: { canGoBack: false, canGoForward: false },
    IdentifyToNode: undefined,
    ImportCommunityInterfaces: undefined,
    ImportIdentity: undefined,
    ImportProfile: undefined,
    ImportTheme: defaultTheme,
    InitialWindowOptions: undefined,
    ListActiveDownloads: [],
    ListDownloads: [],
    ListIdentities: [],
    ListInterfaces: [],
    ListNodes: [],
    ListProfiles: [],
    ListSystemFonts: ["system-ui", "sans-serif"],
    Navigate: { ...emptyPage, url: "about:", html: "<article class='about-page'>About</article>" },
    NavigateFresh: emptyPage,
    OpenDownloadPath: undefined,
    OpenFreshURL: emptyPage,
    OpenNewWindow: undefined,
    OpenURL: emptyPage,
    PickDownloadDir: "/tmp/downloads",
    PickIdentityExportPath: "",
    PickIdentityFile: "",
    PickPluginDir: "",
    PickPluginWasm: "",
    PickPluginZip: "",
    PluginManager: undefined,
    PrepareForWake: { droppedLinks: 0, expiredPaths: 0 },
    PreviewSuggestedCommunityInterfaces: [],
    ProfileName: "default",
    ProfilePath: "/tmp/profile",
    RecordPluginNetworkFetch: undefined,
    ReloadReticulumConfig: undefined,
    RemoveFavorite: undefined,
    RenameIdentity: undefined,
    RenderRaw: "",
    RenderRawBase64: "",
    ResetBrowser: undefined,
    ResetDatabase: undefined,
    ResetSettings: undefined,
    ResetWindowState: undefined,
    ResolveMicronLink: "",
    RestartReticulum: undefined,
    RetryDownload: undefined,
    RunSelfCheck: {
      stackUp: { passed: true },
      configGood: { passed: true },
      dbGood: { passed: true },
      readWriteGood: { passed: true },
      downloadsGood: { passed: true },
      interfaces: { passed: true },
      discovery: { passed: true },
      pageFetch: { passed: true },
      allPassed: true,
      meshEnabled: false,
    },
    SaveDownload: undefined,
    SaveReticulumConfigText: undefined,
    SaveTabs: [],
    SaveTextToDownloadDir: undefined,
    SaveWindowState: undefined,
    SetActiveIdentity: undefined,
    SetApp: undefined,
    SetBrowserPrefs: undefined,
    SetDownloadDir: undefined,
    SetEnableTransport: undefined,
    SetInterfaceEnabled: undefined,
    SetShareInstance: undefined,
    SetKeybinds: undefined,
    SetLogLevel: undefined,
    SetNativeTitlebar: undefined,
    SetPluginManager: undefined,
    SetTheme: undefined,
    ShareApk: undefined,
    ShowConfigDir: undefined,
    ShowDownloadDir: undefined,
    Shutdown: undefined,
    StartApkShareServer: undefined,
    StartReticulum: undefined,
    StopApkShareServer: undefined,
    StopReticulum: undefined,
    Store: undefined,
    SyncMobileChrome: undefined,
    TakePendingDeepLink: "",
    ToggleFullscreen: undefined,
  };
}

const bindingNames = Object.keys(browserserviceDefaults());

/** Build a vi.fn map for every BrowserService export. */
export function createBrowserserviceMocks(
  overrides: Record<string, BindingMock | unknown> = {},
): Record<string, BindingMock> {
  const defaults = browserserviceDefaults();
  const mocks: Record<string, BindingMock> = {};
  for (const name of bindingNames) {
    const override = overrides[name];
    if (typeof override === "function") {
      mocks[name] = override as BindingMock;
      continue;
    }
    const value = override !== undefined ? override : defaults[name];
    mocks[name] = vi.fn(async () => value);
  }
  return mocks;
}

export function createPluginhostMocks(
  overrides: Record<string, BindingMock | unknown> = {},
): Record<string, BindingMock> {
  const defaults: Record<string, unknown> = {
    AddTrustedPublisher: undefined,
    DisablePlugin: undefined,
    EmitPluginEvent: undefined,
    EnablePlugin: undefined,
    GetContributions: { panels: [], commands: [], devtools: [], urlSchemes: [] },
    GetPlugin: undefined,
    GetPluginStorage: "",
    InstallPluginFromDir: undefined,
    InstallPluginFromWasm: undefined,
    InstallPluginFromZip: undefined,
    InvokeCommand: undefined,
    ListPlugins: [],
    PluginFetch: undefined,
    PluginWasmCall: "",
    PluginsDir: "/tmp/plugins",
    PreviewPluginInstallFromDir: undefined,
    PreviewPluginInstallFromWasm: undefined,
    PreviewPluginInstallFromZip: undefined,
    ReportPluginError: undefined,
    SetPluginStorage: undefined,
    UninstallPlugin: undefined,
  };
  const mocks: Record<string, BindingMock> = {};
  for (const [name, value] of Object.entries({ ...defaults, ...overrides })) {
    if (typeof value === "function") {
      mocks[name] = value as BindingMock;
      continue;
    }
    mocks[name] = vi.fn(async () => value);
  }
  return mocks;
}

export type EventHandler = (event: { data?: unknown }) => void;

/** In-memory Wails Events/System stub for createApp tests. */
export function createRuntimeMock() {
  const handlers = new Map<string, Set<EventHandler>>();
  return {
    handlers,
    Events: {
      On: vi.fn((name: string, fn: EventHandler) => {
        let set = handlers.get(name);
        if (!set) {
          set = new Set();
          handlers.set(name, set);
        }
        set.add(fn);
        return () => set!.delete(fn);
      }),
      Emit: vi.fn((name: string, data?: unknown) => {
        const set = handlers.get(name);
        if (!set) {
          return;
        }
        for (const fn of set) {
          fn({ data });
        }
      }),
    },
    System: {
      IsDesktop: vi.fn(() => true),
      IsMobile: vi.fn(() => false),
    },
    emit(name: string, data?: unknown) {
      const set = handlers.get(name);
      if (!set) {
        return;
      }
      for (const fn of set) {
        fn({ data });
      }
    },
  };
}
