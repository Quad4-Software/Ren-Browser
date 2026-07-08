<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { Share2, Wifi, Square } from "@lucide/svelte";
  import { formatBytes } from "$lib/browser/download-progress";
  import { formatBindingError } from "$lib/browser/binding-errors.js";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    GetApkShareInfo,
    GetApkShareSession,
    ShareApk,
    StartApkShareServer,
    StopApkShareServer,
  } from "../../../bindings/renbrowser/internal/app/browserservice.js";
  import type {
    ApkShareInfo,
    ApkShareSession,
  } from "../../../bindings/renbrowser/internal/app/models.js";

  let info = $state<ApkShareInfo | null>(null);
  let session = $state<ApkShareSession | null>(null);
  let loading = $state(false);
  let busy = $state(false);
  let error = $state("");
  let copied = $state(false);

  async function load() {
    loading = true;
    error = "";
    try {
      info = await GetApkShareInfo();
      session = await GetApkShareSession();
      if (info?.error && !info.available) {
        error = info.error;
      }
      if (session?.error) {
        error = session.error;
      }
    } catch (err) {
      error = formatBindingError(err, t("settings.shareApkLoadFailed"));
      info = null;
      session = null;
    } finally {
      loading = false;
    }
  }

  async function shareViaApps() {
    busy = true;
    error = "";
    try {
      await ShareApk();
    } catch (err) {
      error = formatBindingError(err, t("settings.shareApkShareFailed"));
    } finally {
      busy = false;
    }
  }

  async function startSharing() {
    busy = true;
    error = "";
    copied = false;
    try {
      session = await StartApkShareServer();
      if (session?.error) {
        error = session.error;
      }
    } catch (err) {
      error = formatBindingError(err, t("settings.shareApkStartFailed"));
    } finally {
      busy = false;
    }
  }

  async function stopSharing() {
    busy = true;
    error = "";
    copied = false;
    try {
      await StopApkShareServer();
      session = await GetApkShareSession();
    } catch (err) {
      error = formatBindingError(err, t("settings.shareApkStopFailed"));
    } finally {
      busy = false;
    }
  }

  async function copyUrl() {
    const url = session?.url?.trim();
    if (!url) {
      return;
    }
    try {
      await navigator.clipboard.writeText(url);
      copied = true;
    } catch {
      error = t("settings.shareApkCopyFailed");
    }
  }

  onMount(() => {
    void load();
  });

  onDestroy(() => {
    if (session?.active) {
      void StopApkShareServer();
    }
  });
</script>

<div class="share-apk">
  {#if loading}
    <p class="hint">{t("settings.shareApkLoading")}</p>
  {:else if info?.available}
    <p class="hint">{t("settings.shareApkHint")}</p>
    <p class="meta">
      {t("settings.shareApkVersion", {
        version: info.version,
        size: formatBytes(info.size),
      })}
    </p>

    <div class="actions">
      <button type="button" class="action-btn" disabled={busy} onclick={() => void shareViaApps()}>
        <Share2 size={16} />
        {t("settings.shareApkShareSheet")}
      </button>
      {#if session?.active}
        <button
          type="button"
          class="action-btn danger"
          disabled={busy}
          onclick={() => void stopSharing()}
        >
          <Square size={16} />
          {t("settings.shareApkStop")}
        </button>
      {:else}
        <button
          type="button"
          class="action-btn primary"
          disabled={busy}
          onclick={() => void startSharing()}
        >
          <Wifi size={16} />
          {t("settings.shareApkStart")}
        </button>
      {/if}
    </div>

    {#if session?.active && session.url}
      <div class="session">
        <p class="session-label">{t("settings.shareApkUrl")}</p>
        <code class="session-url">{session.url}</code>
        <div class="session-actions">
          <button type="button" class="action-btn" disabled={busy} onclick={() => void copyUrl()}>
            {copied ? t("settings.shareApkCopied") : t("settings.shareApkCopyUrl")}
          </button>
        </div>
        {#if session.qrDataURL}
          <figure class="qr-wrap">
            <img
              class="qr"
              src={session.qrDataURL}
              alt={t("settings.shareApkScan")}
              width="256"
              height="256"
            />
            <figcaption>{t("settings.shareApkScan")}</figcaption>
          </figure>
        {/if}
      </div>
    {/if}
  {:else}
    <p class="hint">{error || t("settings.shareApkUnavailable")}</p>
  {/if}

  {#if error && info?.available}
    <p class="error">{error}</p>
  {/if}
</div>

<style>
  .share-apk {
    display: grid;
    gap: 0.75rem;
    min-width: 0;
    max-width: 100%;
  }

  .hint,
  .meta {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.9rem;
    line-height: 1.45;
    word-break: break-word;
  }

  .actions,
  .session-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    min-width: 0;
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    min-width: 0;
    max-width: 100%;
    padding: 0.55rem 0.8rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface);
    color: var(--ren-text);
    cursor: pointer;
  }

  .action-btn.primary {
    border-color: color-mix(in srgb, var(--ren-accent) 55%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-accent) 12%, var(--ren-surface));
  }

  .action-btn.danger {
    border-color: color-mix(in srgb, #d64545 45%, var(--ren-border));
  }

  .action-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .session {
    display: grid;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: color-mix(in srgb, var(--ren-surface) 88%, transparent);
    min-width: 0;
  }

  .session-label {
    margin: 0;
    font-size: 0.85rem;
    color: var(--ren-muted);
  }

  .session-url {
    display: block;
    padding: 0.55rem 0.65rem;
    border-radius: var(--ren-radius);
    background: var(--ren-content-bg);
    border: 1px solid var(--ren-border);
    font-size: 0.82rem;
    overflow-wrap: anywhere;
    word-break: break-all;
  }

  .qr-wrap {
    display: grid;
    gap: 0.35rem;
    justify-items: center;
    margin: 0.25rem 0 0;
  }

  .qr {
    display: block;
    width: min(100%, 256px);
    height: auto;
    border-radius: var(--ren-radius);
    background: #fff;
    padding: 0.35rem;
  }

  figcaption {
    margin: 0;
    font-size: 0.82rem;
    color: var(--ren-muted);
    text-align: center;
  }

  .error {
    margin: 0;
    color: #d64545;
    font-size: 0.88rem;
  }
</style>
