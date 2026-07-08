<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import {
    ArrowLeft,
    ArrowRight,
    Compass,
    Download,
    Fingerprint,
    History,
    RefreshCw,
    Search,
    Settings,
    Terminal,
    X,
  } from "@lucide/svelte";
  import DownloadsMenu, { type DownloadRow } from "$lib/components/DownloadsMenu.svelte";
  import type { DownloadProgressView } from "$lib/browser/download-progress";
  import { t } from "$lib/i18n/i18n.svelte";
  import { pluginLabel } from "$lib/plugins/plugin-label.js";
  import PluginLucideIcon from "$lib/components/PluginLucideIcon.svelte";
  import type { ActivePanel, PluginPanelContribution } from "$lib/plugins/api-types.js";
  import { panelKey } from "$lib/plugins/registry.js";

  type Props = {
    url: string;
    canGoBack: boolean;
    canGoForward: boolean;
    activePanel: ActivePanel;
    pluginPanels?: PluginPanelContribution[];
    devToolsEnabled: boolean;
    downloadsOpen: boolean;
    downloads: DownloadRow[];
    activeDownloads?: DownloadProgressView[];
    downloadDir: string;
    canIdentify: boolean;
    identifying: boolean;
    onNavigate: (url: string) => void;
    onBack: () => void;
    onForward: () => void;
    onReload: () => void;
    onPanel: (panel: ActivePanel) => void;
    onDownloadPage: () => void;
    onToggleDownloads: () => void;
    onCloseDownloads: () => void;
    onOpenDownload: (path: string) => void;
    onReadDownload?: (path: string) => void;
    onOpenDownloadFolder: () => void;
    onCancelDownload?: (id: string) => void;
    onDismissDownload?: (id: string) => void;
    onRetryDownload?: (id: string) => void;
    retryingDownloadIds?: ReadonlySet<string>;
    onClearDownloadHistory?: () => void;
    clearingDownloadHistory?: boolean;
    onIdentify: () => void;
  };

  let {
    url = $bindable(""),
    canGoBack,
    canGoForward,
    activePanel,
    pluginPanels = [],
    devToolsEnabled,
    downloadsOpen,
    downloads,
    activeDownloads = [],
    downloadDir,
    canIdentify,
    identifying,
    onNavigate,
    onBack,
    onForward,
    onReload,
    onPanel,
    onDownloadPage,
    onToggleDownloads,
    onCloseDownloads,
    onOpenDownload,
    onReadDownload = () => {},
    onOpenDownloadFolder,
    onCancelDownload = () => {},
    onDismissDownload = () => {},
    onRetryDownload = () => {},
    retryingDownloadIds,
    onClearDownloadHistory = () => {},
    clearingDownloadHistory = false,
    onIdentify,
  }: Props = $props();

  const activeDownloadCount = $derived(
    activeDownloads.filter(
      (item) =>
        item.status === "pending" || item.status === "downloading" || item.status === "retrying",
    ).length,
  );

  function submit(event: Event) {
    event.preventDefault();
    onNavigate(url);
  }
</script>

<header class="chrome">
  <div class="nav-cluster">
    <button
      class="ren-icon-btn"
      aria-label={t("chrome.back")}
      disabled={!canGoBack}
      onclick={onBack}
    >
      <ArrowLeft size={16} />
    </button>
    <button
      class="ren-icon-btn"
      aria-label={t("chrome.forward")}
      disabled={!canGoForward}
      onclick={onForward}
    >
      <ArrowRight size={16} />
    </button>
    <button class="ren-icon-btn" aria-label={t("chrome.reload")} onclick={onReload}>
      <RefreshCw size={16} />
    </button>
  </div>

  <form class="url-form" onsubmit={submit}>
    <input
      class="url-input ren-input"
      bind:value={url}
      placeholder={t("chrome.urlPlaceholder")}
      spellcheck="false"
      autocomplete="off"
    />
  </form>

  <div class="tool-cluster">
    <button
      class="ren-icon-btn mobile-nav-dup"
      class:active={activePanel === "search"}
      aria-label={t("chrome.search")}
      onclick={() => onPanel("search")}
    >
      <Search size={16} />
    </button>
    <button
      class="ren-icon-btn mobile-nav-dup"
      class:active={activePanel === "history"}
      aria-label={t("chrome.history")}
      onclick={() => onPanel("history")}
    >
      <History size={16} />
    </button>
    <button
      class="ren-icon-btn mobile-nav-dup"
      class:active={activePanel === "discovery"}
      aria-label={t("chrome.discovery")}
      onclick={() => onPanel("discovery")}
    >
      <Compass size={16} />
    </button>
    <div class="downloads-anchor mobile-nav-dup">
      <button
        class="ren-icon-btn"
        class:active={downloadsOpen}
        aria-label={t("chrome.downloads")}
        aria-expanded={downloadsOpen}
        onclick={onToggleDownloads}
      >
        <Download size={16} />
        {#if activeDownloadCount > 0}
          <span class="download-badge"
            >{activeDownloadCount > 99 ? "99+" : activeDownloadCount}</span
          >
        {/if}
      </button>
      <DownloadsMenu
        open={downloadsOpen}
        active={activeDownloads}
        {downloads}
        {downloadDir}
        {onDownloadPage}
        onOpenFile={onOpenDownload}
        onReadFile={onReadDownload}
        onOpenFolder={onOpenDownloadFolder}
        onCancelActive={onCancelDownload}
        onDismissActive={onDismissDownload}
        onRetryActive={onRetryDownload}
        retryingIds={retryingDownloadIds}
        onClearHistory={onClearDownloadHistory}
        clearingHistory={clearingDownloadHistory}
        onClose={onCloseDownloads}
      />
    </div>
    {#if canIdentify}
      <button
        class="ren-icon-btn always-show"
        aria-label={t("chrome.identify")}
        title={t("chrome.identifyTitle")}
        disabled={identifying}
        onclick={onIdentify}
      >
        <Fingerprint size={16} />
      </button>
    {/if}
    {#each pluginPanels as panel (panel.pluginId + ":" + panel.id)}
      {@const key = panelKey(panel.pluginId, panel.id)}
      <button
        class="ren-icon-btn mobile-nav-dup"
        class:active={activePanel === key}
        aria-label={pluginLabel(panel.pluginId, panel.title)}
        title={pluginLabel(panel.pluginId, panel.title)}
        onclick={() => onPanel(key)}
      >
        <span class="plugin-icon">
          <PluginLucideIcon name={panel.icon} size={16} />
        </span>
      </button>
    {/each}
    {#if devToolsEnabled}
      <button
        class="ren-icon-btn mobile-nav-dup"
        class:active={activePanel === "devtools"}
        aria-label={t("chrome.devtools")}
        onclick={() => onPanel("devtools")}
      >
        <Terminal size={16} />
      </button>
    {/if}
    <button
      class="ren-icon-btn mobile-nav-dup"
      class:active={activePanel === "settings"}
      aria-label={t("chrome.settings")}
      onclick={() => onPanel("settings")}
    >
      <Settings size={16} />
    </button>
    {#if activePanel !== "browser"}
      <button
        class="ren-icon-btn mobile-nav-dup"
        aria-label={t("chrome.closePanel")}
        onclick={() => onPanel("browser")}
      >
        <X size={16} />
      </button>
    {/if}
  </div>
</header>

<style>
  .chrome {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 0.65rem;
    align-items: center;
    padding: 0.65rem 0.85rem;
    background: var(--ren-chrome-bg);
    border-bottom: 1px solid var(--ren-border);
  }

  .nav-cluster,
  .tool-cluster {
    display: flex;
    gap: 0.2rem;
    align-items: center;
  }

  .downloads-anchor {
    position: relative;
  }

  .download-badge {
    position: absolute;
    bottom: -2px;
    right: -2px;
    min-width: 15px;
    height: 15px;
    padding: 0 3px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    background: var(--ren-accent);
    color: #fff;
    font-size: 0.6rem;
    font-weight: 700;
    line-height: 1;
    border: 1.5px solid var(--ren-chrome-bg);
  }

  .url-form {
    min-width: 0;
  }

  .url-input {
    border-radius: 999px;
    background: var(--ren-surface-muted);
    border-color: var(--ren-border);
  }

  @media (max-width: 768px) {
    .chrome {
      grid-template-columns: 1fr;
    }

    .nav-cluster {
      display: none;
    }

    /* These actions (search, discovery, history, devtools, settings, downloads,
       close panel) are already reachable from the bottom mobile
       nav bar, so hide the duplicates here to keep the URL bar row clean. */
    .mobile-nav-dup {
      display: none;
    }
  }
  .plugin-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 0.72rem;
    font-weight: 700;
    line-height: 1;
  }
</style>
