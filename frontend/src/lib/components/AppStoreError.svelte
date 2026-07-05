<script lang="ts">
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

<div class="overlay" role="alertdialog" aria-modal="true" aria-labelledby="store-error-title">
  <div class="panel">
    <PageErrorState
      error={detail}
      errorKind={storeKind}
      currentURL={path}
      {onRetry}
      onResetDatabase={copy.showResetDatabase ? onResetDatabase : undefined}
    />
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    z-index: 1300;
    display: grid;
    place-items: center;
    padding: 1.5rem;
    background: rgb(0 0 0 / 0.5);
  }

  .panel {
    width: min(34rem, 100%);
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 4px);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
  }
</style>
