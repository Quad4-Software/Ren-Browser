<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onMount } from "svelte";
  import { Events } from "@wailsio/runtime";
  import { createApp } from "$lib/app/create-app.svelte";
  import PluginToast from "$lib/components/PluginToast.svelte";
  import AppDesktopChrome from "$lib/components/AppDesktopChrome.svelte";
  import AppWorkspace from "$lib/components/AppWorkspace.svelte";
  import AppConfirmDialogs from "$lib/components/AppConfirmDialogs.svelte";
  import InitialSetupModal from "$lib/components/InitialSetupModal.svelte";
  import MobileUrlBar from "$lib/components/MobileUrlBar.svelte";
  import MobileNav from "$lib/components/MobileNav.svelte";
  import DownloadsMenu from "$lib/components/DownloadsMenu.svelte";
  import AppStoreError from "$lib/components/AppStoreError.svelte";
  import {
    isEditableKeyboardTarget,
    keyboardChromeState,
    parseNativeKeyboardEvent,
    visualViewportKeyboardInset,
    type NativeKeyboardState,
  } from "$lib/browser/mobile-keyboard";

  const app = createApp();

  let imeInset = $state(0);
  let keyboardOpen = $state(false);
  let nativeKeyboard = $state<NativeKeyboardState>({ visible: false, height: 0 });
  let focusedEditable = $state(false);

  function syncKeyboardChrome() {
    const next = keyboardChromeState({
      visualInset: visualViewportKeyboardInset(window.innerHeight, window.visualViewport),
      native: nativeKeyboard,
      focusedEditable,
    });
    imeInset = next.imeInset;
    keyboardOpen = next.keyboardOpen;
  }

  function onFocusIn(event: FocusEvent) {
    focusedEditable = isEditableKeyboardTarget(event.target);
    syncKeyboardChrome();
  }

  function onFocusOut() {
    queueMicrotask(() => {
      focusedEditable = isEditableKeyboardTarget(document.activeElement);
      syncKeyboardChrome();
    });
  }

  onMount(() => {
    const viewport = window.visualViewport;
    const onViewport = () => syncKeyboardChrome();
    viewport?.addEventListener("resize", onViewport);
    viewport?.addEventListener("scroll", onViewport);
    window.addEventListener("resize", onViewport);
    const offKeyboard = Events.On("common:keyboard", (event: { data?: unknown }) => {
      nativeKeyboard = parseNativeKeyboardEvent(event);
      syncKeyboardChrome();
    });
    syncKeyboardChrome();
    return () => {
      viewport?.removeEventListener("resize", onViewport);
      viewport?.removeEventListener("scroll", onViewport);
      window.removeEventListener("resize", onViewport);
      offKeyboard?.();
    };
  });

  onMount(app.mount);
</script>

<svelte:window onfocusin={onFocusIn} onfocusout={onFocusOut} />

<div
  class="app-shell"
  class:mobile-ui={app.mobileUI}
  class:keyboard-open={app.mobileUI && keyboardOpen}
  style:--ime-inset="{app.mobileUI ? imeInset : 0}px"
>
  <PluginToast message={app.pluginToast} isError={app.pluginToastIsError} />

  {#if app.mobileUI}
    <MobileUrlBar
      bind:url={app.url}
      tabCount={app.tabs.length}
      canIdentify={app.canIdentify}
      identifying={app.identifying}
      atTabLimit={app.atTabLimit}
      onNavigate={app.openPage}
      onHome={app.mobileHome}
      onNewTab={app.newTab}
      onOpenTabs={app.openMobileTabs}
      onIdentify={app.requestIdentify}
    />
  {:else}
    <AppDesktopChrome {app} />
  {/if}

  <AppWorkspace {app} />

  {#if app.mobileUI && !app.mobileTabsOpen}
    <MobileNav
      activePanel={app.activePanel}
      mobileDevTools={app.mobileDevTools}
      downloadsOpen={app.downloadsOpen}
      activeDownloadCount={app.activeDownloadViews.filter(
        (item) => item.status === "pending" || item.status === "downloading",
      ).length}
      onPanel={app.setPanel}
      onToggleDownloads={app.toggleDownloads}
    />
  {/if}

  {#if app.mobileUI}
    <DownloadsMenu
      open={app.downloadsOpen}
      active={app.activeDownloadViews}
      downloads={app.downloads}
      downloadDir={app.downloadDir}
      variant="sheet"
      onDownloadPage={app.downloadCurrentPage}
      onOpenFile={app.openDownload}
      onReadFile={app.readDownload}
      onOpenFolder={app.openDownloadFolder}
      onCancelActive={app.cancelActiveDownload}
      onDismissActive={app.dismissActiveDownload}
      onRetryActive={app.retryActiveDownload}
      retryingIds={app.retryingDownloadIds}
      onClearHistory={app.clearDownloadHistory}
      clearingHistory={app.clearingDownloadHistory}
      onClose={() => (app.downloadsOpen = false)}
    />
  {/if}

  <AppConfirmDialogs {app} />

  <InitialSetupModal {app} />

  {#if app.storeErrorVisible}
    <AppStoreError
      kind={app.storeHealth.kind ?? "database_corrupt"}
      detail={app.storeHealth.detail ?? ""}
      path={app.storeHealth.path}
      onResetDatabase={app.requestResetDatabase}
      onRetry={() => void app.loadStoreHealth()}
    />
  {/if}
</div>

<style>
  .app-shell {
    height: 100vh;
    display: grid;
    grid-template-rows: auto auto 1fr auto;
    background: var(--ren-surface-bg);
    min-width: 0;
    max-width: 100%;
    overflow-x: clip;
  }

  .app-shell.mobile-ui {
    height: 100dvh;
    max-height: 100dvh;
    grid-template-rows: auto 1fr auto;
    overflow: hidden;
    padding-bottom: var(--ime-inset, 0px);
  }

  .app-shell.mobile-ui :global(.side-pane) {
    overflow-x: clip;
    width: 100%;
    max-width: 100%;
  }

  .app-shell.mobile-ui :global(input.ren-input),
  .app-shell.mobile-ui :global(textarea.ren-input),
  .app-shell.mobile-ui :global(textarea),
  .app-shell.mobile-ui :global(select) {
    font-size: 16px;
  }
</style>
