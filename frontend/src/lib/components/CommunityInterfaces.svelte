<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Globe, RefreshCw } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  export type CommunityInterface = {
    id: number;
    name: string;
    type: string;
    typeName: string;
    network: string;
    host: string;
    port: number | null;
    status: string;
    config: string;
    installed: boolean;
  };

  type Props = {
    items: CommunityInterface[];
    loading: boolean;
    importing: boolean;
    error: string;
    filter: string;
    selected: Set<number>;
    onFilter: (value: string) => void;
    onRefresh: () => void;
    onToggle: (id: number) => void;
    onImport: () => void;
  };

  let {
    items,
    loading,
    importing,
    error,
    filter = $bindable(),
    selected,
    onFilter,
    onRefresh,
    onToggle,
    onImport,
  }: Props = $props();

  const filtered = $derived.by(() => {
    const q = filter.trim().toLowerCase();
    if (!q) {
      return items;
    }
    return items.filter(
      (item) =>
        item.name.toLowerCase().includes(q) ||
        item.typeName.toLowerCase().includes(q) ||
        item.network.toLowerCase().includes(q) ||
        item.host.toLowerCase().includes(q),
    );
  });

  const selectedCount = $derived(selected.size);
</script>

<section class="community">
  <div class="header">
    <h3>{t("community.title")}</h3>
    <p class="hint">{t("community.hint")}</p>
  </div>

  <div class="toolbar">
    <input
      class="search"
      type="search"
      placeholder={t("community.searchPlaceholder")}
      bind:value={filter}
      oninput={() => onFilter(filter)}
    />
    <button
      type="button"
      class="icon-btn"
      aria-label={t("community.refresh")}
      onclick={onRefresh}
      disabled={loading}
    >
      <span class:spin={loading}>
        <RefreshCw size={16} />
      </span>
    </button>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <ul class="list">
    {#if loading && items.length === 0}
      <li class="empty">{t("community.loading")}</li>
    {:else if filtered.length === 0}
      <li class="empty">
        <EmptyState
          title={t("community.noInterfaces")}
          description={t("community.noInterfacesDescription")}
        >
          <Globe size={22} />
        </EmptyState>
      </li>
    {:else}
      {#each filtered as item (item.id)}
        <li class:installed={item.installed}>
          <label>
            <input
              type="checkbox"
              checked={selected.has(item.id)}
              disabled={item.installed || importing}
              onchange={() => onToggle(item.id)}
            />
            <span class="body">
              <span class="name">{item.name}</span>
              <span class="meta">
                {item.typeName} · {item.network}
                {#if item.host}
                  · {item.host}{#if item.port}:{item.port}{/if}
                {/if}
                {#if item.installed}
                  · {t("common.installed")}
                {/if}
              </span>
            </span>
          </label>
        </li>
      {/each}
    {/if}
  </ul>

  <button
    type="button"
    class="import-btn"
    onclick={onImport}
    disabled={importing || selectedCount === 0}
  >
    {importing
      ? t("community.adding")
      : selectedCount > 0
        ? t("community.addSelected", { count: selectedCount })
        : t("community.addSelectedEmpty")}
  </button>
</section>

<style>
  .community {
    display: grid;
    gap: 0.65rem;
  }

  .header h3 {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .hint {
    margin: 0.2rem 0 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
  }

  .toolbar {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.45rem;
  }

  .search {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 2.5rem;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    cursor: pointer;
  }

  .icon-btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  :global(.spin) {
    display: inline-flex;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.85rem;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
    max-height: 40vh;
    overflow: auto;
  }

  .list li {
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface-raised);
  }

  .list li.installed {
    opacity: 0.72;
  }

  label {
    display: flex;
    gap: 0.55rem;
    align-items: flex-start;
    padding: 0.6rem 0.7rem;
    cursor: pointer;
  }

  .body {
    display: grid;
    gap: 0.15rem;
    min-width: 0;
  }

  .name {
    font-weight: 500;
    color: var(--ren-fg);
  }

  .meta {
    color: var(--ren-muted);
    font-size: 0.8rem;
    word-break: break-word;
  }

  .empty {
    border: none;
    background: transparent;
    padding: 0;
  }

  .import-btn {
    border: 1px solid var(--ren-accent);
    background: var(--ren-accent);
    color: #fff;
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.6rem 0.85rem;
    font: inherit;
    cursor: pointer;
  }

  .import-btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
</style>
