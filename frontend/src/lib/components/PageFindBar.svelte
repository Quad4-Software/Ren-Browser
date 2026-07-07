<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Search, X } from "@lucide/svelte";
  import {
    clearFindHighlights,
    highlightFindMatches,
    scrollToFindMatch,
  } from "$lib/browser/find-in-page";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    open: boolean;
    onClose: () => void;
    contentRoot: HTMLElement | undefined;
  };

  let { open, onClose, contentRoot }: Props = $props();

  let query = $state("");
  let matchCount = $state(0);
  let activeIndex = $state(0);

  $effect(() => {
    if (!open) {
      query = "";
      matchCount = 0;
      activeIndex = 0;
      if (contentRoot) {
        clearFindHighlights(contentRoot);
      }
    }
  });

  $effect(() => {
    if (!open || !contentRoot) {
      return;
    }
    matchCount = highlightFindMatches(contentRoot, query);
    activeIndex = 0;
    if (query && matchCount > 0) {
      scrollToFindMatch(contentRoot, 0);
    }
  });

  function nextMatch(forward: boolean) {
    if (!contentRoot || matchCount === 0) {
      return;
    }
    activeIndex = forward ? activeIndex + 1 : activeIndex - 1;
    activeIndex = ((activeIndex % matchCount) + matchCount) % matchCount;
    scrollToFindMatch(contentRoot, activeIndex);
  }

  function onInput(event: Event) {
    query = (event.currentTarget as HTMLInputElement).value;
  }

  function onKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      onClose();
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      nextMatch(!event.shiftKey);
    }
  }
</script>

{#if open}
  <div class="find-bar">
    <Search size={14} />
    <input
      class="find-input"
      type="search"
      placeholder={t("find.placeholder")}
      value={query}
      oninput={onInput}
      onkeydown={onKeyDown}
    />
    <span class="count">
      {#if query}
        {matchCount > 0
          ? t("common.matchCount", { current: activeIndex + 1, total: matchCount })
          : t("common.noMatches")}
      {/if}
    </span>
    <button class="ren-icon-btn" aria-label={t("find.previous")} onclick={() => nextMatch(false)}>
      ↑
    </button>
    <button class="ren-icon-btn" aria-label={t("find.next")} onclick={() => nextMatch(true)}>
      ↓
    </button>
    <button class="ren-icon-btn" aria-label={t("find.close")} onclick={onClose}>
      <X size={14} />
    </button>
  </div>
{/if}

<style>
  .find-bar {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    min-width: 0;
    overflow: hidden;
    padding: 0.45rem 0.75rem;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    color: var(--ren-muted);
  }

  .find-input {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.35rem 0.55rem;
    font: inherit;
  }

  .count {
    flex-shrink: 0;
    font-size: 0.78rem;
    min-width: 3rem;
    text-align: right;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
