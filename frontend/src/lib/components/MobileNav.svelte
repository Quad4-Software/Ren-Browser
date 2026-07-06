<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Compass, Download, History, Home, Settings, Terminal } from "@lucide/svelte";
  import type { ActivePanel, PluginPanelContribution } from "$lib/plugins/api-types.js";
  import { panelKey } from "$lib/plugins/registry.js";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    activePanel: ActivePanel;
    pluginPanels?: PluginPanelContribution[];
    mobileDevTools: boolean;
    downloadsOpen: boolean;
    onPanel: (panel: ActivePanel) => void;
    onToggleDownloads: () => void;
  };

  let {
    activePanel,
    pluginPanels = [],
    mobileDevTools,
    downloadsOpen,
    onPanel,
    onToggleDownloads,
  }: Props = $props();
</script>

<nav class="mobile-nav" aria-label={t("mobileNav.label")}>
  <button class:active={activePanel === "browser"} onclick={() => onPanel("browser")}>
    <Home size={18} />
    <span>{t("mobileNav.browse")}</span>
  </button>
  <button class:active={activePanel === "discovery"} onclick={() => onPanel("discovery")}>
    <Compass size={18} />
    <span>{t("mobileNav.discover")}</span>
  </button>
  <button class:active={activePanel === "history"} onclick={() => onPanel("history")}>
    <History size={18} />
    <span>{t("mobileNav.history")}</span>
  </button>
  <button
    class:active={downloadsOpen}
    onclick={onToggleDownloads}
    aria-label={t("mobileNav.downloads")}
  >
    <Download size={18} />
    <span>{t("mobileNav.downloads")}</span>
  </button>
  {#each pluginPanels as panel (panel.pluginId + ":" + panel.id)}
    {@const key = panelKey(panel.pluginId, panel.id)}
    <button class:active={activePanel === key} onclick={() => onPanel(key)}>
      <span class="plugin-dot">{panel.title.slice(0, 1)}</span>
      <span>{panel.title}</span>
    </button>
  {/each}
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
    grid-auto-flow: column;
    grid-auto-columns: minmax(3.5rem, 1fr);
    overflow-x: auto;
    gap: 0.1rem;
    padding: 0.35rem 0.25rem calc(0.35rem + env(safe-area-inset-bottom));
    background: var(--ren-chrome-bg);
    border-top: 1px solid var(--ren-border);
  }

  button {
    display: grid;
    justify-items: center;
    gap: 0.12rem;
    border: none;
    background: transparent;
    color: var(--ren-muted);
    font: inherit;
    font-size: 0.62rem;
    padding: 0.35rem 0.1rem;
    border-radius: 10px;
    transition:
      background 0.15s ease,
      color 0.15s ease;
  }

  button:hover {
    background: var(--ren-tab-hover);
    color: var(--ren-fg-secondary);
  }

  button.active {
    color: #fff;
    background: var(--ren-accent);
  }

  .plugin-dot {
    width: 18px;
    height: 18px;
    display: grid;
    place-items: center;
    border-radius: 999px;
    background: var(--ren-surface-raised);
    font-size: 0.7rem;
    font-weight: 700;
  }

  button.active .plugin-dot {
    background: rgba(255, 255, 255, 0.2);
  }

  @media (max-width: 768px) {
    .mobile-nav {
      display: grid;
    }
  }
</style>
