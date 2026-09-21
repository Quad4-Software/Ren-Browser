<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Check, Compass, ListFilter, Snail, Star } from "@lucide/svelte";
  import { DropdownMenu } from "bits-ui";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { displayName } from "$lib/brand";
  import { t } from "$lib/i18n/i18n.svelte";

  type Node = {
    hash: string;
    name: string;
    hops: number;
    lastSeen: number;
    announces: number;
  };

  type Props = {
    nodes: Node[];
    favorites: string[];
    slowMode: boolean;
    onOpen: (url: string) => void;
    onFavorite: (url: string) => void;
    onSlowModeChange: (value: boolean) => void;
  };

  type SortMode = "newest" | "announces" | "name";

  let { nodes, favorites, slowMode, onOpen, onFavorite, onSlowModeChange }: Props = $props();

  let query = $state("");
  let favoritesOnly = $state(false);
  let sortMode = $state<SortMode>("newest");

  const favoriteKeys = $derived.by(() => {
    const keys: Record<string, true> = {};
    for (const fav of favorites) {
      keys[fav] = true;
      const hash = fav.split(":")[0];
      if (hash) {
        keys[hash] = true;
      }
    }
    return keys;
  });

  function isFavorite(hash: string): boolean {
    return !!favoriteKeys[hash] || !!favoriteKeys[`${hash}:/page/index.mu`];
  }

  const pool = $derived.by(() =>
    favoritesOnly ? nodes.filter((node) => isFavorite(node.hash)) : nodes,
  );

  const filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    const matched = q
      ? pool.filter((node) => `${node.name} ${node.hash}`.toLowerCase().includes(q))
      : pool.slice();
    if (sortMode === "announces") {
      matched.sort((a, b) => b.announces - a.announces || b.lastSeen - a.lastSeen);
    } else if (sortMode === "name") {
      matched.sort(
        (a, b) => (a.name || a.hash).localeCompare(b.name || b.hash) || b.lastSeen - a.lastSeen,
      );
    } else {
      matched.sort((a, b) => b.lastSeen - a.lastSeen);
    }
    return matched;
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
    const seen = t("common.lastSeen", { when: formatSeen(node.lastSeen) });
    if (node.announces > 0) {
      const count =
        node.announces === 1
          ? t("discovery.announcesCount", { count: node.announces })
          : t("discovery.announcesCountPlural", { count: node.announces });
      return `${seen} · ${count}`;
    }
    return seen;
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
        <DropdownMenu.Root>
          <DropdownMenu.Trigger
            class={`ren-icon-btn filter-btn${sortMode !== "newest" ? " active" : ""}`}
            aria-label={t("discovery.sort")}
            title={t("discovery.sort")}
          >
            <ListFilter size={16} />
          </DropdownMenu.Trigger>
          <DropdownMenu.Portal>
            <DropdownMenu.Content
              class="discovery-sort-menu"
              align="end"
              sideOffset={5}
              collisionPadding={8}
            >
              <DropdownMenu.RadioGroup bind:value={sortMode}>
                <DropdownMenu.RadioItem value="newest" textValue={t("discovery.sortNewest")}>
                  {#snippet children({ checked })}
                    <span class="sort-check"
                      >{#if checked}<Check size={14} />{/if}</span
                    >
                    {t("discovery.sortNewest")}
                  {/snippet}
                </DropdownMenu.RadioItem>
                <DropdownMenu.RadioItem value="announces" textValue={t("discovery.sortAnnounces")}>
                  {#snippet children({ checked })}
                    <span class="sort-check"
                      >{#if checked}<Check size={14} />{/if}</span
                    >
                    {t("discovery.sortAnnounces")}
                  {/snippet}
                </DropdownMenu.RadioItem>
                <DropdownMenu.RadioItem value="name" textValue={t("discovery.sortName")}>
                  {#snippet children({ checked })}
                    <span class="sort-check"
                      >{#if checked}<Check size={14} />{/if}</span
                    >
                    {t("discovery.sortName")}
                  {/snippet}
                </DropdownMenu.RadioItem>
              </DropdownMenu.RadioGroup>
            </DropdownMenu.Content>
          </DropdownMenu.Portal>
        </DropdownMenu.Root>
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
          <div class="node-entry">
            <button type="button" class="node-open" onclick={() => openNode(node)}>
              <span class="row">
                <span class="name">{node.name || t("discovery.unnamedSite")}</span>
                {#if node.hops >= 0}
                  <span class="hops-badge">{formatHops(node.hops)}</span>
                {/if}
              </span>
              <span class="meta">{formatMeta(node)}</span>
            </button>
            <button
              type="button"
              class="fav"
              class:faved={isFavorite(node.hash)}
              aria-label={t("discovery.favoriteSite")}
              aria-pressed={isFavorite(node.hash)}
              onclick={() => onFavorite(`${node.hash}:/page/index.mu`)}
            >
              <Star size={14} fill={isFavorite(node.hash) ? "currentColor" : "none"} />
            </button>
          </div>
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

  .node-entry {
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

  .node-entry:hover {
    border-color: var(--ren-border-strong);
    background: var(--ren-tab-hover);
  }

  .node-open {
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
    color: var(--ren-muted);
    flex-shrink: 0;
    align-self: center;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    margin-right: 0.45rem;
    border: none;
    border-radius: 8px;
    padding: 0;
    background: transparent;
    cursor: pointer;
  }

  .fav.faved {
    color: var(--ren-accent);
  }

  .fav:hover {
    color: var(--ren-accent);
    background: var(--ren-chrome-bg);
  }

  .fav:focus-visible,
  .node-open:focus-visible {
    outline: 2px solid var(--ren-accent);
    outline-offset: 2px;
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

  :global(.discovery-sort-menu) {
    z-index: 1100;
    min-width: 10rem;
    padding: 0.35rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    display: grid;
    gap: 0.15rem;
  }

  :global(.discovery-sort-menu [role="menuitemradio"]) {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    text-align: left;
    border: none;
    background: transparent;
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.45rem 0.65rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
  }

  :global(.discovery-sort-menu [data-highlighted]) {
    background: var(--ren-tab-hover);
  }

  :global(.discovery-sort-menu .sort-check) {
    width: 1rem;
    display: inline-flex;
    color: var(--ren-accent);
    flex-shrink: 0;
  }
</style>
