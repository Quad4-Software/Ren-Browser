<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  /* eslint-disable svelte/no-at-html-tags -- live micron preview */
  import { onMount } from "svelte";
  import { Download } from "@lucide/svelte";
  import { handlePageLinkClick } from "$lib/browser/page-links";
  import { downloadPageContent, downloadText } from "$lib/browser/download";
  import { formatBindingError } from "$lib/browser/binding-errors.js";
  import { micronShellStyle } from "$lib/browser/url";
  import {
    ensureMicronWasmReady,
    listAvailableMicronWasmParsers,
    parseMicronHeaderColors,
    renderClientMicronPage,
    resolveEffectiveMicronEngine,
    type MicronEffectiveEngine,
    type MicronRendererPreference,
  } from "$lib/micron/render-page";
  import { isWebAssemblySupported } from "$lib/micron/wasm-loader";
  import type { MicronWasmParserListEntry } from "$lib/micron/wasm-store";
  import PageContextMenu from "$lib/components/PageContextMenu.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type EditorParserPreference = Extract<MicronRendererPreference, "auto" | "js" | "wasm">;

  type Props = {
    source: string;
    currentURL: string;
    micronWasmEnabled: boolean;
    micronWasmParserId: string;
    micronWasmReady: boolean;
    onSourceChange: (source: string) => void;
    onNavigate: (url: string) => void;
  };

  let {
    source,
    currentURL,
    micronWasmEnabled,
    micronWasmParserId,
    micronWasmReady,
    onSourceChange,
    onNavigate,
  }: Props = $props();

  const SNAP_POINTS = [30, 40, 50, 60, 70];
  const MIN_RATIO = 25;
  const MAX_RATIO = 75;
  const SNAP_THRESHOLD = 4;

  let previewHtml = $state("");
  let pageFg = $state("");
  let pageBg = $state("");
  let previewError = $state("");
  let previewEl: HTMLElement | undefined = $state();
  let menu = $state<{ x: number; y: number } | null>(null);
  let renderRaf: number | undefined;
  let sourceInput: HTMLTextAreaElement | undefined = $state();
  let splitEl = $state<HTMLDivElement | null>(null);
  let dividerEl = $state<HTMLButtonElement | null>(null);
  let sourceRatio = $state(50);
  let dragging = $state(false);
  let verticalLayout = $state(false);
  let parserChoice = $state("auto");
  let wasmParsers = $state<MicronWasmParserListEntry[]>([]);
  let localWasmReady = $state(false);

  const wasmSupported = isWebAssemblySupported();
  const shellStyle = $derived(micronShellStyle("micron", pageFg, pageBg));
  const parserOptions = $derived.by(() => {
    const options = [
      { value: "auto", label: t("settings.rendererAuto") },
      { value: "js", label: t("settings.rendererJs") },
    ];
    if (wasmSupported && micronWasmEnabled) {
      for (const parser of wasmParsers) {
        options.push({
          value: `wasm:${parser.id}`,
          label: `${t("settings.rendererWasm")} · ${parser.label}`,
        });
      }
    }
    return options;
  });

  function parsedChoice(): { engine: EditorParserPreference; wasmId: string } {
    if (parserChoice.startsWith("wasm:")) {
      return { engine: "wasm", wasmId: parserChoice.slice(5) };
    }
    return {
      engine: parserChoice as EditorParserPreference,
      wasmId: micronWasmParserId,
    };
  }

  function snapRatio(value: number): number {
    const clamped = Math.min(MAX_RATIO, Math.max(MIN_RATIO, value));
    let closest = SNAP_POINTS[0];
    let minDist = Math.abs(clamped - closest);
    for (const point of SNAP_POINTS) {
      const dist = Math.abs(clamped - point);
      if (dist < minDist) {
        minDist = dist;
        closest = point;
      }
    }
    if (minDist <= SNAP_THRESHOLD) {
      return closest;
    }
    return clamped;
  }

  function updateLayoutMode() {
    verticalLayout = (splitEl?.clientWidth ?? 0) <= 900;
  }

  function resolvePreviewEngine(): MicronEffectiveEngine {
    const { engine } = parsedChoice();
    return resolveEffectiveMicronEngine(engine, {
      wasmEnabled: micronWasmEnabled,
      wasmAvailable: wasmSupported,
      wasmReady: localWasmReady || micronWasmReady,
      hasServerHtml: false,
    });
  }

  async function renderPreviewNow() {
    try {
      const { engine: preference, wasmId } = parsedChoice();
      if (preference === "wasm" && micronWasmEnabled && wasmSupported) {
        localWasmReady = await ensureMicronWasmReady(micronWasmEnabled, wasmId);
        if (!localWasmReady) {
          previewError = t("editor.parserWasmUnavailable");
          previewHtml = renderClientMicronPage(previewURL(), source, "js");
          const colors = parseMicronHeaderColors(source);
          pageFg = colors.fg;
          pageBg = colors.bg;
          return;
        }
      }
      const engine = resolvePreviewEngine();
      previewHtml = renderClientMicronPage(previewURL(), source, engine);
      const colors = parseMicronHeaderColors(source);
      pageFg = colors.fg;
      pageBg = colors.bg;
      previewError = "";
    } catch (err) {
      previewError = formatBindingError(err, "Preview failed");
      previewHtml = "";
    }
  }

  function previewURL() {
    return currentURL === "editor:" || currentURL === "editor"
      ? "editor:/page/editor.mu"
      : currentURL;
  }

  function scheduleRender() {
    if (renderRaf !== undefined) {
      cancelAnimationFrame(renderRaf);
    }
    renderRaf = requestAnimationFrame(() => {
      renderRaf = undefined;
      void renderPreviewNow();
    });
  }

  async function onParserChange(event: Event) {
    parserChoice = (event.currentTarget as HTMLSelectElement).value;
    localWasmReady = false;
    await renderPreviewNow();
  }

  function onInput(event: Event) {
    const value = (event.currentTarget as HTMLTextAreaElement).value;
    onSourceChange(value);
    scheduleRender();
  }

  function onDividerPointerDown(event: PointerEvent) {
    if (!dividerEl) {
      return;
    }
    dragging = true;
    dividerEl.setPointerCapture(event.pointerId);
  }

  function onDividerPointerMove(event: PointerEvent) {
    if (!dragging || !splitEl) {
      return;
    }
    const rect = splitEl.getBoundingClientRect();
    const next = verticalLayout
      ? ((event.clientY - rect.top) / rect.height) * 100
      : ((event.clientX - rect.left) / rect.width) * 100;
    sourceRatio = Math.min(MAX_RATIO, Math.max(MIN_RATIO, next));
  }

  function onDividerPointerUp(event: PointerEvent) {
    if (dividerEl?.hasPointerCapture(event.pointerId)) {
      dividerEl.releasePointerCapture(event.pointerId);
    }
    if (dragging) {
      sourceRatio = snapRatio(sourceRatio);
    }
    dragging = false;
  }

  $effect(() => {
    void source;
    void currentURL;
    void parserChoice;
    void micronWasmEnabled;
    void micronWasmParserId;
    void micronWasmReady;
    scheduleRender();
    return () => {
      if (renderRaf !== undefined) {
        cancelAnimationFrame(renderRaf);
      }
    };
  });

  $effect(() => {
    if (!splitEl || typeof ResizeObserver === "undefined") {
      return;
    }
    updateLayoutMode();
    const observer = new ResizeObserver(() => updateLayoutMode());
    observer.observe(splitEl);
    return () => observer.disconnect();
  });

  onMount(() => {
    void listAvailableMicronWasmParsers().then((entries) => {
      wasmParsers = entries;
    });
  });

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY };
  }

  function closeMenu() {
    menu = null;
  }

  function exportIndexMu() {
    if (!source.trim()) {
      return;
    }
    downloadText("index.mu", source, "text/plain");
  }

  async function downloadPage() {
    closeMenu();
    await downloadPageContent(currentURL, "editor", source);
  }

  async function handlePreviewClick(event: MouseEvent) {
    if (!previewEl) {
      return;
    }
    try {
      await handlePageLinkClick(event, previewEl, currentURL, onNavigate);
    } catch (err) {
      console.error("[MicronEditor] link click failed", err);
      previewError = formatBindingError(err, "Preview failed");
    }
  }
</script>

<section class="editor">
  <div
    class="split"
    class:vertical={verticalLayout}
    class:dragging
    bind:this={splitEl}
    style:--source-ratio="{sourceRatio}%"
  >
    <div class="pane source-pane">
      <label class="pane-label" for="micron-source">{t("editor.source")}</label>
      <textarea
        id="micron-source"
        class="source-input"
        value={source}
        spellcheck="false"
        bind:this={sourceInput}
        oninput={onInput}
        oncontextmenu={openMenu}
      ></textarea>
    </div>

    <button
      type="button"
      class="divider"
      class:vertical={verticalLayout}
      bind:this={dividerEl}
      aria-label={t("editor.resizePanes")}
      onpointerdown={onDividerPointerDown}
      onpointermove={onDividerPointerMove}
      onpointerup={onDividerPointerUp}
      onpointercancel={onDividerPointerUp}
    ></button>

    <div class="pane preview-pane">
      <div class="pane-header">
        <span class="pane-label">{t("editor.preview")}</span>
        <div class="preview-controls">
          <label class="parser-select">
            <span class="parser-label">{t("editor.parser")}</span>
            <select class="ren-select" value={parserChoice} onchange={onParserChange}>
              {#each parserOptions as option (option.value)}
                <option value={option.value}>{option.label}</option>
              {/each}
            </select>
          </label>
          <button
            type="button"
            class="export-btn ren-icon-btn"
            aria-label={t("editor.exportIndexMu")}
            title={t("editor.exportIndexMu")}
            disabled={!source.trim()}
            onclick={exportIndexMu}
          >
            <Download size={14} />
          </button>
        </div>
      </div>
      {#if previewError}
        <div class="preview-error">{previewError}</div>
      {/if}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div
        class="preview"
        style={shellStyle}
        bind:this={previewEl}
        onclick={handlePreviewClick}
        oncontextmenu={openMenu}
        role="document"
      >
        {@html previewHtml}
      </div>
    </div>
  </div>
</section>

{#if menu}
  <PageContextMenu
    x={menu.x}
    y={menu.y}
    canViewSource={source.length > 0}
    onViewSource={() => {
      closeMenu();
      sourceInput?.focus();
    }}
    onDownload={downloadPage}
    onClose={closeMenu}
  />
{/if}

<style>
  .editor {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: var(--ren-content-bg);
  }

  .split {
    flex: 1;
    min-height: 0;
    display: flex;
    width: 100%;
  }

  .split.vertical {
    flex-direction: column;
  }

  .split.dragging {
    user-select: none;
    cursor: col-resize;
  }

  .split.vertical.dragging {
    cursor: row-resize;
  }

  .pane {
    min-width: 0;
    min-height: 0;
    display: grid;
    grid-template-rows: auto 1fr;
    overflow: hidden;
  }

  .source-pane {
    flex: 0 0 var(--source-ratio);
  }

  .preview-pane {
    flex: 1;
    background: #000;
  }

  .split.vertical .source-pane {
    flex: 0 0 var(--source-ratio);
    width: 100%;
  }

  .split.vertical .preview-pane {
    flex: 1;
    width: 100%;
  }

  .divider {
    flex: 0 0 5px;
    cursor: col-resize;
    background: transparent;
    position: relative;
    border: none;
    padding: 0;
    z-index: 1;
  }

  .divider.vertical {
    width: 100%;
    height: 5px;
    cursor: row-resize;
    flex: 0 0 5px;
  }

  .divider::after {
    content: "";
    position: absolute;
    inset: 0;
    background: var(--ren-border);
    opacity: 0.55;
    transition: opacity 0.12s ease;
  }

  .divider:hover::after,
  .split.dragging .divider::after {
    opacity: 1;
    background: var(--ren-accent);
  }

  .pane-label {
    padding: 0.45rem 0.85rem;
    color: var(--ren-muted);
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
  }

  .pane-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.35rem;
    border-bottom: 1px solid color-mix(in srgb, #ffffff 12%, transparent);
    background: #09090b;
  }

  .pane-header .pane-label {
    flex: 0 0 auto;
    border-bottom: none;
    color: color-mix(in srgb, #ffffff 55%, transparent);
    background: transparent;
  }

  .preview-controls {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    min-width: 0;
    flex: 1;
    justify-content: flex-end;
    padding-right: 0.35rem;
  }

  .parser-select {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    min-width: 0;
  }

  .parser-label {
    color: color-mix(in srgb, #ffffff 55%, transparent);
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    white-space: nowrap;
  }

  .parser-select select {
    max-width: 14rem;
    min-width: 0;
    width: auto;
    font-size: 0.78rem;
    padding-top: 0.2rem;
    padding-bottom: 0.2rem;
    padding-left: 0.35rem;
  }

  .export-btn {
    flex: 0 0 auto;
    color: color-mix(in srgb, #ffffff 70%, transparent);
  }

  .export-btn:hover:not(:disabled) {
    color: #fff;
    background: color-mix(in srgb, #ffffff 12%, transparent);
  }

  .export-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .source-input {
    width: 100%;
    height: 100%;
    resize: none;
    border: none;
    padding: 1rem;
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.88rem;
    line-height: 1.45;
    outline: none;
  }

  .preview {
    overflow: auto;
    padding: 1rem 1.25rem 2rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    line-height: 1.25;
    min-height: 0;
  }

  .preview-error {
    padding: 0.5rem 0.85rem;
    color: var(--ren-danger);
    font-size: 0.85rem;
    border-bottom: 1px solid color-mix(in srgb, #ffffff 12%, transparent);
  }
</style>
