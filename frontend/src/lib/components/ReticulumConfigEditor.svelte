<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  type Props = {
    configText: string;
    configPath: string;
    saving: boolean;
    error: string;
    onChange: (text: string) => void;
    onSave: () => void;
    onReload: () => void;
  };

  let {
    configText = $bindable(),
    configPath,
    saving,
    error,
    onChange,
    onSave,
    onReload,
  }: Props = $props();
</script>

<section class="config-editor">
  <div class="header">
    <h3>Reticulum config</h3>
    <p class="hint">{configPath}</p>
  </div>

  <textarea
    class="editor"
    bind:value={configText}
    spellcheck="false"
    oninput={() => onChange(configText)}
    aria-label="Reticulum configuration file"
  ></textarea>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="actions">
    <button type="button" onclick={onReload} disabled={saving}>Reload</button>
    <button type="button" class="primary" onclick={onSave} disabled={saving}>
      {saving ? "Saving..." : "Save and restart interfaces"}
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
