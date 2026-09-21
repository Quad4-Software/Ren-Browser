<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import TabBar from "$lib/components/TabBar.svelte";
  import BrowserChrome from "$lib/components/BrowserChrome.svelte";
  import WindowControls from "$lib/components/WindowControls.svelte";
  import { handleTitlebarDoubleClick } from "$lib/browser/window-actions";
  import type { AppController } from "$lib/app/create-app.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();

  const vertical = $derived(app.tabLayout === "left" && !app.mobileUI);

  async function handleStripDoubleClick() {
    if (app.nativeTitlebar || app.mobileUI) {
      return;
    }
    const result = await handleTitlebarDoubleClick();
    if (!result.ok) {
      app.showPluginToast(result.error || t("window.actionFailed"), { isError: true });
    }
  }
</script>

{#if vertical}
  <div class="chrome-stack">
    {#if !app.nativeTitlebar}
      <div
        class="win-strip"
        style:--wails-draggable="drag"
        aria-hidden="true"
        ondblclick={handleStripDoubleClick}
      >
        <div class="win-strip-controls" style:--wails-draggable="no-drag">
          <WindowControls />
        </div>
      </div>
    {/if}
    {@render browserChrome()}
  </div>
{:else}
  <TabBar
    tabs={app.tabs}
    nativeTitlebar={app.nativeTitlebar}
    mobileUI={app.mobileUI}
    showWindowControls={app.desktopChrome}
    tabHoverPreviews={app.tabHoverPreviews}
    micronEngine={app.effectiveMicronEngine}
    splitViewOpen={app.splitViewOpen}
    splitTabId={app.splitTabId}
    onSelect={app.setActiveTab}
    onClose={app.closeTab}
    onNew={app.newTab}
    onReorder={app.reorderTabs}
    onReload={app.reloadTab}
    onDuplicate={app.duplicateTab}
    onFavorite={app.favoriteTab}
    onViewSource={app.viewSourceTab}
    onDownload={app.downloadTab}
    onSplit={app.splitTabView}
    onCloseSplit={app.closeSplitView}
    onCloseOthers={app.closeOtherTabs}
    onCloseRight={app.closeTabsToRight}
    onCloseAll={app.requestCloseAllTabs}
    onTogglePin={app.togglePinTab}
    onWindowChromeError={(message) => app.showPluginToast(message, { isError: true })}
  />
  {@render browserChrome()}
{/if}

{#snippet browserChrome()}
  <BrowserChrome
    bind:url={app.url}
    canGoBack={app.canGoBack}
    canGoForward={app.canGoForward}
    activePanel={app.activePanel}
    pluginPanels={app.pluginContributions.panels}
    devToolsEnabled={app.mobileDevTools}
    downloadsOpen={app.downloadsOpen}
    downloads={app.downloads}
    activeDownloads={app.activeDownloadViews}
    downloadDir={app.downloadDir}
    canIdentify={app.canIdentify}
    identifying={app.identifying}
    onNavigate={app.openPage}
    onBack={app.goBack}
    onForward={app.goForward}
    onReload={() => app.openPage(app.url)}
    onDownloadPage={app.downloadCurrentPage}
    onToggleDownloads={app.toggleDownloads}
    onCloseDownloads={() => (app.downloadsOpen = false)}
    onOpenDownload={app.openDownload}
    onReadDownload={app.readDownload}
    onOpenDownloadFolder={app.openDownloadFolder}
    onCancelDownload={app.cancelActiveDownload}
    onDismissDownload={app.dismissActiveDownload}
    onRetryDownload={app.retryActiveDownload}
    retryingDownloadIds={app.retryingDownloadIds}
    onClearDownloadHistory={app.clearDownloadHistory}
    clearingDownloadHistory={app.clearingDownloadHistory}
    onIdentify={app.requestIdentify}
    onPanel={app.setPanel}
  />
{/snippet}

<style>
  .chrome-stack {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }

  .win-strip {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    min-height: 1.9rem;
    background: var(--ren-chrome-bg);
  }

  .win-strip-controls {
    flex-shrink: 0;
  }
</style>
