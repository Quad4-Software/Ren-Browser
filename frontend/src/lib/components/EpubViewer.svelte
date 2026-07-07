<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  /* eslint-disable svelte/no-at-html-tags -- sanitized epub chapter HTML only */
  import { ChevronLeft, ChevronRight, List } from "@lucide/svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    DOCUMENT_PARSE_TIMEOUT_MS,
    formatDocumentError,
    withTimeout,
  } from "$lib/documents/async";
  import { decodeBase64ToUint8Array } from "$lib/documents/types";
  import { parseEpub, revokeEpubBlobUrls, type EpubBook } from "$lib/documents/epub";

  type Props = {
    binaryB64: string;
    onRetry?: () => void;
  };

  let { binaryB64, onRetry = () => {} }: Props = $props();

  let book: EpubBook | undefined = $state();
  let chapterIndex = $state(0);
  let loading = $state(false);
  let error = $state("");
  let tocOpen = $state(false);

  const chapter = $derived(book?.chapters[chapterIndex]);

  $effect(() => {
    const payload = binaryB64;
    let cancelled = false;

    if (!payload.trim()) {
      loading = false;
      error = t("documents.missingData");
      book = undefined;
      return;
    }

    loading = true;
    error = "";
    if (book) {
      revokeEpubBlobUrls(book);
    }
    book = undefined;
    chapterIndex = 0;

    void (async () => {
      try {
        const bytes = decodeBase64ToUint8Array(payload);
        const parsed = await withTimeout(
          parseEpub(bytes),
          DOCUMENT_PARSE_TIMEOUT_MS,
          "EPUB parse",
        );
        if (cancelled) {
          revokeEpubBlobUrls(parsed);
          return;
        }
        book = parsed;
      } catch (err) {
        if (cancelled) {
          return;
        }
        console.error("[EpubViewer] load failed", err);
        error = formatDocumentError(err, t("documents.loadFailed"));
      } finally {
        if (!cancelled) {
          loading = false;
        }
      }
    })();

    return () => {
      cancelled = true;
      if (book) {
        revokeEpubBlobUrls(book);
      }
      book = undefined;
    };
  });
</script>

<section class="epub-viewer">
  <header class="toolbar">
    <button type="button" class="toc-btn" aria-expanded={tocOpen} onclick={() => (tocOpen = !tocOpen)}>
      <List size={16} />
      <span>{t("documents.toc")}</span>
    </button>
    <div class="nav">
      <button
        type="button"
        aria-label={t("documents.prevChapter")}
        disabled={chapterIndex <= 0}
        onclick={() => (chapterIndex -= 1)}
      >
        <ChevronLeft size={16} />
      </button>
      <span class="status">{chapter?.title ?? ""}</span>
      <button
        type="button"
        aria-label={t("documents.nextChapter")}
        disabled={!book || chapterIndex >= book.chapters.length - 1}
        onclick={() => (chapterIndex += 1)}
      >
        <ChevronRight size={16} />
      </button>
    </div>
  </header>

  {#if tocOpen && book}
    <nav class="toc" aria-label={t("documents.toc")}>
      <ul>
        {#each book.chapters as item, index (item.id)}
          <li>
            <button
              type="button"
              class:active={index === chapterIndex}
              onclick={() => {
                chapterIndex = index;
                tocOpen = false;
              }}
            >
              {item.title}
            </button>
          </li>
        {/each}
      </ul>
    </nav>
  {/if}

  <div class="viewport">
    {#if loading}
      <div class="state">{t("documents.loading")}</div>
    {:else if error}
      <div class="state-panel error">
        <p>{error}</p>
        <button type="button" class="retry-btn" onclick={onRetry}>{t("documents.retry")}</button>
      </div>
    {:else if chapter}
      <article class="chapter">
        <h1>{chapter.title}</h1>
        <div class="chapter-body">{@html chapter.html}</div>
      </article>
    {/if}
  </div>
</section>

<style>
  .epub-viewer {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    background: var(--ren-content-bg);
  }

  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    flex-wrap: wrap;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
  }

  .toc-btn,
  .nav button,
  .retry-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.3rem 0.55rem;
    font: inherit;
    cursor: pointer;
  }

  .nav {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    min-width: 0;
    flex: 1;
    justify-content: flex-end;
  }

  .nav button:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .status {
    font-size: 0.84rem;
    color: var(--ren-muted);
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 14rem;
  }

  .toc {
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    max-height: 10rem;
    overflow: auto;
  }

  .toc ul {
    list-style: none;
    margin: 0;
    padding: 0.35rem;
    display: grid;
    gap: 0.2rem;
  }

  .toc button {
    width: 100%;
    text-align: left;
    border: none;
    background: transparent;
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.4rem 0.55rem;
    font: inherit;
    cursor: pointer;
  }

  .toc button:hover,
  .toc button.active {
    background: var(--ren-tab-hover);
  }

  .viewport {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 1rem 1.25rem 2rem;
  }

  .chapter {
    max-width: 42rem;
    margin: 0 auto;
    line-height: 1.6;
  }

  .chapter h1 {
    margin: 0 0 1rem;
    font-size: 1.35rem;
  }

  .chapter-body :global(img) {
    max-width: 100%;
    height: auto;
  }

  .chapter-body :global(a) {
    color: var(--ren-accent);
    pointer-events: none;
    text-decoration: underline;
  }

  .state {
    margin: auto;
    color: var(--ren-muted);
    font-size: 0.9rem;
    text-align: center;
  }

  .state-panel {
    margin: auto;
    max-width: 28rem;
    text-align: center;
    display: grid;
    gap: 0.75rem;
    padding: 1rem;
  }

  .state-panel p {
    margin: 0;
    overflow-wrap: anywhere;
  }

  .state-panel.error p {
    color: var(--ren-danger, #e5484d);
  }
</style>
