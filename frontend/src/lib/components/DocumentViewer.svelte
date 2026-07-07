<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import EpubViewer from "$lib/components/EpubViewer.svelte";
  import PdfViewer from "$lib/components/PdfViewer.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import { isDocumentContentType } from "$lib/documents/types";

  type Props = {
    contentType: string;
    binaryB64: string;
    pageError?: string;
    onRetry?: () => void;
  };

  let { contentType, binaryB64, pageError = "", onRetry = () => {} }: Props = $props();

  const missingData = $derived(
    isDocumentContentType(contentType) && !binaryB64.trim() && !pageError.trim(),
  );
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
  {:else if contentType === "pdf"}
    <PdfViewer {binaryB64} {onRetry} />
  {:else}
    <EpubViewer {binaryB64} {onRetry} />
  {/if}
{/if}

<style>
  .document-error {
    margin: auto;
    max-width: 28rem;
    padding: 2rem 1.25rem;
    text-align: center;
    display: grid;
    gap: 0.75rem;
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
