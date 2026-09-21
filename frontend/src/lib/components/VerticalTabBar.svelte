<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onDestroy } from "svelte";
  import { DropdownMenu, Popover } from "bits-ui";
  import { ChevronRight, Pin, Plus, X } from "@lucide/svelte";
  import { handleTitlebarDoubleClick } from "$lib/browser/window-actions";
  import {
    groupTabsForDisplay,
    MAX_TABS,
    TAB_GROUP_COLORS,
    type Tab,
    type TabGroup,
  } from "$lib/browser/url";
  import TabPreviewThumb from "$lib/components/TabPreviewThumb.svelte";
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
    | "closeBelow"
    | "closeAll"
    | "pin"
    | "unpin";

  type Props = {
    tabs: Tab[];
    tabGroups: TabGroup[];
    nativeTitlebar: boolean;
    mobileUI: boolean;
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
    onCreateGroup: (tabId: string) => void;
    onAssignGroup: (tabId: string, groupId: string | undefined) => void;
    onRenameGroup: (groupId: string, name: string) => void;
    onGroupColor: (groupId: string, color: string) => void;
    onToggleGroupCollapsed: (groupId: string) => void;
    onRemoveGroup: (groupId: string) => void;
    onWindowChromeError?: (message: string) => void;
  };

  let {
    tabs,
    tabGroups,
    nativeTitlebar,
    mobileUI,
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
    onCreateGroup,
    onAssignGroup,
    onRenameGroup,
    onGroupColor,
    onToggleGroupCollapsed,
    onRemoveGroup,
    onWindowChromeError = () => {},
  }: Props = $props();

  let dragId = $state<string | null>(null);
  let menu = $state<{ x: number; y: number; tabId: string } | null>(null);
  let groupMenu = $state<{ x: number; y: number; groupId: string } | null>(null);
  let editingGroupId = $state<string | null>(null);
  let editingName = $state("");

  let listEl = $state<HTMLDivElement | null>(null);
  let hoverTabId = $state<string | null>(null);
  let previewAnchorEl = $state<HTMLElement | null>(null);

  const PREVIEW_OFFSET = 6;
  const PREVIEW_HOVER_DELAY_MS = 400;

  let previewTimer: ReturnType<typeof setTimeout> | undefined;

  const displayItems = $derived(groupTabsForDisplay(tabs, tabGroups));
  const hoverTab = $derived(hoverTabId ? tabs.find((tab) => tab.id === hoverTabId) : null);
  const atTabLimit = $derived(tabs.length >= MAX_TABS);

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

  function groupDisplayName(group: TabGroup): string {
    return group.name || t("tab.groupUnnamed");
  }

  const menuTabId = $derived(menu?.tabId ?? null);
  const menuTab = $derived(menuTabId ? tabs.find((tab) => tab.id === menuTabId) : null);
  const menuTabIndex = $derived(menuTabId ? tabs.findIndex((tab) => tab.id === menuTabId) : -1);
  const canCloseBelow = $derived(
    menuTabIndex >= 0 && tabs.slice(menuTabIndex + 1).some((tab) => !tab.pinned),
  );
  const canCloseOthers = $derived(tabs.length > 1);
  const showCloseSplit = $derived(splitViewOpen);

  const menuGroupId = $derived(groupMenu?.groupId ?? null);
  const menuGroup = $derived(
    menuGroupId ? tabGroups.find((group) => group.id === menuGroupId) : null,
  );

  const menuAnchor = $derived.by(() => {
    const current = menu;
    if (!current) {
      return null;
    }
    return { getBoundingClientRect: () => new DOMRect(current.x, current.y, 0, 0) };
  });

  const groupMenuAnchor = $derived.by(() => {
    const current = groupMenu;
    if (!current) {
      return null;
    }
    return { getBoundingClientRect: () => new DOMRect(current.x, current.y, 0, 0) };
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

  function dropTabId(event: DragEvent): string | null {
    return dragId ?? event.dataTransfer?.getData("text/plain") ?? null;
  }

  function handleDropOnTab(event: DragEvent, targetId: string) {
    event.preventDefault();
    event.stopPropagation();
    const fromId = dropTabId(event);
    dragId = null;
    if (!fromId || fromId === targetId) {
      return;
    }
    onReorder(fromId, targetId);
  }

  function handleDropOnGroup(event: DragEvent, groupId: string) {
    event.preventDefault();
    event.stopPropagation();
    const fromId = dropTabId(event);
    dragId = null;
    if (!fromId) {
      return;
    }
    onAssignGroup(fromId, groupId);
  }

  function openMenu(event: MouseEvent, tabId: string) {
    event.preventDefault();
    menu = { x: event.clientX, y: event.clientY, tabId };
  }

  function openMenuForTab(tabEl: HTMLElement, tabId: string) {
    const rect = tabEl.getBoundingClientRect();
    menu = { x: rect.left + 8, y: rect.bottom, tabId };
  }

  function openGroupMenu(event: MouseEvent, groupId: string) {
    event.preventDefault();
    groupMenu = { x: event.clientX, y: event.clientY, groupId };
  }

  function handleTabKeydown(event: KeyboardEvent, tabId: string) {
    const current = event.currentTarget as HTMLElement;
    if (
      event.key === "ArrowUp" ||
      event.key === "ArrowDown" ||
      event.key === "Home" ||
      event.key === "End"
    ) {
      event.preventDefault();
      const tabEls = Array.from(listEl?.querySelectorAll<HTMLElement>(".tab") ?? []);
      const index = tabEls.indexOf(current);
      if (index < 0) {
        return;
      }
      const next =
        event.key === "ArrowUp"
          ? index > 0
            ? index - 1
            : tabEls.length - 1
          : event.key === "ArrowDown"
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

  function startRenameGroup(group: TabGroup) {
    editingGroupId = group.id;
    editingName = group.name;
  }

  function commitRenameGroup() {
    if (editingGroupId) {
      onRenameGroup(editingGroupId, editingName);
    }
    editingGroupId = null;
  }

  function handleRenameKeydown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      event.preventDefault();
      commitRenameGroup();
    } else if (event.key === "Escape") {
      event.preventDefault();
      editingGroupId = null;
    }
  }

  function closeMenu() {
    menu = null;
  }

  function closeGroupMenu() {
    groupMenu = null;
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
      case "closeBelow":
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

  async function handleDragAreaDoubleClick() {
    if (nativeTitlebar || mobileUI) {
      return;
    }
    const result = await handleTitlebarDoubleClick();
    if (!result.ok) {
      onWindowChromeError(result.error || t("window.actionFailed"));
    }
  }
</script>

<aside class="vtabs" class:native-titlebar={nativeTitlebar} aria-label={t("tab.list")}>
  <div
    class="vtabs-list"
    role="list"
    bind:this={listEl}
    style:--wails-draggable="no-drag"
    ondragover={handleDragOver}
  >
    {#each displayItems as item (item.type === "group" ? `g:${item.group.id}` : item.tab.id)}
      {#if item.type === "group"}
        <div
          class="tab-group"
          role="listitem"
          class:collapsed={item.group.collapsed}
          class:dragging={item.tabs.some((member) => member.id === dragId)}
          style:--group-color={item.group.color}
          oncontextmenu={(event) => openGroupMenu(event, item.group.id)}
          ondragover={handleDragOver}
          ondrop={(event) => handleDropOnGroup(event, item.group.id)}
        >
          <button
            type="button"
            class="group-header"
            aria-expanded={!item.group.collapsed}
            aria-label={groupDisplayName(item.group)}
            onclick={() => onToggleGroupCollapsed(item.group.id)}
          >
            <ChevronRight size={12} class="group-chevron" />
            <span class="group-dot" aria-hidden="true"></span>
            {#if editingGroupId === item.group.id}
              <input
                class="group-name-input"
                bind:value={editingName}
                maxlength="48"
                placeholder={t("tab.groupName")}
                aria-label={t("tab.renameGroup")}
                onclick={(event) => event.stopPropagation()}
                onkeydown={handleRenameKeydown}
                onblur={commitRenameGroup}
              />
            {:else}
              <span
                class="group-name"
                aria-hidden="true"
                ondblclick={(event) => {
                  event.stopPropagation();
                  startRenameGroup(item.group);
                }}>{groupDisplayName(item.group)}</span
              >
            {/if}
            <span class="group-count">{item.tabs.length}</span>
          </button>
          {#if !item.group.collapsed}
            {#each item.tabs as tab (tab.id)}
              {@render tabRow(tab, true)}
            {/each}
          {/if}
        </div>
      {:else}
        {@render tabRow(item.tab, false)}
      {/if}
    {/each}
  </div>

  <div
    class="vtabs-footer"
    role="presentation"
    style:--wails-draggable={nativeTitlebar ? "no-drag" : "drag"}
    ondblclick={handleDragAreaDoubleClick}
  >
    <button
      class="new-tab ren-icon-btn"
      aria-label={t("tab.newTab")}
      disabled={atTabLimit}
      title={atTabLimit ? t("tab.limitReached", { max: MAX_TABS }) : t("tab.newTab")}
      onclick={onNew}
      style:--wails-draggable="no-drag"
    >
      <Plus size={14} />
    </button>
  </div>
</aside>

{#snippet tabRow(tab: Tab, grouped: boolean)}
  <div
    class="tab-item"
    role="listitem"
    class:active={tab.active}
    class:pinned={tab.pinned}
    class:grouped
    class:split={splitViewOpen && splitTabId === tab.id}
    class:split-primary={splitViewOpen && tab.active}
    class:dragging={dragId === tab.id}
    draggable="true"
    style:--wails-draggable="no-drag"
    oncontextmenu={(event) => openMenu(event, tab.id)}
    onmouseenter={(event) => showTabPreview(tab.id, event.currentTarget)}
    onmouseleave={hideTabPreview}
    ondragstart={(event) => handleDragStart(event, tab.id)}
    ondragend={handleDragEnd}
    ondragover={handleDragOver}
    ondrop={(event) => handleDropOnTab(event, tab.id)}
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
      {/if}
      <span class="title">{tab.title || t("tab.new")}</span>
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
{/snippet}

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
      {#if menuTab && !menuTab.pinned}
        <DropdownMenu.Separator class="tab-context-separator" />
        {#if menuTab.groupId}
          <DropdownMenu.Item
            textValue={t("tab.removeFromGroup")}
            onSelect={() => {
              const tabId = menu?.tabId;
              closeMenu();
              if (tabId) {
                onAssignGroup(tabId, undefined);
              }
            }}
          >
            {t("tab.removeFromGroup")}
          </DropdownMenu.Item>
        {/if}
        <DropdownMenu.Sub>
          <DropdownMenu.SubTrigger class="tab-context-subtrigger" textValue={t("tab.addToGroup")}>
            {t("tab.addToGroup")}
          </DropdownMenu.SubTrigger>
          <DropdownMenu.SubContent class="tab-context-menu" sideOffset={4} collisionPadding={8}>
            <DropdownMenu.Item
              textValue={t("tab.newGroup")}
              onSelect={() => {
                const tabId = menu?.tabId;
                closeMenu();
                if (tabId) {
                  onCreateGroup(tabId);
                }
              }}
            >
              {t("tab.newGroup")}
            </DropdownMenu.Item>
            {#each tabGroups as group (group.id)}
              {#if group.id !== menuTab?.groupId}
                <DropdownMenu.Item
                  textValue={groupDisplayName(group)}
                  onSelect={() => {
                    const tabId = menu?.tabId;
                    closeMenu();
                    if (tabId) {
                      onAssignGroup(tabId, group.id);
                    }
                  }}
                >
                  <span class="menu-group-dot" style:background={group.color}></span>
                  {groupDisplayName(group)}
                </DropdownMenu.Item>
              {/if}
            {/each}
          </DropdownMenu.SubContent>
        </DropdownMenu.Sub>
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
      {#if canCloseBelow}
        <DropdownMenu.Item textValue={t("tab.closeBelow")} onSelect={() => runAction("closeBelow")}>
          {t("tab.closeBelow")}
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

<DropdownMenu.Root
  open={groupMenu !== null}
  onOpenChange={(next) => {
    if (!next) {
      closeGroupMenu();
    }
  }}
>
  <DropdownMenu.Portal>
    <DropdownMenu.Content
      class="tab-context-menu"
      customAnchor={groupMenuAnchor}
      align="start"
      collisionPadding={8}
    >
      <DropdownMenu.Item
        textValue={t("tab.renameGroup")}
        onSelect={() => {
          if (menuGroup) {
            startRenameGroup(menuGroup);
          }
          closeGroupMenu();
        }}
      >
        {t("tab.renameGroup")}
      </DropdownMenu.Item>
      <DropdownMenu.Sub>
        <DropdownMenu.SubTrigger class="tab-context-subtrigger" textValue={t("tab.groupColor")}>
          {t("tab.groupColor")}
        </DropdownMenu.SubTrigger>
        <DropdownMenu.SubContent class="tab-context-menu" sideOffset={4} collisionPadding={8}>
          {#each TAB_GROUP_COLORS as color (color)}
            <DropdownMenu.Item
              textValue={color}
              onSelect={() => {
                const groupId = groupMenu?.groupId;
                closeGroupMenu();
                if (groupId) {
                  onGroupColor(groupId, color);
                }
              }}
            >
              <span
                class="menu-group-dot"
                class:selected={menuGroup?.color === color}
                style:background={color}
              ></span>
              {color}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.SubContent>
      </DropdownMenu.Sub>
      <DropdownMenu.Separator class="tab-context-separator" />
      <DropdownMenu.Item
        class="tab-context-danger"
        textValue={t("tab.ungroup")}
        onSelect={() => {
          const groupId = groupMenu?.groupId;
          closeGroupMenu();
          if (groupId) {
            onRemoveGroup(groupId);
          }
        }}
      >
        {t("tab.ungroup")}
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
      side="right"
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
  .vtabs {
    display: flex;
    flex-direction: column;
    width: var(--vtabs-width, 15rem);
    min-width: 11rem;
    max-width: 20rem;
    height: 100%;
    min-height: 0;
    background: var(--ren-chrome-bg);
    border-right: 1px solid var(--ren-border);
    overflow: hidden;
  }

  .vtabs-list {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 0.4rem 0.4rem 0.6rem;
  }

  .tab-item {
    box-sizing: border-box;
    display: flex;
    align-items: center;
    width: 100%;
    min-width: 0;
    border: 1px solid transparent;
    border-radius: 8px;
    background: transparent;
    color: var(--ren-muted);
    font: inherit;
    font-size: 0.86rem;
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease;
  }

  .tab {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    flex: 1;
    min-width: 0;
    border: none;
    border-radius: 8px;
    padding: 0.42rem 0.15rem 0.42rem 0.55rem;
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
    font-weight: 500;
  }

  .tab-item.split-primary,
  .tab-item.active.split-primary {
    box-shadow: inset 2px 0 0 var(--ren-accent);
  }

  .tab-item.split:not(.active) {
    box-shadow: inset 2px 0 0 color-mix(in srgb, var(--ren-accent) 55%, var(--ren-muted));
  }

  .tab-item.pinned .tab {
    cursor: pointer;
  }

  .tab-item.pinned.active {
    box-shadow: inset 2px 0 0 var(--ren-accent);
  }

  .pin-glyph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1rem;
    flex-shrink: 0;
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

  .tab-item.grouped .title {
    padding-left: 0.15rem;
  }

  .tab-group {
    border-left: 3px solid var(--group-color, transparent);
    border-radius: 6px;
    margin: 1px 0;
  }

  .group-header {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    width: 100%;
    min-width: 0;
    border: none;
    border-radius: 6px;
    padding: 0.32rem 0.4rem;
    background: transparent;
    color: var(--ren-fg-secondary);
    font: inherit;
    font-size: 0.78rem;
    font-weight: 600;
    text-align: left;
    cursor: pointer;
  }

  .group-header:hover {
    background: var(--ren-tab-hover);
  }

  .group-header:focus-visible {
    outline: 2px solid var(--ren-accent);
    outline-offset: -2px;
  }

  .group-header :global(.group-chevron) {
    flex-shrink: 0;
    transition: transform 0.12s ease;
    transform: rotate(90deg);
    color: var(--group-color, var(--ren-muted));
  }

  .tab-group.collapsed .group-header :global(.group-chevron) {
    transform: rotate(0deg);
  }

  .group-dot {
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    flex-shrink: 0;
    background: var(--group-color, var(--ren-muted));
  }

  .group-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .group-name-input {
    flex: 1;
    min-width: 0;
    border: 1px solid var(--ren-accent);
    border-radius: 4px;
    padding: 0.05rem 0.3rem;
    background: var(--ren-surface-bg);
    color: var(--ren-fg);
    font: inherit;
    font-size: 0.78rem;
  }

  .group-count {
    flex-shrink: 0;
    color: var(--ren-muted);
    font-size: 0.72rem;
    font-weight: 500;
  }

  .close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    width: 1.35rem;
    height: 1.35rem;
    margin-inline-end: 0.15rem;
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

  .vtabs-footer {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    padding: 0.35rem 0.4rem calc(0.35rem + env(safe-area-inset-bottom));
    border-top: 1px solid var(--ren-border);
    min-height: 2rem;
  }

  .new-tab:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }

  :global(.menu-group-dot) {
    display: inline-block;
    width: 0.6rem;
    height: 0.6rem;
    border-radius: 50%;
    margin-right: 0.4rem;
    vertical-align: -0.05rem;
  }

  :global(.menu-group-dot.selected) {
    outline: 2px solid var(--ren-fg);
    outline-offset: 1px;
  }

  :global(.tab-context-subtrigger) {
    text-align: left;
    border: none;
    background: transparent;
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.45rem 0.65rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    width: 100%;
  }

  :global(.tab-context-subtrigger[data-highlighted]) {
    background: var(--ren-tab-hover);
  }
</style>
