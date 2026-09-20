<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { DropdownMenu } from "bits-ui";
  import { FileCode, Download } from "@lucide/svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    x: number;
    y: number;
    canViewSource: boolean;
    onViewSource: () => void;
    onDownload: () => void;
    onClose: () => void;
  };

  let { x, y, canViewSource, onViewSource, onDownload, onClose }: Props = $props();

  const anchor = $derived({
    getBoundingClientRect: () => new DOMRect(x, y, 0, 0),
  });
</script>

<DropdownMenu.Root
  open={true}
  onOpenChange={(next) => {
    if (!next) {
      onClose();
    }
  }}
>
  <DropdownMenu.Portal>
    <DropdownMenu.Content class="page-context-menu" customAnchor={anchor} align="start">
      {#if canViewSource}
        <DropdownMenu.Item textValue={t("content.viewSource")} onSelect={onViewSource}>
          <FileCode size={14} />
          <span>{t("content.viewSource")}</span>
        </DropdownMenu.Item>
      {/if}
      <DropdownMenu.Item textValue={t("content.downloadPage")} onSelect={onDownload}>
        <Download size={14} />
        <span>{t("content.downloadPage")}</span>
      </DropdownMenu.Item>
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  :global(.page-context-menu) {
    z-index: 1100;
    min-width: 10rem;
    max-width: calc(100vw - 1rem);
    padding: 0.35rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    display: grid;
    gap: 0.15rem;
  }

  :global(.page-context-menu [data-bits-ui-dropdown-menu-item]),
  :global(.page-context-menu [role="menuitem"]) {
    display: flex;
    align-items: center;
    gap: 0.45rem;
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

  :global(.page-context-menu [data-highlighted]) {
    background: var(--ren-tab-hover);
  }
</style>
