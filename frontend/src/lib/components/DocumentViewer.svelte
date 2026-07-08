<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { t } from "$lib/i18n/i18n.svelte";
  import { isDocumentContentType } from "$lib/documents/types";
  import type { Component } from "svelte";

  type ViewerProps = {
    binaryB64: string;
    onRetry?: () => void;
  };

  type Props = {
    contentType: string;
    binaryB64: string;
    pageError?: string;
    onRetry?: () => void;
  };

  let { contentType, binaryB64, pageError = "", onRetry = () => {} }: Props = $props();

  let ViewerComponent = $state<Component<ViewerProps> | null>(null);

  const missingData = $derived(
    isDocumentContentType(contentType) && !binaryB64.trim() && !pageError.trim(),
  );

  $effect(() => {
    if (!isDocumentContentType(contentType) || pageError.trim() || missingData) {
      ViewerComponent = null;
      return;
    }
    let cancelled = false;
    const loader =
      contentType === "pdf"
        ? import("$lib/components/PdfViewer.svelte")
        : import("$lib/components/EpubViewer.svelte");
    void loader.then((module) => {
      if (!cancelled) {
        ViewerComponent = module.default;
      }
    });
    return () => {
      cancelled = true;
    };
  });
</script>

{#if isDocumentContentType(contentType)}
  {#if pageError.trim()}
    <div class="document-error">
      <p>{pageError}</p>
      <button type="button" onclick={onRetry}>{t("documents.retry")}</button>
    </div>
  {:else if missingData}
    <div class="document-error">
      <p>{t("documents.missingData")}</p>
      <button type="button" onclick={onRetry}>{t("documents.retry")}</button>
    </div>
  {:else if ViewerComponent}
    <ViewerComponent {binaryB64} {onRetry} />
  {:else}
    <div class="document-loading">{t("documents.loading")}</div>
  {/if}
{/if}

<style>
  .document-error,
  .document-loading {
    margin: auto;
    max-width: 28rem;
    padding: 2rem 1.25rem;
    text-align: center;
  }

  .document-error {
    display: grid;
    gap: 0.75rem;
  }

  .document-loading {
    color: var(--ren-muted);
    font-size: 0.9rem;
  }

  .document-error p {
    margin: 0;
    color: var(--ren-danger, #e5484d);
    overflow-wrap: anywhere;
  }

  .document-error button {
    justify-self: center;
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    border-radius: 8px;
    padding: 0.45rem 0.85rem;
    font: inherit;
    cursor: pointer;
  }
</style>
