<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Minus, Square, X } from "@lucide/svelte";
  import { System, Window } from "@wailsio/runtime";

  const desktop = System.IsDesktop();

  async function minimize() {
    await Window.Minimise();
  }

  async function maximize() {
    await Window.ToggleMaximise();
  }

  async function close() {
    await Window.Close();
  }
</script>

{#if desktop}
  <div class="window-controls" style:--wails-draggable="no-drag">
    <button type="button" class="win-btn" aria-label="Minimize" onclick={minimize}>
      <Minus size={14} />
    </button>
    <button type="button" class="win-btn" aria-label="Maximize" onclick={maximize}>
      <Square size={12} />
    </button>
    <button type="button" class="win-btn close" aria-label="Close" onclick={close}>
      <X size={14} />
    </button>
  </div>
{/if}

<style>
  .window-controls {
    display: inline-flex;
    align-items: center;
    gap: 0.15rem;
    flex-shrink: 0;
    margin-bottom: 0.15rem;
  }

  .win-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border: none;
    border-radius: 8px;
    background: transparent;
    color: var(--ren-muted);
    cursor: pointer;
    transition:
      background 0.15s ease,
      color 0.15s ease;
  }

  .win-btn:hover {
    background: var(--ren-tab-hover);
    color: var(--ren-fg);
  }

  .win-btn.close:hover {
    background: var(--ren-danger);
    color: #fff;
  }
</style>
