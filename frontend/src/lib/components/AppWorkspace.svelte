<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import MobileTabsPage from "$lib/components/MobileTabsPage.svelte";
  import AppPagePane from "$lib/components/AppPagePane.svelte";
  import AppSidePanel from "$lib/components/AppSidePanel.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();

  const panelOpen = $derived(app.activePanel !== "browser");
  const overlaySidebars = $derived(app.overlaySidebars && !app.mobileUI);
  const useOverlay = $derived(overlaySidebars && panelOpen);
  const useSplit = $derived(panelOpen && !app.mobileUI && !overlaySidebars);
</script>

<main
  class="workspace"
  class:split={useSplit}
  class:overlay-panel={useOverlay}
  class:mobile-panel={app.mobileUI && panelOpen && !app.mobileTabsOpen}
  class:mobile-tabs={app.mobileUI && app.mobileTabsOpen}
>
  {#if app.mobileUI && app.mobileTabsOpen}
    <MobileTabsPage
      tabs={app.tabs}
      activeTabId={app.activeTabId}
      atTabLimit={app.atTabLimit}
      onSelect={app.mobileSelectTab}
      onClose={app.closeTab}
      onCloseAll={app.requestCloseAllTabs}
      onNew={app.newTab}
      onDismiss={app.closeMobileTabs}
    />
  {:else}
    <section class="page-pane">
      <AppPagePane {app} />
    </section>

    {#if useOverlay}
      <button
        type="button"
        class="panel-scrim"
        aria-label={t("chrome.closePanel")}
        onclick={() => app.setPanel("browser")}
      ></button>
    {/if}

    <AppSidePanel {app} />
  {/if}
</main>

<style>
  .workspace {
    position: relative;
    min-height: 0;
    min-width: 0;
    max-width: 100%;
    overflow-x: clip;
    display: grid;
    grid-template-columns: 1fr;
  }

  .workspace.split {
    grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.8fr);
  }

  .workspace.overlay-panel {
    grid-template-columns: 1fr;
  }

  .workspace.overlay-panel :global(.side-pane) {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    width: min(380px, 92vw);
    z-index: 120;
    max-height: none;
    height: 100%;
    border-top: none;
    box-shadow: var(--ren-shadow);
  }

  .panel-scrim {
    position: absolute;
    inset: 0;
    z-index: 110;
    border: 0;
    padding: 0;
    margin: 0;
    background: color-mix(in srgb, #000 35%, transparent);
    cursor: pointer;
  }

  .workspace.mobile-panel {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr;
  }

  .workspace.mobile-panel .page-pane {
    display: none;
  }

  .workspace.mobile-panel :global(.side-pane) {
    max-height: none;
    height: 100%;
    border-left: none;
    box-shadow: none;
  }

  .workspace.mobile-tabs {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr;
  }

  .workspace.mobile-tabs .page-pane,
  .workspace.mobile-tabs :global(.side-pane) {
    display: none;
  }

  .page-pane {
    min-height: 0;
    min-width: 0;
    border-top: 1px solid var(--ren-border);
  }

  @media (max-width: 900px) {
    .workspace.split {
      grid-template-columns: 1fr;
      grid-template-rows: 1fr auto;
    }

    .workspace.split :global(.side-pane) {
      max-height: 45vh;
      border-left: none;
    }
  }
</style>
