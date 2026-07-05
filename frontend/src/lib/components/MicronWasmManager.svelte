<script lang="ts">
  import { onMount } from "svelte";
  import { Plus, Trash2 } from "@lucide/svelte";
  import {
    BUNDLED_MICRON_WASM_PARSER_ID,
    addMicronWasmParserFromGitHub,
    addMicronWasmParserFromUpload,
    bundledMicronWasmReleaseTag,
    listMicronWasmParsers,
    removeMicronWasmParser,
    type MicronWasmParserListEntry,
  } from "$lib/micron/wasm-store";
  import {
    invalidateNomadMicronWasmPreload,
    isMicronWasmBundled,
    isWebAssemblySupported,
    preloadNomadMicronWasm,
    probeBundledMicronWasmByteLength,
  } from "$lib/micron/wasm-loader";

  type Props = {
    selectedParserId: string;
    wasmEnabled: boolean;
    onSelectParser: (parserId: string) => void;
    onWasmReadyChange: (ready: boolean) => void;
  };

  let { selectedParserId, wasmEnabled, onSelectParser, onWasmReadyChange }: Props = $props();

  let parsers = $state<MicronWasmParserListEntry[]>([]);
  let releaseTag = $state("");
  let busy = $state(false);
  let error = $state("");
  let uploadInput: HTMLInputElement | undefined = $state();

  const wasmSupported = isWebAssemblySupported();
  const wasmBundled = isMicronWasmBundled();
  const defaultTag = bundledMicronWasmReleaseTag();

  async function refreshList() {
    const bundledBytes = wasmBundled ? await probeBundledMicronWasmByteLength() : 0;
    parsers = await listMicronWasmParsers({
      includeBundled: wasmBundled,
      bundledByteLength: bundledBytes,
    });
  }

  async function activateParser(parserId: string) {
    error = "";
    onSelectParser(parserId);
    if (!wasmEnabled || !wasmSupported) {
      onWasmReadyChange(false);
      return;
    }
    invalidateNomadMicronWasmPreload();
    const ready = await preloadNomadMicronWasm(parserId);
    onWasmReadyChange(ready);
    if (!ready && parserId !== BUNDLED_MICRON_WASM_PARSER_ID) {
      error = "Parser failed to load. Micron pages will fall back to JavaScript.";
    }
  }

  async function fetchFromGitHub() {
    const tag = releaseTag.trim();
    if (!tag) {
      error = "Enter a release tag (for example v1.0.6).";
      return;
    }
    busy = true;
    error = "";
    try {
      const id = await addMicronWasmParserFromGitHub(tag);
      await refreshList();
      await activateParser(id);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  }

  async function onUploadSelected(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    busy = true;
    error = "";
    try {
      const id = await addMicronWasmParserFromUpload(file);
      await refreshList();
      await activateParser(id);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
      input.value = "";
    }
  }

  async function removeParser(parserId: string) {
    if (!confirm("Remove this WASM parser from local storage?")) {
      return;
    }
    busy = true;
    error = "";
    try {
      await removeMicronWasmParser(parserId);
      await refreshList();
      const fallback =
        parsers.find((entry) => entry.id === BUNDLED_MICRON_WASM_PARSER_ID)?.id ??
        parsers[0]?.id ??
        BUNDLED_MICRON_WASM_PARSER_ID;
      const next = selectedParserId === parserId ? fallback : selectedParserId;
      await activateParser(next);
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      busy = false;
    }
  }

  function formatBytes(bytes: number): string {
    if (!bytes) {
      return "";
    }
    if (bytes < 1024) {
      return `${bytes} B`;
    }
    if (bytes < 1024 * 1024) {
      return `${Math.round(bytes / 1024)} KB`;
    }
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }

  onMount(() => {
    releaseTag = defaultTag;
    void refreshList();
  });
</script>

<div class="wasm-manager">
  {#if !wasmSupported}
    <p class="warn">
      This webview does not support WebAssembly. Micron pages use the JavaScript parser instead.
    </p>
  {:else if !wasmBundled}
    <p class="hint">
      No bundled micron-parser-go.wasm in this build. Fetch a release below, upload a .wasm file, or
      run the vendor fetch script to install the default parser and wasm_exec.js.
    </p>
  {/if}

  {#if parsers.length > 0}
    <ul class="parser-list">
      {#each parsers as parser (parser.id)}
        <li class:selected={selectedParserId === parser.id}>
          <label class="parser-row">
            <input
              type="radio"
              name="micron-wasm-parser"
              value={parser.id}
              checked={selectedParserId === parser.id}
              disabled={busy || !wasmEnabled || !wasmSupported}
              onchange={() => void activateParser(parser.id)}
            />
            <span class="parser-meta">
              <span class="parser-label">{parser.label}</span>
              <span class="parser-detail">
                {parser.source}
                {#if parser.releaseTag}
                  · {parser.releaseTag}
                {/if}
                {#if parser.byteLength}
                  · {formatBytes(parser.byteLength)}
                {/if}
              </span>
            </span>
          </label>
          {#if parser.removable}
            <button
              type="button"
              class="remove-btn"
              aria-label="Remove parser"
              disabled={busy}
              onclick={() => void removeParser(parser.id)}
            >
              <Trash2 size={15} />
            </button>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}

  <div class="add-row">
    <input
      type="text"
      placeholder={defaultTag}
      bind:value={releaseTag}
      disabled={busy || !wasmSupported}
      spellcheck="false"
      autocomplete="off"
    />
    <button type="button" disabled={busy || !wasmSupported} onclick={() => void fetchFromGitHub()}>
      Fetch release
    </button>
  </div>

  <div class="upload-row">
    <input
      bind:this={uploadInput}
      type="file"
      accept=".wasm,application/wasm"
      class="file-input"
      disabled={busy || !wasmSupported}
      onchange={onUploadSelected}
    />
    <button
      type="button"
      class="upload-btn"
      disabled={busy || !wasmSupported}
      onclick={() => uploadInput?.click()}
    >
      <Plus size={15} />
      Upload .wasm
    </button>
  </div>

  <p class="hint">
    Custom WASM modules are stored locally. Only install parsers you trust. Failed WASM loads fall
    back to micron-parser-js automatically.
  </p>

  {#if error}
    <p class="error">{error}</p>
  {/if}
</div>

<style>
  .wasm-manager {
    display: grid;
    gap: 0.65rem;
  }

  .parser-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
  }

  .parser-list li {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.35rem;
    align-items: center;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface-raised);
    padding: 0.45rem 0.55rem;
  }

  .parser-list li.selected {
    border-color: color-mix(in srgb, var(--ren-accent) 45%, var(--ren-border));
  }

  .parser-row {
    display: flex;
    align-items: flex-start;
    gap: 0.55rem;
    cursor: pointer;
    min-width: 0;
  }

  .parser-meta {
    display: grid;
    gap: 0.1rem;
    min-width: 0;
  }

  .parser-label {
    color: var(--ren-fg);
    font-size: 0.9rem;
    word-break: break-word;
  }

  .parser-detail {
    color: var(--ren-muted);
    font-size: 0.78rem;
    word-break: break-all;
  }

  .remove-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--ren-danger);
    cursor: pointer;
    padding: 0.2rem;
    line-height: 0;
  }

  .remove-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .add-row,
  .upload-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.45rem;
  }

  .file-input {
    display: none;
  }

  .upload-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
  }

  .warn {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.85rem;
  }

  .hint {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
  }

  .error {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.85rem;
  }

  input,
  button {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
  }

  button {
    cursor: pointer;
    white-space: nowrap;
  }

  button:disabled,
  input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
