<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Download, FolderOpen, RotateCcw, X } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    formatBytes,
    formatEta,
    formatSpeed,
    type DownloadProgressView,
  } from "$lib/browser/download-progress";

  export type DownloadRow = {
    name: string;
    path: string;
    size: number;
    modifiedAt: number;
  };

  type Props = {
    open: boolean;
    active?: DownloadProgressView[];
    downloads: DownloadRow[];
    downloadDir: string;
    variant?: "dropdown" | "sheet";
    onDownloadPage: () => void;
    onOpenFile: (path: string) => void;
    onOpenFolder: () => void;
    onCancelActive?: (id: string) => void;
    onDismissActive?: (id: string) => void;
    onRetryActive?: (id: string) => void;
    onClose: () => void;
  };

  let {
    open,
    active = [],
    downloads,
    downloadDir,
    variant = "dropdown",
    onDownloadPage,
    onOpenFile,
    onOpenFolder,
    onCancelActive = () => {},
    onDismissActive = () => {},
    onRetryActive = () => {},
    onClose,
  }: Props = $props();

  function formatWhen(ts: number): string {
    if (!ts) {
      return "";
    }
    return new Date(ts * 1000).toLocaleString();
  }

  function progressPercent(item: DownloadProgressView): number | null {
    if (item.total <= 0) {
      return null;
    }
    return Math.max(0, Math.min(100, (item.received / item.total) * 100));
  }

  function metaLine(item: DownloadProgressView): string {
    const parts: string[] = [];
    parts.push(
      item.total > 0
        ? `${formatBytes(item.received)} / ${formatBytes(item.total)}`
        : formatBytes(item.received),
    );
    const speed = formatSpeed(item.speedBps);
    if (speed) {
      parts.push(speed);
    }
    const eta = formatEta(item.etaSeconds);
    if (eta) {
      parts.push(t("downloads.etaLeft", { eta }));
    }
    return parts.join(" \u00b7 ");
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
      {#if active.length > 0}
        <ul class="active-list">
          {#each active as item (item.id)}
            {@const percent = progressPercent(item)}
            <li class="active-row" class:error={item.status === "failed"}>
              <div class="active-head">
                <span class="name" class:pending={item.status === "pending"}>{item.name}</span>
                {#if item.status === "pending" || item.status === "downloading"}
                  <button
                    type="button"
                    class="icon-btn"
                    aria-label={t("downloads.cancel")}
                    onclick={() => onCancelActive(item.id)}
                  >
                    <X size={13} />
                  </button>
                {:else if item.status === "failed" || item.status === "interrupted"}
                  <button
                    type="button"
                    class="icon-btn"
                    aria-label={t("downloads.retry")}
                    onclick={() => onRetryActive(item.id)}
                  >
                    <RotateCcw size={13} />
                  </button>
                  <button
                    type="button"
                    class="icon-btn"
                    aria-label={t("downloads.dismiss")}
                    onclick={() => onDismissActive(item.id)}
                  >
                    <X size={13} />
                  </button>
                {:else}
                  <button
                    type="button"
                    class="icon-btn"
                    aria-label={t("downloads.dismiss")}
                    onclick={() => onDismissActive(item.id)}
                  >
                    <X size={13} />
                  </button>
                {/if}
              </div>
              {#if item.status === "failed"}
                <span class="error-text">{item.error || t("downloads.downloadFailed")}</span>
              {:else if item.status === "interrupted"}
                <span class="meta">{t("downloads.interrupted")}</span>
              {:else if item.status === "retrying"}
                <span class="meta"
                  >{t("downloads.retrying", {
                    attempt: item.attempt ?? 1,
                    max: item.maxAttempts ?? 1,
                  })}</span
                >
              {:else if item.status === "canceled"}
                <span class="meta">{t("downloads.canceled")}</span>
              {:else if item.status === "completed"}
                <span class="meta success">{t("downloads.fileSaved")}</span>
              {:else if item.status === "pending"}
                <span class="meta">{t("downloads.starting")}</span>
              {:else}
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    class:indeterminate={percent === null}
                    style={percent !== null ? `width:${percent}%` : ""}
                  ></div>
                </div>
                <span class="meta">{metaLine(item)}</span>
              {/if}
            </li>
          {/each}
        </ul>
      {/if}

      {#if downloads.length === 0 && active.length === 0}
        <EmptyState
          title={t("downloads.noDownloads")}
          description={t("downloads.noDownloadsDescription")}
        >
          <Download size={22} />
        </EmptyState>
      {:else if downloads.length > 0}
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
    width: 100%;
    min-width: 0;
    padding: 0.5rem 0.75rem;
    font-size: 0.86rem;
  }

  .page-btn span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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

  .active-list {
    border-bottom: 1px solid var(--ren-border);
    padding-bottom: 0.55rem;
    margin-bottom: 0.1rem;
  }

  .active-row {
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    border-radius: 10px;
    padding: 0.55rem 0.65rem;
    display: grid;
    gap: 0.3rem;
  }

  .active-row.error {
    border-color: color-mix(in srgb, var(--ren-danger, #e5484d) 55%, var(--ren-border));
  }

  .active-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .active-head .name {
    flex: 1;
    min-width: 0;
    font-weight: 600;
    font-size: 0.86rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .active-head .name.pending {
    color: var(--ren-muted);
  }

  .icon-btn {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 20px;
    height: 20px;
    border: none;
    border-radius: 999px;
    background: transparent;
    color: var(--ren-muted);
    cursor: pointer;
  }

  .icon-btn:hover {
    background: var(--ren-tab-hover);
    color: var(--ren-fg);
  }

  .progress-track {
    height: 5px;
    border-radius: 999px;
    background: var(--ren-surface-muted);
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    border-radius: 999px;
    background: var(--ren-accent);
    transition: width 0.25s ease;
  }

  .progress-fill.indeterminate {
    width: 40%;
    animation: indeterminate 1.1s ease-in-out infinite;
  }

  @keyframes indeterminate {
    0% {
      transform: translateX(-100%);
    }
    100% {
      transform: translateX(250%);
    }
  }

  .error-text {
    color: var(--ren-danger, #e5484d);
    font-size: 0.78rem;
    overflow-wrap: anywhere;
    word-break: break-word;
  }

  .meta.success {
    color: var(--ren-accent);
  }

  .file-row {
    width: 100%;
    min-width: 0;
    text-align: left;
    padding: 0.65rem 0.75rem;
    display: grid;
    gap: 0.2rem;
  }

  .name {
    min-width: 0;
    font-weight: 600;
    font-size: 0.88rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta {
    min-width: 0;
    color: var(--ren-muted);
    font-size: 0.78rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
