<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  /* eslint-disable svelte/no-at-html-tags -- live micron preview */
  import {
    resolveLinkURL,
    resolveMicronNavigation,
    resolveNomadDataURL,
  } from "$lib/browser/micron-links";
  import { micronShellStyle } from "$lib/browser/url";
  import { parseMicronHeaderColors, renderClientMicronPage } from "$lib/micron/render-page";
  import PageContextMenu from "$lib/components/PageContextMenu.svelte";
  import { downloadPageContent } from "$lib/browser/download";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    source: string;
    currentURL: string;
    onSourceChange: (source: string) => void;
    onNavigate: (url: string) => void;
  };

  let { source, currentURL, onSourceChange, onNavigate }: Props = $props();

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
  let renderTimer: ReturnType<typeof setTimeout> | undefined;
  let sourceInput: HTMLTextAreaElement | undefined = $state();
  let splitEl = $state<HTMLDivElement | null>(null);
  let dividerEl = $state<HTMLButtonElement | null>(null);
  let sourceRatio = $state(50);
  let dragging = $state(false);
  let verticalLayout = $state(false);

  const shellStyle = $derived(micronShellStyle("micron", pageFg, pageBg));

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

  function renderPreviewNow() {
    try {
      const renderURL =
        currentURL === "editor:" || currentURL === "editor"
          ? "editor:/page/editor.mu"
          : currentURL;
      previewHtml = renderClientMicronPage(renderURL, source, "js");
      const colors = parseMicronHeaderColors(source);
      pageFg = colors.fg;
      pageBg = colors.bg;
      previewError = "";
    } catch (err) {
      previewError = err instanceof Error ? err.message : String(err);
      previewHtml = "";
    }
  }

  function scheduleRender() {
    if (renderTimer) {
      clearTimeout(renderTimer);
    }
    renderTimer = setTimeout(renderPreviewNow, 200);
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
    scheduleRender();
    return () => {
      if (renderTimer) {
        clearTimeout(renderTimer);
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

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY };
  }

  function closeMenu() {
    menu = null;
  }

  async function downloadPage() {
    closeMenu();
    await downloadPageContent(currentURL, "editor", source);
  }

  async function handlePreviewClick(event: MouseEvent) {
    const target = event.target as HTMLElement | null;
    if (!target || !previewEl) {
      return;
    }

    const nodeLink = target.closest("[data-action='openNode']");
    if (nodeLink) {
      event.preventDefault();
      const destination = nodeLink.getAttribute("data-destination");
      if (!destination) {
        return;
      }
      const fieldsSpec = nodeLink.getAttribute("data-fields");
      const next = await resolveMicronNavigation(previewEl, currentURL, destination, fieldsSpec);
      if (next) {
        onNavigate(next);
      }
      return;
    }

    const nomadAnchor = target.closest("a[data-nomad-url]");
    if (nomadAnchor) {
      event.preventDefault();
      const dataUrl = nomadAnchor.getAttribute("data-nomad-url");
      if (dataUrl) {
        onNavigate(resolveNomadDataURL(currentURL, dataUrl));
      }
      return;
    }

    const anchor = target.closest("a");
    if (!anchor) {
      return;
    }
    const href = anchor.getAttribute("href");
    if (!href || href.startsWith("http://") || href.startsWith("https://")) {
      return;
    }
    event.preventDefault();
    const next = resolveLinkURL(currentURL, href);
    if (next) {
      onNavigate(next);
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
      <div class="pane-label">{t("editor.preview")}</div>
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

  .preview-pane .pane-label {
    border-bottom-color: color-mix(in srgb, #ffffff 12%, transparent);
    color: color-mix(in srgb, #ffffff 55%, transparent);
    background: #09090b;
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
