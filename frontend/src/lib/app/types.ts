// SPDX-License-Identifier: MIT
import type { CommunityInterface } from "../../../bindings/renbrowser/internal/rns/models.js";
import type { DownloadRow } from "$lib/components/DownloadsMenu.svelte";
import type { KeybindSettings } from "$lib/browser/keybinds";
import type { MicronRendererPreference } from "$lib/micron/render-page";
import type { ActivePanel, ContributionsSnapshot } from "$lib/plugins/api-types.js";
import type { ThemeSettings } from "$lib/theme/tokens";
import type { Tab, TabPage } from "$lib/browser/url";
import type { ActiveDownloadRow } from "$lib/browser/download-progress";

export type InitialSetupStep = "welcome" | "suggested" | "pick" | "config";

export type Node = {
  hash: string;
  name: string;
  hops: number;
  lastSeen: number;
};

export type PageResponse = {
  url: string;
  nodeHash: string;
  path: string;
  contentType: string;
  html: string;
  raw: string;
  binaryB64?: string;
  pageFg?: string;
  pageBg?: string;
  durationMs: number;
  fromCache?: boolean;
  cachedAt?: number;
  hops?: number;
  error?: string;
  errorKind?: string;
};

export type DevLogEntry = {
  time: number;
  level: string;
  message: string;
  detail?: string;
};

export type NetworkEntry = {
  time: number;
  url: string;
  nodeHash: string;
  path: string;
  durationMs: number;
  bytes: number;
  fromCache: boolean;
  hops: number;
  interface?: string;
  error?: string;
};

export type InterfaceRow = {
  name: string;
  type: string;
  enabled: boolean;
  online: boolean;
  txBytes: number;
  rxBytes: number;
};

export type ReticulumStatusRow = {
  enableTransport: boolean;
  shareInstance: boolean;
  connectedToSharedInstance: boolean;
  sharedInstanceMode: string;
  transportActive: boolean;
};

export type TabSnapshot = {
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

export type StoreHealth = {
  ok: boolean;
  kind?: string;
  detail?: string;
  path: string;
};

export type SandboxStatus = {
  type: string;
  enabled: boolean;
  requested?: boolean;
  supported?: boolean;
  auto?: boolean;
  disabledByEnv?: boolean;
  reason?: string;
  abi?: number;
  seccompEnabled?: boolean;
  seccompSupported?: boolean;
  seccompReason?: string;
  inFlatpak?: boolean;
  inAppImage?: boolean;
  inContainer?: boolean;
  containerRuntime?: string;
  webkitSandbox?: string;
  webkitSandboxNote?: string;
  onAndroid?: boolean;
};

export type HistoryEntry = {
  id: number;
  url: string;
  title: string;
  nodeHash: string;
  visitedAt: number;
};

export type AppState = {
  activePanel: ActivePanel;
  pluginContributions: ContributionsSnapshot;
  pluginToast: string;
  pluginToastIsError: boolean;
  pluginsDir: string;
  pluginGrantedById: Record<string, string[]>;
  url: string;
  loading: boolean;
  html: string;
  contentType: string;
  error: string;
  errorKind: string;
  durationMs: number;
  hops: number;
  pageFg: string;
  pageBg: string;
  nodes: Node[];
  logs: DevLogEntry[];
  network: NetworkEntry[];
  favorites: string[];
  history: HistoryEntry[];
  interfaces: InterfaceRow[];
  configPath: string;
  logLevel: number;
  systemFonts: string[];
  theme: ThemeSettings;
  keybinds: KeybindSettings;
  downloadDir: string;
  downloads: DownloadRow[];
  activeDownloads: ActiveDownloadRow[];
  clearingDownloadHistory: boolean;
  downloadsOpen: boolean;
  findOpen: boolean;
  canGoBack: boolean;
  canGoForward: boolean;
  lastRaw: string;
  binaryB64: string;
  pagePath: string;
  fromCache: boolean;
  cachedAt: number;
  showSource: boolean;
  openLinksInNewTab: boolean;
  nativeTitlebar: boolean;
  uiLanguage: string;
  docsLanguage: string;
  micronRenderer: MicronRendererPreference;
  micronWasmEnabled: boolean;
  micronWasmReady: boolean;
  micronWasmAvailable: boolean;
  micronWasmParserId: string;
  micronWasmParserLabel: string;
  identifying: boolean;
  identifyConfirmOpen: boolean;
  resetDbConfirmOpen: boolean;
  closeAllConfirmOpen: boolean;
  shutdownConfirmOpen: boolean;
  clearHistoryConfirmOpen: boolean;
  publicMode: boolean;
  serverMode: boolean;
  storeHealth: StoreHealth;
  sandboxStatus: SandboxStatus;
  meshOnline: boolean;
  splitViewOpen: boolean;
  splitTabId: string | null;
  splitRatio: number;
  compactViewport: boolean;
  configText: string;
  configSaving: boolean;
  configError: string;
  pageCacheEntries: number;
  pageCacheMax: number;
  pageCacheRAMEntries: number;
  pageCacheDiskEntries: number;
  pageCacheClearing: boolean;
  pageCacheEnabled: boolean;
  communityItems: CommunityInterface[];
  communityLoading: boolean;
  communityImporting: boolean;
  communityError: string;
  communityFromBundle: boolean;
  communityFilter: string;
  initialSetupOpen: boolean;
  initialSetupStep: InitialSetupStep;
  suggestedItems: CommunityInterface[];
  suggestedLoading: boolean;
  initialSetupBusy: boolean;
  initialSetupError: string;
  discoverySlowMode: boolean;
  mobileDevTools: boolean;
  tabHoverPreviews: boolean;
  mobileTabsOpen: boolean;
  settingsSectionsCollapsed: Record<string, boolean>;
  tabs: Tab[];
};

export type { Tab, TabPage };
