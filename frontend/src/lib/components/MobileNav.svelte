<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Compass, Download, History, Home, Settings, Terminal } from "@lucide/svelte";

  type Panel = "browser" | "discovery" | "history" | "devtools" | "settings";

  type Props = {
    activePanel: Panel;
    downloadsOpen: boolean;
    onPanel: (panel: Panel) => void;
    onToggleDownloads: () => void;
  };

  let { activePanel, downloadsOpen, onPanel, onToggleDownloads }: Props = $props();
</script>

<nav class="mobile-nav" aria-label="Mobile navigation">
  <button class:active={activePanel === "browser"} onclick={() => onPanel("browser")}>
    <Home size={18} />
    <span>Browse</span>
  </button>
  <button class:active={activePanel === "discovery"} onclick={() => onPanel("discovery")}>
    <Compass size={18} />
    <span>Discover</span>
  </button>
  <button class:active={activePanel === "history"} onclick={() => onPanel("history")}>
    <History size={18} />
    <span>History</span>
  </button>
  <button class:active={downloadsOpen} onclick={onToggleDownloads} aria-label="Downloads">
    <Download size={18} />
    <span>Downloads</span>
  </button>
  <button class:active={activePanel === "devtools"} onclick={() => onPanel("devtools")}>
    <Terminal size={18} />
    <span>Devtools</span>
  </button>
  <button class:active={activePanel === "settings"} onclick={() => onPanel("settings")}>
    <Settings size={18} />
    <span>Settings</span>
  </button>
</nav>

<style>
  .mobile-nav {
    display: none;
    position: sticky;
    bottom: 0;
    grid-template-columns: repeat(6, minmax(0, 1fr));
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

  @media (max-width: 768px) {
    .mobile-nav {
      display: grid;
    }
  }
</style>
