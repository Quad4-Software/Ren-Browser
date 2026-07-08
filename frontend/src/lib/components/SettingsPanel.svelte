<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { FolderOpen, Network, RefreshCw, ShieldCheck, ShieldOff } from "@lucide/svelte";
  import Toggle from "$lib/components/Toggle.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import MicronWasmManager from "$lib/components/MicronWasmManager.svelte";
  import ReticulumConfigEditor from "$lib/components/ReticulumConfigEditor.svelte";
  import CommunityInterfaces, {
    type CommunityInterface,
  } from "$lib/components/CommunityInterfaces.svelte";
  import IdentityPanel from "$lib/components/IdentityPanel.svelte";
  import ShareApkPanel from "$lib/components/ShareApkPanel.svelte";
  import ExtensionsPanel from "$lib/components/ExtensionsPanel.svelte";
  import SettingsSection from "$lib/components/SettingsSection.svelte";
  import type { MicronRendererPreference } from "$lib/micron/render-page";
  import { isWebAssemblySupported } from "$lib/micron/wasm-loader";
  import type { ThemeSettings } from "$lib/theme/tokens";
  import {
    KEYBIND_ACTIONS,
    chordFromEvent,
    formatChord,
    keybindLabel,
    setKeybindRecording,
    type KeybindAction,
    type KeybindSettings,
  } from "$lib/browser/keybinds";
  import {
    detectOSLocale,
    localeNativeName,
    localeLabel,
    resolveLocale,
    SUPPORTED_LOCALES,
    t,
  } from "$lib/i18n/i18n.svelte";
  import { System } from "@wailsio/runtime";

  type InterfaceRow = {
    name: string;
    type: string;
    enabled: boolean;
    online: boolean;
    txBytes: number;
    rxBytes: number;
  };

  type SandboxStatusRow = {
    type: string;
    enabled: boolean;
    reason?: string;
  };

  type Props = {
    theme: ThemeSettings;
    systemFonts: string[];
    keybinds: KeybindSettings;
    uiLanguage: string;
    interfaces: InterfaceRow[];
    configPath: string;
    downloadDir: string;
    openLinksInNewTab: boolean;
    tabHoverPreviews: boolean;
    nativeTitlebar: boolean;
    micronRenderer: MicronRendererPreference;
    micronWasmEnabled: boolean;
    micronWasmParserId: string;
    desktopChrome: boolean;
    mobileUI: boolean;
    mobileDevTools: boolean;
    publicMode: boolean;
    configText: string;
    configSaving: boolean;
    configError: string;
    communityItems: CommunityInterface[];
    communityLoading: boolean;
    communityImporting: boolean;
    communityError: string;
    communityFromBundle: boolean;
    communityFilter: string;
    communitySelected: Set<number>;
    pageCacheEntries: number;
    pageCacheMax: number;
    pageCacheClearing: boolean;
    pageCacheEnabled: boolean;
    sandboxStatus: SandboxStatusRow;
    onChange: (theme: ThemeSettings) => void;
    onChangeKeybinds: (keybinds: KeybindSettings) => void;
    onChangeUILanguage: (value: string) => void;
    onChangeDownloadDir: (dir: string) => void;
    onPickDownloadDir: () => void;
    onChangeOpenLinksInNewTab: (value: boolean) => void;
    onChangeTabHoverPreviews: (value: boolean) => void;
    onChangeMobileDevTools: (value: boolean) => void;
    onOpenSearch?: () => void;
    onChangeNativeTitlebar: (value: boolean) => void;
    onChangeMicronRenderer: (value: MicronRendererPreference) => void;
    onChangeMicronWasmEnabled: (value: boolean) => void;
    onChangeMicronWasmParser: (parserId: string) => void;
    onMicronWasmReadyChange: (ready: boolean) => void;
    onResetDefaults: () => void;
    onShutdown: () => void;
    onToggleInterface: (name: string, enabled: boolean) => void;
    onExportTheme: () => void;
    onImportTheme: (json: string) => void;
    onConfigChange: (text: string) => void;
    onConfigSave: () => void;
    onConfigReload: () => void;
    onOpenConfigDir?: () => void;
    onCommunityRefresh: () => void;
    onCommunityFilter: (value: string) => void;
    onCommunityToggle: (id: number) => void;
    onCommunityImport: () => void;
    onClearPageCache: () => void;
    onChangePageCacheEnabled: (value: boolean) => void;
    sectionsCollapsed: Record<string, boolean>;
    onChangeSectionsCollapsed: (sections: Record<string, boolean>) => void;
    pluginsDir?: string;
    onPluginsChanged?: () => void;
  };

  let {
    theme = $bindable(),
    systemFonts,
    keybinds,
    uiLanguage,
    interfaces,
    configPath,
    downloadDir = $bindable(),
    openLinksInNewTab,
    tabHoverPreviews,
    nativeTitlebar,
    micronRenderer,
    micronWasmEnabled,
    micronWasmParserId,
    desktopChrome,
    mobileUI,
    mobileDevTools,
    publicMode,
    configText = $bindable(),
    configSaving,
    configError,
    communityItems,
    communityLoading,
    communityImporting,
    communityError,
    communityFromBundle,
    communityFilter = $bindable(),
    communitySelected,
    onChange,
    onChangeKeybinds,
    onChangeUILanguage,
    onChangeDownloadDir,
    onPickDownloadDir,
    onChangeOpenLinksInNewTab,
    onChangeTabHoverPreviews,
    onChangeMobileDevTools,
    onOpenSearch = () => {},
    onChangeNativeTitlebar,
    onChangeMicronRenderer,
    onChangeMicronWasmEnabled,
    onChangeMicronWasmParser,
    onMicronWasmReadyChange,
    onResetDefaults,
    onShutdown,
    onToggleInterface,
    onExportTheme,
    onImportTheme,
    onConfigChange,
    onConfigSave,
    onConfigReload,
    onOpenConfigDir,
    onCommunityRefresh,
    onCommunityFilter,
    onCommunityToggle,
    onCommunityImport,
    onClearPageCache,
    onChangePageCacheEnabled,
    sectionsCollapsed,
    onChangeSectionsCollapsed,
    onPluginsChanged,
    pluginsDir = "",
    pageCacheEntries = 0,
    pageCacheMax = 128,
    pageCacheClearing = false,
    pageCacheEnabled = true,
    sandboxStatus = { type: "none", enabled: false },
  }: Props = $props();

  let recordingAction = $state<KeybindAction | null>(null);

  const keybindActions = KEYBIND_ACTIONS;

  function sandboxTypeLabel(type: string): string {
    if (type === "landlock") {
      return t("settings.sandboxTypeLandlock");
    }
    return t("settings.sandboxTypeNone");
  }

  function sandboxReasonLabel(reason?: string): string {
    if (!reason) {
      return "";
    }
    if (reason.includes("WebKitGTK")) {
      return t("settings.sandboxReasonDesktopWebkit");
    }
    if (reason.includes("kernel does not support Landlock")) {
      return t("settings.sandboxReasonUnsupportedKernel");
    }
    if (reason.includes("--no-landlock")) {
      return t("settings.sandboxReasonNoLandlockFlag");
    }
    if (reason.includes("_LANDLOCK")) {
      return t("settings.sandboxReasonEnvDisabled");
    }
    if (reason.startsWith("not supported on")) {
      return t("settings.sandboxReasonUnsupportedPlatform");
    }
    return reason;
  }

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

  function toggleSettingsSection(id: string) {
    onChangeSectionsCollapsed({
      ...sectionsCollapsed,
      [id]: !sectionsCollapsed[id],
    });
  }

  function sectionCollapsed(id: string): boolean {
    return sectionsCollapsed[id] === true;
  }

  const isAndroid = System.IsAndroid();
</script>

<svelte:window onkeydown={recordKeybind} />

<section class="settings" class:mobile={mobileUI}>
  <SettingsSection
    id="appearance"
    title={t("settings.appearance")}
    heading="h2"
    collapsed={sectionCollapsed("appearance")}
    onToggle={toggleSettingsSection}
  >
    <label>
      <span>{t("language.title")}</span>
      <select
        class="ren-select"
        value={uiLanguage}
        onchange={(event) => onChangeUILanguage((event.currentTarget as HTMLSelectElement).value)}
      >
        <option value=""
          >{t("language.system", {
            locale: localeNativeName(resolveLocale(detectOSLocale())),
          })}</option
        >
        {#each SUPPORTED_LOCALES as locale (locale.code)}
          <option value={locale.code}>{localeLabel(locale.code)}</option>
        {/each}
      </select>
    </label>
    <p class="hint">{t("language.hint")}</p>

    <label>
      <span>{t("settings.themeMode")}</span>
      <select
        class="ren-select"
        value={theme.mode}
        onchange={(event) =>
          update("mode", (event.currentTarget as HTMLSelectElement).value as ThemeSettings["mode"])}
      >
        <option value="dark">{t("settings.themeDark")}</option>
        <option value="light">{t("settings.themeLight")}</option>
        <option value="system">{t("settings.themeSystem")}</option>
      </select>
    </label>

    <label class="accent-picker">
      <span>{t("settings.accent")}</span>
      <input
        class="accent-swatch"
        type="color"
        value={theme.accent}
        oninput={(event) => update("accent", (event.currentTarget as HTMLInputElement).value)}
      />
    </label>

    <label>
      <span>{t("settings.fontFamily")}</span>
      <select
        class="ren-select"
        value={theme.fontFamily}
        onchange={(event) => update("fontFamily", (event.currentTarget as HTMLSelectElement).value)}
      >
        {#each fontOptions as font (font)}
          <option value={font}>{font}</option>
        {/each}
      </select>
    </label>

    <label>
      <span>{t("settings.fontSize", { size: theme.fontSize })}</span>
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
      label={t("settings.compactToolbar")}
      checked={theme.compactToolbar}
      onchange={(value) => update("compactToolbar", value)}
    />
  </SettingsSection>

  <SettingsSection
    id="customTokens"
    title={t("settings.customTokens")}
    collapsed={sectionCollapsed("customTokens")}
    onToggle={toggleSettingsSection}
  >
    <label>
      <span>{t("settings.borderColor")}</span>
      <input
        type="text"
        placeholder={t("settings.borderPlaceholder")}
        value={theme.customTokens.border ?? ""}
        oninput={(event) => updateToken("border", (event.currentTarget as HTMLInputElement).value)}
      />
    </label>

    <div class="theme-io">
      <button onclick={onExportTheme}>{t("settings.exportTheme")}</button>
      <label class="file-btn">
        {t("settings.importTheme")}
        <input type="file" accept="application/json,.json" onchange={importThemeFile} />
      </label>
    </div>
  </SettingsSection>

  <SettingsSection
    id="browsing"
    title={t("settings.browsing")}
    collapsed={sectionCollapsed("browsing")}
    onToggle={toggleSettingsSection}
  >
    <Toggle
      label={t("settings.openLinksInNewTab")}
      checked={openLinksInNewTab}
      onchange={onChangeOpenLinksInNewTab}
    />

    <Toggle
      label={t("settings.mobileDevTools")}
      checked={mobileDevTools}
      onchange={onChangeMobileDevTools}
    />

    {#if mobileUI}
      <button type="button" class="panel-link" onclick={onOpenSearch}>
        {t("settings.openSearch")}
      </button>
    {/if}

    {#if desktopChrome}
      <Toggle
        label={t("settings.tabHoverPreviews")}
        checked={tabHoverPreviews}
        onchange={onChangeTabHoverPreviews}
      />
      <Toggle
        label={t("settings.nativeTitlebar")}
        checked={nativeTitlebar}
        onchange={onChangeNativeTitlebar}
      />
    {/if}
  </SettingsSection>

  <SettingsSection
    id="pageCache"
    title={t("settings.pageCache")}
    collapsed={sectionCollapsed("pageCache")}
    onToggle={toggleSettingsSection}
  >
    <p class="hint">
      {t("settings.pageCacheHint")}
    </p>
    <Toggle
      label={t("settings.pageCacheEnabled")}
      checked={pageCacheEnabled}
      onchange={onChangePageCacheEnabled}
    />
    <div class="cache-row">
      <span class="meta"
        >{t("common.entries", { current: pageCacheEntries, max: pageCacheMax })}</span
      >
      <button
        type="button"
        class="reset-btn"
        disabled={pageCacheClearing}
        onclick={onClearPageCache}
      >
        {pageCacheClearing ? t("common.clearing") : t("settings.clearPageCache")}
      </button>
    </div>
  </SettingsSection>

  <SettingsSection
    id="micron"
    title={t("settings.micronPages")}
    collapsed={sectionCollapsed("micron")}
    onToggle={toggleSettingsSection}
  >
    <p class="hint">
      {t("settings.micronHint")}
    </p>

    {#if !isWebAssemblySupported()}
      <p class="warn">
        {t("settings.wasmUnavailable")}
      </p>
    {/if}

    {#if isWebAssemblySupported()}
      <Toggle
        label={t("settings.micronWasmEnabled")}
        checked={micronWasmEnabled}
        onchange={onChangeMicronWasmEnabled}
      />
    {/if}

    <label>
      <span>{t("settings.micronRenderer")}</span>
      <select
        class="ren-select"
        value={micronRenderer}
        onchange={(event) =>
          onChangeMicronRenderer(
            (event.currentTarget as HTMLSelectElement).value as MicronRendererPreference,
          )}
      >
        <option value="auto">{t("settings.rendererAuto")}</option>
        {#if isWebAssemblySupported() && micronWasmEnabled}
          <option value="wasm">{t("settings.rendererWasm")}</option>
        {/if}
        <option value="go">{t("settings.rendererGo")}</option>
        <option value="js">{t("settings.rendererJs")}</option>
      </select>
    </label>

    {#if isWebAssemblySupported() && micronWasmEnabled}
      <MicronWasmManager
        selectedParserId={micronWasmParserId}
        wasmEnabled={micronWasmEnabled}
        onSelectParser={onChangeMicronWasmParser}
        onWasmReadyChange={onMicronWasmReadyChange}
      />
    {/if}

    <div class="reset-row">
      <button type="button" class="reset-btn" onclick={onResetDefaults}
        >{t("settings.resetDefaults")}</button
      >
    </div>
  </SettingsSection>

  <SettingsSection
    id="downloads"
    title={t("settings.downloads")}
    collapsed={sectionCollapsed("downloads")}
    onToggle={toggleSettingsSection}
  >
    <label>
      <span>{t("settings.downloadFolder")}</span>
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
          aria-label={t("settings.chooseDownloadFolder")}
          onclick={onPickDownloadDir}
        >
          <FolderOpen size={16} />
        </button>
      </div>
    </label>
  </SettingsSection>

  <SettingsSection
    id="keybinds"
    title={t("settings.keyboardShortcuts")}
    collapsed={sectionCollapsed("keybinds")}
    onToggle={toggleSettingsSection}
  >
    {#if !mobileUI}
      <ul class="keybinds">
        {#each keybindActions as action (action)}
          <li>
            <span>{keybindLabel(action)}</span>
            <button
              type="button"
              class="keybind-btn"
              class:recording={recordingAction === action}
              onclick={() => startRecording(action)}
            >
              {recordingAction === action
                ? t("common.pressKeys")
                : formatChord(keybinds.bindings[action])}
            </button>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="hint">{t("settings.keyboardShortcutsDesktopOnly")}</p>
    {/if}
  </SettingsSection>

  <SettingsSection
    id="extensions"
    title={t("extensions.title")}
    collapsed={sectionCollapsed("extensions")}
    onToggle={toggleSettingsSection}
  >
    <ExtensionsPanel {pluginsDir} showTitle={false} onChanged={onPluginsChanged} />
  </SettingsSection>

  <SettingsSection
    id="community"
    title={t("community.title")}
    collapsed={sectionCollapsed("community")}
    onToggle={toggleSettingsSection}
  >
    {#snippet actions()}
      <button
        type="button"
        class="ren-icon-btn"
        aria-label={t("community.refresh")}
        title={desktopChrome ? t("community.refresh") : undefined}
        disabled={communityLoading}
        onclick={(event) => {
          event.stopPropagation();
          onCommunityRefresh();
        }}
      >
        <span class:spin={communityLoading}>
          <RefreshCw size={16} />
        </span>
      </button>
    {/snippet}
    <CommunityInterfaces
      showTitle={false}
      {desktopChrome}
      items={communityItems}
      loading={communityLoading}
      importing={communityImporting}
      error={communityError}
      fromBundle={communityFromBundle}
      bind:filter={communityFilter}
      selected={communitySelected}
      onFilter={onCommunityFilter}
      onRefresh={onCommunityRefresh}
      onToggle={onCommunityToggle}
      onImport={onCommunityImport}
    />
  </SettingsSection>

  <SettingsSection
    id="identity"
    title={t("identity.title")}
    collapsed={sectionCollapsed("identity")}
    onToggle={toggleSettingsSection}
  >
    <IdentityPanel showTitle={false} />
  </SettingsSection>

  <SettingsSection
    id="reticulumConfig"
    title={t("config.title")}
    collapsed={sectionCollapsed("reticulumConfig")}
    onToggle={toggleSettingsSection}
  >
    <ReticulumConfigEditor
      showTitle={false}
      bind:configText
      {configPath}
      saving={configSaving}
      error={configError}
      onChange={onConfigChange}
      onSave={onConfigSave}
      onReload={onConfigReload}
      {onOpenConfigDir}
    />
  </SettingsSection>

  <SettingsSection
    id="reticulumInterfaces"
    title={t("settings.reticulumInterfaces")}
    collapsed={sectionCollapsed("reticulumInterfaces")}
    onToggle={toggleSettingsSection}
  >
    <p class="hint">{t("settings.reticulumInterfacesHint")}</p>

    <ul class="ifaces">
      {#if interfaces.length === 0}
        <li class="ifaces-empty">
          <EmptyState
            title={t("settings.noInterfaces")}
            description={t("settings.noInterfacesDescription")}
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
              {iface.type} · {iface.online ? t("common.online") : t("common.offline")} · {t(
                "common.txRx",
                {
                  tx: formatBytes(iface.txBytes),
                  rx: formatBytes(iface.rxBytes),
                },
              )}
            </span>
          </li>
        {/each}
      {/if}
    </ul>
  </SettingsSection>

  <SettingsSection
    id="security"
    title={t("settings.security")}
    collapsed={sectionCollapsed("security")}
    onToggle={toggleSettingsSection}
  >
    <div class="sandbox-card" class:active={sandboxStatus.enabled}>
      <div class="sandbox-head">
        <span class="sandbox-icon" aria-hidden="true">
          {#if sandboxStatus.enabled}
            <ShieldCheck size={20} strokeWidth={2} />
          {:else}
            <ShieldOff size={20} strokeWidth={2} />
          {/if}
        </span>
        <div class="sandbox-copy">
          <span class="sandbox-name">{sandboxTypeLabel(sandboxStatus.type)}</span>
          <span class="sandbox-subtitle">{t("settings.sandboxSubtitle")}</span>
        </div>
        <span
          class="sandbox-badge"
          class:enabled={sandboxStatus.enabled}
          class:disabled={!sandboxStatus.enabled}
        >
          {sandboxStatus.enabled ? t("settings.sandboxEnabled") : t("settings.sandboxDisabled")}
        </span>
      </div>
      {#if sandboxReasonLabel(sandboxStatus.reason)}
        <p class="sandbox-note">{sandboxReasonLabel(sandboxStatus.reason)}</p>
      {:else if sandboxStatus.enabled}
        <p class="sandbox-note active">{t("settings.sandboxActiveHint")}</p>
      {/if}
    </div>
  </SettingsSection>

  {#if isAndroid}
    <SettingsSection
      id="shareApk"
      title={t("settings.shareApk")}
      collapsed={sectionCollapsed("shareApk")}
      onToggle={toggleSettingsSection}
    >
      <ShareApkPanel />
    </SettingsSection>
  {/if}

  {#if !publicMode}
    <SettingsSection
      id="application"
      title={t("settings.application")}
      collapsed={sectionCollapsed("application")}
      onToggle={toggleSettingsSection}
    >
      <p class="hint">{t("settings.shutdownHint")}</p>
      <div class="reset-row">
        <button type="button" class="reset-btn" onclick={onShutdown}
          >{t("settings.shutdown")}</button
        >
      </div>
    </SettingsSection>
  {/if}
</section>

<style>
  .settings {
    height: 100%;
    width: 100%;
    max-width: 100%;
    min-width: 0;
    overflow: auto;
    overflow-x: hidden;
    padding: 1rem;
    display: grid;
    gap: 0.85rem;
    background: var(--ren-content-bg);
  }

  label {
    display: grid;
    gap: 0.35rem;
    min-width: 0;
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
  button {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
    max-width: 100%;
    min-width: 0;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease;
  }

  input[type="text"],
  input:not([type]) {
    width: 100%;
  }

  .settings :global(select),
  .settings :global(.ren-select) {
    width: 100%;
    max-width: 100%;
    min-width: 0;
  }

  input:focus {
    outline: none;
    border-color: var(--ren-focus);
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--ren-focus) 28%, transparent);
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
    min-width: 0;
  }

  .theme-io button,
  .theme-io .file-btn {
    flex: 1 1 auto;
    min-width: 0;
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
    min-width: 0;
  }

  .cache-row .meta {
    flex: 1 1 8rem;
    min-width: 0;
  }

  .cache-row .reset-btn {
    width: auto;
    flex: 1 1 auto;
    min-width: 0;
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
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 0.45rem;
  }

  .download-dir input {
    min-width: 0;
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

  .panel-link {
    width: 100%;
    text-align: left;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface-raised);
    color: var(--ren-fg);
    padding: 0.65rem 0.8rem;
    font: inherit;
    cursor: pointer;
    transition:
      border-color 0.15s ease,
      background 0.15s ease;
  }

  .panel-link:hover {
    border-color: var(--ren-border-strong);
    background: var(--ren-tab-hover);
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
    flex-wrap: wrap;
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    padding: 0.55rem 0.75rem;
    background: var(--ren-surface-raised);
    min-width: 0;
  }

  .keybinds li > span {
    flex: 1 1 8rem;
    min-width: 0;
  }

  .keybind-btn {
    min-width: 0;
    flex: 1 1 6rem;
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

  .sandbox-card {
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-surface-raised);
    padding: 0.85rem 0.9rem;
    display: grid;
    gap: 0.65rem;
  }

  .sandbox-card.active {
    border-color: color-mix(in srgb, var(--ren-success, #3d9a5f) 35%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-success, #3d9a5f) 6%, var(--ren-surface-raised));
  }

  .sandbox-head {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    flex-wrap: wrap;
    min-width: 0;
  }

  .sandbox-icon {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.25rem;
    height: 2.25rem;
    border-radius: 999px;
    border: 1px solid var(--ren-border);
    background: color-mix(in srgb, var(--ren-muted) 10%, transparent);
    color: var(--ren-muted);
  }

  .sandbox-card.active .sandbox-icon {
    color: var(--ren-success, #3d9a5f);
    border-color: color-mix(in srgb, var(--ren-success, #3d9a5f) 40%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-success, #3d9a5f) 12%, transparent);
  }

  .sandbox-copy {
    flex: 1;
    min-width: 0;
    display: grid;
    gap: 0.15rem;
  }

  .sandbox-name {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--ren-fg);
    line-height: 1.25;
  }

  .sandbox-subtitle {
    font-size: 0.8rem;
    color: var(--ren-muted);
    line-height: 1.35;
  }

  .sandbox-badge {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    border-radius: 999px;
    padding: 0.18rem 0.55rem;
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    border: 1px solid transparent;
    white-space: normal;
    text-align: center;
    max-width: 100%;
  }

  .settings.mobile .theme-io {
    flex-direction: column;
  }

  .settings.mobile .theme-io button,
  .settings.mobile .theme-io .file-btn {
    width: 100%;
    justify-content: center;
  }

  .settings.mobile .cache-row {
    flex-direction: column;
    align-items: stretch;
  }

  .settings.mobile .cache-row .reset-btn {
    width: 100%;
  }

  .settings.mobile .keybind-btn {
    width: 100%;
    flex-basis: 100%;
  }

  .settings.mobile .sandbox-badge {
    margin-left: auto;
  }

  .sandbox-badge.enabled {
    color: var(--ren-success, #3d9a5f);
    border-color: color-mix(in srgb, var(--ren-success, #3d9a5f) 45%, transparent);
    background: color-mix(in srgb, var(--ren-success, #3d9a5f) 12%, transparent);
  }

  .sandbox-badge.disabled {
    color: var(--ren-muted);
    border-color: var(--ren-border);
    background: color-mix(in srgb, var(--ren-muted) 10%, transparent);
  }

  .sandbox-note {
    margin: 0;
    padding-top: 0.65rem;
    border-top: 1px solid var(--ren-border);
    color: var(--ren-muted);
    font-size: 0.82rem;
    line-height: 1.45;
  }

  .sandbox-note.active {
    color: color-mix(in srgb, var(--ren-success, #3d9a5f) 75%, var(--ren-muted));
  }

  :global(.spin) {
    display: inline-flex;
    animation: community-refresh-spin 0.8s linear infinite;
  }

  @keyframes community-refresh-spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
