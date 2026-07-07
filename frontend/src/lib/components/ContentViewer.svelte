<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  /* eslint-disable svelte/no-at-html-tags -- renders trusted mesh page content */
  import { ArrowLeft, FileCode, Globe, X } from "@lucide/svelte";
  import { handlePageLinkClick } from "$lib/browser/page-links";
  import { micronShellStyle } from "$lib/browser/url";
  import { renderDocsPage } from "$lib/browser/docs-render";
  import {
    parseMicronHeaderColors,
    renderClientMicronPage,
    usesClientMicronRenderer,
    type MicronEffectiveEngine,
  } from "$lib/micron/render-page";
  import { attachMicronMultilineExpansion } from "$lib/micron/multiline";
  import PageFindBar from "$lib/components/PageFindBar.svelte";
  import PageContextMenu from "$lib/components/PageContextMenu.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import PageErrorState from "$lib/components/PageErrorState.svelte";
  import DocumentViewer from "$lib/components/DocumentViewer.svelte";
  import { displayName } from "$lib/brand";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    downloadFailureMessage,
    downloadPageContent,
    isDownloadCanceledError,
    isFileURL,
    pageDownloadName,
    type DownloadResult,
  } from "$lib/browser/download";
  import { isDocumentContentType } from "$lib/documents/types";
  import {
    attachMobileGestures,
    type MobileGestureProgress,
  } from "$lib/browser/mobile-gestures.js";

  type Props = {
    html: string;
    contentType: string;
    loading: boolean;
    error: string;
    errorKind?: string;
    currentURL: string;
    raw?: string;
    binaryB64?: string;
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
    onDownloadResult?: (result: DownloadResult) => void;
    mobileGestures?: boolean;
    canGoBack?: boolean;
    onBack?: () => void;
  };

  let {
    html,
    contentType,
    loading,
    error,
    errorKind = "",
    currentURL,
    raw = "",
    binaryB64 = "",
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
    onDownloadResult = () => {},
    mobileGestures = false,
    canGoBack = false,
    onBack = () => {},
  }: Props = $props();

  let viewerEl: HTMLElement | undefined = $state();
  let contentEl: HTMLElement | undefined = $state();
  let menu = $state<{ x: number; y: number } | null>(null);
  let dismissedCacheKey = $state("");
  let multilineHintVisible = $state(false);
  let gesture = $state<MobileGestureProgress>({
    pullOffset: 0,
    pullTriggered: false,
    backOffset: 0,
    backTriggered: false,
  });

  const cacheBannerKey = $derived(`${fromCache}:${cachedAt}`);
  const isDocument = $derived(isDocumentContentType(contentType));
  const isMicron = $derived(contentType === "micron");
  const isAbout = $derived(contentType === "about");
  const isLicense = $derived(contentType === "license");
  const isDocs = $derived(contentType === "docs");
  const isInternalPage = $derived(isAbout || isLicense || isDocs);

  const displayHtml = $derived.by(() => {
    if (isDocs && raw.trim()) {
      return renderDocsPage(raw, currentURL);
    }
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
  const gestureTransform = $derived.by(() => {
    if (gesture.backOffset > 0) {
      return `translateX(${gesture.backOffset}px)`;
    }
    if (gesture.pullOffset > 0) {
      return `translateY(${gesture.pullOffset}px)`;
    }
    return "";
  });

  function openMenu(event: MouseEvent) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY };
  }

  function closeMenu() {
    menu = null;
  }

  async function runDownload(url: string, contentTypeForSave: string, payload: string) {
    const isFile = isFileURL(url);
    if (isFile) {
      // Mesh file fetches can take a while; confirm the click landed right
      // away so it doesn't look like nothing happened while it's in flight.
      onDownloadResult({ ok: true, pending: true, message: t("downloads.downloading") });
    }
    try {
      await downloadPageContent(url, contentTypeForSave, payload);
      onDownloadResult({
        ok: true,
        message: isFile ? t("downloads.fileSaved") : t("downloads.saved"),
      });
    } catch (err) {
      console.error("[ContentViewer] download failed", url, err);
      if (isDownloadCanceledError(err)) {
        onDownloadResult({
          ok: false,
          message: "",
          canceled: true,
          name: pageDownloadName(url, contentTypeForSave),
        });
        return;
      }
      onDownloadResult({
        ok: false,
        message: downloadFailureMessage(err, t("downloads.downloadFailed")),
      });
    }
  }

  async function downloadPage() {
    closeMenu();
    const payload = raw || html;
    if (!payload && !isFileURL(currentURL)) {
      return;
    }
    await runDownload(currentURL, contentType, payload);
  }

  async function handleClick(event: MouseEvent) {
    if (showSource || !contentEl) {
      return;
    }
    try {
      await handlePageLinkClick(event, contentEl, currentURL, async (next) => {
        if (isFileURL(next)) {
          await runDownload(next, "file", "");
          return;
        }
        onNavigate(next);
      });
    } catch (err) {
      console.error("[ContentViewer] link click failed", err);
      if (isDownloadCanceledError(err)) {
        onDownloadResult({
          ok: false,
          message: "",
          canceled: true,
          name: pageDownloadName(currentURL, contentType),
        });
        return;
      }
      onDownloadResult({
        ok: false,
        message: downloadFailureMessage(err, t("downloads.downloadFailed")),
      });
    }
  }

  $effect(() => {
    const root = contentEl;
    const active = isMicron && !showSource && !loading && !error && Boolean(displayHtml);

    if (!root || !active) {
      multilineHintVisible = false;
      return;
    }

    const expansion = attachMicronMultilineExpansion(root, {
      onArmed: () => {
        multilineHintVisible = true;
      },
      onDisarmed: () => {
        multilineHintVisible = false;
      },
      onExpanded: () => {
        multilineHintVisible = false;
      },
    });

    return () => expansion.teardown();
  });

  $effect(() => {
    const surface = viewerEl;
    const enabled = mobileGestures;
    if (!surface || !enabled) {
      gesture = {
        pullOffset: 0,
        pullTriggered: false,
        backOffset: 0,
        backTriggered: false,
      };
      return;
    }

    const attachment = attachMobileGestures(surface, {
      getCanGoBack: () => canGoBack,
      getScrollTop: () => contentEl?.scrollTop ?? 0,
      isActive: () => !loading && !showSource,
      onRefresh: onRetry,
      onBack,
      onProgress: (progress) => {
        gesture = progress;
      },
    });

    return () => attachment.teardown();
  });
</script>

<section
  class="viewer"
  class:micron={isMicron && !showSource}
  class:about={isInternalPage}
  class:mobile-gestures={mobileGestures}
  bind:this={viewerEl}
>
  {#if mobileGestures && gesture.backOffset > 0}
    <div
      class="gesture-back-indicator"
      class:triggered={gesture.backTriggered}
      aria-hidden="true"
    ></div>
  {/if}

  {#if mobileGestures && gesture.pullOffset > 0}
    <div class="gesture-pull-indicator" class:triggered={gesture.pullTriggered} aria-hidden="true">
      <span
        >{gesture.pullTriggered ? t("content.releaseToRefresh") : t("content.pullToRefresh")}</span
      >
    </div>
  {/if}

  <div class="gesture-body" style:transform={gestureTransform}>
    <PageFindBar open={findOpen && !showSource && !isDocument} onClose={onFindClose} contentRoot={contentEl} />

    {#if showCacheBanner}
      <div class="cache-banner">
        <span class="cache-text">{t("content.cachedBanner", { when: cacheLabel })}</span>
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
    {:else if isDocument}
      <DocumentViewer
        {contentType}
        {binaryB64}
        pageError={error}
        onRetry={onRetry}
      />
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
        <EmptyState title={displayName} description={t("content.emptyDescription")}>
          <Globe size={22} />
        </EmptyState>
      </div>
    {/if}
  </div>
</section>

{#if multilineHintVisible}
  <div class="multiline-hint" aria-live="polite">{t("content.multilineHint")}</div>
{/if}

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
    min-width: 0;
    overflow-x: hidden;
    background: var(--ren-content-bg);
    position: relative;
  }

  .viewer.mobile-gestures {
    touch-action: pan-y;
  }

  .gesture-body {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    min-width: 0;
    will-change: transform;
  }

  .gesture-pull-indicator,
  .gesture-back-indicator {
    position: absolute;
    z-index: 2;
    pointer-events: none;
  }

  .gesture-pull-indicator {
    top: 0.35rem;
    left: 50%;
    transform: translateX(-50%);
    padding: 0.35rem 0.75rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--ren-chrome-bg) 88%, transparent);
    border: 1px solid var(--ren-border);
    color: var(--ren-muted);
    font-size: 0.78rem;
    white-space: nowrap;
  }

  .gesture-pull-indicator.triggered {
    color: var(--ren-accent);
    border-color: color-mix(in srgb, var(--ren-accent) 45%, var(--ren-border));
  }

  .gesture-back-indicator {
    top: 0;
    bottom: 0;
    left: 0;
    width: 4px;
    background: linear-gradient(
      90deg,
      color-mix(in srgb, var(--ren-accent) 70%, transparent),
      transparent
    );
    opacity: 0.55;
  }

  .gesture-back-indicator.triggered {
    width: 6px;
    opacity: 1;
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
    flex-wrap: wrap;
    min-width: 0;
    padding: 0.55rem 1rem;
    background: color-mix(in srgb, var(--ren-accent) 12%, var(--ren-chrome-bg));
    border-bottom: 1px solid var(--ren-border);
    color: var(--ren-fg);
    font-size: 0.88rem;
  }

  .cache-text {
    flex: 1 1 10rem;
    min-width: 0;
    overflow-wrap: anywhere;
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
    flex-shrink: 0;
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
    flex-wrap: wrap;
    min-width: 0;
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
    min-width: 0;
  }

  .back-btn {
    border: none;
    background: transparent;
    color: var(--ren-fg);
    cursor: pointer;
    padding: 0.2rem 0.35rem;
    border-radius: 8px;
    max-width: 100%;
  }

  .back-btn span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
    min-width: 0;
    overflow: auto;
    overflow-x: hidden;
    padding: 1rem 1.25rem 2rem;
    line-height: 1.55;
    overflow-wrap: anywhere;
  }

  .content :global(img),
  .content :global(video),
  .content :global(canvas),
  .content :global(svg) {
    max-width: 100%;
    height: auto;
  }

  .content :global(table) {
    display: block;
    max-width: 100%;
    overflow-x: auto;
  }

  .content :global(pre) {
    max-width: 100%;
    overflow-x: auto;
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

  .content.micron :global(input.Mu-armed) {
    outline: 1px dashed #fbbf24;
    outline-offset: 1px;
  }

  .content.micron :global(textarea.Mu-multiline) {
    outline: 1px solid #34d399;
    outline-offset: 1px;
    resize: vertical;
  }

  .multiline-hint {
    position: fixed;
    right: 0.75rem;
    bottom: 0.75rem;
    z-index: 20;
    pointer-events: none;
    padding: 0.35rem 0.55rem;
    border-radius: 6px;
    background: #fcd34d;
    color: #18181b;
    font-size: 0.78rem;
    box-shadow: 0 2px 8px rgb(0 0 0 / 25%);
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

  .content :global(.docs-nav a),
  .content :global(.docs-lang-switch a) {
    color: var(--ren-accent);
    text-decoration: none;
  }

  .content :global(.docs-nav a:hover),
  .content :global(.docs-lang-switch a:hover),
  .content :global(.docs-body a:hover) {
    text-decoration: underline;
  }

  .content :global(.docs-body a) {
    color: var(--ren-accent);
    text-decoration: none;
  }

  .content :global(.docs-body .docs-external-ref) {
    color: var(--ren-muted);
    cursor: default;
    text-decoration: underline dotted;
  }

  .content :global(.docs-body p) {
    margin: 0.75rem 0;
  }

  .content :global(.docs-body ul),
  .content :global(.docs-body ol) {
    margin: 0.75rem 0;
    padding-left: 1.5rem;
  }

  .content :global(.docs-body li) {
    margin: 0.35rem 0;
  }

  .content :global(.docs-body code) {
    font-family: var(--ren-mono, ui-monospace, monospace);
    font-size: 0.9em;
    padding: 0.1rem 0.35rem;
    border-radius: 4px;
    background: color-mix(in srgb, var(--ren-chrome-bg) 80%, transparent);
  }

  .content :global(.docs-body table) {
    width: 100%;
    border-collapse: collapse;
    margin: 1rem 0;
  }

  .content :global(.docs-body th),
  .content :global(.docs-body td) {
    border: 1px solid var(--ren-border);
    padding: 0.45rem 0.6rem;
    text-align: left;
    vertical-align: top;
  }

  .content :global(.docs-body blockquote) {
    margin: 0.75rem 0;
    padding-left: 1rem;
    border-left: 3px solid var(--ren-border);
    color: var(--ren-muted);
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
