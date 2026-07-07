<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { TriangleAlert } from "@lucide/svelte";
  import { buildCrashDebugLog } from "$lib/browser/crash-log";
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    message: string;
    cause?: unknown;
    closing?: boolean;
    onReload: () => void;
    onClose: () => void;
  };

  let { message, cause, closing = false, onReload, onClose }: Props = $props();

  let copyState = $state<"idle" | "copied" | "failed">("idle");

  const copyLabel = $derived(
    copyState === "copied"
      ? t("crash.copyLogsCopied")
      : copyState === "failed"
        ? t("crash.copyLogsFailed")
        : t("crash.copyLogs"),
  );

  async function copyDebugLogs() {
    try {
      await navigator.clipboard.writeText(buildCrashDebugLog(message, cause));
      copyState = "copied";
      window.setTimeout(() => {
        copyState = "idle";
      }, 2000);
    } catch {
      copyState = "failed";
      window.setTimeout(() => {
        copyState = "idle";
      }, 2500);
    }
  }
</script>

<div class="crash-page" role="alertdialog" aria-modal="true" aria-labelledby="crash-title">
  <div class="panel">
    <div class="icon" aria-hidden="true">
      <TriangleAlert size={28} strokeWidth={1.75} />
    </div>
    <h1 id="crash-title">{t("crash.title")}</h1>
    <p class="description">{t("crash.description")}</p>
    {#if message}
      <pre class="message">{message}</pre>
    {/if}
    <div class="actions">
      <button type="button" class="secondary" onclick={() => void copyDebugLogs()}>
        {copyLabel}
      </button>
      <button type="button" class="primary" onclick={onReload}>{t("crash.reload")}</button>
      <button type="button" class="danger" disabled={closing} onclick={onClose}>
        {t("crash.closeApp")}
      </button>
    </div>
  </div>
</div>

<style>
  .crash-page {
    min-height: 100vh;
    min-height: 100dvh;
    display: grid;
    place-items: center;
    padding: 1.5rem;
    background: var(--ren-surface-bg);
    color: var(--ren-fg);
  }

  .panel {
    width: min(28rem, 100%);
    text-align: center;
  }

  .icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 3.5rem;
    height: 3.5rem;
    border-radius: 999px;
    margin-bottom: 1rem;
    background: color-mix(in srgb, var(--ren-danger) 14%, var(--ren-chrome-bg));
    color: var(--ren-danger);
  }

  h1 {
    margin: 0 0 0.5rem;
    font-size: 1.2rem;
    font-weight: 600;
  }

  .description {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.92rem;
    line-height: 1.5;
  }

  .message {
    margin: 1rem 0 0;
    padding: 0.75rem 0.85rem;
    border: 1px solid var(--ren-border);
    border-radius: 10px;
    background: var(--ren-input-bg);
    color: var(--ren-fg-secondary);
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.78rem;
    line-height: 1.45;
    text-align: left;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 9rem;
    overflow: auto;
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.55rem;
    margin-top: 1.15rem;
  }

  .primary,
  .secondary,
  .danger {
    border-radius: 10px;
    padding: 0.52rem 0.95rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    transition:
      background 0.15s ease,
      border-color 0.15s ease,
      color 0.15s ease;
  }

  .primary {
    border: 1px solid var(--ren-accent);
    background: var(--ren-accent);
    color: #fff;
  }

  .primary:hover {
    background: var(--ren-accent-hover);
    border-color: var(--ren-accent-hover);
  }

  .secondary {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
  }

  .secondary:hover {
    background: var(--ren-tab-hover);
  }

  .danger {
    border: 1px solid color-mix(in srgb, var(--ren-danger) 55%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-danger) 12%, var(--ren-chrome-bg));
    color: var(--ren-danger);
  }

  .danger:hover:not(:disabled) {
    background: color-mix(in srgb, var(--ren-danger) 20%, var(--ren-chrome-bg));
  }

  .danger:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
