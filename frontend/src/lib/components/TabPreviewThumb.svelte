<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { LoaderCircle } from "@lucide/svelte";
  import { expandHexColor, micronPageColors, type Tab } from "$lib/browser/url";
  import { normalizePageErrorKind, pageErrorContent } from "$lib/browser/errors";
  import {
    PREVIEW_REF_HEIGHT,
    PREVIEW_REF_WIDTH,
    previewScaleForBox,
    wrapPreviewSrcdoc,
  } from "$lib/browser/preview-srcdoc";
  import {
    parseMicronHeaderColors,
    renderClientMicronPage,
    usesClientMicronRenderer,
    type MicronEffectiveEngine,
  } from "$lib/micron/render-page";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    tab: Tab;
    label?: string;
    class?: string;
    micronEngine?: MicronEffectiveEngine;
  };

  let { tab, label = "", class: className = "", micronEngine = "js" }: Props = $props();

  let thumbEl = $state<HTMLDivElement | null>(null);
  let boxWidth = $state(0);

  function previewLabel(): string {
    const title = tab.title.trim();
    if (title) {
      return title;
    }
    const url = tab.url.trim();
    if (url) {
      return url;
    }
    return t("tab.new");
  }

  const isMicron = $derived(tab.page?.contentType === "micron");

  const previewHtml = $derived.by(() => {
    const page = tab.page;
    if (!page || page.error) {
      return "";
    }
    const serverHtml = page.html?.trim() ?? "";
    if (serverHtml) {
      return serverHtml;
    }
    if (
      isMicron &&
      page.lastRaw?.trim() &&
      usesClientMicronRenderer(micronEngine) &&
      tab.url.trim()
    ) {
      try {
        return renderClientMicronPage(tab.url, page.lastRaw, micronEngine);
      } catch {
        return "";
      }
    }
    return "";
  });

  const hasPreviewHtml = $derived(previewHtml.length > 0);

  const previewColors = $derived.by(() => {
    const page = tab.page;
    if (!page) {
      return { fg: "", bg: "" };
    }
    let fg = page.pageFg ?? "";
    let bg = page.pageBg ?? "";
    if (isMicron && page.lastRaw?.trim()) {
      try {
        const parsed = parseMicronHeaderColors(page.lastRaw);
        if (parsed.fg) {
          fg = parsed.fg;
        }
        if (parsed.bg) {
          bg = parsed.bg;
        }
      } catch {
        // keep stored page colors
      }
    }
    if (isMicron) {
      return micronPageColors(fg, bg);
    }
    return {
      fg: fg.trim() ? `#${expandHexColor(fg)}` : "",
      bg: bg.trim() ? `#${expandHexColor(bg)}` : "",
    };
  });

  const previewSrcdoc = $derived(
    hasPreviewHtml
      ? wrapPreviewSrcdoc(previewHtml, {
          fg: previewColors.fg,
          bg: previewColors.bg,
        })
      : "",
  );

  const previewScale = $derived(previewScaleForBox(boxWidth));

  const displayLabel = $derived(label || previewLabel());

  $effect(() => {
    const el = thumbEl;
    if (!el) {
      return;
    }
    const sync = () => {
      boxWidth = el.clientWidth;
    };
    sync();
    const observer = new ResizeObserver(() => {
      sync();
    });
    observer.observe(el);
    return () => observer.disconnect();
  });
</script>

<div
  bind:this={thumbEl}
  class="thumb {className}"
  class:has-html={hasPreviewHtml}
  style:background={previewColors.bg || "var(--ren-surface-muted)"}
  style:color={previewColors.fg || "var(--ren-fg)"}
>
  {#if tab.loading}
    <div class="thumb-state">
      <span class="thumb-spinner">
        <LoaderCircle size={20} />
      </span>
      <span>{t("common.loading")}</span>
    </div>
  {:else if tab.page?.error}
    <div class="thumb-state error">
      <span class="thumb-state-title">
        {pageErrorContent(
          normalizePageErrorKind(tab.page.errorKind, tab.page.error),
          tab.page.error,
        ).title}
      </span>
    </div>
  {:else if hasPreviewHtml}
    <div class="thumb-viewport" aria-hidden="true">
      <iframe
        class="thumb-iframe"
        title={displayLabel}
        srcdoc={previewSrcdoc}
        sandbox=""
        tabindex="-1"
        style:width="{PREVIEW_REF_WIDTH}px"
        style:height="{PREVIEW_REF_HEIGHT}px"
        style:transform="scale({previewScale})"
      ></iframe>
    </div>
  {:else}
    <span class="thumb-title">{displayLabel}</span>
  {/if}
</div>

<style>
  .thumb {
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    width: 100%;
    min-height: 0;
    background: var(--ren-surface-muted);
  }

  .thumb.has-html {
    padding: 0;
  }

  .thumb-viewport {
    position: relative;
    width: 100%;
    height: 100%;
    min-height: inherit;
    overflow: hidden;
  }

  .thumb-iframe {
    position: absolute;
    top: 0;
    left: 0;
    border: none;
    transform-origin: top left;
    pointer-events: none;
    background: transparent;
  }

  .thumb-state {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    text-align: center;
    padding: 0.5rem;
    font-size: 0.78rem;
    color: var(--ren-muted);
  }

  .thumb-state.error {
    color: var(--ren-danger);
  }

  .thumb-state-title {
    font-weight: 600;
    font-size: 0.82rem;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .thumb-spinner {
    display: inline-flex;
    animation: spin 0.9s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .thumb-title {
    font-weight: 600;
    font-size: 0.88rem;
    line-height: 1.3;
    padding: 0.75rem;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
