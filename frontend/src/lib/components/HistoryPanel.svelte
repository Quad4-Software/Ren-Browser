<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Clock, History, Trash2 } from "@lucide/svelte";
  import { SvelteDate } from "svelte/reactivity";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type HistoryEntry = {
    id: number;
    url: string;
    title: string;
    nodeHash: string;
    visitedAt: number;
  };

  type HistoryGroup = {
    dateLabel: string;
    entries: HistoryEntry[];
  };

  type Props = {
    history: HistoryEntry[];
    onOpen: (url: string) => void;
    onClear: () => void;
  };

  let { history, onOpen, onClear }: Props = $props();

  let query = $state("");

  const filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) {
      return history;
    }
    return history.filter((entry) => {
      const hay = `${historyLabel(entry)} ${entry.url} ${entry.nodeHash}`.toLowerCase();
      return hay.includes(q);
    });
  });

  const grouped = $derived.by(() => {
    const groups: HistoryGroup[] = [];
    let currentLabel = "";
    let currentEntries: HistoryEntry[] = [];

    for (const entry of filtered) {
      const label = humanDate(entry.visitedAt);
      if (label !== currentLabel) {
        if (currentEntries.length > 0) {
          groups.push({ dateLabel: currentLabel, entries: currentEntries });
        }
        currentLabel = label;
        currentEntries = [entry];
      } else {
        currentEntries.push(entry);
      }
    }

    if (currentEntries.length > 0) {
      groups.push({ dateLabel: currentLabel, entries: currentEntries });
    }

    return groups;
  });

  const placeholder = $derived(
    history.length > 0
      ? t("common.searchCount", { count: history.length, noun: t("history.pages") })
      : t("common.search", { noun: t("history.pages") }),
  );

  function sameDay(a: Date, b: Date): boolean {
    return (
      a.getFullYear() === b.getFullYear() &&
      a.getMonth() === b.getMonth() &&
      a.getDate() === b.getDate()
    );
  }

  function humanDate(ts: number): string {
    if (!ts) {
      return t("common.unknownDate");
    }
    const date = new Date(ts * 1000);
    const today = new SvelteDate();
    const yesterday = new SvelteDate();
    yesterday.setDate(today.getDate() - 1);

    if (sameDay(date, today)) {
      return t("common.today");
    }
    if (sameDay(date, yesterday)) {
      return t("common.yesterday");
    }

    return date.toLocaleDateString(undefined, {
      weekday: "long",
      month: "long",
      day: "numeric",
      year: date.getFullYear() !== today.getFullYear() ? "numeric" : undefined,
    });
  }

  function formatTime(ts: number): string {
    if (!ts) {
      return t("common.unknown");
    }
    return new Date(ts * 1000).toLocaleTimeString(undefined, {
      hour: "numeric",
      minute: "2-digit",
    });
  }

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
</script>

<section class="history-panel">
  <header>
    <div class="title-row">
      <div>
        <h2>{t("history.title")}</h2>
        <p class="subtitle">{t("history.subtitle")}</p>
      </div>
      {#if history.length > 0}
        <button
          type="button"
          class="ren-icon-btn clear-btn"
          aria-label={t("history.clear")}
          title={t("history.clear")}
          onclick={onClear}
        >
          <Trash2 size={16} />
        </button>
      {/if}
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

  {#if history.length === 0}
    <EmptyState title={t("history.noHistory")} description={t("history.noHistoryDescription")}>
      <History size={22} />
    </EmptyState>
  {:else if filtered.length === 0}
    <EmptyState
      title={t("history.noMatching")}
      description={t("common.nothingMatches", { query: query.trim() })}
    >
      <Clock size={22} />
    </EmptyState>
  {:else}
    <div class="groups">
      {#each grouped as group (group.dateLabel)}
        <div class="date-separator" role="presentation">
          <span class="line"></span>
          <span class="label">{group.dateLabel}</span>
          <span class="line"></span>
        </div>
        <ul>
          {#each group.entries as entry (entry.id)}
            <li>
              <button onclick={() => onOpen(entry.url)}>
                <span class="name">{historyLabel(entry)}</span>
                <span class="meta">{entry.url}</span>
                <span class="meta">{formatTime(entry.visitedAt)}</span>
              </button>
            </li>
          {/each}
        </ul>
      {/each}
    </div>
  {/if}
</section>

<style>
  .history-panel {
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

  .clear-btn {
    flex-shrink: 0;
    margin-left: auto;
  }

  .search {
    margin-bottom: 1rem;
  }

  .groups {
    display: grid;
    gap: 0.75rem;
  }

  .date-separator {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin: 0.25rem 0;
  }

  .date-separator .line {
    flex: 1;
    height: 1px;
    background: var(--ren-border);
  }

  .date-separator .label {
    flex-shrink: 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
    font-weight: 500;
    white-space: nowrap;
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
