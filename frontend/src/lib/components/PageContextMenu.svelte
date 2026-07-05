<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
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
</script>

<svelte:window onclick={onClose} />

<div
  class="context-menu"
  style:left="{x}px"
  style:top="{y}px"
  role="menu"
  tabindex="0"
  onclick={(event) => event.stopPropagation()}
  onkeydown={(event) => {
    if (event.key === "Escape") {
      onClose();
    }
  }}
>
  {#if canViewSource}
    <button role="menuitem" onclick={onViewSource}>
      <FileCode size={14} />
      <span>{t("content.viewSource")}</span>
    </button>
  {/if}
  <button role="menuitem" onclick={onDownload}>
    <Download size={14} />
    <span>{t("content.downloadPage")}</span>
  </button>
</div>

<style>
  .context-menu {
    position: fixed;
    z-index: 1000;
    min-width: 10rem;
    padding: 0.35rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    display: grid;
    gap: 0.15rem;
  }

  .context-menu button {
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

  .context-menu button:hover {
    background: var(--ren-tab-hover);
  }
</style>
