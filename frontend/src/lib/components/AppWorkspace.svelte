<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import MobileTabsPage from "$lib/components/MobileTabsPage.svelte";
  import AppPagePane from "$lib/components/AppPagePane.svelte";
  import AppSidePanel from "$lib/components/AppSidePanel.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import { readSidebarWidth, writeSidebarWidth } from "$lib/browser/layout-prefs";
  import {
    capturePointer,
    clampSidebarWidth,
    createDragScheduler,
    releasePointer,
    SIDEBAR_DEFAULT_PX,
    sidebarWidthFromPointer,
  } from "$lib/browser/pane-resize";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();

  const panelOpen = $derived(app.activePanel !== "browser");
  const overlaySidebars = $derived(app.overlaySidebars && !app.mobileUI);
  const useOverlay = $derived(overlaySidebars && panelOpen);
  const useSplit = $derived(panelOpen && !app.mobileUI && !overlaySidebars);

  let sidebarWidth = $state(readSidebarWidth());
  let resizing = $state(false);
  let workspaceEl = $state<HTMLElement | null>(null);
  let resizerEl = $state<HTMLButtonElement | null>(null);

  const drag = createDragScheduler((event) => {
    if (!workspaceEl) {
      return;
    }
    const next = sidebarWidthFromPointer(event, workspaceEl.getBoundingClientRect());
    workspaceEl.style.setProperty("--sidebar-width", `${next}px`);
  });

  function commitWidth() {
    if (!workspaceEl) {
      return;
    }
    const raw = Number.parseFloat(workspaceEl.style.getPropertyValue("--sidebar-width"));
    if (!Number.isFinite(raw)) {
      return;
    }
    const next = clampSidebarWidth(raw, workspaceEl.getBoundingClientRect().width);
    sidebarWidth = next;
    writeSidebarWidth(next);
  }

  function onResizePointerDown(event: PointerEvent) {
    if (!workspaceEl || !resizerEl) {
      return;
    }
    resizing = true;
    capturePointer(resizerEl, event);
    workspaceEl.style.setProperty("--sidebar-width", `${sidebarWidth}px`);
  }

  function onResizePointerMove(event: PointerEvent) {
    if (!resizing) {
      return;
    }
    drag.move(event);
  }

  function onResizePointerUp(event: PointerEvent) {
    drag.cancel();
    releasePointer(resizerEl, event);
    if (resizing) {
      commitWidth();
    }
    resizing = false;
  }

  function onResizeKeydown(event: KeyboardEvent) {
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") {
      return;
    }
    event.preventDefault();
    const span = workspaceEl?.getBoundingClientRect().width ?? SIDEBAR_DEFAULT_PX * 2;
    const delta = event.key === "ArrowLeft" ? 16 : -16;
    const next = clampSidebarWidth(sidebarWidth + delta, span);
    sidebarWidth = next;
    writeSidebarWidth(next);
  }

  function resetSidebarWidth() {
    const span = workspaceEl?.getBoundingClientRect().width ?? SIDEBAR_DEFAULT_PX * 2;
    const next = clampSidebarWidth(SIDEBAR_DEFAULT_PX, span);
    sidebarWidth = next;
    writeSidebarWidth(next);
  }
</script>

<svelte:window
  onpointermove={resizing ? onResizePointerMove : undefined}
  onpointerup={resizing ? onResizePointerUp : undefined}
  onpointercancel={resizing ? onResizePointerUp : undefined}
/>

<main
  class="workspace"
  class:split={useSplit}
  class:overlay-panel={useOverlay}
  class:mobile-panel={app.mobileUI && panelOpen && !app.mobileTabsOpen}
  class:mobile-tabs={app.mobileUI && app.mobileTabsOpen}
  class:resizing
  bind:this={workspaceEl}
  style:--sidebar-width="{sidebarWidth}px"
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

    {#if useSplit}
      <button
        type="button"
        class="sidebar-resizer"
        bind:this={resizerEl}
        aria-label={t("chrome.resizeSidebar")}
        onpointerdown={onResizePointerDown}
        onpointermove={onResizePointerMove}
        onpointerup={onResizePointerUp}
        onpointercancel={onResizePointerUp}
        onkeydown={onResizeKeydown}
        ondblclick={resetSidebarWidth}
      ></button>
    {/if}

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
    grid-template-columns: minmax(0, 1fr) 5px var(--sidebar-width, 360px);
  }

  .workspace.resizing {
    cursor: col-resize;
    user-select: none;
  }

  .workspace.resizing :global(iframe) {
    pointer-events: none;
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

  .sidebar-resizer {
    width: 5px;
    padding: 0;
    border: none;
    background: transparent;
    cursor: col-resize;
    position: relative;
    touch-action: none;
    z-index: 2;
  }

  .sidebar-resizer::before {
    content: "";
    position: absolute;
    inset: 0 -4px;
  }

  .sidebar-resizer::after {
    content: "";
    position: absolute;
    inset: 0;
    background: var(--ren-border);
    opacity: 0.55;
    transition: opacity 0.12s ease;
  }

  .sidebar-resizer:hover::after,
  .workspace.resizing .sidebar-resizer::after {
    opacity: 1;
    background: var(--ren-accent);
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
    height: 100%;
    display: flex;
    flex-direction: column;
    border-top: 1px solid var(--ren-border);
    contain: layout;
  }

  .page-pane > :global(*) {
    flex: 1;
    min-height: 0;
    min-width: 0;
  }

  @media (max-width: 900px) {
    .workspace.split {
      grid-template-columns: 1fr;
      grid-template-rows: 1fr auto;
    }

    .workspace.split .sidebar-resizer {
      display: none;
    }

    .workspace.split :global(.side-pane) {
      max-height: 45vh;
      border-left: none;
    }
  }
</style>
