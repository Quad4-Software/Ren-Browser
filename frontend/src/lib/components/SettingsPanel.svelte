<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { FolderOpen, Network } from "@lucide/svelte";
  import Toggle from "$lib/components/Toggle.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import MicronWasmManager from "$lib/components/MicronWasmManager.svelte";
  import ReticulumConfigEditor from "$lib/components/ReticulumConfigEditor.svelte";
  import CommunityInterfaces, {
    type CommunityInterface,
  } from "$lib/components/CommunityInterfaces.svelte";
  import ExtensionsPanel from "$lib/components/ExtensionsPanel.svelte";
  import type { MicronRendererPreference } from "$lib/micron/render-page";
  import { isWebAssemblySupported } from "$lib/micron/wasm-loader";
  import type { ThemeSettings } from "$lib/theme/tokens";
  import {
    KEYBIND_LABELS,
    chordFromEvent,
    formatChord,
    setKeybindRecording,
    type KeybindAction,
    type KeybindSettings,
  } from "$lib/browser/keybinds";

  type InterfaceRow = {
    name: string;
    type: string;
    enabled: boolean;
    online: boolean;
    txBytes: number;
    rxBytes: number;
  };

  type Props = {
    theme: ThemeSettings;
    systemFonts: string[];
    keybinds: KeybindSettings;
    interfaces: InterfaceRow[];
    configPath: string;
    downloadDir: string;
    openLinksInNewTab: boolean;
    nativeTitlebar: boolean;
    micronRenderer: MicronRendererPreference;
    micronWasmEnabled: boolean;
    micronWasmParserId: string;
    desktopChrome: boolean;
    mobileUI: boolean;
    configText: string;
    configSaving: boolean;
    configError: string;
    communityItems: CommunityInterface[];
    communityLoading: boolean;
    communityImporting: boolean;
    communityError: string;
    communityFilter: string;
    communitySelected: Set<number>;
    pageCacheEntries: number;
    pageCacheMax: number;
    pageCacheClearing: boolean;
    onChange: (theme: ThemeSettings) => void;
    onChangeKeybinds: (keybinds: KeybindSettings) => void;
    onChangeDownloadDir: (dir: string) => void;
    onPickDownloadDir: () => void;
    onChangeOpenLinksInNewTab: (value: boolean) => void;
    onChangeNativeTitlebar: (value: boolean) => void;
    onChangeMicronRenderer: (value: MicronRendererPreference) => void;
    onChangeMicronWasmEnabled: (value: boolean) => void;
    onChangeMicronWasmParser: (parserId: string) => void;
    onMicronWasmReadyChange: (ready: boolean) => void;
    onResetDefaults: () => void;
    onToggleInterface: (name: string, enabled: boolean) => void;
    onExportTheme: () => void;
    onImportTheme: (json: string) => void;
    onConfigChange: (text: string) => void;
    onConfigSave: () => void;
    onConfigReload: () => void;
    onCommunityRefresh: () => void;
    onCommunityFilter: (value: string) => void;
    onCommunityToggle: (id: number) => void;
    onCommunityImport: () => void;
    onClearPageCache: () => void;
    pluginsDir?: string;
    onPluginsChanged?: () => void;
  };

  let {
    theme = $bindable(),
    systemFonts,
    keybinds,
    interfaces,
    configPath,
    downloadDir = $bindable(),
    openLinksInNewTab,
    nativeTitlebar,
    micronRenderer,
    micronWasmEnabled,
    micronWasmParserId,
    desktopChrome,
    mobileUI,
    configText = $bindable(),
    configSaving,
    configError,
    communityItems,
    communityLoading,
    communityImporting,
    communityError,
    communityFilter = $bindable(),
    communitySelected,
    onChange,
    onChangeKeybinds,
    onChangeDownloadDir,
    onPickDownloadDir,
    onChangeOpenLinksInNewTab,
    onChangeNativeTitlebar,
    onChangeMicronRenderer,
    onChangeMicronWasmEnabled,
    onChangeMicronWasmParser,
    onMicronWasmReadyChange,
    onResetDefaults,
    onToggleInterface,
    onExportTheme,
    onImportTheme,
    onConfigChange,
    onConfigSave,
    onConfigReload,
    onCommunityRefresh,
    onCommunityFilter,
    onCommunityToggle,
    onCommunityImport,
    onClearPageCache,
    onPluginsChanged,
    pluginsDir = "",
    pageCacheEntries = 0,
    pageCacheMax = 128,
    pageCacheClearing = false,
  }: Props = $props();

  let recordingAction = $state<KeybindAction | null>(null);

  const keybindActions = Object.keys(KEYBIND_LABELS) as KeybindAction[];

  function update<K extends keyof ThemeSettings>(key: K, value: ThemeSettings[K]) {
    theme = { ...theme, [key]: value };
    onChange(theme);
  }

  function updateToken(key: string, value: string) {
    theme = {
      ...theme,
      customTokens: { ...theme.customTokens, [key]: value },
    };
    onChange(theme);
  }

  const fontOptions = $derived.by(() => {
    const fonts = [...systemFonts];
    if (theme.fontFamily && !fonts.includes(theme.fontFamily)) {
      fonts.unshift(theme.fontFamily);
    }
    return fonts;
  });

  function importThemeFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) {
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === "string") {
        onImportTheme(reader.result);
      }
    };
    reader.readAsText(file);
  }

  function startRecording(action: KeybindAction) {
    recordingAction = action;
    setKeybindRecording(true);
  }

  function recordKeybind(event: KeyboardEvent) {
    if (!recordingAction) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    if (event.key === "Escape") {
      recordingAction = null;
      setKeybindRecording(false);
      return;
    }
    const chord = chordFromEvent(event);
    if (!chord || chord === "mod" || chord === "mod+shift" || chord === "mod+alt") {
      return;
    }
    onChangeKeybinds({
      bindings: { ...keybinds.bindings, [recordingAction]: chord },
    });
    recordingAction = null;
    setKeybindRecording(false);
  }

  function formatBytes(bytes: number): string {
    if (!bytes) {
      return "0 B";
    }
    const units = ["B", "KB", "MB", "GB"];
    let value = bytes;
    let unit = 0;
    while (value >= 1024 && unit < units.length - 1) {
      value /= 1024;
      unit++;
    }
    const rounded = value >= 10 || unit === 0 ? Math.round(value) : Math.round(value * 10) / 10;
    return `${rounded} ${units[unit]}`;
  }
</script>

<svelte:window onkeydown={recordKeybind} />

<section class="settings">
  <h2>Appearance</h2>

  <label>
    <span>Theme mode</span>
    <select
      value={theme.mode}
      onchange={(event) =>
        update("mode", (event.currentTarget as HTMLSelectElement).value as ThemeSettings["mode"])}
    >
      <option value="dark">Dark</option>
      <option value="light">Light</option>
      <option value="system">System</option>
    </select>
  </label>

  <label class="accent-picker">
    <span>Accent</span>
    <input
      class="accent-swatch"
      type="color"
      value={theme.accent}
      oninput={(event) => update("accent", (event.currentTarget as HTMLInputElement).value)}
    />
  </label>

  <label>
    <span>Font family</span>
    <select
      value={theme.fontFamily}
      onchange={(event) => update("fontFamily", (event.currentTarget as HTMLSelectElement).value)}
    >
      {#each fontOptions as font (font)}
        <option value={font}>{font}</option>
      {/each}
    </select>
  </label>

  <label>
    <span>Font size ({theme.fontSize}px)</span>
    <input
      type="range"
      min="12"
      max="20"
      value={theme.fontSize}
      oninput={(event) =>
        update("fontSize", Number((event.currentTarget as HTMLInputElement).value))}
    />
  </label>

  <Toggle
    label="Compact toolbar"
    checked={theme.compactToolbar}
    onchange={(value) => update("compactToolbar", value)}
  />

  <h3>Custom tokens</h3>
  <label>
    <span>Border color</span>
    <input
      type="text"
      placeholder="#2a3140"
      value={theme.customTokens.border ?? ""}
      oninput={(event) => updateToken("border", (event.currentTarget as HTMLInputElement).value)}
    />
  </label>

  <div class="theme-io">
    <button onclick={onExportTheme}>Export theme JSON</button>
    <label class="file-btn">
      Import theme
      <input type="file" accept="application/json,.json" onchange={importThemeFile} />
    </label>
  </div>

  <h3>Browsing</h3>
  <Toggle
    label="Open sites in a new tab"
    checked={openLinksInNewTab}
    onchange={onChangeOpenLinksInNewTab}
  />

  <h3>Page cache</h3>
  <p class="hint">
    Cached mesh pages load instantly. Clear the cache if you need fresh content from the network.
  </p>
  <div class="cache-row">
    <span class="meta">{pageCacheEntries} / {pageCacheMax} entries</span>
    <button type="button" class="reset-btn" disabled={pageCacheClearing} onclick={onClearPageCache}>
      {pageCacheClearing ? "Clearing..." : "Clear page cache"}
    </button>
  </div>

  {#if desktopChrome}
    <Toggle
      label="Use native title bar"
      checked={nativeTitlebar}
      onchange={onChangeNativeTitlebar}
    />
  {/if}

  <h3>Micron pages</h3>
  <p class="hint">
    JavaScript uses micron-parser-js in the browser. WebAssembly uses micron-parser-go modules you
    select below. Go uses the server renderer. If WASM is unavailable or fails to load, pages fall
    back to JavaScript automatically.
  </p>

  {#if !isWebAssemblySupported()}
    <p class="warn">
      WebAssembly is not available in this webview. Only JavaScript and Go renderers can be used.
    </p>
  {/if}

  {#if isWebAssemblySupported()}
    <Toggle
      label="Enable Micron WebAssembly engine"
      checked={micronWasmEnabled}
      onchange={onChangeMicronWasmEnabled}
    />
  {/if}

  <label>
    <span>Micron renderer (.mu)</span>
    <select
      value={micronRenderer}
      onchange={(event) =>
        onChangeMicronRenderer(
          (event.currentTarget as HTMLSelectElement).value as MicronRendererPreference,
        )}
    >
      <option value="auto">Auto (WASM, then Go, then JS)</option>
      {#if isWebAssemblySupported() && micronWasmEnabled}
        <option value="wasm">WebAssembly (micron-parser-go)</option>
      {/if}
      <option value="go">Go (micron-parser-go server)</option>
      <option value="js">JavaScript (micron-parser-js)</option>
    </select>
  </label>

  {#if isWebAssemblySupported() && micronWasmEnabled}
    <h3>WASM parsers</h3>
    <MicronWasmManager
      selectedParserId={micronWasmParserId}
      wasmEnabled={micronWasmEnabled}
      onSelectParser={onChangeMicronWasmParser}
      onWasmReadyChange={onMicronWasmReadyChange}
    />
  {/if}

  <div class="reset-row">
    <button type="button" class="reset-btn" onclick={onResetDefaults}>Reset to defaults</button>
  </div>

  <h3>Downloads</h3>
  <label>
    <span>Download folder</span>
    <div class="download-dir">
      <input
        type="text"
        bind:value={downloadDir}
        spellcheck="false"
        onblur={() => onChangeDownloadDir(downloadDir)}
      />
      <button
        type="button"
        class="folder-btn"
        aria-label="Choose download folder"
        onclick={onPickDownloadDir}
      >
        <FolderOpen size={16} />
      </button>
    </div>
  </label>

  <h3>Keyboard shortcuts</h3>
  {#if !mobileUI}
    <ul class="keybinds">
      {#each keybindActions as action (action)}
        <li>
          <span>{KEYBIND_LABELS[action]}</span>
          <button
            type="button"
            class="keybind-btn"
            class:recording={recordingAction === action}
            onclick={() => startRecording(action)}
          >
            {recordingAction === action ? "Press keys..." : formatChord(keybinds.bindings[action])}
          </button>
        </li>
      {/each}
    </ul>
  {:else}
    <p class="hint">Keyboard shortcuts are available on desktop builds.</p>
  {/if}

  <ExtensionsPanel {pluginsDir} onChanged={onPluginsChanged} />

  <CommunityInterfaces
    items={communityItems}
    loading={communityLoading}
    importing={communityImporting}
    error={communityError}
    bind:filter={communityFilter}
    selected={communitySelected}
    onFilter={onCommunityFilter}
    onRefresh={onCommunityRefresh}
    onToggle={onCommunityToggle}
    onImport={onCommunityImport}
  />

  <ReticulumConfigEditor
    bind:configText
    {configPath}
    saving={configSaving}
    error={configError}
    onChange={onConfigChange}
    onSave={onConfigSave}
    onReload={onConfigReload}
  />

  <h3>Reticulum interfaces</h3>
  <p class="hint">Toggle interfaces configured above or imported from the directory.</p>

  <ul class="ifaces">
    {#if interfaces.length === 0}
      <li class="ifaces-empty">
        <EmptyState
          title="No interfaces configured"
          description="Add Reticulum interfaces in your config file to connect to the mesh."
        >
          <Network size={22} />
        </EmptyState>
      </li>
    {:else}
      {#each interfaces as iface (iface.name)}
        <li>
          <Toggle
            label={iface.name}
            checked={iface.enabled}
            onchange={(value) => onToggleInterface(iface.name, value)}
          />
          <span class="meta">
            {iface.type} · {iface.online ? "online" : "offline"} · tx {formatBytes(iface.txBytes)} · rx
            {formatBytes(iface.rxBytes)}
          </span>
        </li>
      {/each}
    {/if}
  </ul>
</section>

<style>
  .settings {
    height: 100%;
    overflow: auto;
    padding: 1rem;
    display: grid;
    gap: 0.85rem;
    background: var(--ren-content-bg);
  }

  h2,
  h3 {
    margin: 0;
  }

  h3 {
    margin-top: 0.5rem;
    color: var(--ren-muted);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  label {
    display: grid;
    gap: 0.35rem;
  }

  .accent-picker {
    display: grid;
    gap: 0.35rem;
  }

  .accent-swatch {
    width: 100%;
    height: 2.25rem;
    padding: 0;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: none;
    cursor: pointer;
  }

  .accent-swatch::-webkit-color-swatch-wrapper {
    padding: 0;
  }

  .accent-swatch::-webkit-color-swatch {
    border: none;
    border-radius: calc(var(--ren-radius) + 1px);
  }

  .accent-swatch::-moz-color-swatch {
    border: none;
    border-radius: calc(var(--ren-radius) + 1px);
  }

  span {
    color: var(--ren-muted);
    font-size: 0.9rem;
  }

  input,
  select,
  button {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease;
  }

  input:focus,
  select:focus {
    outline: none;
    border-color: var(--ren-focus);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--ren-focus) 28%, transparent);
  }

  select {
    appearance: none;
    background-image:
      linear-gradient(45deg, transparent 50%, var(--ren-muted) 50%),
      linear-gradient(135deg, var(--ren-muted) 50%, transparent 50%);
    background-position:
      calc(100% - 16px) calc(50% - 2px),
      calc(100% - 11px) calc(50% - 2px);
    background-size:
      5px 5px,
      5px 5px;
    background-repeat: no-repeat;
    padding-right: 2rem;
  }

  select option {
    background: var(--ren-input-bg);
    color: var(--ren-fg);
  }

  input[type="color"] {
    padding: 0;
    min-height: 2.25rem;
  }

  button {
    cursor: pointer;
  }

  .theme-io {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .reset-row {
    display: flex;
  }

  .cache-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .cache-row .reset-btn {
    width: auto;
    flex: 1 1 auto;
    min-width: 10rem;
  }

  .reset-btn {
    width: 100%;
    text-align: center;
    color: var(--ren-danger);
    border-color: color-mix(in srgb, var(--ren-danger) 35%, var(--ren-border));
  }

  .reset-btn:hover {
    background: color-mix(in srgb, var(--ren-danger) 12%, var(--ren-chrome-bg));
  }

  .download-dir {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.45rem;
  }

  .folder-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 2.5rem;
    padding-inline: 0.55rem;
  }

  .file-btn {
    position: relative;
    overflow: hidden;
    display: inline-flex;
    align-items: center;
    cursor: pointer;
  }

  .file-btn input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }

  .hint {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
    word-break: break-all;
  }

  .warn {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.85rem;
  }

  .ifaces {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.5rem;
  }

  .ifaces li {
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    padding: 0.65rem 0.75rem;
    display: grid;
    gap: 0.2rem;
    background: var(--ren-surface-raised);
  }

  .ifaces-empty {
    border: none;
    padding: 0;
    background: transparent;
  }

  .keybinds {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
  }

  .keybinds li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    padding: 0.55rem 0.75rem;
    background: var(--ren-surface-raised);
  }

  .keybind-btn {
    min-width: 8rem;
    text-align: center;
    cursor: pointer;
  }

  .keybind-btn.recording {
    background: var(--ren-accent);
    border-color: var(--ren-accent);
    color: #fff;
  }

  .meta {
    font-size: 0.8rem;
  }
</style>
