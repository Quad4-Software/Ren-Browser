<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import {
    AlertTriangle,
    Database,
    HardDrive,
    ServerCrash,
    ServerOff,
    Unplug,
    WifiOff,
  } from "@lucide/svelte";
  import {
    normalizePageErrorKind,
    pageErrorContent,
    type PageErrorKind,
  } from "$lib/browser/errors";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    error: string;
    errorKind?: string;
    currentURL?: string;
    onRetry?: () => void;
    onResetDatabase?: () => void;
  };

  let { error, errorKind = "", currentURL = "", onRetry, onResetDatabase }: Props = $props();

  const kind = $derived(normalizePageErrorKind(errorKind, error));
  const copy = $derived(pageErrorContent(kind, error));

  function iconFor(kind: PageErrorKind) {
    switch (kind) {
      case "connection_failed":
        return ServerOff;
      case "connection_lost":
        return Unplug;
      case "not_found":
        return WifiOff;
      case "internal":
        return ServerCrash;
      case "storage_full":
        return HardDrive;
      case "database_corrupt":
        return Database;
      default:
        return AlertTriangle;
    }
  }

  const Icon = $derived(iconFor(kind));
</script>

<div class="error-page" class:warning={copy.tone === "warning"}>
  <div class="icon" aria-hidden="true">
    <Icon size={24} />
  </div>
  <h2>{copy.title}</h2>
  {#if copy.description}
    <p class="description">{copy.description}</p>
  {/if}
  {#if currentURL}
    <p class="url">{currentURL}</p>
  {/if}
  <div class="actions">
    {#if copy.showRetry && onRetry}
      <button type="button" class="primary" onclick={onRetry}>{t("errors.tryAgain")}</button>
    {/if}
    {#if copy.showResetDatabase && onResetDatabase}
      <button type="button" class="danger" onclick={onResetDatabase}>{t("errors.resetDatabase")}</button>
    {/if}
  </div>
</div>

<style>
  .error-page {
    margin: auto;
    max-width: 30rem;
    padding: 2.5rem 1.5rem;
    text-align: center;
    color: var(--ren-fg);
  }

  .error-page.warning {
    --error-accent: color-mix(in srgb, var(--ren-accent) 70%, #d97706);
  }

  .error-page:not(.warning) {
    --error-accent: var(--ren-danger);
  }

  .icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 3.25rem;
    height: 3.25rem;
    border-radius: 999px;
    margin-bottom: 0.85rem;
    background: color-mix(in srgb, var(--error-accent) 14%, var(--ren-chrome-bg));
    color: var(--error-accent);
  }

  h2 {
    margin: 0 0 0.5rem;
    font-size: 1.15rem;
    font-weight: 600;
  }

  .description {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.92rem;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .url {
    margin: 0.85rem 0 0;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.78rem;
    color: var(--ren-muted);
    word-break: break-all;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.55rem;
    margin-top: 1.1rem;
  }

  .primary,
  .danger {
    border-radius: 8px;
    padding: 0.48rem 0.95rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
  }

  .primary {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
  }

  .primary:hover {
    background: var(--ren-tab-hover);
  }

  .danger {
    border: 1px solid color-mix(in srgb, var(--ren-danger) 55%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-danger) 12%, var(--ren-chrome-bg));
    color: var(--ren-danger);
  }

  .danger:hover {
    background: color-mix(in srgb, var(--ren-danger) 20%, var(--ren-chrome-bg));
  }
</style>
