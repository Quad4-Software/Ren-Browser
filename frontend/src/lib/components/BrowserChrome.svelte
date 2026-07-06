<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import {
    ArrowLeft,
    ArrowRight,
    Compass,
    Download,
    Fingerprint,
    History,
    Moon,
    RefreshCw,
    Settings,
    Sun,
    Terminal,
    X,
  } from "@lucide/svelte";
  import DownloadsMenu, { type DownloadRow } from "$lib/components/DownloadsMenu.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import type { ActivePanel, PluginPanelContribution } from "$lib/plugins/api-types.js";
  import { panelKey } from "$lib/plugins/registry.js";

  type Props = {
    url: string;
    canGoBack: boolean;
    canGoForward: boolean;
    activePanel: ActivePanel;
    pluginPanels?: PluginPanelContribution[];
    themeMode: "dark" | "light";
    downloadsOpen: boolean;
    downloads: DownloadRow[];
    downloadDir: string;
    canIdentify: boolean;
    identifying: boolean;
    onNavigate: (url: string) => void;
    onBack: () => void;
    onForward: () => void;
    onReload: () => void;
    onPanel: (panel: ActivePanel) => void;
    onToggleTheme: () => void;
    onDownloadPage: () => void;
    onToggleDownloads: () => void;
    onCloseDownloads: () => void;
    onOpenDownload: (path: string) => void;
    onOpenDownloadFolder: () => void;
    onIdentify: () => void;
  };

  let {
    url = $bindable(""),
    canGoBack,
    canGoForward,
    activePanel,
    pluginPanels = [],
    themeMode,
    downloadsOpen,
    downloads,
    downloadDir,
    canIdentify,
    identifying,
    onNavigate,
    onBack,
    onForward,
    onReload,
    onPanel,
    onToggleTheme,
    onDownloadPage,
    onToggleDownloads,
    onCloseDownloads,
    onOpenDownload,
    onOpenDownloadFolder,
    onIdentify,
  }: Props = $props();

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
    <div class="downloads-anchor mobile-nav-dup">
      <button
        class="ren-icon-btn"
        class:active={downloadsOpen}
        aria-label={t("chrome.downloads")}
        aria-expanded={downloadsOpen}
        onclick={onToggleDownloads}
      >
        <Download size={16} />
      </button>
      <DownloadsMenu
        open={downloadsOpen}
        {downloads}
        {downloadDir}
        {onDownloadPage}
        onOpenFile={onOpenDownload}
        onOpenFolder={onOpenDownloadFolder}
        onClose={onCloseDownloads}
      />
    </div>
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
    {#each pluginPanels as panel (panel.pluginId + ":" + panel.id)}
      {@const key = panelKey(panel.pluginId, panel.id)}
      <button
        class="ren-icon-btn mobile-nav-dup"
        class:active={activePanel === key}
        aria-label={panel.title}
        title={panel.title}
        onclick={() => onPanel(key)}
      >
        <span class="plugin-icon">{panel.title.slice(0, 1)}</span>
      </button>
    {/each}
    <button
      class="ren-icon-btn mobile-nav-dup"
      class:active={activePanel === "devtools"}
      aria-label={t("chrome.devtools")}
      onclick={() => onPanel("devtools")}
    >
      <Terminal size={16} />
    </button>
    <button
      class="ren-icon-btn mobile-nav-dup"
      class:active={activePanel === "settings"}
      aria-label={t("chrome.settings")}
      onclick={() => onPanel("settings")}
    >
      <Settings size={16} />
    </button>
    <button
      class="ren-icon-btn mobile-nav-dup"
      aria-label={t("chrome.toggleTheme")}
      onclick={onToggleTheme}
    >
      {#if themeMode === "dark"}
        <Sun size={16} />
      {:else}
        <Moon size={16} />
      {/if}
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

    /* These actions (history, discovery, devtools, settings, downloads,
       theme, close panel) are already reachable from the bottom mobile
       nav bar, so hide the duplicates here to keep the URL bar row clean. */
    .mobile-nav-dup {
      display: none;
    }
  }
  .plugin-icon {
    font-size: 0.72rem;
    font-weight: 700;
    line-height: 1;
  }
</style>
