<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { AlertDialog } from "bits-ui";
  import { t } from "$lib/i18n/i18n.svelte";

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
    confirmLabel = t("common.ok"),
    cancelLabel = t("common.cancel"),
    onConfirm,
    onCancel,
  }: Props = $props();

  let confirmed = false;
</script>

<AlertDialog.Root
  {open}
  onOpenChange={(next) => {
    if (!next) {
      if (!confirmed) {
        onCancel();
      }
      confirmed = false;
    }
  }}
>
  <AlertDialog.Portal>
    <AlertDialog.Overlay class="confirm-backdrop" />
    <AlertDialog.Content class="confirm-dialog" interactOutsideBehavior="close">
      <AlertDialog.Title id="confirm-dialog-title" class="confirm-title">
        {title}
      </AlertDialog.Title>
      <AlertDialog.Description id="confirm-dialog-message" class="confirm-message">
        {message}
      </AlertDialog.Description>
      <div class="confirm-actions">
        <AlertDialog.Cancel type="button" class="confirm-cancel-btn"
          >{cancelLabel}</AlertDialog.Cancel
        >
        <AlertDialog.Action
          type="button"
          class="confirm-confirm-btn"
          onclick={() => {
            confirmed = true;
            onConfirm();
          }}
        >
          {confirmLabel}
        </AlertDialog.Action>
      </div>
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>

<style>
  :global(.confirm-backdrop) {
    position: fixed;
    inset: 0;
    z-index: 1200;
    background: rgb(0 0 0 / 0.45);
  }

  :global(.confirm-dialog) {
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

  :global(.confirm-title) {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--ren-fg);
  }

  :global(.confirm-message) {
    margin: 0;
    font-size: 0.92rem;
    line-height: 1.45;
    color: var(--ren-fg-secondary);
  }

  :global(.confirm-actions) {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding-top: 0.15rem;
  }

  :global(.confirm-cancel-btn),
  :global(.confirm-confirm-btn) {
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

  :global(.confirm-cancel-btn) {
    background: transparent;
    color: var(--ren-fg);
  }

  :global(.confirm-cancel-btn:hover) {
    background: var(--ren-tab-hover);
  }

  :global(.confirm-confirm-btn) {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
    color: #fff;
  }

  :global(.confirm-confirm-btn:hover) {
    background: var(--ren-accent-hover);
    border-color: var(--ren-accent-hover);
  }
</style>
