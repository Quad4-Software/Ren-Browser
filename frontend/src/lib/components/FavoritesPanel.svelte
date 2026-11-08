<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Star } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import type { Node } from "$lib/app/types";
  import { tabTitleFromURL } from "$lib/browser/url";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    favorites: string[];
    nodes: Node[];
    onOpen: (url: string) => void;
    onRemove: (url: string) => void;
  };

  let { favorites, nodes, onOpen, onRemove }: Props = $props();

  let query = $state("");

  // The backend returns favorites oldest-first (added_at ASC); show newest first.
  const ordered = $derived(favorites.slice().reverse());

  const filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) {
      return ordered;
    }
    return ordered.filter((favUrl) => {
      const hay = `${tabTitleFromURL(favUrl, nodes)} ${favUrl}`.toLowerCase();
      return hay.includes(q);
    });
  });

  const placeholder = $derived(
    favorites.length > 0
      ? t("common.searchCount", { count: favorites.length, noun: t("discovery.sites") })
      : t("common.search", { noun: t("discovery.sites") }),
  );
</script>

<section class="favorites-panel">
  <header>
    <div class="title-row">
      <div>
        <h2>{t("favorites.title")}</h2>
        <p class="subtitle">{t("favorites.subtitle")}</p>
      </div>
    </div>
    <input
      class="search ren-input"
      type="search"
      bind:value={query}
      {placeholder}
      spellcheck="false"
      autocomplete="off"
    />
  </header>

  {#if favorites.length === 0}
    <EmptyState title={t("favorites.empty")} description={t("favorites.emptyDescription")}>
      <Star size={22} />
    </EmptyState>
  {:else if filtered.length === 0}
    <EmptyState
      title={t("discovery.noMatching")}
      description={t("common.nothingMatches", { query: query.trim() })}
    >
      <Star size={22} />
    </EmptyState>
  {:else}
    <ul>
      {#each filtered as favUrl (favUrl)}
        <li>
          <div class="fav-entry">
            <button type="button" class="fav-open" onclick={() => onOpen(favUrl)}>
              <span class="name">{tabTitleFromURL(favUrl, nodes)}</span>
              <span class="meta">{favUrl}</span>
            </button>
            <div class="fav-actions">
              <button
                type="button"
                class="fav-remove"
                aria-label={t("favorites.remove")}
                title={t("favorites.remove")}
                onclick={() => onRemove(favUrl)}
              >
                <Star size={14} fill="currentColor" />
              </button>
            </div>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .favorites-panel {
    height: 100%;
    overflow: auto;
    overflow-x: hidden;
    padding: 1rem;
    background: var(--ren-content-bg);
  }

  .title-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.65rem;
    margin-bottom: 0.75rem;
  }

  header h2 {
    margin: 0 0 0.25rem;
    font-size: 1.05rem;
    font-weight: 600;
  }

  .subtitle {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.88rem;
  }

  .search {
    margin-bottom: 1rem;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
  }

  .fav-entry {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: stretch;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface-raised);
    transition:
      border-color 0.15s ease,
      background 0.15s ease;
  }

  .fav-entry:hover {
    border-color: var(--ren-border-strong);
    background: var(--ren-tab-hover);
  }

  .fav-open {
    width: 100%;
    min-width: 0;
    text-align: left;
    border: none;
    background: transparent;
    color: var(--ren-fg);
    padding: 0.8rem 0.95rem;
    cursor: pointer;
    display: grid;
    gap: 0.25rem;
  }

  .fav-actions {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 0.45rem 0 0;
    flex-shrink: 0;
  }

  .fav-remove {
    color: var(--ren-accent);
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border: none;
    border-radius: 8px;
    padding: 0;
    background: transparent;
    cursor: pointer;
  }

  .fav-remove:hover {
    background: var(--ren-chrome-bg);
  }

  .fav-remove:focus-visible,
  .fav-open:focus-visible {
    outline: 2px solid var(--ren-accent);
    outline-offset: 2px;
  }

  .name {
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta {
    color: var(--ren-muted);
    font-size: 0.85em;
    word-break: break-all;
  }
</style>
