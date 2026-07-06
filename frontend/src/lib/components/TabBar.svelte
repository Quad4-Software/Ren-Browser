<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { Pin, Plus, X } from "@lucide/svelte";
  import { clampMenuPosition } from "$lib/browser/context-menu";
  import { MAX_TABS, TAB_GAP_PX, type Tab, tabsAreaWidth, tabWidthForTab } from "$lib/browser/url";
  import TabPreviewThumb from "$lib/components/TabPreviewThumb.svelte";
  import WindowControls from "$lib/components/WindowControls.svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type MenuAction =
    | "reload"
    | "duplicate"
    | "favorite"
    | "viewSource"
    | "download"
    | "split"
    | "closeSplit"
    | "close"
    | "closeOthers"
    | "closeRight"
    | "closeAll"
    | "pin"
    | "unpin";

  type Props = {
    tabs: Tab[];
    nativeTitlebar: boolean;
    mobileUI: boolean;
    showWindowControls?: boolean;
    tabHoverPreviews: boolean;
    splitTabId: string | null;
    splitViewOpen: boolean;
    onSelect: (id: string) => void;
    onClose: (id: string) => void;
    onNew: () => void;
    onReorder: (fromId: string, toId: string) => void;
    onReload: (id: string) => void;
    onDuplicate: (id: string) => void;
    onFavorite: (id: string) => void;
    onViewSource: (id: string) => void;
    onDownload: (id: string) => void;
    onSplit: (id: string) => void;
    onCloseSplit: () => void;
    onCloseOthers: (id: string) => void;
    onCloseRight: (id: string) => void;
    onCloseAll: () => void;
    onTogglePin: (id: string) => void;
  };

  let {
    tabs,
    nativeTitlebar,
    mobileUI,
    showWindowControls = true,
    tabHoverPreviews,
    splitTabId,
    splitViewOpen,
    onSelect,
    onClose,
    onNew,
    onReorder,
    onReload,
    onDuplicate,
    onFavorite,
    onViewSource,
    onDownload,
    onSplit,
    onCloseSplit,
    onCloseOthers,
    onCloseRight,
    onCloseAll,
    onTogglePin,
  }: Props = $props();

  let dragId = $state<string | null>(null);
  let menu = $state<{ x: number; y: number; tabId: string } | null>(null);
  let menuEl = $state<HTMLDivElement | null>(null);
  let menuPos = $state({ x: 0, y: 0 });
  const DRAG_STRIP_MIN_PX = 88;
  const CONTROLS_RESERVED_PX = 104;

  let tabbarEl = $state<HTMLDivElement | null>(null);
  let tabsSlotEl = $state<HTMLDivElement | null>(null);
  let controlsSlotEl = $state<HTMLDivElement | null>(null);
  let newTabEl = $state<HTMLButtonElement | null>(null);
  let tabsSlotWidth = $state(0);
  let controlsSlotWidth = $state(0);
  let newTabWidth = $state(0);
  let hoverTabId = $state<string | null>(null);
  let previewPos = $state({ left: 0, top: 0 });

  const PREVIEW_WIDTH = 220;
  const PREVIEW_OFFSET = 6;
  const PREVIEW_HOVER_DELAY_MS = 400;

  let previewTimer: ReturnType<typeof setTimeout> | undefined;

  const hoverTab = $derived(hoverTabId ? tabs.find((tab) => tab.id === hoverTabId) : null);

  const tabsRowMaxWidth = $derived(
    mobileUI ? tabsSlotWidth : Math.max(0, tabsSlotWidth - DRAG_STRIP_MIN_PX),
  );
  const windowControlsWidth = $derived(
    nativeTitlebar || mobileUI ? "0px" : `${controlsSlotWidth || CONTROLS_RESERVED_PX}px`,
  );
  const tabsStripWidth = $derived(tabsAreaWidth(tabsRowMaxWidth, newTabWidth));
  const atTabLimit = $derived(tabs.length >= MAX_TABS);

  function widthForTab(tab: Tab): number {
    return tabWidthForTab(tabsStripWidth, tabs, tab);
  }

  function pinnedGlyph(tab: Tab): string {
    if (!tab.url) {
      return "•";
    }
    const title = tab.title.trim();
    if (title) {
      return title.charAt(0).toUpperCase();
    }
    return "•";
  }

  const menuTabId = $derived(menu?.tabId ?? null);
  const menuTab = $derived(menuTabId ? tabs.find((tab) => tab.id === menuTabId) : null);
  const menuTabIndex = $derived(menuTabId ? tabs.findIndex((tab) => tab.id === menuTabId) : -1);
  const canCloseRight = $derived(
    menuTabIndex >= 0 && tabs.slice(menuTabIndex + 1).some((tab) => !tab.pinned),
  );
  const canCloseOthers = $derived(tabs.length > 1);
  const showCloseSplit = $derived(splitViewOpen);

  $effect(() => {
    if (!menu || !menuEl) {
      return;
    }
    const rect = menuEl.getBoundingClientRect();
    menuPos = clampMenuPosition(menu.x, menu.y, rect.width, rect.height);
  });

  $effect(() => {
    const bar = tabbarEl;
    const slot = tabsSlotEl;
    const controls = controlsSlotEl;
    const newBtn = newTabEl;
    if (!bar || !slot) {
      return;
    }
    const syncSizes = () => {
      tabsSlotWidth = slot.clientWidth;
      controlsSlotWidth = controls?.offsetWidth ?? 0;
      newTabWidth = newBtn?.offsetWidth ?? 0;
    };
    const observer = new ResizeObserver(() => {
      syncSizes();
    });
    observer.observe(bar);
    observer.observe(slot);
    if (controls) {
      observer.observe(controls);
    }
    if (newBtn) {
      observer.observe(newBtn);
    }
    syncSizes();
    return () => observer.disconnect();
  });

  $effect(() => {
    void tabs.length;
    void tabs.map((tab) => `${tab.id}:${tab.pinned}:${tab.active}`).join("|");
    const slot = tabsSlotEl;
    const newBtn = newTabEl;
    const controls = controlsSlotEl;
    if (!slot) {
      return;
    }
    const id = requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        tabsSlotWidth = slot.clientWidth;
        controlsSlotWidth = controls?.offsetWidth ?? 0;
        newTabWidth = newBtn?.offsetWidth ?? 0;
      });
    });
    return () => cancelAnimationFrame(id);
  });

  function handleDragStart(event: DragEvent, tabId: string) {
    if ((event.target as HTMLElement).closest(".close")) {
      event.preventDefault();
      return;
    }
    hideTabPreview();
    dragId = tabId;
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
      event.dataTransfer.setData("text/plain", tabId);
    }
  }

  function handleDragEnd() {
    dragId = null;
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
  }

  function handleDrop(event: DragEvent, targetId: string) {
    event.preventDefault();
    event.stopPropagation();
    const fromId = dragId ?? event.dataTransfer?.getData("text/plain");
    if (!fromId || fromId === targetId) {
      dragId = null;
      return;
    }
    onReorder(fromId, targetId);
    dragId = null;
  }

  function openMenu(event: MouseEvent, tabId: string) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY, tabId };
  }

  function closeMenu() {
    menu = null;
  }

  function clearPreviewTimer() {
    if (previewTimer !== undefined) {
      clearTimeout(previewTimer);
      previewTimer = undefined;
    }
  }

  function showTabPreview(tabId: string, target: HTMLElement) {
    if (mobileUI || !tabHoverPreviews) {
      return;
    }
    clearPreviewTimer();
    previewTimer = setTimeout(() => {
      previewTimer = undefined;
      const rect = target.getBoundingClientRect();
      let left = rect.left + rect.width / 2 - PREVIEW_WIDTH / 2;
      const margin = 8;
      left = Math.max(margin, Math.min(left, window.innerWidth - PREVIEW_WIDTH - margin));
      hoverTabId = tabId;
      previewPos = { left, top: rect.bottom + PREVIEW_OFFSET };
    }, PREVIEW_HOVER_DELAY_MS);
  }

  function hideTabPreview() {
    clearPreviewTimer();
    hoverTabId = null;
  }

  onDestroy(() => {
    clearPreviewTimer();
  });

  function runAction(action: MenuAction) {
    if (!menu) {
      return;
    }
    const tabId = menu.tabId;
    closeMenu();
    switch (action) {
      case "reload":
        onReload(tabId);
        break;
      case "duplicate":
        onDuplicate(tabId);
        break;
      case "favorite":
        onFavorite(tabId);
        break;
      case "viewSource":
        onViewSource(tabId);
        break;
      case "download":
        onDownload(tabId);
        break;
      case "split":
        onSplit(tabId);
        break;
      case "closeSplit":
        onCloseSplit();
        break;
      case "close":
        onClose(tabId);
        break;
      case "closeOthers":
        onCloseOthers(tabId);
        break;
      case "closeRight":
        onCloseRight(tabId);
        break;
      case "closeAll":
        onCloseAll();
        break;
      case "pin":
      case "unpin":
        onTogglePin(tabId);
        break;
    }
  }
</script>

<svelte:window onclick={closeMenu} />

<div
  class="tabbar"
  class:native-titlebar={nativeTitlebar}
  class:mobile-ui={mobileUI}
  class:frameless-desktop={!nativeTitlebar && !mobileUI}
  style:--window-controls-width={windowControlsWidth}
  bind:this={tabbarEl}
>
  <div class="tabs-slot" bind:this={tabsSlotEl}>
    <div class="tabs-row" style:max-width="{tabsRowMaxWidth}px">
      <div
        class="tabs"
        role="tablist"
        tabindex="0"
        style:--tab-gap="{TAB_GAP_PX}px"
        style:--wails-draggable="no-drag"
        ondragover={handleDragOver}
      >
        {#each tabs as tab (tab.id)}
          {@const tabItemWidth = widthForTab(tab)}
          <button
            class="tab"
            class:active={tab.active}
            class:pinned={tab.pinned}
            class:split={splitViewOpen && splitTabId === tab.id}
            class:split-primary={splitViewOpen && tab.active}
            class:dragging={dragId === tab.id}
            role="tab"
            aria-selected={tab.active}
            title={tab.title}
            draggable="true"
            style:--tab-width="{tabItemWidth}px"
            style:--wails-draggable="no-drag"
            onclick={() => onSelect(tab.id)}
            oncontextmenu={(event) => openMenu(event, tab.id)}
            onmouseenter={(event) => showTabPreview(tab.id, event.currentTarget)}
            onmouseleave={hideTabPreview}
            ondragstart={(event) => handleDragStart(event, tab.id)}
            ondragend={handleDragEnd}
            ondragover={handleDragOver}
            ondrop={(event) => handleDrop(event, tab.id)}
          >
            {#if tab.pinned}
              <span class="pin-glyph" aria-hidden="true">
                {#if tab.url}
                  {pinnedGlyph(tab)}
                {:else}
                  <Pin size={12} />
                {/if}
              </span>
            {:else}
              <span class="title">{tab.title}</span>
              <span
                class="close"
                role="button"
                tabindex="0"
                aria-label={t("tab.close")}
                onclick={(event) => {
                  event.stopPropagation();
                  onClose(tab.id);
                }}
                onkeydown={(event) => {
                  if (event.key === "Enter") {
                    event.stopPropagation();
                    onClose(tab.id);
                  }
                }}
              >
                <X size={14} />
              </span>
            {/if}
          </button>
        {/each}
      </div>

      <button
        class="new-tab ren-icon-btn"
        bind:this={newTabEl}
        aria-label={t("tab.newTab")}
        disabled={atTabLimit}
        title={atTabLimit ? t("tab.limitReached", { max: MAX_TABS }) : t("tab.newTab")}
        onclick={onNew}
        style:--wails-draggable="no-drag"
      >
        <Plus size={14} />
      </button>
    </div>

    {#if !mobileUI}
      <div
        class="drag-strip"
        style:--wails-draggable={nativeTitlebar ? "no-drag" : "drag"}
        aria-hidden="true"
      ></div>
    {/if}
  </div>

  {#if showWindowControls && !nativeTitlebar && !mobileUI}
    <div class="controls-slot" bind:this={controlsSlotEl}>
      <WindowControls />
    </div>
  {/if}
</div>

{#if menu}
  <div
    class="context-menu"
    bind:this={menuEl}
    style:left="{menuPos.x}px"
    style:top="{menuPos.y}px"
    role="menu"
    tabindex="0"
    onclick={(event) => event.stopPropagation()}
    onkeydown={(event) => {
      if (event.key === "Escape") {
        closeMenu();
      }
    }}
  >
    <button role="menuitem" onclick={() => runAction("reload")}>{t("tab.reload")}</button>
    <button role="menuitem" onclick={() => runAction("duplicate")}>{t("tab.duplicate")}</button>
    <button role="menuitem" onclick={() => runAction("favorite")}>{t("tab.favorite")}</button>
    {#if menuTab?.pinned}
      <button role="menuitem" onclick={() => runAction("unpin")}>{t("tab.unpin")}</button>
    {:else}
      <button role="menuitem" onclick={() => runAction("pin")}>{t("tab.pin")}</button>
    {/if}
    <button role="menuitem" onclick={() => runAction("viewSource")}>{t("tab.viewSource")}</button>
    <button role="menuitem" onclick={() => runAction("download")}>{t("tab.downloadPage")}</button>
    <button role="menuitem" onclick={() => runAction("split")}>{t("tab.split")}</button>
    {#if showCloseSplit}
      <button role="menuitem" onclick={() => runAction("closeSplit")}>{t("tab.closeSplit")}</button>
    {/if}
    <hr />
    {#if !menuTab?.pinned}
      <button role="menuitem" onclick={() => runAction("close")}>{t("tab.closeTab")}</button>
    {/if}
    {#if canCloseOthers}
      <button role="menuitem" onclick={() => runAction("closeOthers")}
        >{t("tab.closeOthers")}</button
      >
    {/if}
    {#if canCloseRight}
      <button role="menuitem" onclick={() => runAction("closeRight")}>{t("tab.closeRight")}</button>
    {/if}
    <button role="menuitem" class="danger" onclick={() => runAction("closeAll")}
      >{t("tab.closeAll")}</button
    >
  </div>
{/if}

{#if hoverTab && !mobileUI}
  <div
    class="tab-preview-popover"
    style:left="{previewPos.left}px"
    style:top="{previewPos.top}px"
    role="tooltip"
  >
    <TabPreviewThumb tab={hoverTab} label={hoverTab.title} class="tab-preview-thumb" />
    <div class="tab-preview-footer">
      <span class="tab-preview-title">{hoverTab.title || hoverTab.url || t("tab.new")}</span>
      {#if hoverTab.url && hoverTab.url !== hoverTab.title}
        <span class="tab-preview-url">{hoverTab.url}</span>
      {/if}
    </div>
  </div>
{/if}

<style>
  .tabbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: flex-end;
    gap: 0.35rem;
    padding: 0.45rem 0.5rem 0 0.85rem;
    background: var(--ren-chrome-bg);
    border-bottom: 1px solid var(--ren-border);
    min-width: 0;
    min-height: 2.5rem;
    overflow: hidden;
  }

  .tabbar.frameless-desktop {
    position: relative;
    grid-template-columns: minmax(0, 1fr);
    padding-right: calc(0.5rem + var(--window-controls-width, 6.5rem));
  }

  .tabbar.native-titlebar,
  .tabbar.mobile-ui {
    grid-template-columns: minmax(0, 1fr);
  }

  .tabbar.mobile-ui {
    padding-top: env(safe-area-inset-top);
  }

  .tabbar.mobile-ui .drag-strip {
    display: none;
  }

  .controls-slot {
    flex-shrink: 0;
    min-width: max-content;
  }

  .tabbar.frameless-desktop .controls-slot {
    position: absolute;
    right: 0.5rem;
    bottom: 0;
    z-index: 20;
    pointer-events: auto;
  }

  .tabbar.frameless-desktop .tabs-slot {
    grid-column: 1;
    min-width: 0;
  }

  .tabs-slot {
    min-width: 0;
    display: flex;
    align-items: flex-end;
    overflow: hidden;
  }

  .tabs-row {
    display: flex;
    align-items: flex-end;
    gap: 0.35rem;
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
  }

  .tabs {
    display: flex;
    gap: var(--tab-gap);
    min-width: 0;
    flex: 0 1 auto;
    overflow: hidden;
  }

  .tab {
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    width: var(--tab-width);
    min-width: 0;
    max-width: var(--tab-width);
    flex: 0 1 var(--tab-width);
    border: 1px solid transparent;
    border-radius: 10px 10px 0 0;
    background: transparent;
    color: var(--ren-muted);
    padding: 0.5rem 0.55rem 0.5rem 0.7rem;
    cursor: grab;
    font: inherit;
    font-size: 0.86rem;
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease,
      width 0.12s ease;
  }

  .tab:hover:not(.active) {
    background: var(--ren-tab-hover);
    color: var(--ren-fg-secondary);
  }

  .tab.dragging {
    opacity: 0.55;
  }

  .tab.active {
    background: var(--ren-tab-active);
    color: var(--ren-fg);
    border-color: var(--ren-border);
    border-bottom-color: transparent;
    font-weight: 500;
  }

  .tab.split-primary,
  .tab.active.split-primary {
    box-shadow: inset 0 -2px 0 var(--ren-accent);
  }

  .tab.split:not(.active) {
    box-shadow: inset 0 -2px 0 color-mix(in srgb, var(--ren-accent) 55%, var(--ren-muted));
  }

  .tab.pinned {
    justify-content: center;
    padding-inline: 0.35rem;
    cursor: pointer;
    flex: 0 0 var(--tab-width);
    min-width: var(--tab-width);
  }

  .tab.pinned.active {
    box-shadow: inset 0 -2px 0 var(--ren-accent);
  }

  .pin-glyph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1rem;
    font-size: 0.72rem;
    font-weight: 600;
    line-height: 1;
    color: var(--ren-muted);
  }

  .tab.active .pin-glyph {
    color: var(--ren-fg);
  }

  .title {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-align: left;
  }

  .close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 1.35rem;
    height: 1.35rem;
    margin-inline-end: -0.15rem;
    border-radius: 6px;
    opacity: 0;
    cursor: pointer;
    color: var(--ren-muted);
    transition:
      opacity 0.12s ease,
      background 0.12s ease,
      color 0.12s ease;
  }

  .tab:hover .close,
  .tab:focus-within .close,
  .tab.active .close {
    opacity: 0.8;
  }

  .close:hover {
    opacity: 1;
    background: var(--ren-tab-hover);
    color: var(--ren-fg);
  }

  .new-tab {
    flex-shrink: 0;
    margin-bottom: 0.15rem;
  }

  .new-tab:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  .drag-strip {
    flex: 1 1 0;
    min-width: 5.5rem;
    align-self: stretch;
    margin-bottom: 0.15rem;
  }

  .context-menu {
    position: fixed;
    z-index: 1000;
    min-width: 11.5rem;
    max-width: calc(100vw - 1rem);
    padding: 0.35rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    display: grid;
    gap: 0.15rem;
  }

  .context-menu button {
    text-align: left;
    border: none;
    background: transparent;
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.45rem 0.65rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
  }

  .context-menu button:hover {
    background: var(--ren-tab-hover);
  }

  .context-menu button.danger {
    color: var(--ren-danger);
  }

  .context-menu hr {
    border: none;
    border-top: 1px solid var(--ren-border);
    margin: 0.15rem 0;
  }

  .tab-preview-popover {
    position: fixed;
    z-index: 950;
    width: 220px;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    overflow: hidden;
    pointer-events: none;
  }

  :global(.tab-preview-thumb.thumb) {
    width: 100%;
    height: 10.5rem;
    border-bottom: 1px solid var(--ren-border);
  }

  .tab-preview-footer {
    padding: 0.5rem 0.6rem;
    display: grid;
    gap: 0.15rem;
    min-width: 0;
  }

  .tab-preview-title {
    font-weight: 600;
    font-size: 0.82rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tab-preview-url {
    color: var(--ren-muted);
    font-size: 0.72rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
