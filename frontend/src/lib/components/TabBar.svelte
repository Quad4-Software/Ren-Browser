<script lang="ts">
  import { Pin, Plus, X } from "@lucide/svelte";
  import { MAX_TABS, TAB_GAP_PX, type Tab, tabsAreaWidth, tabWidthForTab } from "$lib/browser/url";
  import WindowControls from "$lib/components/WindowControls.svelte";

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
  const DRAG_STRIP_MIN_PX = 88;

  let tabbarEl = $state<HTMLDivElement | null>(null);
  let tabsSlotEl = $state<HTMLDivElement | null>(null);
  let newTabEl = $state<HTMLButtonElement | null>(null);
  let tabsSlotWidth = $state(0);
  let newTabWidth = $state(0);

  const tabsRowMaxWidth = $derived(
    mobileUI ? tabsSlotWidth : Math.max(0, tabsSlotWidth - DRAG_STRIP_MIN_PX),
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
    const bar = tabbarEl;
    const slot = tabsSlotEl;
    const newBtn = newTabEl;
    if (!bar || !slot) {
      return;
    }
    const syncSizes = () => {
      tabsSlotWidth = slot.clientWidth;
      newTabWidth = newBtn?.offsetWidth ?? 0;
    };
    const observer = new ResizeObserver(() => {
      syncSizes();
    });
    observer.observe(bar);
    observer.observe(slot);
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
    if (!slot) {
      return;
    }
    const id = requestAnimationFrame(() => {
      tabsSlotWidth = slot.clientWidth;
      newTabWidth = newBtn?.offsetWidth ?? 0;
    });
    return () => cancelAnimationFrame(id);
  });

  function handleDragStart(tabId: string) {
    dragId = tabId;
  }

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
  }

  function handleDrop(targetId: string) {
    if (!dragId || dragId === targetId) {
      dragId = null;
      return;
    }
    onReorder(dragId, targetId);
    dragId = null;
  }

  function openMenu(event: MouseEvent, tabId: string) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY, tabId };
  }

  function closeMenu() {
    menu = null;
  }

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

<div class="tabbar" class:native-titlebar={nativeTitlebar} class:mobile-ui={mobileUI} bind:this={tabbarEl}>
  <div class="tabs-slot" bind:this={tabsSlotEl}>
    <div class="tabs-row" style:max-width="{tabsRowMaxWidth}px">
      <div
        class="tabs"
        role="tablist"
        style:--tab-gap="{TAB_GAP_PX}px"
        style:--wails-draggable={nativeTitlebar ? "no-drag" : "no-drag"}
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
            onclick={() => onSelect(tab.id)}
            oncontextmenu={(event) => openMenu(event, tab.id)}
            ondragstart={() => handleDragStart(tab.id)}
            ondragover={handleDragOver}
            ondrop={() => handleDrop(tab.id)}
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
                aria-label="Close tab"
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
                <X size={12} />
              </span>
            {/if}
          </button>
        {/each}
      </div>

      <button
        class="new-tab ren-icon-btn"
        bind:this={newTabEl}
        aria-label="New tab"
        disabled={atTabLimit}
        title={atTabLimit ? `Tab limit (${MAX_TABS}) reached` : "New tab"}
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

  {#if !nativeTitlebar}
    <div class="controls-slot">
      <WindowControls />
    </div>
  {/if}
</div>

{#if menu}
  <div
    class="context-menu"
    style:left="{menu.x}px"
    style:top="{menu.y}px"
    role="menu"
    tabindex="0"
    onclick={(event) => event.stopPropagation()}
    onkeydown={(event) => {
      if (event.key === "Escape") {
        closeMenu();
      }
    }}
  >
    <button role="menuitem" onclick={() => runAction("reload")}>Reload</button>
    <button role="menuitem" onclick={() => runAction("duplicate")}>Duplicate</button>
    <button role="menuitem" onclick={() => runAction("favorite")}>Favorite</button>
    {#if menuTab?.pinned}
      <button role="menuitem" onclick={() => runAction("unpin")}>Unpin tab</button>
    {:else}
      <button role="menuitem" onclick={() => runAction("pin")}>Pin tab</button>
    {/if}
    <button role="menuitem" onclick={() => runAction("viewSource")}>View source</button>
    <button role="menuitem" onclick={() => runAction("download")}>Download page</button>
    <button role="menuitem" onclick={() => runAction("split")}>Split tab</button>
    {#if showCloseSplit}
      <button role="menuitem" onclick={() => runAction("closeSplit")}>Close split</button>
    {/if}
    <hr />
    {#if !menuTab?.pinned}
      <button role="menuitem" onclick={() => runAction("close")}>Close tab</button>
    {/if}
    {#if canCloseOthers}
      <button role="menuitem" onclick={() => runAction("closeOthers")}>Close other tabs</button>
    {/if}
    {#if canCloseRight}
      <button role="menuitem" onclick={() => runAction("closeRight")}
        >Close tabs to the right</button
      >
    {/if}
    <button role="menuitem" class="danger" onclick={() => runAction("closeAll")}
      >Close all tabs</button
    >
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
  }

  .tabbar.native-titlebar {
    grid-template-columns: minmax(0, 1fr);
  }

  .tabbar.mobile-ui {
    padding-top: env(safe-area-inset-top);
  }

  .tabbar.mobile-ui .tabs-slot {
    grid-template-columns: minmax(0, 1fr);
  }

  .controls-slot {
    flex-shrink: 0;
    min-width: max-content;
  }

  .tabs-slot {
    min-width: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(4.5rem, 5.5rem);
    align-items: flex-end;
    gap: 0;
    overflow: hidden;
  }

  .tabs-row {
    display: flex;
    align-items: flex-end;
    gap: 0.35rem;
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
    min-width: var(--tab-width);
    max-width: var(--tab-width);
    flex: 0 0 var(--tab-width);
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
    flex-shrink: 0;
    opacity: 0;
    transition: opacity 0.12s ease;
    color: var(--ren-muted);
  }

  .tab:hover .close,
  .tab:focus-within .close,
  .tab.active .close {
    opacity: 0.8;
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
    min-width: 0;
    align-self: stretch;
    margin-bottom: 0.15rem;
  }

  .context-menu {
    position: fixed;
    z-index: 1000;
    min-width: 11.5rem;
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
</style>
