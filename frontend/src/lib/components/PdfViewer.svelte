<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { ChevronLeft, ChevronRight, ZoomIn, ZoomOut } from "@lucide/svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import { formatDocumentError } from "$lib/documents/async";
  import { decodeBase64ToUint8Array } from "$lib/documents/types";
  import { fitWidthScale, loadPdfDocument, renderPdfPage } from "$lib/documents/pdf";
  import type { PDFDocumentProxy } from "pdfjs-dist";

  type Props = {
    binaryB64: string;
    onRetry?: () => void;
  };

  let { binaryB64, onRetry = () => {} }: Props = $props();

  let viewportEl: HTMLElement | undefined = $state();
  let canvasEl: HTMLCanvasElement | undefined = $state();
  let doc: PDFDocumentProxy | undefined = $state();
  let page = $state(1);
  let numPages = $state(0);
  let scale = $state(1);
  let loading = $state(false);
  let error = $state("");

  async function paintPage() {
    if (!doc || !canvasEl) {
      return;
    }
    await renderPdfPage(doc, page, canvasEl, scale);
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
    let cancelled = false;

    if (!payload.trim()) {
      loading = false;
      error = t("documents.missingData");
      doc = undefined;
      return;
    }

    loading = true;
    error = "";
    void doc?.cleanup();
    doc = undefined;
    page = 1;
    numPages = 0;

    void (async () => {
      try {
        const bytes = decodeBase64ToUint8Array(payload);
        const loaded = await loadPdfDocument(bytes);
        if (cancelled) {
          await loaded.doc.cleanup();
          return;
        }
        doc = loaded.doc;
        numPages = loaded.numPages;
        loading = false;
        await fitToWidth();
      } catch (err) {
        if (cancelled) {
          return;
        }
        console.error("[PdfViewer] load failed", err);
        error = formatDocumentError(err, t("documents.loadFailed"));
        loading = false;
      }
    })();

    return () => {
      cancelled = true;
      void doc?.cleanup();
      doc = undefined;
    };
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
</script>

<section class="pdf-viewer">
  <header class="toolbar">
    <div class="nav">
      <button type="button" aria-label={t("documents.prevPage")} disabled={page <= 1} onclick={prevPage}>
        <ChevronLeft size={16} />
      </button>
      <span class="status">{t("documents.pageOf", { current: String(page), total: String(numPages || 1) })}</span>
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
    </div>
  </header>

  <div class="viewport" bind:this={viewportEl}>
    {#if loading}
      <div class="state">{t("documents.loading")}</div>
    {:else if error}
      <div class="state-panel error">
        <p>{error}</p>
        <button type="button" class="retry-btn" onclick={onRetry}>{t("documents.retry")}</button>
      </div>
    {:else}
      <canvas bind:this={canvasEl} class="page-canvas"></canvas>
    {/if}
  </div>
</section>

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
    overflow: auto;
    padding: 1rem;
    display: flex;
    justify-content: center;
  }

  .page-canvas {
    max-width: 100%;
    height: auto;
    box-shadow: 0 2px 12px rgb(0 0 0 / 18%);
    background: #fff;
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
