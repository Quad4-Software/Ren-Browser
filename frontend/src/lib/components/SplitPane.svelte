<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import type { Snippet } from "svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    capturePointer,
    clampSplitRatio,
    createDragScheduler,
    releasePointer,
    splitRatioFromPointer,
  } from "$lib/browser/pane-resize";

  type Props = {
    ratio: number;
    onRatioChange: (ratio: number) => void;
    primary: Snippet;
    secondary: Snippet;
  };

  let { ratio, onRatioChange, primary, secondary }: Props = $props();

  let dragging = $state(false);
  let rootEl = $state<HTMLDivElement | null>(null);
  let dividerEl = $state<HTMLButtonElement | null>(null);

  const drag = createDragScheduler((event) => {
    if (!rootEl) {
      return;
    }
    const next = splitRatioFromPointer(event, rootEl.getBoundingClientRect());
    rootEl.style.setProperty("--split-ratio", `${next}%`);
  });

  function commitRatio() {
    if (!rootEl) {
      return;
    }
    const raw = Number.parseFloat(rootEl.style.getPropertyValue("--split-ratio"));
    if (Number.isFinite(raw)) {
      onRatioChange(raw);
    }
  }

  function onPointerDown(event: PointerEvent) {
    if (!rootEl || !dividerEl) {
      return;
    }
    dragging = true;
    capturePointer(dividerEl, event);
    rootEl.style.setProperty("--split-ratio", `${ratio}%`);
  }

  function onPointerMove(event: PointerEvent) {
    if (!dragging) {
      return;
    }
    drag.move(event);
  }

  function onPointerUp(event: PointerEvent) {
    drag.cancel();
    releasePointer(dividerEl, event);
    if (dragging) {
      commitRatio();
    }
    dragging = false;
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") {
      return;
    }
    event.preventDefault();
    const delta = event.key === "ArrowLeft" ? -2 : 2;
    onRatioChange(clampSplitRatio(ratio + delta));
  }
</script>

<svelte:window
  onpointermove={dragging ? onPointerMove : undefined}
  onpointerup={dragging ? onPointerUp : undefined}
  onpointercancel={dragging ? onPointerUp : undefined}
/>

<div class="split-root" class:dragging bind:this={rootEl} style:--split-ratio="{ratio}%">
  <div class="pane primary">
    {@render primary()}
  </div>
  <button
    type="button"
    class="divider"
    bind:this={dividerEl}
    aria-label={t("split.resize")}
    onpointerdown={onPointerDown}
    onpointermove={onPointerMove}
    onpointerup={onPointerUp}
    onpointercancel={onPointerUp}
    onkeydown={onKeydown}
  ></button>
  <div class="pane secondary">
    {@render secondary()}
  </div>
</div>

<style>
  .split-root {
    display: flex;
    min-height: 0;
    height: 100%;
    width: 100%;
    contain: layout;
  }

  .split-root.dragging {
    cursor: col-resize;
    user-select: none;
  }

  .split-root.dragging :global(iframe) {
    pointer-events: none;
  }

  .pane {
    min-width: 0;
    min-height: 0;
    overflow: hidden;
    contain: layout;
  }

  .pane.primary {
    flex: 0 0 var(--split-ratio, 52%);
  }

  .pane.secondary {
    flex: 1;
    border-left: 1px solid var(--ren-border);
  }

  .divider {
    width: 5px;
    flex-shrink: 0;
    cursor: col-resize;
    background: transparent;
    position: relative;
    border: none;
    padding: 0;
    touch-action: none;
  }

  .divider::before {
    content: "";
    position: absolute;
    inset: 0 -4px;
  }

  .divider::after {
    content: "";
    position: absolute;
    inset: 0;
    background: var(--ren-border);
    opacity: 0.55;
    transition: opacity 0.12s ease;
  }

  .divider:hover::after,
  .split-root.dragging .divider::after {
    opacity: 1;
    background: var(--ren-accent);
  }
</style>
