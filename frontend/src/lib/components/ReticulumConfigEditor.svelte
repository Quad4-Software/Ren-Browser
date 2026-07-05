<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { t } from "$lib/i18n/i18n.svelte";

  type Props = {
    configText: string;
    configPath: string;
    saving: boolean;
    error: string;
    onChange: (text: string) => void;
    onSave: () => void;
    onReload: () => void;
    onExport?: () => void;
    showTitle?: boolean;
  };

  let {
    configText = $bindable(),
    configPath,
    saving,
    error,
    onChange,
    onSave,
    onReload,
    onExport,
    showTitle = true,
  }: Props = $props();

  function exportConfig() {
    if (onExport) {
      onExport();
      return;
    }
    const blob = new Blob([configText], { type: "text/plain;charset=utf-8" });
    const href = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = href;
    a.download = "reticulum.conf";
    a.click();
    URL.revokeObjectURL(href);
  }
</script>

<section class="config-editor">
  {#if showTitle}
    <div class="header">
      <h3>{t("config.title")}</h3>
      <p class="hint">{configPath}</p>
    </div>
  {:else if configPath}
    <p class="hint path-only">{configPath}</p>
  {/if}

  <textarea
    class="editor"
    bind:value={configText}
    spellcheck="false"
    oninput={() => onChange(configText)}
    aria-label={t("config.ariaLabel")}
  ></textarea>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="actions">
    <button type="button" onclick={onReload} disabled={saving}>{t("config.reload")}</button>
    <button type="button" onclick={exportConfig} disabled={saving || !configText}
      >{t("config.export")}</button
    >
    <button type="button" class="primary" onclick={onSave} disabled={saving}>
      {saving ? t("common.saving") : t("config.save")}
    </button>
  </div>
</section>

<style>
  .config-editor {
    display: grid;
    gap: 0.65rem;
  }

  .header h3 {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .hint {
    margin: 0.2rem 0 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
    word-break: break-all;
  }

  .path-only {
    margin: 0;
  }

  .editor {
    min-height: 12rem;
    max-height: 40vh;
    resize: vertical;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.8rem;
    line-height: 1.45;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: var(--ren-radius);
    padding: 0.65rem 0.75rem;
  }

  .editor:focus {
    outline: none;
    border-color: var(--ren-focus);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--ren-focus) 28%, transparent);
  }

  .error {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.85rem;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  button {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
    cursor: pointer;
  }

  button:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  button.primary {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
    color: #fff;
  }
</style>
