<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { AlertDialog } from "bits-ui";
  import PageErrorState from "$lib/components/PageErrorState.svelte";
  import { isStoreBlockingKind, pageErrorContent, type StoreErrorKind } from "$lib/browser/errors";

  type Props = {
    kind: string;
    detail: string;
    path: string;
    onResetDatabase: () => void;
    onRetry?: () => void;
  };

  let { kind, detail, path, onResetDatabase, onRetry }: Props = $props();

  const storeKind = $derived(
    isStoreBlockingKind(kind) ? (kind as StoreErrorKind) : ("database_corrupt" as StoreErrorKind),
  );
  const copy = $derived(pageErrorContent(storeKind, detail));
</script>

<AlertDialog.Root open={true}>
  <AlertDialog.Portal>
    <AlertDialog.Overlay class="appstore-error-overlay" />
    <AlertDialog.Content
      class="appstore-error-panel"
      interactOutsideBehavior="ignore"
      onEscapeKeydown={(e) => e.preventDefault()}
    >
      <AlertDialog.Title class="appstore-error-sr">{copy.title}</AlertDialog.Title>
      {#if copy.description}
        <AlertDialog.Description class="appstore-error-sr">
          {copy.description}
        </AlertDialog.Description>
      {/if}
      <PageErrorState
        error={detail}
        errorKind={storeKind}
        currentURL={path}
        {onRetry}
        onResetDatabase={copy.showResetDatabase ? onResetDatabase : undefined}
      />
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>

<style>
  :global(.appstore-error-overlay) {
    position: fixed;
    inset: 0;
    z-index: 1300;
    background: rgb(0 0 0 / 0.5);
  }

  :global(.appstore-error-panel) {
    position: fixed;
    top: 50%;
    left: 50%;
    z-index: 1301;
    width: min(34rem, calc(100vw - 3rem));
    max-height: calc(100dvh - 3rem);
    overflow: auto;
    transform: translate(-50%, -50%);
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 4px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
  }

  :global(.appstore-error-sr) {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0 0 0 0);
    white-space: nowrap;
    border: 0;
  }
</style>
