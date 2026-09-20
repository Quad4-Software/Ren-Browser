<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Dialog } from "bits-ui";
  import { ChevronLeft, X } from "@lucide/svelte";
  import { t } from "$lib/i18n/i18n.svelte";

  type Chapter = {
    id: string;
    title: string;
  };

  type Props = {
    chapters: Chapter[];
    activeIndex: number;
    variant?: "sidebar" | "drawer";
    onSelect: (index: number) => void;
    onClose?: () => void;
    onCollapse?: () => void;
  };

  let {
    chapters,
    activeIndex,
    variant = "sidebar",
    onSelect,
    onClose = () => {},
    onCollapse = () => {},
  }: Props = $props();

  function selectChapter(index: number) {
    onSelect(index);
    if (variant === "drawer") {
      onClose();
    }
  }
</script>

{#if variant === "drawer"}
  <Dialog.Root
    open={true}
    onOpenChange={(next) => {
      if (!next) {
        onClose();
      }
    }}
  >
    <Dialog.Portal>
      <Dialog.Overlay class="toc-drawer-scrim" />
      <Dialog.Content class="toc-drawer">
        <header class="toc-header">
          <Dialog.Title class="toc-header-title">{t("documents.toc")}</Dialog.Title>
          <button
            type="button"
            class="toc-icon-btn"
            aria-label={t("documents.closeToc")}
            onclick={onClose}
          >
            <X size={16} />
          </button>
        </header>
        <ul class="toc-list">
          {#each chapters as item, index (item.id)}
            <li>
              <button
                type="button"
                class:active={index === activeIndex}
                onclick={() => selectChapter(index)}
              >
                {item.title}
              </button>
            </li>
          {/each}
        </ul>
      </Dialog.Content>
    </Dialog.Portal>
  </Dialog.Root>
{:else}
  <nav class="toc-panel sidebar" aria-label={t("documents.toc")}>
    <header class="toc-title-row">
      <h2 class="toc-title">{t("documents.toc")}</h2>
      <button
        type="button"
        class="toc-icon-btn"
        aria-label={t("documents.collapseToc")}
        onclick={onCollapse}
      >
        <ChevronLeft size={16} />
      </button>
    </header>
    <ul class="toc-list">
      {#each chapters as item, index (item.id)}
        <li>
          <button
            type="button"
            class:active={index === activeIndex}
            aria-current={index === activeIndex ? "location" : undefined}
            onclick={() => selectChapter(index)}
          >
            {item.title}
          </button>
        </li>
      {/each}
    </ul>
  </nav>
{/if}

<style>
  :global(.toc-drawer-scrim) {
    position: fixed;
    inset: 0;
    z-index: 120;
    background: rgb(0 0 0 / 0.45);
  }

  .toc-panel {
    display: flex;
    flex-direction: column;
    min-height: 0;
    background: var(--ren-chrome-bg);
  }

  .toc-panel.sidebar {
    width: 100%;
    height: 100%;
    border-right: 1px solid var(--ren-border);
  }

  :global(.toc-drawer) {
    position: fixed;
    top: 0;
    left: 0;
    z-index: 121;
    width: min(18rem, 88vw);
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-right: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
  }

  .toc-title-row,
  .toc-header {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    min-height: 2.5rem;
    padding: 0.35rem 0.35rem 0.35rem 0.75rem;
    border-bottom: 1px solid var(--ren-border);
    flex-shrink: 0;
  }

  .toc-title,
  :global(.toc-header-title) {
    margin: 0;
    flex: 1;
    min-width: 0;
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--ren-fg);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .toc-icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    padding: 0;
    margin: 0;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 8px;
    cursor: pointer;
    flex-shrink: 0;
  }

  :global(.toc-drawer) .toc-header {
    padding: 0.35rem 0.25rem 0.35rem 0.75rem;
  }

  :global(.toc-drawer) .toc-icon-btn {
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 6px;
  }

  .toc-list {
    list-style: none;
    margin: 0;
    padding: 0.35rem;
    overflow: auto;
    flex: 1;
    display: grid;
    gap: 0.2rem;
    align-content: start;
  }

  .toc-list button {
    width: 100%;
    text-align: left;
    border: none;
    background: transparent;
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.45rem 0.55rem;
    font: inherit;
    font-size: 0.86rem;
    cursor: pointer;
  }

  .toc-list button:hover,
  .toc-list button.active {
    background: var(--ren-tab-hover);
  }
</style>
