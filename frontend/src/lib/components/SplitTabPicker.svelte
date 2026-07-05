<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Columns2, X } from "@lucide/svelte";
  import type { Tab } from "$lib/browser/url";

  type Props = {
    tabs: Tab[];
    activeTabId: string;
    onSelect: (id: string) => void;
    onClose: () => void;
  };

  let { tabs, activeTabId, onSelect, onClose }: Props = $props();

  const candidates = $derived(tabs.filter((tab) => tab.id !== activeTabId));
</script>

<div class="split-picker">
  <header>
    <div class="heading">
      <Columns2 size={18} />
      <h2>Split view</h2>
    </div>
    <button
      type="button"
      class="close-btn ren-icon-btn"
      aria-label="Close split view"
      onclick={onClose}
    >
      <X size={16} />
    </button>
  </header>

  <p class="hint">Choose a tab to show beside the active tab.</p>

  {#if candidates.length === 0}
    <p class="empty">Open another tab to split with.</p>
  {:else}
    <ul class="tab-list">
      {#each candidates as tab (tab.id)}
        <li>
          <button type="button" class="tab-choice" onclick={() => onSelect(tab.id)}>
            <span class="title">{tab.title}</span>
            <span class="url">{tab.url || "New tab"}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .split-picker {
    height: 100%;
    display: grid;
    grid-template-rows: auto auto 1fr;
    gap: 0.75rem;
    padding: 1rem 1.1rem;
    background: var(--ren-surface-bg);
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .heading {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--ren-fg);
  }

  h2 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
  }

  .hint,
  .empty {
    margin: 0;
    font-size: 0.88rem;
    color: var(--ren-muted);
  }

  .tab-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
    align-content: start;
    overflow: auto;
  }

  .tab-choice {
    width: 100%;
    text-align: left;
    border: 1px solid var(--ren-border);
    border-radius: 10px;
    background: var(--ren-chrome-bg);
    color: var(--ren-fg);
    padding: 0.7rem 0.8rem;
    display: grid;
    gap: 0.2rem;
    font: inherit;
    cursor: pointer;
    transition:
      background 0.15s ease,
      border-color 0.15s ease;
  }

  .tab-choice:hover {
    background: var(--ren-tab-hover);
    border-color: var(--ren-border-strong);
  }

  .title {
    font-size: 0.9rem;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .url {
    font-size: 0.78rem;
    color: var(--ren-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
