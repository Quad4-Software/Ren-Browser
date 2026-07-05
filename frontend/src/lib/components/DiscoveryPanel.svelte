<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Compass, Star } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { displayName } from "$lib/brand";

  type Node = {
    hash: string;
    name: string;
    hops: number;
    lastSeen: number;
  };

  type Props = {
    nodes: Node[];
    favorites: string[];
    onOpen: (url: string) => void;
    onFavorite: (url: string) => void;
  };

  let { nodes, favorites, onOpen, onFavorite }: Props = $props();

  let query = $state("");

  const filtered = $derived(() => {
    const q = query.trim().toLowerCase();
    if (!q) {
      return nodes;
    }
    return nodes.filter((node) => {
      const hay = `${node.name} ${node.hash}`.toLowerCase();
      return hay.includes(q);
    });
  });

  const placeholder = $derived(
    nodes.length > 0 ? `Search ${nodes.length} sites...` : "Search sites...",
  );

  function openNode(node: Node) {
    onOpen(`${node.hash}:/page/index.mu`);
  }

  function formatSeen(ts: number): string {
    if (!ts) {
      return "recently";
    }
    return new Date(ts * 1000).toLocaleString();
  }

  function isFavorite(hash: string): boolean {
    const url = `${hash}:/page/index.mu`;
    return favorites.some((f) => f.startsWith(hash) || f === url);
  }
</script>

<section class="discovery">
  <header>
    <h2>Discovery</h2>
    <p>Browse sites on the mesh network.</p>
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
    <EmptyState
      title="No sites discovered yet"
      description="{displayName} is scanning the mesh. Sites will appear here when nodes announce themselves."
    >
      <Compass size={22} />
    </EmptyState>
  {:else if filtered().length === 0}
    <EmptyState title="No matching sites" description={'Nothing matches "' + query.trim() + '".'}>
      <Compass size={22} />
    </EmptyState>
  {:else}
    <ul>
      {#each filtered() as node (node.hash)}
        <li>
          <button onclick={() => openNode(node)}>
            <span class="row">
              <span class="name">{node.name || "Unnamed site"}</span>
              <span
                class="fav"
                role="button"
                tabindex="0"
                aria-label="Favorite site"
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
            <span class="meta">Last seen {formatSeen(node.lastSeen)}</span>
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

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
  }

  button {
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

  button:hover {
    border-color: var(--ren-border-strong);
    background: var(--ren-tab-hover);
  }

  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.5rem;
  }

  .name {
    font-weight: 600;
  }

  .fav {
    color: var(--ren-accent);
  }

  .meta {
    color: var(--ren-muted);
    font-size: 0.85em;
  }
</style>
