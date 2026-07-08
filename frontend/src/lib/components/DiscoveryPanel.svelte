<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Compass, Snail, Star } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { displayName } from "$lib/brand";
  import { t } from "$lib/i18n/i18n.svelte";

  type Node = {
    hash: string;
    name: string;
    hops: number;
    lastSeen: number;
  };

  type Props = {
    nodes: Node[];
    favorites: string[];
    slowMode: boolean;
    onOpen: (url: string) => void;
    onFavorite: (url: string) => void;
    onSlowModeChange: (value: boolean) => void;
  };

  let { nodes, favorites, slowMode, onOpen, onFavorite, onSlowModeChange }: Props = $props();

  let query = $state("");
  let favoritesOnly = $state(false);

  function isFavorite(hash: string): boolean {
    const url = `${hash}:/page/index.mu`;
    return favorites.some((f) => f.startsWith(hash) || f === url);
  }

  const pool = $derived.by(() =>
    favoritesOnly ? nodes.filter((node) => isFavorite(node.hash)) : nodes,
  );

  const filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) {
      return pool;
    }
    return pool.filter((node) => {
      const hay = `${node.name} ${node.hash}`.toLowerCase();
      return hay.includes(q);
    });
  });

  const placeholder = $derived(
    pool.length > 0
      ? t("common.searchCount", { count: pool.length, noun: t("discovery.nodes") })
      : t("common.search", { noun: t("discovery.nodes") }),
  );

  function openNode(node: Node) {
    onOpen(`${node.hash}:/page/index.mu`);
  }

  function formatSeen(ts: number): string {
    if (!ts) {
      return t("common.recently");
    }
    return new Date(ts * 1000).toLocaleString();
  }

  function formatHops(hops: number): string {
    if (hops < 0) {
      return "";
    }
    return hops === 1
      ? t("devtools.hopsCount", { count: hops })
      : t("devtools.hopsCountPlural", { count: hops });
  }

  function formatMeta(node: Node): string {
    return t("common.lastSeen", { when: formatSeen(node.lastSeen) });
  }

  const scanningDescription = $derived(t("discovery.scanning", { app: displayName }));
</script>

<section class="discovery">
  <header>
    <div class="title-row">
      <div>
        <h2>{t("discovery.title")}</h2>
        <p>{t("discovery.subtitle")}</p>
      </div>
      <div class="header-actions">
        <button
          type="button"
          class="ren-icon-btn filter-btn"
          class:active={favoritesOnly}
          aria-label={favoritesOnly
            ? t("discovery.favoritesOnlyOn")
            : t("discovery.favoritesOnlyOff")}
          title={t("discovery.favoritesOnly")}
          onclick={() => (favoritesOnly = !favoritesOnly)}
        >
          <Star size={16} fill={favoritesOnly ? "currentColor" : "none"} />
        </button>
        <button
          type="button"
          class="ren-icon-btn filter-btn"
          class:active={slowMode}
          aria-label={slowMode ? t("discovery.slowModeOn") : t("discovery.slowModeOff")}
          title={t("discovery.slowMode")}
          onclick={() => onSlowModeChange(!slowMode)}
        >
          <Snail size={16} />
        </button>
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

  {#if nodes.length === 0}
    <EmptyState title={t("discovery.noSites")} description={scanningDescription}>
      <Compass size={22} />
    </EmptyState>
  {:else if favoritesOnly && pool.length === 0}
    <EmptyState
      title={t("discovery.noFavorites")}
      description={t("discovery.noFavoritesDescription")}
    >
      <Star size={22} />
    </EmptyState>
  {:else if filtered.length === 0}
    <EmptyState
      title={t("discovery.noMatching")}
      description={t("common.nothingMatches", { query: query.trim() })}
    >
      <Compass size={22} />
    </EmptyState>
  {:else}
    <ul>
      {#each filtered as node (node.hash)}
        <li>
          <button onclick={() => openNode(node)}>
            <span class="row">
              <span class="name">{node.name || t("discovery.unnamedSite")}</span>
              {#if node.hops >= 0}
                <span class="hops-badge">{formatHops(node.hops)}</span>
              {/if}
              <span
                class="fav"
                role="button"
                tabindex="0"
                aria-label={t("discovery.favoriteSite")}
                onclick={(event) => {
                  event.stopPropagation();
                  onFavorite(`${node.hash}:/page/index.mu`);
                }}
                onkeydown={(event) => {
                  if (event.key === "Enter") {
                    event.stopPropagation();
                    onFavorite(`${node.hash}:/page/index.mu`);
                  }
                }}
              >
                <Star size={14} fill={isFavorite(node.hash) ? "currentColor" : "none"} />
              </span>
            </span>
            <span class="meta">{formatMeta(node)}</span>
          </button>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .discovery {
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

  header p {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.88rem;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex-shrink: 0;
    margin-left: auto;
  }

  .filter-btn.active {
    color: var(--ren-accent);
    background: color-mix(in srgb, var(--ren-accent) 14%, var(--ren-chrome-bg));
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

  ul button {
    width: 100%;
    text-align: left;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    padding: 0.8rem 0.95rem;
    cursor: pointer;
    display: grid;
    gap: 0.25rem;
    transition:
      border-color 0.15s ease,
      background 0.15s ease;
  }

  ul button:hover {
    border-color: var(--ren-border-strong);
    background: var(--ren-tab-hover);
  }

  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
  }

  .name {
    flex: 1;
    min-width: 0;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .fav {
    color: var(--ren-accent);
    flex-shrink: 0;
  }

  .hops-badge {
    flex-shrink: 0;
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--ren-muted);
    border: 1px solid var(--ren-border);
    border-radius: 999px;
    padding: 0.1rem 0.45rem;
    white-space: nowrap;
  }

  .meta {
    color: var(--ren-muted);
    font-size: 0.85em;
    word-break: break-all;
  }
</style>
