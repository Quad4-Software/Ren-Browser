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

  type Props = {
    url: string;
    canGoBack: boolean;
    canGoForward: boolean;
    activePanel: "browser" | "discovery" | "history" | "devtools" | "settings";
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
    onPanel: (panel: Props["activePanel"]) => void;
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
    <button class="ren-icon-btn" aria-label="Back" disabled={!canGoBack} onclick={onBack}>
      <ArrowLeft size={16} />
    </button>
    <button class="ren-icon-btn" aria-label="Forward" disabled={!canGoForward} onclick={onForward}>
      <ArrowRight size={16} />
    </button>
    <button class="ren-icon-btn" aria-label="Reload" onclick={onReload}>
      <RefreshCw size={16} />
    </button>
  </div>

  <form class="url-form" onsubmit={submit}>
    <input
      class="url-input ren-input"
      bind:value={url}
      placeholder="nodehash:/page/index.mu"
      spellcheck="false"
      autocomplete="off"
    />
  </form>

  <div class="tool-cluster">
    {#if canIdentify}
      <button
        class="ren-icon-btn"
        aria-label="Identify to node"
        title="Identify yourself to this node"
        disabled={identifying}
        onclick={onIdentify}
      >
        <Fingerprint size={16} />
      </button>
    {/if}
    <div class="downloads-anchor">
      <button
        class="ren-icon-btn"
        class:active={downloadsOpen}
        aria-label="Downloads"
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
      class="ren-icon-btn"
      class:active={activePanel === "history"}
      aria-label="History"
      onclick={() => onPanel("history")}
    >
      <History size={16} />
    </button>
    <button
      class="ren-icon-btn"
      class:active={activePanel === "discovery"}
      aria-label="Discovery"
      onclick={() => onPanel("discovery")}
    >
      <Compass size={16} />
    </button>
    <button
      class="ren-icon-btn"
      class:active={activePanel === "devtools"}
      aria-label="Developer tools"
      onclick={() => onPanel("devtools")}
    >
      <Terminal size={16} />
    </button>
    <button
      class="ren-icon-btn"
      class:active={activePanel === "settings"}
      aria-label="Settings"
      onclick={() => onPanel("settings")}
    >
      <Settings size={16} />
    </button>
    <button class="ren-icon-btn" aria-label="Toggle theme" onclick={onToggleTheme}>
      {#if themeMode === "dark"}
        <Sun size={16} />
      {:else}
        <Moon size={16} />
      {/if}
    </button>
    {#if activePanel !== "browser"}
      <button class="ren-icon-btn" aria-label="Close panel" onclick={() => onPanel("browser")}>
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
  }
</style>
