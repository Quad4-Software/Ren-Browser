<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Compass, Download, History, Home, Settings, Terminal } from "@lucide/svelte";
  import type { ActivePanel } from "$lib/plugins/api-types.js";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    activePanel: ActivePanel;
    mobileDevTools: boolean;
    downloadsOpen: boolean;
    activeDownloadCount?: number;
    onPanel: (panel: ActivePanel) => void;
    onToggleDownloads: () => void;
  };

  let {
    activePanel,
    mobileDevTools,
    downloadsOpen,
    activeDownloadCount = 0,
    onPanel,
    onToggleDownloads,
  }: Props = $props();
</script>

<nav class="mobile-nav" aria-label={t("mobileNav.label")}>
  <button class:active={activePanel === "browser"} onclick={() => onPanel("browser")}>
    <Home size={18} />
    <span>{t("mobileNav.browse")}</span>
  </button>
  <button class:active={activePanel === "history"} onclick={() => onPanel("history")}>
    <History size={18} />
    <span>{t("mobileNav.history")}</span>
  </button>
  <button class:active={activePanel === "discovery"} onclick={() => onPanel("discovery")}>
    <Compass size={18} />
    <span>{t("mobileNav.discover")}</span>
  </button>
  <button
    class="downloads-btn"
    class:active={downloadsOpen}
    onclick={onToggleDownloads}
    aria-label={t("mobileNav.downloads")}
  >
    <span class="icon-wrap">
      <Download size={18} />
      {#if activeDownloadCount > 0}
        <span class="download-badge">{activeDownloadCount > 99 ? "99+" : activeDownloadCount}</span>
      {/if}
    </span>
    <span>{t("mobileNav.downloads")}</span>
  </button>
  {#if mobileDevTools}
    <button class:active={activePanel === "devtools"} onclick={() => onPanel("devtools")}>
      <Terminal size={18} />
      <span>{t("mobileNav.devtools")}</span>
    </button>
  {/if}
  <button class:active={activePanel === "settings"} onclick={() => onPanel("settings")}>
    <Settings size={18} />
    <span>{t("mobileNav.settings")}</span>
  </button>
</nav>

<style>
  .mobile-nav {
    display: none;
    position: sticky;
    bottom: 0;
    z-index: 100;
    width: 100%;
    max-width: 100%;
    grid-auto-flow: column;
    grid-auto-columns: minmax(3.5rem, 1fr);
    overflow-x: auto;
    overscroll-behavior-x: contain;
    -webkit-overflow-scrolling: touch;
    gap: 0.1rem;
    padding: 0.35rem 0.25rem calc(0.35rem + env(safe-area-inset-bottom));
    background: var(--ren-chrome-bg);
    border-top: 1px solid var(--ren-border);
  }

  button {
    display: grid;
    justify-items: center;
    gap: 0.12rem;
    min-width: 0;
    border: none;
    background: transparent;
    color: var(--ren-muted);
    font: inherit;
    font-size: 0.62rem;
    padding: 0.35rem 0.2rem;
    border-radius: 10px;
    overflow: visible;
    transition:
      background 0.15s ease,
      color 0.15s ease;
  }

  button > span:not(.icon-wrap) {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  button:hover {
    background: var(--ren-tab-hover);
    color: var(--ren-fg-secondary);
  }

  button.active {
    color: #fff;
    background: var(--ren-accent);
  }

  .icon-wrap {
    position: relative;
    display: inline-flex;
    overflow: visible;
  }

  .download-badge {
    position: absolute;
    bottom: -3px;
    right: -6px;
    min-width: 14px;
    height: 14px;
    padding: 0 3px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    background: var(--ren-accent);
    color: #fff;
    font-size: 0.56rem;
    font-weight: 700;
    line-height: 1;
    border: 1.5px solid var(--ren-chrome-bg);
  }

  :global(.app-shell.mobile-ui) .mobile-nav {
    display: grid;
  }
</style>
