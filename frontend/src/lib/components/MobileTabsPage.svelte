<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Plus, X } from "@lucide/svelte";
  import { type Tab } from "$lib/browser/url";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    tabs: Tab[];
    activeTabId: string;
    atTabLimit: boolean;
    onSelect: (id: string) => void;
    onClose: (id: string) => void;
    onNew: () => void;
    onDismiss: () => void;
  };

  let { tabs, activeTabId, atTabLimit, onSelect, onClose, onNew, onDismiss }: Props = $props();

  function previewLabel(tab: Tab): string {
    const title = tab.title.trim();
    if (title) {
      return title;
    }
    const url = tab.url.trim();
    if (url) {
      return url;
    }
    return t("tab.new");
  }

  function previewSubtitle(tab: Tab): string {
    const url = tab.url.trim();
    if (url && url !== tab.title.trim()) {
      return url;
    }
    if (tab.page?.error) {
      return tab.page.error;
    }
    return tab.page?.contentType || "";
  }
</script>

<section class="tabs-page" aria-label={t("mobileTabs.title")}>
  <header class="tabs-header">
    <h2>{t("mobileTabs.title")}</h2>
    <div class="header-actions">
      <button
        type="button"
        class="new-btn"
        disabled={atTabLimit}
        onclick={onNew}
        aria-label={t("tab.newTab")}
      >
        <Plus size={16} />
        <span>{t("tab.newTab")}</span>
      </button>
      <button
        type="button"
        class="ren-icon-btn"
        aria-label={t("chrome.closePanel")}
        onclick={onDismiss}
      >
        <X size={18} />
      </button>
    </div>
  </header>

  <div class="tabs-grid">
    {#each tabs as tab (tab.id)}
      <article class="tab-card" class:active={tab.id === activeTabId}>
        <button type="button" class="card-main" onclick={() => onSelect(tab.id)}>
          <div
            class="preview"
            style:background={tab.page?.pageBg || "var(--ren-surface-muted)"}
            style:color={tab.page?.pageFg || "var(--ren-fg)"}
          >
            <span class="preview-title">{previewLabel(tab)}</span>
            {#if tab.page?.error}
              <span class="preview-meta error">{tab.page.error}</span>
            {:else if tab.page?.html}
              <span class="preview-meta">{tab.page.contentType || t("mobileTabs.pageLoaded")}</span>
            {/if}
          </div>
          <div class="card-footer">
            <span class="card-title">{previewLabel(tab)}</span>
            <span class="card-url">{previewSubtitle(tab)}</span>
          </div>
        </button>
        {#if tabs.length > 1}
          <button
            type="button"
            class="close-btn ren-icon-btn"
            aria-label={t("tab.close")}
            onclick={() => onClose(tab.id)}
          >
            <X size={14} />
          </button>
        {/if}
      </article>
    {/each}
  </div>
</section>

<style>
  .tabs-page {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--ren-content-bg);
    overflow: hidden;
  }

  .tabs-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.85rem 1rem;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
  }

  h2 {
    margin: 0;
    font-size: 1.05rem;
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .new-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 10px;
    padding: 0.4rem 0.65rem;
    font: inherit;
    font-size: 0.82rem;
    cursor: pointer;
  }

  .new-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .tabs-grid {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 0.85rem;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(10.5rem, 1fr));
    gap: 0.75rem;
    align-content: start;
  }

  .tab-card {
    position: relative;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-surface-raised);
    overflow: hidden;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease;
  }

  .tab-card.active {
    border-color: var(--ren-accent);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--ren-accent) 35%, transparent);
  }

  .card-main {
    width: 100%;
    border: none;
    background: transparent;
    color: inherit;
    padding: 0;
    cursor: pointer;
    text-align: left;
    font: inherit;
  }

  .preview {
    min-height: 6.5rem;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    border-bottom: 1px solid var(--ren-border);
  }

  .preview-title {
    font-weight: 600;
    font-size: 0.88rem;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .preview-meta {
    font-size: 0.75rem;
    opacity: 0.85;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .preview-meta.error {
    color: var(--ren-danger);
    opacity: 1;
  }

  .card-footer {
    padding: 0.55rem 0.65rem;
    display: grid;
    gap: 0.15rem;
  }

  .card-title {
    font-weight: 600;
    font-size: 0.82rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .card-url {
    color: var(--ren-muted);
    font-size: 0.72rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .close-btn {
    position: absolute;
    top: 0.35rem;
    right: 0.35rem;
    background: color-mix(in srgb, var(--ren-chrome-bg) 88%, transparent);
  }
</style>
