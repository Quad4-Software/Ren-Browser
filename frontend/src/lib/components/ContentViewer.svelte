<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  /* eslint-disable svelte/no-at-html-tags -- renders trusted mesh page content */
  import { ArrowLeft, FileCode, Globe, X } from "@lucide/svelte";
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
  import PageFindBar from "$lib/components/PageFindBar.svelte";
  import PageContextMenu from "$lib/components/PageContextMenu.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import PageErrorState from "$lib/components/PageErrorState.svelte";
  import { displayName } from "$lib/brand";
  import { t } from "$lib/i18n/i18n.svelte";
  import { downloadPageContent, isFileURL } from "$lib/browser/download";

  type Props = {
    html: string;
    contentType: string;
    loading: boolean;
    error: string;
    errorKind?: string;
    currentURL: string;
    raw?: string;
    pageFg?: string;
    pageBg?: string;
    fromCache?: boolean;
    cachedAt?: number;
    showSource?: boolean;
    findOpen?: boolean;
    micronEngine?: MicronEffectiveEngine;
    onNavigate: (url: string) => void;
    onRetry: () => void;
    onReloadFresh: () => void;
    onShowSourceChange: (show: boolean) => void;
    onFindClose?: () => void;
  };

  let {
    html,
    contentType,
    loading,
    error,
    errorKind = "",
    currentURL,
    raw = "",
    pageFg = "",
    pageBg = "",
    fromCache = false,
    cachedAt = 0,
    showSource = false,
    findOpen = false,
    micronEngine = "js",
    onNavigate,
    onRetry,
    onReloadFresh,
    onShowSourceChange,
    onFindClose = () => {},
  }: Props = $props();

  let contentEl: HTMLElement | undefined = $state();
  let menu = $state<{ x: number; y: number } | null>(null);
  let dismissedCacheKey = $state("");

  const cacheBannerKey = $derived(`${fromCache}:${cachedAt}`);
  const isMicron = $derived(contentType === "micron");
  const isAbout = $derived(contentType === "about");
  const isLicense = $derived(contentType === "license");
  const isDocs = $derived(contentType === "docs");
  const isInternalPage = $derived(isAbout || isLicense || isDocs);

  const displayHtml = $derived.by(() => {
    if (showSource || !isMicron || !usesClientMicronRenderer(micronEngine) || !raw.trim()) {
      return html;
    }
    try {
      return renderClientMicronPage(currentURL, raw, micronEngine);
    } catch {
      return html;
    }
  });

  const displayFg = $derived.by(() => {
    if (showSource || !isMicron || !usesClientMicronRenderer(micronEngine) || !raw.trim()) {
      return pageFg;
    }
    try {
      return parseMicronHeaderColors(raw).fg;
    } catch {
      return pageFg;
    }
  });

  const displayBg = $derived.by(() => {
    if (showSource || !isMicron || !usesClientMicronRenderer(micronEngine) || !raw.trim()) {
      return pageBg;
    }
    try {
      return parseMicronHeaderColors(raw).bg;
    } catch {
      return pageBg;
    }
  });

  const shellStyle = $derived(micronShellStyle(contentType, displayFg, displayBg));
  const showCacheBanner = $derived(
    fromCache &&
      dismissedCacheKey !== cacheBannerKey &&
      !loading &&
      !error &&
      (displayHtml || showSource),
  );
  const canViewSource = $derived(raw.trim().length > 0);
  const cacheLabel = $derived(
    cachedAt > 0 ? new Date(cachedAt).toLocaleString() : t("common.recently"),
  );

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY };
  }

  function closeMenu() {
    menu = null;
  }

  async function downloadPage() {
    closeMenu();
    const payload = raw || html;
    if (!payload && !isFileURL(currentURL)) {
      return;
    }
    await downloadPageContent(currentURL, contentType, payload);
  }

  async function handleClick(event: MouseEvent) {
    if (showSource) {
      return;
    }
    const target = event.target as HTMLElement | null;
    if (!target || !contentEl) {
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
      const next = await resolveMicronNavigation(contentEl, currentURL, destination, fieldsSpec);
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
    if (!next) {
      return;
    }
    if (isFileURL(next)) {
      await downloadPageContent(next, "file", "");
      return;
    }
    onNavigate(next);
  }
</script>

<section class="viewer" class:micron={isMicron && !showSource} class:about={isInternalPage}>
  <PageFindBar open={findOpen && !showSource} onClose={onFindClose} contentRoot={contentEl} />

  {#if showCacheBanner}
    <div class="cache-banner">
      <span>{t("content.cachedBanner", { when: cacheLabel })}</span>
      <div class="cache-actions">
        <button type="button" onclick={onReloadFresh}>{t("content.loadFresh")}</button>
        <button
          type="button"
          class="cache-dismiss"
          aria-label={t("content.dismissCache")}
          onclick={() => (dismissedCacheKey = cacheBannerKey)}
        >
          <X size={17} />
        </button>
      </div>
    </div>
  {/if}

  {#if showSource && canViewSource}
    <div class="source-bar">
      <button type="button" class="back-btn" onclick={() => onShowSourceChange(false)}>
        <ArrowLeft size={14} />
        <span>{t("content.backToPage")}</span>
      </button>
      <span class="source-label">
        <FileCode size={14} />
        {t("content.pageSource")}
      </span>
    </div>
    <pre class="source-view" oncontextmenu={openMenu}>{raw}</pre>
  {:else if loading}
    <div class="progress" aria-hidden="true"></div>
    <div class="state">{t("content.loadingPage")}</div>
  {:else if error}
    <PageErrorState {error} {errorKind} {currentURL} {onRetry} />
  {:else if displayHtml}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
      class="content"
      class:micron={isMicron}
      data-content-type={contentType}
      style={shellStyle}
      bind:this={contentEl}
      onclick={handleClick}
      oncontextmenu={openMenu}
      role="document"
    >
      {@html displayHtml}
    </div>
  {:else}
    <div class="state">
      <EmptyState
        title={displayName}
        description={t("content.emptyDescription")}
      >
        <Globe size={22} />
      </EmptyState>
    </div>
  {/if}
</section>

{#if menu}
  <PageContextMenu
    x={menu.x}
    y={menu.y}
    {canViewSource}
    onViewSource={() => {
      closeMenu();
      onShowSourceChange(true);
    }}
    onDownload={downloadPage}
    onClose={closeMenu}
  />
{/if}

<style>
  .viewer {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: var(--ren-content-bg);
    position: relative;
  }

  .viewer.micron {
    background: #000;
  }

  .viewer.about {
    background: var(--ren-content-bg);
  }

  .cache-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.55rem 1rem;
    background: color-mix(in srgb, var(--ren-accent) 12%, var(--ren-chrome-bg));
    border-bottom: 1px solid var(--ren-border);
    color: var(--ren-fg);
    font-size: 0.88rem;
  }

  .cache-banner button:not(.cache-dismiss) {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.35rem 0.7rem;
    font: inherit;
    font-size: 0.82rem;
    cursor: pointer;
    white-space: nowrap;
  }

  .cache-actions {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    flex-shrink: 0;
  }

  .cache-dismiss {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.15rem;
    border: none;
    background: transparent;
    color: var(--ren-muted);
    cursor: pointer;
    line-height: 0;
    transition: color 0.12s ease;
  }

  .cache-dismiss:hover {
    color: var(--ren-fg);
  }

  .source-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.55rem 1rem;
    border-bottom: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
  }

  .back-btn,
  .source-label {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.88rem;
  }

  .back-btn {
    border: none;
    background: transparent;
    color: var(--ren-fg);
    cursor: pointer;
    padding: 0.2rem 0.35rem;
    border-radius: 8px;
  }

  .back-btn:hover {
    background: var(--ren-tab-hover);
  }

  .source-label {
    color: var(--ren-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.78rem;
  }

  .source-view {
    flex: 1;
    margin: 0;
    padding: 1rem 1.25rem 2rem;
    overflow: auto;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.88rem;
    line-height: 1.45;
    white-space: pre-wrap;
    word-break: break-word;
    background: var(--ren-input-bg);
    color: var(--ren-fg);
  }

  .progress {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--ren-accent), transparent);
    animation: pulse 1.1s ease-in-out infinite;
  }

  @keyframes pulse {
    0% {
      transform: translateX(-100%);
    }
    100% {
      transform: translateX(100%);
    }
  }

  .content {
    flex: 1;
    overflow: auto;
    padding: 1rem 1.25rem 2rem;
    line-height: 1.55;
  }

  .content.micron {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    line-height: 1.25;
  }

  .content.micron :global(input[type="text"]),
  .content.micron :global(input[type="password"]),
  .content.micron :global(textarea) {
    font-family: inherit;
    font-size: 1em;
    line-height: 1;
    padding: 0;
    margin: 0;
    border: 0;
    border-bottom: 1px solid currentColor;
    border-radius: 0;
    background: transparent;
    color: inherit;
    caret-color: currentColor;
    box-sizing: content-box;
  }

  .state {
    margin: auto;
    text-align: center;
    color: var(--ren-muted);
    padding: 2rem;
  }

  .content :global(mark.ren-find-hit) {
    background: color-mix(in srgb, var(--ren-accent) 45%, transparent);
    color: inherit;
    border-radius: 2px;
  }

  .content :global(mark.ren-find-hit.ren-find-active) {
    outline: 2px solid var(--ren-accent);
  }

  .content :global(.about-page) {
    max-width: 40rem;
    margin: 0 auto;
  }

  .content :global(.about-page h1) {
    margin: 0 0 0.35rem;
    font-size: 1.6rem;
  }

  .content :global(.about-tagline) {
    color: var(--ren-muted);
    margin: 0 0 1.25rem;
  }

  .content :global(.about-table) {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 1rem;
  }

  .content :global(.about-table th),
  .content :global(.about-table td) {
    text-align: left;
    padding: 0.55rem 0.65rem;
    border-bottom: 1px solid var(--ren-border);
    vertical-align: top;
  }

  .content :global(.about-table th) {
    width: 9rem;
    color: var(--ren-muted);
    font-weight: 500;
  }

  .content :global(.about-hint) {
    color: var(--ren-muted);
    font-size: 0.9rem;
  }

  .content :global(.docs-page) {
    max-width: 48rem;
    margin: 0 auto;
    line-height: 1.55;
  }

  .content :global(.docs-nav) {
    margin-bottom: 1rem;
    color: var(--ren-muted);
    font-size: 0.9rem;
  }

  .content :global(.docs-lang-list) {
    padding-left: 1.25rem;
  }

  .content :global(.docs-body h1) {
    margin-top: 0;
    font-size: 1.6rem;
  }

  .content :global(.docs-body h2) {
    margin-top: 1.5rem;
    font-size: 1.2rem;
  }

  .content :global(.docs-body pre) {
    overflow-x: auto;
    padding: 0.75rem;
    border-radius: 6px;
    background: color-mix(in srgb, var(--ren-bg) 85%, var(--ren-fg));
  }

  .content :global(.docs-hint) {
    color: var(--ren-muted);
    font-size: 0.9rem;
  }

  .content :global(.license-page) {
    max-width: 48rem;
    margin: 0 auto;
  }

  .content :global(.license-page h1) {
    margin: 0 0 0.35rem;
    font-size: 1.6rem;
  }

  .content :global(.license-spdx) {
    color: var(--ren-muted);
    margin: 0 0 1rem;
    font-size: 0.9rem;
  }

  .content :global(.license-text) {
    white-space: pre-wrap;
    word-break: break-word;
    font-family: var(--ren-mono, ui-monospace, monospace);
    font-size: 0.82rem;
    line-height: 1.5;
    padding: 1rem;
    border: 1px solid var(--ren-border);
    border-radius: 8px;
    background: color-mix(in srgb, var(--ren-chrome-bg) 70%, transparent);
    overflow-x: auto;
  }

  .content :global(.license-hint) {
    color: var(--ren-muted);
    font-size: 0.9rem;
    margin-top: 1rem;
  }
</style>
