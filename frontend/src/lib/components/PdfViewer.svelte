<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { ChevronLeft, ChevronRight, RotateCw, ZoomIn, ZoomOut } from "@lucide/svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import { resolveDocumentErrorMessage } from "$lib/documents/async";
  import { handleDocumentArrowKeys } from "$lib/documents/document-keys";
  import { decodeBase64ToUint8Array } from "$lib/documents/types";
  import { buildIsolatedPdfShell, ISOLATED_FRAME_SANDBOX } from "$lib/documents/isolated-html";
  import { resolvedReaderTheme, type ReaderTheme } from "$lib/documents/reader-theme";
  import { fitWidthScale, loadPdfDocument, renderPdfPage } from "$lib/documents/pdf";
  import { cycleReaderRotation, type ReaderRotation } from "$lib/documents/reader-layout";
  import { attachReaderSwipe } from "$lib/documents/reader-swipe";
  import type { PDFDocumentProxy } from "pdfjs-dist";

  type Props = {
    binaryB64: string;
    onRetry?: () => void;
  };

  let { binaryB64, onRetry = () => {} }: Props = $props();

  let viewportEl: HTMLElement | undefined = $state();
  let frameEl: HTMLIFrameElement | undefined = $state();
  let pageScrollEl: HTMLElement | null = $state(null);
  let doc: PDFDocumentProxy | undefined = $state();
  let page = $state(1);
  let numPages = $state(0);
  let scale = $state(1);
  let rotation = $state<ReaderRotation>(0);
  let loading = $state(false);
  let error = $state("");
  let loadSeq = 0;
  let ownedDoc: PDFDocumentProxy | undefined;
  let readerTheme = $state<ReaderTheme>(resolvedReaderTheme());

  $effect(() => {
    const root = document.documentElement;
    const syncTheme = () => {
      readerTheme = resolvedReaderTheme();
    };
    syncTheme();
    const observer = new MutationObserver(syncTheme);
    observer.observe(root, { attributes: true, attributeFilter: ["data-theme"] });
    return () => observer.disconnect();
  });

  function dropOwnedDoc() {
    if (ownedDoc) {
      void ownedDoc.cleanup();
      ownedDoc = undefined;
    }
  }

  onDestroy(() => {
    dropOwnedDoc();
  });

  function scrollRootFromFrame(frame: HTMLIFrameElement | undefined): HTMLElement | null {
    const doc = frame?.contentDocument;
    if (!doc) {
      return null;
    }
    return doc.scrollingElement instanceof HTMLElement ? doc.scrollingElement : doc.documentElement;
  }

  function ensurePdfCanvas(
    frame: HTMLIFrameElement,
    theme: ReaderTheme,
    rotate: number,
  ): HTMLCanvasElement | null {
    const idoc = frame.contentDocument;
    if (!idoc) {
      return null;
    }
    let canvas = idoc.getElementById("page-canvas");
    const shellTheme = frame.dataset.readerTheme;
    const shellRotation = frame.dataset.readerRotation ?? "0";
    if (
      !(canvas instanceof HTMLCanvasElement) ||
      shellTheme !== theme ||
      shellRotation !== String(rotate)
    ) {
      idoc.open();
      idoc.write(buildIsolatedPdfShell(theme, rotate));
      idoc.close();
      frame.dataset.readerTheme = theme;
      frame.dataset.readerRotation = String(rotate);
      canvas = idoc.getElementById("page-canvas");
    }
    pageScrollEl = scrollRootFromFrame(frame);
    return canvas instanceof HTMLCanvasElement ? canvas : null;
  }

  async function paintPage() {
    if (!doc || !frameEl) {
      return;
    }
    const canvas = ensurePdfCanvas(frameEl, readerTheme, rotation);
    if (!canvas) {
      return;
    }
    await renderPdfPage(doc, page, canvas, scale, rotation);
  }

  async function fitToWidth() {
    if (!doc || !viewportEl) {
      return;
    }
    const pdfPage = await doc.getPage(page);
    scale = fitWidthScale(pdfPage, viewportEl.clientWidth - 32);
    await paintPage();
  }

  $effect(() => {
    const payload = binaryB64;
    const seq = ++loadSeq;

    dropOwnedDoc();
    doc = undefined;
    page = 1;
    numPages = 0;

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
        const loaded = await loadPdfDocument(bytes);
        if (seq !== loadSeq) {
          await loaded.doc.cleanup();
          return;
        }
        ownedDoc = loaded.doc;
        doc = loaded.doc;
        numPages = loaded.numPages;
        loading = false;
        await fitToWidth();
      } catch (err) {
        if (seq !== loadSeq) {
          return;
        }
        console.error("[PdfViewer] load failed", err);
        error = resolveDocumentErrorMessage(err, t);
        loading = false;
      }
    })();
  });

  async function prevPage() {
    if (page <= 1) {
      return;
    }
    page -= 1;
    await fitToWidth();
  }

  async function nextPage() {
    if (!doc || page >= numPages) {
      return;
    }
    page += 1;
    await fitToWidth();
  }

  async function zoomIn() {
    scale = Math.min(scale * 1.2, 4);
    await paintPage();
  }

  async function zoomOut() {
    scale = Math.max(scale / 1.2, 0.4);
    await paintPage();
  }

  function onKeyDown(event: KeyboardEvent) {
    if (loading || error) {
      return;
    }
    handleDocumentArrowKeys(event, {
      viewport: pageScrollEl ?? viewportEl,
      onPrev: () => {
        void prevPage();
      },
      onNext: () => {
        void nextPage();
      },
      canPrev: page > 1,
      canNext: !!doc && page < numPages,
    });
  }

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
      isEnabled: () => !loading && !error && !!doc,
      canPrev: () => page > 1,
      canNext: () => !!doc && page < numPages,
      onPrev: () => {
        void prevPage();
      },
      onNext: () => {
        void nextPage();
      },
    });
    return () => swipe.teardown();
  });

  $effect(() => {
    if (!loading && !error && doc && viewportEl) {
      viewportEl.focus({ preventScroll: true });
    }
  });

  $effect(() => {
    const frame = frameEl;
    void readerTheme;
    void rotation;
    if (!frame || loading || error || !doc) {
      pageScrollEl = null;
      return;
    }
    const init = () => {
      void paintPage();
    };
    if (frame.contentDocument?.readyState === "complete") {
      init();
    } else {
      frame.addEventListener("load", init, { once: true });
    }
    return () => {
      pageScrollEl = null;
    };
  });
</script>

<div class="pdf-viewer">
  <header class="toolbar">
    <div class="nav">
      <button
        type="button"
        aria-label={t("documents.prevPage")}
        disabled={page <= 1}
        onclick={prevPage}
      >
        <ChevronLeft size={16} />
      </button>
      <span class="status"
        >{t("documents.pageOf", { current: String(page), total: String(numPages || 1) })}</span
      >
      <button
        type="button"
        aria-label={t("documents.nextPage")}
        disabled={!doc || page >= numPages}
        onclick={nextPage}
      >
        <ChevronRight size={16} />
      </button>
    </div>
    <div class="zoom">
      <button type="button" aria-label={t("documents.zoomOut")} onclick={zoomOut}>
        <ZoomOut size={16} />
      </button>
      <button type="button" aria-label={t("documents.zoomIn")} onclick={zoomIn}>
        <ZoomIn size={16} />
      </button>
      <button type="button" class="fit-btn" onclick={fitToWidth}>{t("documents.fitWidth")}</button>
      <button
        type="button"
        aria-label={t("documents.rotate")}
        onclick={() => {
          rotation = cycleReaderRotation(rotation);
          void paintPage();
        }}
      >
        <RotateCw size={16} />
      </button>
    </div>
  </header>

  <div
    class="viewport"
    bind:this={viewportEl}
    role="region"
    aria-label={t("documents.pageOf", { current: String(page), total: String(numPages || 1) })}
    tabindex="-1"
  >
    {#if loading}
      <div class="state">{t("documents.loading")}</div>
    {:else if error}
      <div class="state-panel error">
        <p>{error}</p>
        <button type="button" class="retry-btn" onclick={onRetry}>{t("documents.retry")}</button>
      </div>
    {:else}
      <iframe
        bind:this={frameEl}
        class="pdf-frame"
        title={t("documents.pageOf", { current: String(page), total: String(numPages || 1) })}
        sandbox={ISOLATED_FRAME_SANDBOX}
        src="about:blank"
      ></iframe>
    {/if}
  </div>
</div>

<style>
  .pdf-viewer {
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

  .nav,
  .zoom {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
  }

  .toolbar button,
  .retry-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.3rem 0.45rem;
    font: inherit;
    cursor: pointer;
  }

  .toolbar button:disabled {
    opacity: 0.45;
    cursor: default;
  }

  .fit-btn {
    padding-inline: 0.65rem;
    font-size: 0.82rem;
  }

  .status {
    font-size: 0.84rem;
    color: var(--ren-muted);
    min-width: 6rem;
    text-align: center;
  }

  .viewport {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    padding: 0;
    display: flex;
    flex-direction: column;
    outline: none;
    touch-action: pan-y;
  }

  .pdf-frame {
    flex: 1;
    width: 100%;
    min-height: 0;
    border: none;
    background: transparent;
  }

  .state {
    margin: auto;
    color: var(--ren-muted);
    font-size: 0.9rem;
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
