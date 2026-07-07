<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import MobileTabsPage from "$lib/components/MobileTabsPage.svelte";
  import AppPagePane from "$lib/components/AppPagePane.svelte";
  import AppSidePanel from "$lib/components/AppSidePanel.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();
</script>

<main
  class="workspace"
  class:split={app.activePanel !== "browser" && !app.mobileUI}
  class:mobile-panel={app.mobileUI && app.activePanel !== "browser" && !app.mobileTabsOpen}
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

    <AppSidePanel {app} />
  {/if}
</main>

<style>
  .workspace {
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
