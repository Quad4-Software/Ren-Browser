<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  type Props = {
    open: boolean;
    title: string;
    message: string;
    confirmLabel?: string;
    cancelLabel?: string;
    onConfirm: () => void;
    onCancel: () => void;
  };

  let {
    open,
    title,
    message,
    confirmLabel = "OK",
    cancelLabel = "Cancel",
    onConfirm,
    onCancel,
  }: Props = $props();

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      onCancel();
    }
  }
</script>

<svelte:window onkeydown={open ? handleKeyDown : undefined} />

{#if open}
  <button type="button" class="backdrop" aria-label="Close dialog" onclick={onCancel}></button>
  <div
    class="dialog"
    role="alertdialog"
    aria-modal="true"
    aria-labelledby="confirm-dialog-title"
    aria-describedby="confirm-dialog-message"
  >
    <h2 id="confirm-dialog-title">{title}</h2>
    <p id="confirm-dialog-message">{message}</p>
    <div class="actions">
      <button type="button" class="cancel-btn" onclick={onCancel}>{cancelLabel}</button>
      <button type="button" class="confirm-btn" onclick={onConfirm}>{confirmLabel}</button>
    </div>
  </div>
{/if}

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    z-index: 1200;
    border: none;
    background: rgb(0 0 0 / 0.45);
    cursor: default;
  }

  .dialog {
    position: fixed;
    top: 50%;
    left: 50%;
    z-index: 1201;
    width: min(24rem, calc(100vw - 2rem));
    transform: translate(-50%, -50%);
    padding: 1.1rem 1.15rem 1rem;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    display: grid;
    gap: 0.85rem;
  }

  h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--ren-fg);
  }

  p {
    margin: 0;
    font-size: 0.92rem;
    line-height: 1.45;
    color: var(--ren-fg-secondary);
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding-top: 0.15rem;
  }

  .cancel-btn,
  .confirm-btn {
    border: 1px solid var(--ren-border);
    border-radius: 10px;
    padding: 0.5rem 0.85rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    transition:
      background 0.15s ease,
      border-color 0.15s ease,
      color 0.15s ease;
  }

  .cancel-btn {
    background: transparent;
    color: var(--ren-fg);
  }

  .cancel-btn:hover {
    background: var(--ren-tab-hover);
  }

  .confirm-btn {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
    color: #fff;
  }

  .confirm-btn:hover {
    background: var(--ren-accent-hover);
    border-color: var(--ren-accent-hover);
  }
</style>
