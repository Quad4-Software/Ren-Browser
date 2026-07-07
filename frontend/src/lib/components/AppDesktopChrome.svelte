<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import TabBar from "$lib/components/TabBar.svelte";
  import BrowserChrome from "$lib/components/BrowserChrome.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();
</script>

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

<BrowserChrome
  bind:url={app.url}
  canGoBack={app.canGoBack}
  canGoForward={app.canGoForward}
  activePanel={app.activePanel}
  pluginPanels={app.pluginContributions.panels}
  themeMode={app.theme.mode === "light" ? "light" : "dark"}
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
  onToggleTheme={app.toggleTheme}
/>
