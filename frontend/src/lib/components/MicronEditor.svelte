<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  /* eslint-disable svelte/no-at-html-tags -- live micron preview */
  import { RenderRaw } from "../../../bindings/renbrowser/internal/app/browserservice.js";
  import {
    resolveLinkURL,
    resolveMicronNavigation,
    resolveNomadDataURL,
  } from "$lib/browser/micron-links";
  import { micronShellStyle } from "$lib/browser/url";
  import {
    parseMicronHeaderColors,
    renderClientMicronPage,
    usesClientMicronRenderer,
    type MicronEffectiveEngine,
  } from "$lib/micron/render-page";
  import PageContextMenu from "$lib/components/PageContextMenu.svelte";
  import { downloadPageContent } from "$lib/browser/download";

  type Props = {
    source: string;
    currentURL: string;
    micronEngine?: MicronEffectiveEngine;
    micronWasmReady?: boolean;
    onSourceChange: (source: string) => void;
    onNavigate: (url: string) => void;
  };

  let {
    source,
    currentURL,
    micronEngine = "js",
    micronWasmReady = false,
    onSourceChange,
    onNavigate,
  }: Props = $props();

  let previewHtml = $state("");
  let pageFg = $state("");
  let pageBg = $state("");
  let previewError = $state("");
  let previewEl: HTMLElement | undefined = $state();
  let menu = $state<{ x: number; y: number } | null>(null);
  let renderTimer: ReturnType<typeof setTimeout> | undefined;
  let sourceInput: HTMLTextAreaElement | undefined = $state();

  const shellStyle = $derived(micronShellStyle("micron", pageFg, pageBg));

  async function renderPreview() {
    if (usesClientMicronRenderer(micronEngine)) {
      try {
        previewHtml = renderClientMicronPage(currentURL, source, micronEngine);
        const colors = parseMicronHeaderColors(source);
        pageFg = colors.fg;
        pageBg = colors.bg;
        previewError = "";
      } catch (err) {
        previewError = err instanceof Error ? err.message : String(err);
        previewHtml = "";
      }
      return;
    }
    try {
      const page = await RenderRaw("/page/editor.mu", source);
      previewHtml = page.html ?? "";
      pageFg = page.pageFg ?? "";
      pageBg = page.pageBg ?? "";
      previewError = page.error ?? "";
    } catch (err) {
      previewError = err instanceof Error ? err.message : String(err);
      previewHtml = "";
    }
  }

  function scheduleRender() {
    if (renderTimer) {
      clearTimeout(renderTimer);
    }
    renderTimer = setTimeout(() => {
      void renderPreview();
    }, 200);
  }

  function onInput(event: Event) {
    const value = (event.currentTarget as HTMLTextAreaElement).value;
    onSourceChange(value);
    scheduleRender();
  }

  $effect(() => {
    void source;
    void micronEngine;
    void micronWasmReady;
    scheduleRender();
    return () => {
      if (renderTimer) {
        clearTimeout(renderTimer);
      }
    };
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
  <div class="meta">
    <span>micron editor</span>
    <span>source + preview</span>
  </div>

  <div class="split">
    <div class="pane source-pane">
      <label class="pane-label" for="micron-source">Source</label>
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

    <div class="pane preview-pane">
      <div class="pane-label">Preview</div>
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

  .meta {
    display: flex;
    gap: 0.75rem;
    padding: 0.45rem 1rem;
    color: var(--ren-muted);
    border-bottom: 1px solid var(--ren-border);
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    background: var(--ren-chrome-bg);
  }

  .split {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  }

  .pane {
    min-height: 0;
    display: grid;
    grid-template-rows: auto 1fr;
    border-right: 1px solid var(--ren-border);
  }

  .preview-pane {
    border-right: none;
    background: #000;
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

  @media (max-width: 900px) {
    .split {
      grid-template-columns: 1fr;
      grid-template-rows: minmax(180px, 0.9fr) minmax(220px, 1.1fr);
    }

    .pane {
      border-right: none;
      border-bottom: 1px solid var(--ren-border);
    }

    .preview-pane {
      border-bottom: none;
    }
  }
</style>
