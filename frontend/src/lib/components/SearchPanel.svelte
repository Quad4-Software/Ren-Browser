<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Clock, Compass, Search, Star } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import type { HistoryEntry, Node } from "$lib/app/types";
  import { t } from "$lib/i18n/i18n.svelte";

  type SearchKind = "history" | "site" | "favorite";

  type SearchResult = {
    key: string;
    kind: SearchKind;
    title: string;
    subtitle: string;
    url: string;
    hops?: number;
  };

  type SearchGroup = {
    kind: SearchKind;
    label: string;
    results: SearchResult[];
  };

  type Props = {
    history: HistoryEntry[];
    nodes: Node[];
    favorites: string[];
    onOpen: (url: string, highlight?: string) => void;
  };

  let { history, nodes, favorites, onOpen }: Props = $props();

  let query = $state("");

  function historyLabel(entry: HistoryEntry): string {
    if (entry.title) {
      return entry.title;
    }
    const sep = entry.url.indexOf(":/");
    if (sep >= 0) {
      return entry.url.slice(sep + 2) || entry.url;
    }
    return entry.url;
  }

  function favoriteLabel(url: string): string {
    const sep = url.indexOf(":/");
    if (sep >= 0) {
      return url.slice(sep + 2) || url;
    }
    return url;
  }

  function nodeUrl(node: Node): string {
    return `${node.hash}:/page/index.mu`;
  }

  function matchesQuery(hay: string, q: string): boolean {
    return hay.toLowerCase().includes(q);
  }

  function pushResult(
    out: SearchResult[],
    seen: Set<string>,
    result: Omit<SearchResult, "key">,
  ): void {
    const normalized = result.url.trim().toLowerCase();
    if (!normalized || seen.has(normalized)) {
      return;
    }
    seen.add(normalized);
    out.push({ ...result, key: `${result.kind}:${normalized}` });
  }

  const results = $derived.by(() => {
    const q = query.trim().toLowerCase();
    const out: SearchResult[] = [];
    const seen = new Set<string>();

    if (!q) {
      for (const entry of history.slice(0, 15)) {
        pushResult(out, seen, {
          kind: "history",
          title: historyLabel(entry),
          subtitle: entry.url,
          url: entry.url,
        });
      }
      return out;
    }

    for (const entry of history) {
      const hay = `${historyLabel(entry)} ${entry.url} ${entry.nodeHash}`;
      if (!matchesQuery(hay, q)) {
        continue;
      }
      pushResult(out, seen, {
        kind: "history",
        title: historyLabel(entry),
        subtitle: entry.url,
        url: entry.url,
      });
    }

    for (const node of nodes) {
      const url = nodeUrl(node);
      const hay = `${node.name} ${node.hash} ${url}`;
      if (!matchesQuery(hay, q)) {
        continue;
      }
      pushResult(out, seen, {
        kind: "site",
        title: node.name || t("discovery.unnamedSite"),
        subtitle: node.hash,
        url,
        hops: node.hops,
      });
    }

    for (const fav of favorites) {
      const hay = `${favoriteLabel(fav)} ${fav}`;
      if (!matchesQuery(hay, q)) {
        continue;
      }
      pushResult(out, seen, {
        kind: "favorite",
        title: favoriteLabel(fav),
        subtitle: fav,
        url: fav,
      });
    }

    return out;
  });

  const groups = $derived.by(() => {
    const order: SearchKind[] = ["history", "site", "favorite"];
    const labels: Record<SearchKind, string> = {
      history: t("search.history"),
      site: t("search.nodes"),
      favorite: t("search.favorites"),
    };
    const grouped: SearchGroup[] = [];

    for (const kind of order) {
      const items = results.filter((result) => result.kind === kind);
      if (items.length > 0) {
        grouped.push({ kind, label: labels[kind], results: items });
      }
    }

    return grouped;
  });

  const totalSources = $derived(history.length + nodes.length + favorites.length);

  const placeholder = $derived(
    totalSources > 0
      ? t("common.searchCount", { count: totalSources, noun: t("search.entries") })
      : t("common.search", { noun: t("search.entries") }),
  );

  const hasIndexedData = $derived(history.length > 0 || nodes.length > 0 || favorites.length > 0);

  function kindIcon(kind: SearchKind) {
    switch (kind) {
      case "history":
        return Clock;
      case "site":
        return Compass;
      case "favorite":
        return Star;
    }
  }

  function openResult(url: string) {
    onOpen(url, query.trim() || undefined);
  }

  function formatHops(hops: number): string {
    if (hops < 0) {
      return "";
    }
    return hops === 1
      ? t("devtools.hopsCount", { count: hops })
      : t("devtools.hopsCountPlural", { count: hops });
  }
</script>

<section class="search-panel">
  <header>
    <div>
      <h2>{t("search.title")}</h2>
      <p>{t("search.subtitle")}</p>
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

  {#if !hasIndexedData}
    <EmptyState title={t("search.noData")} description={t("search.noDataDescription")}>
      <Search size={22} />
    </EmptyState>
  {:else if query.trim() && results.length === 0}
    <EmptyState
      title={t("search.noMatching")}
      description={t("common.nothingMatches", { query: query.trim() })}
    >
      <Search size={22} />
    </EmptyState>
  {:else if results.length === 0}
    <EmptyState title={t("search.noRecent")} description={t("search.noRecentDescription")}>
      <Search size={22} />
    </EmptyState>
  {:else}
    <div class="groups">
      {#each groups as group (group.kind)}
        {#if query.trim()}
          <h3>{group.label}</h3>
        {:else}
          <h3>{t("search.recent")}</h3>
        {/if}
        <ul>
          {#each group.results as result (result.key)}
            {@const Icon = kindIcon(result.kind)}
            <li>
              <button onclick={() => openResult(result.url)}>
                <span class="row">
                  <span class="icon" aria-hidden="true"><Icon size={14} /></span>
                  <span class="name">{result.title}</span>
                  {#if result.kind === "site" && result.hops !== undefined && result.hops >= 0}
                    <span class="hops-badge">{formatHops(result.hops)}</span>
                  {/if}
                </span>
                <span class="meta">{result.subtitle}</span>
              </button>
            </li>
          {/each}
        </ul>
      {/each}
    </div>
  {/if}
</section>

<style>
  .search-panel {
    height: 100%;
    overflow: auto;
    overflow-x: hidden;
    padding: 1rem;
    background: var(--ren-content-bg);
  }

  header h2 {
    margin: 0 0 0.25rem;
    font-size: 1.05rem;
    font-weight: 600;
  }

  header p {
    margin: 0 0 0.75rem;
    color: var(--ren-muted);
    font-size: 0.88rem;
  }

  .search {
    margin-bottom: 1rem;
  }

  .groups {
    display: grid;
    gap: 0.75rem;
  }

  h3 {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
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
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
  }

  .icon {
    flex-shrink: 0;
    color: var(--ren-accent);
  }

  .name {
    flex: 1;
    min-width: 0;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
