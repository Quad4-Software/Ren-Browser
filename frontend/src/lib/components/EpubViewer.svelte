<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onDestroy } from "svelte";
  import {
    ChevronLeft,
    ChevronRight,
    Menu,
    PanelLeft,
    RotateCw,
    Search,
    ZoomIn,
    ZoomOut,
  } from "@lucide/svelte";
  import DocumentReaderSearch from "$lib/components/DocumentReaderSearch.svelte";
  import DocumentTocPanel from "$lib/components/DocumentTocPanel.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    DOCUMENT_PARSE_TIMEOUT_MS,
    resolveDocumentErrorMessage,
    withTimeout,
  } from "$lib/documents/async";
  import { decodeBase64ToUint8Array } from "$lib/documents/types";
  import { handleDocumentArrowKeys } from "$lib/documents/document-keys";
  import {
    cycleReaderRotation,
    readerFontScaleLabel,
    stepReaderFontScale,
    type ReaderRotation,
  } from "$lib/documents/reader-layout";
  import { attachReaderSwipe } from "$lib/documents/reader-swipe";
  import { parseEpub, revokeEpubBlobUrls, type EpubBook } from "$lib/documents/epub";
  import IsolatedHtmlFrame from "$lib/components/IsolatedHtmlFrame.svelte";

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
  let sidebarOpen = $state(true);
  let searchOpen = $state(false);
  let fontScale = $state(1);
  let rotation = $state<ReaderRotation>(0);
  let viewportEl: HTMLElement | undefined = $state();
  let chapterScrollEl: HTMLElement | null = $state(null);
  let chapterFrameEl: HTMLIFrameElement | null = $state(null);
  let loadSeq = 0;
  let ownedBook: EpubBook | undefined;

  const chapter = $derived(book?.chapters[chapterIndex]);

  function prevChapter() {
    if (chapterIndex > 0) {
      chapterIndex -= 1;
    }
  }

  function nextChapter() {
    if (book && chapterIndex < book.chapters.length - 1) {
      chapterIndex += 1;
    }
  }

  function selectChapter(index: number) {
    chapterIndex = index;
    tocOpen = false;
  }

  function toggleToc() {
    tocOpen = !tocOpen;
  }

  function closeToc() {
    tocOpen = false;
  }

  $effect(() => {
    if (!tocOpen) {
      return;
    }
    const onEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeToc();
      }
    };
    window.addEventListener("keydown", onEscape);
    return () => {
      window.removeEventListener("keydown", onEscape);
    };
  });

  function onKeyDown(event: KeyboardEvent) {
    if (loading || error || searchOpen) {
      return;
    }
    handleDocumentArrowKeys(event, {
      viewport: chapterScrollEl ?? viewportEl,
      onPrev: prevChapter,
      onNext: nextChapter,
      canPrev: chapterIndex > 0,
      canNext: !!book && chapterIndex < book.chapters.length - 1,
    });
  }

  $effect(() => {
    void chapterIndex;
    if (chapterScrollEl) {
      chapterScrollEl.scrollTop = 0;
    } else if (viewportEl) {
      viewportEl.scrollTop = 0;
    }
  });

  $effect(() => {
    const el = viewportEl;
    if (!el) {
      return;
    }
    const handler = (event: KeyboardEvent) => {
      onKeyDown(event);
    };
    el.addEventListener("keydown", handler);
    return () => {
      el.removeEventListener("keydown", handler);
    };
  });

  $effect(() => {
    const el = viewportEl;
    if (!el) {
      return;
    }
    const swipe = attachReaderSwipe(el, {
      isEnabled: () => !loading && !error && !!book && !tocOpen && !searchOpen,
      canPrev: () => chapterIndex > 0,
      canNext: () => !!book && chapterIndex < book.chapters.length - 1,
      onPrev: prevChapter,
      onNext: nextChapter,
    });
    return () => swipe.teardown();
  });

  $effect(() => {
    if (!loading && !error && book && viewportEl && !searchOpen) {
      viewportEl.focus({ preventScroll: true });
    }
  });

  function dropOwnedBook() {
    if (ownedBook) {
      revokeEpubBlobUrls(ownedBook);
      ownedBook = undefined;
    }
  }

  onDestroy(() => {
    dropOwnedBook();
  });

  $effect(() => {
    const payload = binaryB64;
    const seq = ++loadSeq;

    dropOwnedBook();
    book = undefined;
    chapterIndex = 0;
    tocOpen = false;
    searchOpen = false;

    if (!payload.trim()) {
      loading = false;
      error = t("documents.missingData");
      return;
    }

    loading = true;
    error = "";

    void (async () => {
      try {
        const bytes = decodeBase64ToUint8Array(payload);
        const parsed = await withTimeout(parseEpub(bytes), DOCUMENT_PARSE_TIMEOUT_MS, "EPUB parse");
        if (seq !== loadSeq) {
          revokeEpubBlobUrls(parsed);
          return;
        }
        ownedBook = parsed;
        book = parsed;
      } catch (err) {
        if (seq !== loadSeq) {
          return;
        }
        console.error("[EpubViewer] load failed", err);
        error = resolveDocumentErrorMessage(err, t);
      } finally {
        if (seq === loadSeq) {
          loading = false;
        }
      }
    })();
  });
</script>

<div class="epub-viewer" class:toc-open={tocOpen}>
  <header class="toolbar">
    <div class="toolbar-group toolbar-leading">
      <button
        type="button"
        class="icon-btn toc-mobile-btn"
        aria-expanded={tocOpen}
        aria-label={tocOpen ? t("documents.closeToc") : t("documents.openToc")}
        onclick={toggleToc}
      >
        <Menu size={16} />
      </button>
      <button
        type="button"
        class="icon-btn toc-desktop-btn"
        aria-expanded={sidebarOpen}
        aria-label={sidebarOpen ? t("documents.collapseToc") : t("documents.expandToc")}
        onclick={() => {
          sidebarOpen = !sidebarOpen;
        }}
      >
        <PanelLeft size={16} />
      </button>
      <button
        type="button"
        class="icon-btn"
        class:active={searchOpen}
        aria-expanded={searchOpen}
        aria-label={t("documents.search")}
        onclick={() => {
          searchOpen = !searchOpen;
        }}
      >
        <Search size={16} />
      </button>
    </div>

    <div class="toolbar-group toolbar-controls">
      <button
        type="button"
        class="icon-btn"
        aria-label={t("documents.fontSmaller")}
        onclick={() => {
          fontScale = stepReaderFontScale(fontScale, -1);
        }}
      >
        <ZoomOut size={16} />
      </button>
      <span class="control-label">{readerFontScaleLabel(fontScale)}</span>
      <button
        type="button"
        class="icon-btn"
        aria-label={t("documents.fontLarger")}
        onclick={() => {
          fontScale = stepReaderFontScale(fontScale, 1);
        }}
      >
        <ZoomIn size={16} />
      </button>
      <button
        type="button"
        class="icon-btn"
        aria-label={t("documents.rotate")}
        onclick={() => {
          rotation = cycleReaderRotation(rotation);
        }}
      >
        <RotateCw size={16} />
      </button>
    </div>

    <div class="toolbar-group toolbar-nav">
      <button
        type="button"
        class="icon-btn"
        aria-label={t("documents.prevChapter")}
        disabled={chapterIndex <= 0}
        onclick={prevChapter}
      >
        <ChevronLeft size={16} />
      </button>
      <span class="status">{chapter?.title ?? ""}</span>
      <button
        type="button"
        class="icon-btn"
        aria-label={t("documents.nextChapter")}
        disabled={!book || chapterIndex >= book.chapters.length - 1}
        onclick={nextChapter}
      >
        <ChevronRight size={16} />
      </button>
    </div>
  </header>

  {#if searchOpen}
    <DocumentReaderSearch
      frame={chapterFrameEl}
      onClose={() => {
        searchOpen = false;
      }}
    />
  {/if}

  <div class="reader-body">
    {#if book && sidebarOpen}
      <aside class="toc-sidebar">
        <DocumentTocPanel
          chapters={book.chapters}
          activeIndex={chapterIndex}
          variant="sidebar"
          onSelect={selectChapter}
          onCollapse={() => {
            sidebarOpen = false;
          }}
        />
      </aside>
    {/if}

    <div class="reader-main">
      <div
        class="viewport"
        bind:this={viewportEl}
        role="region"
        aria-label={chapter?.title ?? t("documents.toc")}
        tabindex="-1"
      >
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
            <IsolatedHtmlFrame
              html={chapter.html}
              title={chapter.title}
              {fontScale}
              rotation={rotation}
              onScrollRoot={(root) => {
                chapterScrollEl = root;
              }}
              onFrame={(frame) => {
                chapterFrameEl = frame;
              }}
            />
          </article>
        {/if}
      </div>
    </div>
  </div>

  {#if book && tocOpen}
    <DocumentTocPanel
      chapters={book.chapters}
      activeIndex={chapterIndex}
      variant="drawer"
      onSelect={selectChapter}
      onClose={closeToc}
    />
  {/if}
</div>

<style>
  .epub-viewer {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    background: var(--ren-content-bg);
    position: relative;
  }

  .epub-viewer.toc-open {
    overflow: hidden;
  }

  .toolbar {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    gap: 0.35rem;
    padding: 0.45rem 0.75rem;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    flex-shrink: 0;
    min-width: 0;
    overflow-x: auto;
  }

  .toolbar-group {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    min-width: 0;
    flex-shrink: 0;
  }

  .toolbar-leading {
    flex-shrink: 0;
  }

  .toolbar-controls {
    flex: 1;
    justify-content: center;
    min-width: 0;
  }

  .toolbar-nav {
    flex-shrink: 0;
    margin-left: auto;
  }

  .icon-btn,
  .retry-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.85rem;
    height: 1.85rem;
    padding: 0;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 8px;
    font: inherit;
    cursor: pointer;
    flex-shrink: 0;
  }

  .icon-btn.active {
    border-color: var(--ren-accent);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--ren-accent) 35%, transparent);
  }

  .icon-btn:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .control-label {
    font-size: 0.72rem;
    color: var(--ren-muted);
    min-width: 2.25rem;
    text-align: center;
    flex-shrink: 0;
  }

  .status {
    font-size: 0.78rem;
    color: var(--ren-muted);
    min-width: 0;
    max-width: 9rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .reader-body {
    display: flex;
    flex: 1;
    min-height: 0;
    min-width: 0;
  }

  .toc-sidebar {
    display: none;
    width: min(16rem, 34vw);
    min-width: 12rem;
    flex-shrink: 0;
    min-height: 0;
    border-right: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
  }

  .reader-main {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    min-height: 0;
  }

  .viewport {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    padding: 1rem 1.25rem 2rem;
    display: flex;
    flex-direction: column;
    outline: none;
    touch-action: pan-y;
  }

  .chapter {
    max-width: 42rem;
    margin: 0 auto;
    line-height: 1.6;
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  .chapter h1 {
    margin: 0 0 1rem;
    font-size: 1.35rem;
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

  .toc-mobile-btn {
    display: none;
  }

  .toc-desktop-btn {
    display: none;
  }

  @media (min-width: 769px) {
    .toc-sidebar {
      display: flex;
    }

    .toc-desktop-btn {
      display: inline-flex;
    }
  }

  @media (max-width: 768px) {
    .toc-mobile-btn {
      display: inline-flex;
    }

    .toc-sidebar {
      display: none !important;
    }

    .status {
      max-width: 5.5rem;
    }

    .control-label {
      display: none;
    }
  }
</style>
