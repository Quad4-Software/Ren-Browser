<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { LoaderCircle } from "@lucide/svelte";
  import { type Tab } from "$lib/browser/url";
  import { normalizePageErrorKind, pageErrorContent } from "$lib/browser/errors";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    tab: Tab;
    label?: string;
    class?: string;
  };

  let { tab, label = "", class: className = "" }: Props = $props();

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

  function hasPreviewHtml(): boolean {
    return !!tab.page?.html?.trim() && !tab.page?.error;
  }

  const displayLabel = $derived(label || previewLabel());
</script>

<div
  class="thumb {className}"
  class:has-html={hasPreviewHtml()}
  style:background={tab.page?.pageBg || "var(--ren-surface-muted)"}
  style:color={tab.page?.pageFg || "var(--ren-fg)"}
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
  {:else if hasPreviewHtml()}
    <iframe
      class="thumb-iframe"
      title={displayLabel}
      srcdoc={tab.page?.html ?? ""}
      sandbox=""
      tabindex="-1"
    ></iframe>
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
    background: var(--ren-surface-muted);
  }

  .thumb.has-html {
    padding: 0;
  }

  .thumb-iframe {
    width: 400%;
    height: 400%;
    border: none;
    transform: scale(0.25);
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
