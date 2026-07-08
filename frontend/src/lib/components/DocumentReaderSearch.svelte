<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { ChevronDown, ChevronUp, Search, X } from "@lucide/svelte";
  import {
    clearFindHighlights,
    highlightFindMatches,
    scrollToFindMatch,
  } from "$lib/browser/find-in-page";
  import { frameSearchRoot } from "$lib/documents/document-frame-search";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    frame: HTMLIFrameElement | null | undefined;
    onClose: () => void;
  };

  let { frame, onClose }: Props = $props();

  let query = $state("");
  let matchCount = $state(0);
  let activeIndex = $state(0);

  function searchRoot(): HTMLElement | null {
    return frameSearchRoot(frame);
  }

  $effect(() => {
    const root = searchRoot();
    const needle = query;
    if (!root) {
      matchCount = 0;
      activeIndex = 0;
      return;
    }
    matchCount = highlightFindMatches(root, needle);
    activeIndex = 0;
    if (needle.trim() && matchCount > 0) {
      scrollToFindMatch(root, 0);
    }
  });

  function close() {
    const root = searchRoot();
    if (root) {
      clearFindHighlights(root);
    }
    query = "";
    matchCount = 0;
    activeIndex = 0;
    onClose();
  }

  function nextMatch(forward: boolean) {
    const root = searchRoot();
    if (!root || matchCount === 0) {
      return;
    }
    activeIndex = forward ? activeIndex + 1 : activeIndex - 1;
    activeIndex = ((activeIndex % matchCount) + matchCount) % matchCount;
    scrollToFindMatch(root, activeIndex);
  }

  function onKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      close();
      return;
    }
    if (event.key === "Enter") {
      event.preventDefault();
      nextMatch(!event.shiftKey);
    }
  }
</script>

<div class="reader-search">
  <Search size={14} aria-hidden="true" />
  <input
    class="reader-search-input"
    type="search"
    placeholder={t("documents.searchPlaceholder")}
    bind:value={query}
    onkeydown={onKeyDown}
  />
  <span class="reader-search-count" aria-live="polite">
    {#if query.trim()}
      {matchCount > 0
        ? t("documents.searchMatchOf", {
            current: String(activeIndex + 1),
            total: String(matchCount),
          })
        : t("documents.searchNoResults")}
    {/if}
  </span>
  <button
    type="button"
    class="reader-search-icon"
    aria-label={t("documents.searchPrev")}
    disabled={matchCount === 0}
    onclick={() => nextMatch(false)}
  >
    <ChevronUp size={14} />
  </button>
  <button
    type="button"
    class="reader-search-icon"
    aria-label={t("documents.searchNext")}
    disabled={matchCount === 0}
    onclick={() => nextMatch(true)}
  >
    <ChevronDown size={14} />
  </button>
  <button
    type="button"
    class="reader-search-icon"
    aria-label={t("documents.searchClose")}
    onclick={close}
  >
    <X size={14} />
  </button>
</div>

<style>
  .reader-search {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.4rem 0.75rem;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    color: var(--ren-muted);
  }

  .reader-search-input {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.35rem 0.55rem;
    font: inherit;
    font-size: 0.86rem;
  }

  .reader-search-count {
    font-size: 0.75rem;
    white-space: nowrap;
    min-width: 4.5rem;
    text-align: center;
  }

  .reader-search-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    padding: 0;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 8px;
    cursor: pointer;
    flex-shrink: 0;
  }

  .reader-search-icon:disabled {
    opacity: 0.45;
    cursor: default;
  }

  :global(mark.ren-find-hit) {
    background: color-mix(in srgb, var(--ren-accent) 35%, transparent);
    color: inherit;
    border-radius: 2px;
  }

  :global(mark.ren-find-active) {
    outline: 2px solid var(--ren-accent);
    outline-offset: 1px;
  }
</style>
