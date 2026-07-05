<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import type { Snippet } from "svelte";

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

  function onPointerDown(event: PointerEvent) {
    if (!rootEl || !dividerEl) {
      return;
    }
    dragging = true;
    dividerEl.setPointerCapture(event.pointerId);
  }

  function onPointerMove(event: PointerEvent) {
    if (!dragging || !rootEl) {
      return;
    }
    const rect = rootEl.getBoundingClientRect();
    const next = ((event.clientX - rect.left) / rect.width) * 100;
    onRatioChange(Math.min(75, Math.max(25, next)));
  }

  function onPointerUp(event: PointerEvent) {
    if (dividerEl?.hasPointerCapture(event.pointerId)) {
      dividerEl.releasePointerCapture(event.pointerId);
    }
    dragging = false;
  }
</script>

<div class="split-root" class:dragging bind:this={rootEl}>
  <div class="pane primary" style:flex-basis="{ratio}%">
    {@render primary()}
  </div>
  <button
    type="button"
    class="divider"
    bind:this={dividerEl}
    aria-label="Resize split panes"
    onpointerdown={onPointerDown}
    onpointermove={onPointerMove}
    onpointerup={onPointerUp}
    onpointercancel={onPointerUp}
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
  }

  .split-root.dragging {
    cursor: col-resize;
    user-select: none;
  }

  .pane {
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }

  .pane.primary {
    flex-shrink: 0;
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
