<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Download, FolderOpen } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  export type DownloadRow = {
    name: string;
    path: string;
    size: number;
    modifiedAt: number;
  };

  type Props = {
    open: boolean;
    downloads: DownloadRow[];
    downloadDir: string;
    variant?: "dropdown" | "sheet";
    onDownloadPage: () => void;
    onOpenFile: (path: string) => void;
    onOpenFolder: () => void;
    onClose: () => void;
  };

  let {
    open,
    downloads,
    downloadDir,
    variant = "dropdown",
    onDownloadPage,
    onOpenFile,
    onOpenFolder,
    onClose,
  }: Props = $props();

  function formatBytes(bytes: number): string {
    if (!bytes) {
      return "0 B";
    }
    const units = ["B", "KB", "MB", "GB"];
    let value = bytes;
    let unit = 0;
    while (value >= 1024 && unit < units.length - 1) {
      value /= 1024;
      unit++;
    }
    const rounded = value >= 10 || unit === 0 ? Math.round(value) : Math.round(value * 10) / 10;
    return `${rounded} ${units[unit]}`;
  }

  function formatWhen(ts: number): string {
    if (!ts) {
      return "";
    }
    return new Date(ts * 1000).toLocaleString();
  }
</script>

{#if open}
  <button
    type="button"
    class="backdrop"
    class:sheet={variant === "sheet"}
    aria-label={t("downloads.close")}
    onclick={onClose}
  ></button>
  <div
    class="menu"
    class:sheet={variant === "sheet"}
    role="dialog"
    aria-label={t("downloads.title")}
    tabindex="-1"
  >
    <header>
      <h2>{t("downloads.title")}</h2>
      <button type="button" class="page-btn" onclick={onDownloadPage}>
        <Download size={14} />
        <span>{t("downloads.saveCurrentPage")}</span>
      </button>
    </header>

    <div class="list">
      {#if downloads.length === 0}
        <EmptyState
          title={t("downloads.noDownloads")}
          description={t("downloads.noDownloadsDescription")}
        >
          <Download size={22} />
        </EmptyState>
      {:else}
        <ul>
          {#each downloads as item (item.path)}
            <li>
              <button type="button" class="file-row" onclick={() => onOpenFile(item.path)}>
                <span class="name">{item.name}</span>
                <span class="meta">{formatBytes(item.size)} · {formatWhen(item.modifiedAt)}</span>
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <footer>
      <button type="button" class="folder-btn" onclick={onOpenFolder}>
        <FolderOpen size={14} />
        <span class="folder-label">{downloadDir || t("downloads.downloadsFolder")}</span>
      </button>
    </footer>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 900;
    border: none;
    background: transparent;
    cursor: default;
  }

  .backdrop.sheet {
    z-index: 115;
    background: color-mix(in srgb, var(--ren-surface-bg) 35%, transparent);
  }

  .menu {
    position: absolute;
    top: calc(100% + 0.35rem);
    right: 0;
    z-index: 901;
    width: min(22rem, calc(100vw - 1.5rem));
    max-height: min(24rem, calc(100vh - 8rem));
    display: flex;
    flex-direction: column;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    overflow: hidden;
  }

  .menu.sheet {
    position: fixed;
    top: auto;
    left: 0.75rem;
    right: 0.75rem;
    bottom: calc(3.6rem + env(safe-area-inset-bottom));
    width: auto;
    max-height: min(55vh, 28rem);
    z-index: 130;
  }

  header {
    display: grid;
    gap: 0.55rem;
    padding: 0.75rem 0.85rem 0.65rem;
    border-bottom: 1px solid var(--ren-border);
  }

  h2 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
  }

  .page-btn,
  .folder-btn,
  .file-row {
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 10px;
    font: inherit;
    cursor: pointer;
    transition:
      background 0.15s ease,
      border-color 0.15s ease;
  }

  .page-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.45rem;
    padding: 0.5rem 0.75rem;
    font-size: 0.86rem;
  }

  .page-btn:hover,
  .folder-btn:hover,
  .file-row:hover {
    background: var(--ren-tab-hover);
    border-color: var(--ren-border-strong);
  }

  .list {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0.45rem;
    display: grid;
    gap: 0.35rem;
  }

  .file-row {
    width: 100%;
    text-align: left;
    padding: 0.65rem 0.75rem;
    display: grid;
    gap: 0.2rem;
  }

  .name {
    font-weight: 600;
    font-size: 0.88rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta {
    color: var(--ren-muted);
    font-size: 0.78rem;
  }

  footer {
    padding: 0.55rem 0.65rem 0.65rem;
    border-top: 1px solid var(--ren-border);
  }

  .folder-btn {
    width: 100%;
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.5rem 0.65rem;
    font-size: 0.82rem;
  }

  .folder-label {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--ren-muted);
  }
</style>
