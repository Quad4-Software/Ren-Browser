<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { DropdownMenu, Popover } from "bits-ui";
  import { useResizeObserver } from "runed";
  import { Pin, Plus, X } from "@lucide/svelte";
  import { handleTitlebarDoubleClick } from "$lib/browser/window-actions";
  import { MAX_TABS, TAB_GAP_PX, type Tab, tabsAreaWidth, tabWidthForTab } from "$lib/browser/url";
  import TabPreviewThumb from "$lib/components/TabPreviewThumb.svelte";
  import WindowControls from "$lib/components/WindowControls.svelte";
  import type { MicronEffectiveEngine } from "$lib/micron/render-page";
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
    micronEngine?: MicronEffectiveEngine;
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
    onWindowChromeError?: (message: string) => void;
  };

  let {
    tabs,
    nativeTitlebar,
    mobileUI,
    showWindowControls = true,
    tabHoverPreviews,
    micronEngine = "js",
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
    onWindowChromeError = () => {},
  }: Props = $props();

  let dragId = $state<string | null>(null);
  let menu = $state<{ x: number; y: number; tabId: string } | null>(null);
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
  let previewAnchorEl = $state<HTMLElement | null>(null);

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
  const tabLayoutKey = $derived(
    `${tabs.length}:${tabs.filter((tab) => tab.pinned).length}:${tabs.find((tab) => tab.active)?.id ?? ""}`,
  );

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

  const menuAnchor = $derived.by(() => {
    const current = menu;
    if (!current) {
      return null;
    }
    return { getBoundingClientRect: () => new DOMRect(current.x, current.y, 0, 0) };
  });

  function syncSizes() {
    tabsSlotWidth = tabsSlotEl?.clientWidth ?? 0;
    controlsSlotWidth = controlsSlotEl?.offsetWidth ?? 0;
    newTabWidth = newTabEl?.offsetWidth ?? 0;
  }

  useResizeObserver(
    () =>
      ([tabbarEl, tabsSlotEl, controlsSlotEl, newTabEl] as Array<HTMLElement | null>).filter(
        (el): el is HTMLElement => el !== null,
      ),
    () => syncSizes(),
  );

  $effect(() => {
    void tabbarEl;
    void tabsSlotEl;
    void controlsSlotEl;
    void newTabEl;
    syncSizes();
  });

  $effect(() => {
    void tabLayoutKey;
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

  function openMenuForTab(tabEl: HTMLElement, tabId: string) {
    const rect = tabEl.getBoundingClientRect();
    menu = { x: rect.left + 8, y: rect.bottom, tabId };
  }

  function handleTabKeydown(event: KeyboardEvent, tabId: string) {
    const current = event.currentTarget as HTMLElement;
    if (
      event.key === "ArrowLeft" ||
      event.key === "ArrowRight" ||
      event.key === "Home" ||
      event.key === "End"
    ) {
      event.preventDefault();
      const tabEls = Array.from(tabsSlotEl?.querySelectorAll<HTMLElement>(".tab") ?? []);
      const index = tabEls.indexOf(current);
      if (index < 0) {
        return;
      }
      const next =
        event.key === "ArrowLeft"
          ? index > 0
            ? index - 1
            : tabEls.length - 1
          : event.key === "ArrowRight"
            ? index < tabEls.length - 1
              ? index + 1
              : 0
            : event.key === "Home"
              ? 0
              : tabEls.length - 1;
      tabEls[next]?.focus();
      return;
    }
    if (event.key === "ContextMenu" || (event.shiftKey && event.key === "F10")) {
      event.preventDefault();
      openMenuForTab(current, tabId);
    }
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
      previewAnchorEl = target;
      hoverTabId = tabId;
    }, PREVIEW_HOVER_DELAY_MS);
  }

  function hideTabPreview() {
    clearPreviewTimer();
    hoverTabId = null;
    previewAnchorEl = null;
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

  async function handleDragStripDoubleClick() {
    if (nativeTitlebar || mobileUI) {
      return;
    }
    const result = await handleTitlebarDoubleClick();
    if (!result.ok) {
      onWindowChromeError(result.error || t("window.actionFailed"));
    }
  }
</script>

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
        role="list"
        aria-label={t("tab.list")}
        style:--tab-gap="{TAB_GAP_PX}px"
        style:--wails-draggable="no-drag"
        ondragover={handleDragOver}
      >
        {#each tabs as tab (tab.id)}
          {@const tabItemWidth = widthForTab(tab)}
          <div
            class="tab-item"
            role="listitem"
            class:active={tab.active}
            class:pinned={tab.pinned}
            class:split={splitViewOpen && splitTabId === tab.id}
            class:split-primary={splitViewOpen && tab.active}
            class:dragging={dragId === tab.id}
            draggable="true"
            style:--tab-width="{tabItemWidth}px"
            style:--wails-draggable="no-drag"
            oncontextmenu={(event) => openMenu(event, tab.id)}
            onmouseenter={(event) => showTabPreview(tab.id, event.currentTarget)}
            onmouseleave={hideTabPreview}
            ondragstart={(event) => handleDragStart(event, tab.id)}
            ondragend={handleDragEnd}
            ondragover={handleDragOver}
            ondrop={(event) => handleDrop(event, tab.id)}
          >
            <button
              type="button"
              class="tab"
              aria-current={tab.active ? "page" : undefined}
              aria-label={tab.title || tab.url || t("tab.new")}
              tabindex={tab.active ? 0 : -1}
              title={tabHoverPreviews ? undefined : tab.title}
              onclick={() => onSelect(tab.id)}
              onkeydown={(event) => handleTabKeydown(event, tab.id)}
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
              {/if}
            </button>
            {#if !tab.pinned}
              <button
                type="button"
                class="close"
                aria-label={t("tab.close")}
                onclick={() => onClose(tab.id)}
              >
                <X size={14} />
              </button>
            {/if}
          </div>
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
        ondblclick={handleDragStripDoubleClick}
      ></div>
    {/if}
  </div>

  {#if showWindowControls && !nativeTitlebar && !mobileUI}
    <div class="controls-slot" bind:this={controlsSlotEl}>
      <WindowControls />
    </div>
  {/if}
</div>

<DropdownMenu.Root
  open={menu !== null}
  onOpenChange={(next) => {
    if (!next) {
      closeMenu();
    }
  }}
>
  <DropdownMenu.Portal>
    <DropdownMenu.Content
      class="tab-context-menu"
      customAnchor={menuAnchor}
      align="start"
      collisionPadding={8}
    >
      <DropdownMenu.Item textValue={t("tab.reload")} onSelect={() => runAction("reload")}>
        {t("tab.reload")}
      </DropdownMenu.Item>
      <DropdownMenu.Item textValue={t("tab.duplicate")} onSelect={() => runAction("duplicate")}>
        {t("tab.duplicate")}
      </DropdownMenu.Item>
      <DropdownMenu.Item textValue={t("tab.favorite")} onSelect={() => runAction("favorite")}>
        {t("tab.favorite")}
      </DropdownMenu.Item>
      {#if menuTab?.pinned}
        <DropdownMenu.Item textValue={t("tab.unpin")} onSelect={() => runAction("unpin")}>
          {t("tab.unpin")}
        </DropdownMenu.Item>
      {:else}
        <DropdownMenu.Item textValue={t("tab.pin")} onSelect={() => runAction("pin")}>
          {t("tab.pin")}
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item textValue={t("tab.viewSource")} onSelect={() => runAction("viewSource")}>
        {t("tab.viewSource")}
      </DropdownMenu.Item>
      <DropdownMenu.Item textValue={t("tab.downloadPage")} onSelect={() => runAction("download")}>
        {t("tab.downloadPage")}
      </DropdownMenu.Item>
      <DropdownMenu.Item textValue={t("tab.split")} onSelect={() => runAction("split")}>
        {t("tab.split")}
      </DropdownMenu.Item>
      {#if showCloseSplit}
        <DropdownMenu.Item textValue={t("tab.closeSplit")} onSelect={() => runAction("closeSplit")}>
          {t("tab.closeSplit")}
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Separator class="tab-context-separator" />
      {#if !menuTab?.pinned}
        <DropdownMenu.Item textValue={t("tab.closeTab")} onSelect={() => runAction("close")}>
          {t("tab.closeTab")}
        </DropdownMenu.Item>
      {/if}
      {#if canCloseOthers}
        <DropdownMenu.Item
          textValue={t("tab.closeOthers")}
          onSelect={() => runAction("closeOthers")}
        >
          {t("tab.closeOthers")}
        </DropdownMenu.Item>
      {/if}
      {#if canCloseRight}
        <DropdownMenu.Item textValue={t("tab.closeRight")} onSelect={() => runAction("closeRight")}>
          {t("tab.closeRight")}
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item
        class="tab-context-danger"
        textValue={t("tab.closeAll")}
        onSelect={() => runAction("closeAll")}
      >
        {t("tab.closeAll")}
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<Popover.Root
  open={hoverTab !== null && !mobileUI}
  onOpenChange={(next) => {
    if (!next) {
      hideTabPreview();
    }
  }}
>
  <Popover.Portal>
    <Popover.Content
      class="tab-preview-popover"
      customAnchor={previewAnchorEl}
      side="bottom"
      align="center"
      sideOffset={PREVIEW_OFFSET}
      collisionPadding={8}
      trapFocus={false}
      role="tooltip"
      onOpenAutoFocus={(event) => event.preventDefault()}
      onCloseAutoFocus={(event) => event.preventDefault()}
    >
      {#if hoverTab}
        <TabPreviewThumb
          tab={hoverTab}
          label={hoverTab.title}
          class="tab-preview-thumb"
          {micronEngine}
        />
        <div class="tab-preview-footer">
          <span class="tab-preview-title">{hoverTab.title || hoverTab.url || t("tab.new")}</span>
          {#if hoverTab.url && hoverTab.url !== hoverTab.title}
            <span class="tab-preview-url">{hoverTab.url}</span>
          {/if}
        </div>
      {/if}
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>

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

  .tab-item {
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    width: var(--tab-width);
    min-width: 0;
    max-width: var(--tab-width);
    flex: 0 1 var(--tab-width);
    border: 1px solid transparent;
    border-radius: 10px 10px 0 0;
    background: transparent;
    color: var(--ren-muted);
    font: inherit;
    font-size: 0.86rem;
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease,
      width 0.12s ease;
  }

  .tab {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    flex: 1;
    min-width: 0;
    border: none;
    border-radius: 10px 10px 0 0;
    padding: 0.5rem 0.15rem 0.5rem 0.7rem;
    background: transparent;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: grab;
  }

  .tab-item:hover:not(.active) {
    background: var(--ren-tab-hover);
    color: var(--ren-fg-secondary);
  }

  .tab-item.dragging {
    opacity: 0.55;
  }

  .tab-item.active {
    background: var(--ren-tab-active);
    color: var(--ren-fg);
    border-color: var(--ren-border);
    border-bottom-color: transparent;
    font-weight: 500;
  }

  .tab-item.split-primary,
  .tab-item.active.split-primary {
    box-shadow: inset 0 -2px 0 var(--ren-accent);
  }

  .tab-item.split:not(.active) {
    box-shadow: inset 0 -2px 0 color-mix(in srgb, var(--ren-accent) 55%, var(--ren-muted));
  }

  .tab-item.pinned {
    flex: 0 0 var(--tab-width);
    min-width: var(--tab-width);
  }

  .tab-item.pinned .tab {
    justify-content: center;
    padding-inline: 0.35rem;
    cursor: pointer;
  }

  .tab-item.pinned.active {
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

  .tab-item.active .pin-glyph {
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
    border: none;
    border-radius: 6px;
    padding: 0;
    background: transparent;
    opacity: 0;
    cursor: pointer;
    color: var(--ren-muted);
    transition:
      opacity 0.12s ease,
      background 0.12s ease,
      color 0.12s ease;
  }

  .tab-item:hover .close,
  .tab-item:focus-within .close,
  .tab-item.active .close {
    opacity: 0.8;
  }

  .close:hover {
    opacity: 1;
    background: var(--ren-tab-hover);
    color: var(--ren-fg);
  }

  .close:focus-visible {
    opacity: 1;
    outline: 2px solid var(--ren-accent);
    outline-offset: -1px;
  }

  .tab:focus-visible {
    outline: 2px solid var(--ren-accent);
    outline-offset: -2px;
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

  :global(.tab-context-menu) {
    z-index: 1100;
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

  :global(.tab-context-menu [role="menuitem"]) {
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

  :global(.tab-context-menu [data-highlighted]) {
    background: var(--ren-tab-hover);
  }

  :global(.tab-context-menu .tab-context-danger) {
    color: var(--ren-danger);
  }

  :global(.tab-context-separator) {
    border: none;
    border-top: 1px solid var(--ren-border);
    margin: 0.15rem 0;
  }

  :global(.tab-preview-popover) {
    z-index: 1100;
    width: 280px;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    overflow: hidden;
    pointer-events: none;
  }

  :global(.tab-preview-thumb.thumb) {
    width: 100%;
    height: 14rem;
    border-bottom: 1px solid var(--ren-border);
  }

  :global(.tab-preview-thumb.thumb .thumb-viewport) {
    min-height: 14rem;
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
