<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import {
    Box,
    FolderOpen,
    Network,
    Package,
    ShieldCheck,
    ShieldOff,
    Smartphone,
  } from "@lucide/svelte";
  import { Select, Slider } from "bits-ui";
  import { useEventListener } from "runed";
  import Toggle from "$lib/components/Toggle.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import MicronWasmManager from "$lib/components/MicronWasmManager.svelte";
  import ReticulumConfigEditor from "$lib/components/ReticulumConfigEditor.svelte";
  import CommunityInterfaces from "$lib/components/CommunityInterfaces.svelte";
  import type { CommunityInterface } from "../../../bindings/renbrowser/internal/rns/models.js";
  import type { SelfCheckResult } from "../../../bindings/renbrowser/internal/app/models.js";
  import IdentityPanel from "$lib/components/IdentityPanel.svelte";
  import ShareApkPanel from "$lib/components/ShareApkPanel.svelte";
  import ExtensionsPanel from "$lib/components/ExtensionsPanel.svelte";
  import SettingsSection from "$lib/components/SettingsSection.svelte";
  import type { MicronRendererPreference } from "$lib/micron/render-page";
  import type { TabLayout } from "$lib/browser/url";
  import type { MicronImageNodePolicy, MicronImagesMode } from "$lib/micron/images";
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

  type ReticulumStatusRow = {
    enableTransport: boolean;
    shareInstance: boolean;
    connectedToSharedInstance: boolean;
    sharedInstanceMode: string;
    transportActive: boolean;
  };

  type SandboxStatusRow = {
    type: string;
    enabled: boolean;
    reason?: string;
    abi?: number;
    seccompEnabled?: boolean;
    seccompSupported?: boolean;
    seccompReason?: string;
    inFlatpak?: boolean;
    inAppImage?: boolean;
    inContainer?: boolean;
    containerRuntime?: string;
    webkitSandbox?: string;
    webkitSandboxNote?: string;
    onAndroid?: boolean;
  };

  type Props = {
    theme: ThemeSettings;
    systemFonts: string[];
    keybinds: KeybindSettings;
    uiLanguage: string;
    interfaces: InterfaceRow[];
    reticulumStatus: ReticulumStatusRow;
    configPath: string;
    downloadDir: string;
    openLinksInNewTab: boolean;
    tabHoverPreviews: boolean;
    tabLayout: TabLayout;
    nativeTitlebar: boolean;
    micronRenderer: MicronRendererPreference;
    micronWasmEnabled: boolean;
    micronWasmParserId: string;
    micronPreserveLayout: boolean;
    micronImagesMode: MicronImagesMode;
    micronImageNodes: Record<string, string>;
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
    communityFilter: string;
    communitySelected: Set<number>;
    pageCacheRAMEntries: number;
    pageCacheDiskEntries: number;
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
    onChangeTabLayout: (value: TabLayout) => void;
    onChangeMobileDevTools: (value: boolean) => void;
    onOpenSearch?: () => void;
    onChangeNativeTitlebar: (value: boolean) => void;
    onChangeMicronRenderer: (value: MicronRendererPreference) => void;
    onChangeMicronWasmEnabled: (value: boolean) => void;
    onChangeMicronPreserveLayout: (value: boolean) => void;
    onChangeMicronImagesMode: (value: MicronImagesMode) => void;
    onMicronImageNodePolicy: (nodeHash: string, policy: MicronImageNodePolicy | null) => void;
    onChangeMicronWasmParser: (parserId: string) => void | Promise<void>;
    onMicronWasmReadyChange: (ready: boolean) => void;
    onResetDefaults: () => void;
    onResetBrowser: () => void;
    onRestartReticulum: () => void;
    onShutdown: () => void;
    onToggleInterface: (name: string, enabled: boolean) => void;
    onToggleTransport: (enabled: boolean) => void;
    onToggleShareInstance: (enabled: boolean) => void;
    onExportTheme: () => void;
    onImportTheme: (json: string) => void;
    onConfigChange: (text: string) => void;
    onConfigSave: () => void;
    onConfigReload: () => void;
    onOpenConfigDir?: () => void;
    onCommunityFilter: (value: string) => void;
    onCommunityToggle: (id: number) => void;
    onCommunityImport: () => void;
    onClearPageCache: () => void;
    onChangePageCacheEnabled: (value: boolean) => void;
    sectionsCollapsed: Record<string, boolean>;
    onChangeSectionsCollapsed: (sections: Record<string, boolean>) => void;
    pluginsDir?: string;
    onPluginsChanged?: () => void;
    selfTestResult?: SelfCheckResult | null;
    selfTestRunning?: boolean;
    onRunSelfTest?: () => void;
  };

  let {
    theme = $bindable(),
    systemFonts,
    keybinds,
    uiLanguage,
    interfaces,
    reticulumStatus,
    configPath,
    downloadDir = $bindable(),
    openLinksInNewTab,
    tabHoverPreviews,
    tabLayout,
    nativeTitlebar,
    micronRenderer,
    micronWasmEnabled,
    micronWasmParserId,
    micronPreserveLayout,
    micronImagesMode,
    micronImageNodes,
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
    communityFilter = $bindable(),
    communitySelected,
    onChange,
    onChangeKeybinds,
    onChangeUILanguage,
    onChangeDownloadDir,
    onPickDownloadDir,
    onChangeOpenLinksInNewTab,
    onChangeTabHoverPreviews,
    onChangeTabLayout,
    onChangeMobileDevTools,
    onOpenSearch = () => {},
    onChangeNativeTitlebar,
    onChangeMicronRenderer,
    onChangeMicronWasmEnabled,
    onChangeMicronPreserveLayout,
    onChangeMicronImagesMode,
    onMicronImageNodePolicy,
    onChangeMicronWasmParser,
    onMicronWasmReadyChange,
    onResetDefaults,
    onResetBrowser,
    onRestartReticulum,
    onShutdown,
    onToggleInterface,
    onToggleTransport,
    onToggleShareInstance,
    onExportTheme,
    onImportTheme,
    onConfigChange,
    onConfigSave,
    onConfigReload,
    onOpenConfigDir,
    onCommunityFilter,
    onCommunityToggle,
    onCommunityImport,
    onClearPageCache,
    onChangePageCacheEnabled,
    sectionsCollapsed,
    onChangeSectionsCollapsed,
    onPluginsChanged,
    pluginsDir = "",
    pageCacheRAMEntries = 0,
    pageCacheDiskEntries = 0,
    pageCacheClearing = false,
    pageCacheEnabled = true,
    sandboxStatus = { type: "none", enabled: false },
    selfTestResult = null,
    selfTestRunning = false,
    onRunSelfTest = () => {},
  }: Props = $props();

  let recordingAction = $state<KeybindAction | null>(null);

  const keybindActions = KEYBIND_ACTIONS;

  function sandboxTypeLabel(type: string): string {
    if (type === "landlock+seccomp") {
      return t("settings.sandboxTypeLandlockSeccomp");
    }
    if (type === "seccomp") {
      return t("settings.sandboxTypeSeccomp");
    }
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

  function webkitSandboxLabel(state?: string): string {
    switch (state) {
      case "active":
        return t("settings.webkitSandboxActive");
      case "disabled":
        return t("settings.webkitSandboxDisabled");
      default:
        return t("settings.webkitSandboxUnavailable");
    }
  }

  function webkitSandboxNoteLabel(note?: string): string {
    switch (note) {
      case "flatpak":
        return t("settings.webkitSandboxNoteFlatpak");
      case "appimage":
        return t("settings.webkitSandboxNoteAppImage");
      case "env":
        return t("settings.webkitSandboxNoteEnv");
      case "android-webview":
        return t("settings.webkitSandboxNoteAndroid");
      case "not-linux":
        return t("settings.webkitSandboxNoteNotLinux");
      case "container":
        return t("settings.webkitSandboxNoteContainer");
      default:
        return "";
    }
  }

  function containerRuntimeLabel(runtime?: string): string {
    if (!runtime) {
      return t("settings.containerDetected");
    }
    return t("settings.containerRuntime", { runtime });
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

  const languageItems = $derived.by(() => [
    {
      value: "",
      label: t("language.system", {
        locale: localeNativeName(resolveLocale(detectOSLocale())),
      }),
    },
    ...SUPPORTED_LOCALES.map((locale) => ({
      value: locale.code,
      label: localeLabel(locale.code),
    })),
  ]);

  const themeModeItems = $derived.by(() => [
    { value: "dark", label: t("settings.themeDark") },
    { value: "light", label: t("settings.themeLight") },
    { value: "system", label: t("settings.themeSystem") },
  ]);

  const fontItems = $derived(fontOptions.map((font) => ({ value: font, label: font })));

  const tabLayoutItems = $derived([
    { value: "top", label: t("settings.tabLayoutTop") },
    { value: "left", label: t("settings.tabLayoutLeft") },
  ]);

  const micronRendererItems = $derived.by(() => {
    const items = [{ value: "auto", label: t("settings.rendererAuto") }];
    if (isWebAssemblySupported() && micronWasmEnabled) {
      items.push({ value: "wasm", label: t("settings.rendererWasm") });
    }
    items.push({ value: "go", label: t("settings.rendererGo") });
    items.push({ value: "js", label: t("settings.rendererJs") });
    return items;
  });

  const micronImagesItems = $derived.by(() => [
    { value: "ask", label: t("settings.micronImagesAsk") },
    { value: "always", label: t("settings.micronImagesAlways") },
    { value: "off", label: t("settings.micronImagesOff") },
  ]);

  const micronImagePolicyItems = $derived.by(() => [
    { value: "always", label: t("settings.micronImagePolicyAlways") },
    { value: "never", label: t("settings.micronImagePolicyNever") },
  ]);

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

  useEventListener(() => window, "keydown", recordKeybind);
</script>

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
      <Select.Root
        type="single"
        value={uiLanguage}
        items={languageItems}
        onValueChange={(value) => onChangeUILanguage(value)}
      >
        <Select.Trigger class="ren-select" aria-label={t("language.title")}>
          <Select.Value
            placeholder={t("language.system", {
              locale: localeNativeName(resolveLocale(detectOSLocale())),
            })}
          />
        </Select.Trigger>
        <Select.Portal>
          <Select.Content class="ren-select-content" sideOffset={4}>
            <Select.Viewport>
              {#each languageItems as item (item.value)}
                <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item>
              {/each}
            </Select.Viewport>
          </Select.Content>
        </Select.Portal>
      </Select.Root>
    </label>
    <p class="hint">{t("language.hint")}</p>

    <label>
      <span>{t("settings.themeMode")}</span>
      <Select.Root
        type="single"
        value={theme.mode}
        items={themeModeItems}
        onValueChange={(value) => update("mode", value as ThemeSettings["mode"])}
      >
        <Select.Trigger class="ren-select" aria-label={t("settings.themeMode")}>
          <Select.Value />
        </Select.Trigger>
        <Select.Portal>
          <Select.Content class="ren-select-content" sideOffset={4}>
            <Select.Viewport>
              {#each themeModeItems as item (item.value)}
                <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item>
              {/each}
            </Select.Viewport>
          </Select.Content>
        </Select.Portal>
      </Select.Root>
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
      <Select.Root
        type="single"
        value={theme.fontFamily}
        items={fontItems}
        onValueChange={(value) => update("fontFamily", value)}
      >
        <Select.Trigger class="ren-select" aria-label={t("settings.fontFamily")}>
          <Select.Value />
        </Select.Trigger>
        <Select.Portal>
          <Select.Content class="ren-select-content" sideOffset={4}>
            <Select.Viewport>
              {#each fontItems as item (item.value)}
                <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item>
              {/each}
            </Select.Viewport>
          </Select.Content>
        </Select.Portal>
      </Select.Root>
    </label>

    <label>
      <span>{t("settings.fontSize", { size: theme.fontSize })}</span>
      <Slider.Root
        type="single"
        class="settings-slider"
        value={theme.fontSize}
        min={12}
        max={20}
        step={1}
        onValueChange={(value) => update("fontSize", value)}
      >
        <Slider.Range class="settings-slider-range" />
        <Slider.Thumb
          index={0}
          class="settings-slider-thumb"
          aria-label={t("settings.fontSize", { size: theme.fontSize })}
        />
      </Slider.Root>
    </label>

    <Toggle
      label={t("settings.compactToolbar")}
      checked={theme.compactToolbar}
      onchange={(value) => update("compactToolbar", value)}
    />
    {#if !mobileUI}
      <Toggle
        label={t("settings.overlaySidebars")}
        checked={theme.overlaySidebars}
        onchange={(value) => update("overlaySidebars", value)}
      />
      <p class="hint">{t("settings.overlaySidebarsHint")}</p>
    {/if}
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
      <label>
        <span>{t("settings.tabLayout")}</span>
        <Select.Root
          type="single"
          value={tabLayout}
          items={tabLayoutItems}
          onValueChange={(value) => onChangeTabLayout(value as TabLayout)}
        >
          <Select.Trigger class="ren-select" aria-label={t("settings.tabLayout")}>
            <Select.Value />
          </Select.Trigger>
          <Select.Portal>
            <Select.Content class="ren-select-content" sideOffset={4}>
              <Select.Viewport>
                {#each tabLayoutItems as item (item.value)}
                  <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item>
                {/each}
              </Select.Viewport>
            </Select.Content>
          </Select.Portal>
        </Select.Root>
      </label>
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
        >{t("common.pageCacheStats", {
          ram: pageCacheRAMEntries,
          disk: pageCacheDiskEntries,
        })}</span
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
      <Select.Root
        type="single"
        value={micronRenderer}
        items={micronRendererItems}
        onValueChange={(value) => onChangeMicronRenderer(value as MicronRendererPreference)}
      >
        <Select.Trigger class="ren-select" aria-label={t("settings.micronRenderer")}>
          <Select.Value />
        </Select.Trigger>
        <Select.Portal>
          <Select.Content class="ren-select-content" sideOffset={4}>
            <Select.Viewport>
              {#each micronRendererItems as item (item.value)}
                <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item>
              {/each}
            </Select.Viewport>
          </Select.Content>
        </Select.Portal>
      </Select.Root>
    </label>

    <Toggle
      label={t("settings.micronPreserveLayout")}
      checked={micronPreserveLayout}
      onchange={onChangeMicronPreserveLayout}
    />
    <p class="hint">{t("settings.micronPreserveLayoutHint")}</p>

    <label>
      <span>{t("settings.micronImages")}</span>
      <Select.Root
        type="single"
        value={micronImagesMode}
        items={micronImagesItems}
        onValueChange={(value) => onChangeMicronImagesMode(value as MicronImagesMode)}
      >
        <Select.Trigger class="ren-select" aria-label={t("settings.micronImages")}>
          <Select.Value />
        </Select.Trigger>
        <Select.Portal>
          <Select.Content class="ren-select-content" sideOffset={4}>
            <Select.Viewport>
              {#each micronImagesItems as item (item.value)}
                <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item>
              {/each}
            </Select.Viewport>
          </Select.Content>
        </Select.Portal>
      </Select.Root>
    </label>
    <p class="hint">{t("settings.micronImagesHint")}</p>

    {#if Object.keys(micronImageNodes).length > 0}
      <div class="micron-image-nodes" role="group" aria-label={t("settings.micronImageNodes")}>
        <span class="micron-image-nodes-title">{t("settings.micronImageNodes")}</span>
        <ul>
          {#each Object.entries(micronImageNodes).sort( ([a], [b]) => a.localeCompare(b) ) as [hash, policy] (hash)}
            <li>
              <code class="micron-image-node-hash" title={hash}>{hash}</code>
              <Select.Root
                type="single"
                value={policy}
                items={micronImagePolicyItems}
                onValueChange={(value) =>
                  onMicronImageNodePolicy(hash, value as MicronImageNodePolicy)}
              >
                <Select.Trigger
                  class="ren-select micron-image-node-policy"
                  aria-label={t("settings.micronImageNodePolicy")}
                >
                  <Select.Value />
                </Select.Trigger>
                <Select.Portal>
                  <Select.Content class="ren-select-content" sideOffset={4}>
                    <Select.Viewport>
                      {#each micronImagePolicyItems as item (item.value)}
                        <Select.Item value={item.value} label={item.label}>{item.label}</Select.Item
                        >
                      {/each}
                    </Select.Viewport>
                  </Select.Content>
                </Select.Portal>
              </Select.Root>
              <button
                type="button"
                class="micron-image-node-remove"
                aria-label={t("settings.micronImageNodeRemove")}
                onclick={() => onMicronImageNodePolicy(hash, null)}
              >
                {t("settings.micronImageNodeRemove")}
              </button>
            </li>
          {/each}
        </ul>
      </div>
    {/if}

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
              aria-pressed={recordingAction === action}
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
    <CommunityInterfaces
      showTitle={false}
      items={communityItems}
      loading={communityLoading}
      importing={communityImporting}
      error={communityError}
      bind:filter={communityFilter}
      selected={communitySelected}
      onFilter={onCommunityFilter}
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

    <div class="rns-toggles">
      <Toggle
        label={t("settings.enableTransport")}
        checked={reticulumStatus.enableTransport}
        onchange={onToggleTransport}
      />
      <p class="hint">{t("settings.enableTransportHint")}</p>
      <p class="meta">
        {#if reticulumStatus.connectedToSharedInstance}
          {t("settings.transportInactiveSharedClient")}
        {:else if reticulumStatus.transportActive}
          {t("settings.transportActive")}
        {:else}
          {t("settings.transportInactive")}
        {/if}
      </p>

      <Toggle
        label={t("settings.shareInstance")}
        checked={reticulumStatus.shareInstance}
        onchange={onToggleShareInstance}
      />
      <p class="hint">{t("settings.shareInstanceHint")}</p>
      <p class="meta sharedInstanceStatus">
        <span class="label">{t("settings.sharedInstanceStatus")}:</span>
        <span
          class="status-badge"
          class:server={reticulumStatus.sharedInstanceMode === "server"}
          class:client={reticulumStatus.sharedInstanceMode === "client"}
        >
          {#if reticulumStatus.sharedInstanceMode === "server"}
            {t("settings.sharedInstanceModeServer")}
          {:else if reticulumStatus.sharedInstanceMode === "client"}
            {t("settings.sharedInstanceModeClient")}
          {:else}
            {t("settings.sharedInstanceModeDisabled")}
          {/if}
        </span>
      </p>
      <p class="hint">{t("settings.shareInstanceRestartHint")}</p>
    </div>

    <p class="hint" style="margin-top: 1rem;">{t("settings.restartReticulumHint")}</p>
    <div class="reset-row" style="margin-bottom: 1rem;">
      <button type="button" class="reset-btn" onclick={onRestartReticulum}
        >{t("settings.restartReticulum")}</button
      >
    </div>

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
    <div class="security-stack">
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

      <div class="sandbox-card" class:active={sandboxStatus.inFlatpak}>
        <div class="sandbox-head">
          <span class="sandbox-icon" aria-hidden="true">
            <Package size={20} strokeWidth={2} />
          </span>
          <div class="sandbox-copy">
            <span class="sandbox-name">{t("settings.flatpakTitle")}</span>
            <span class="sandbox-subtitle">{t("settings.flatpakSubtitle")}</span>
          </div>
          <span
            class="sandbox-badge"
            class:enabled={sandboxStatus.inFlatpak}
            class:disabled={!sandboxStatus.inFlatpak}
          >
            {sandboxStatus.inFlatpak ? t("settings.envYes") : t("settings.envNo")}
          </span>
        </div>
        <p class="sandbox-note">
          {sandboxStatus.inFlatpak
            ? t("settings.flatpakActiveHint")
            : t("settings.flatpakInactiveHint")}
        </p>
      </div>

      <div
        class="sandbox-card"
        class:active={sandboxStatus.webkitSandbox === "active"}
        class:warn={sandboxStatus.webkitSandbox === "disabled"}
      >
        <div class="sandbox-head">
          <span class="sandbox-icon" aria-hidden="true">
            {#if sandboxStatus.webkitSandbox === "active"}
              <ShieldCheck size={20} strokeWidth={2} />
            {:else}
              <ShieldOff size={20} strokeWidth={2} />
            {/if}
          </span>
          <div class="sandbox-copy">
            <span class="sandbox-name">{t("settings.webkitSandboxTitle")}</span>
            <span class="sandbox-subtitle">{t("settings.webkitSandboxSubtitle")}</span>
          </div>
          <span
            class="sandbox-badge"
            class:enabled={sandboxStatus.webkitSandbox === "active"}
            class:disabled={sandboxStatus.webkitSandbox !== "active"}
            class:warn={sandboxStatus.webkitSandbox === "disabled"}
          >
            {webkitSandboxLabel(sandboxStatus.webkitSandbox)}
          </span>
        </div>
        {#if webkitSandboxNoteLabel(sandboxStatus.webkitSandboxNote)}
          <p class="sandbox-note">{webkitSandboxNoteLabel(sandboxStatus.webkitSandboxNote)}</p>
        {:else if sandboxStatus.webkitSandbox === "active"}
          <p class="sandbox-note active">{t("settings.webkitSandboxActiveHint")}</p>
        {/if}
      </div>

      <div class="sandbox-card" class:active={sandboxStatus.inContainer}>
        <div class="sandbox-head">
          <span class="sandbox-icon" aria-hidden="true">
            <Box size={20} strokeWidth={2} />
          </span>
          <div class="sandbox-copy">
            <span class="sandbox-name">{t("settings.containerTitle")}</span>
            <span class="sandbox-subtitle">{t("settings.containerSubtitle")}</span>
          </div>
          <span
            class="sandbox-badge"
            class:enabled={sandboxStatus.inContainer}
            class:disabled={!sandboxStatus.inContainer}
          >
            {sandboxStatus.inContainer ? t("settings.envYes") : t("settings.envNo")}
          </span>
        </div>
        <p class="sandbox-note">
          {#if sandboxStatus.inContainer}
            {containerRuntimeLabel(sandboxStatus.containerRuntime)}
          {:else}
            {t("settings.containerInactiveHint")}
          {/if}
        </p>
      </div>

      <div class="sandbox-card" class:active={sandboxStatus.onAndroid}>
        <div class="sandbox-head">
          <span class="sandbox-icon" aria-hidden="true">
            <Smartphone size={20} strokeWidth={2} />
          </span>
          <div class="sandbox-copy">
            <span class="sandbox-name">{t("settings.androidTitle")}</span>
            <span class="sandbox-subtitle">{t("settings.androidSubtitle")}</span>
          </div>
          <span
            class="sandbox-badge"
            class:enabled={sandboxStatus.onAndroid}
            class:disabled={!sandboxStatus.onAndroid}
          >
            {sandboxStatus.onAndroid ? t("settings.envYes") : t("settings.envNo")}
          </span>
        </div>
        <p class="sandbox-note">
          {sandboxStatus.onAndroid
            ? t("settings.androidActiveHint")
            : t("settings.androidInactiveHint")}
        </p>
      </div>
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

  <SettingsSection
    id="selfTest"
    title={t("settings.selfTest")}
    collapsed={sectionCollapsed("selfTest")}
    onToggle={toggleSettingsSection}
  >
    <p class="hint">{t("settings.selfTestHint")}</p>
    <div class="reset-row" style="margin-bottom: 1rem;">
      <button type="button" class="reset-btn" onclick={onRunSelfTest} disabled={selfTestRunning}>
        {#if selfTestRunning}
          {t("settings.selfTestRunning")}
        {:else}
          {t("settings.runSelfTest")}
        {/if}
      </button>
    </div>

    {#if selfTestResult}
      <div class="self-test-results">
        <div
          class="self-test-summary"
          class:passed={selfTestResult.allPassed}
          class:failed={!selfTestResult.allPassed}
        >
          {selfTestResult.allPassed ? t("settings.selfTestPassed") : t("settings.selfTestFailed")}
        </div>
        <ul class="self-test-list">
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkStackUp")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.stackUp.passed}
              class:failed={!selfTestResult.stackUp.passed}
            >
              {selfTestResult.stackUp.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if !selfTestResult.stackUp.passed}
              <div class="check-reason">{selfTestResult.stackUp.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkConfigGood")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.configGood.passed}
              class:failed={!selfTestResult.configGood.passed}
            >
              {selfTestResult.configGood.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if !selfTestResult.configGood.passed}
              <div class="check-reason">{selfTestResult.configGood.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkDBGood")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.dbGood.passed}
              class:failed={!selfTestResult.dbGood.passed}
            >
              {selfTestResult.dbGood.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if !selfTestResult.dbGood.passed}
              <div class="check-reason">{selfTestResult.dbGood.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkReadWriteGood")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.readWriteGood.passed}
              class:failed={!selfTestResult.readWriteGood.passed}
            >
              {selfTestResult.readWriteGood.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if !selfTestResult.readWriteGood.passed}
              <div class="check-reason">{selfTestResult.readWriteGood.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkDownloadsGood")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.downloadsGood.passed}
              class:failed={!selfTestResult.downloadsGood.passed}
            >
              {selfTestResult.downloadsGood.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if !selfTestResult.downloadsGood.passed}
              <div class="check-reason">{selfTestResult.downloadsGood.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkInterfaces")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.interfaces.passed}
              class:failed={!selfTestResult.interfaces.passed}
            >
              {selfTestResult.interfaces.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if selfTestResult.interfaces.reason}
              <div class="check-reason">{selfTestResult.interfaces.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkDiscovery")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.discovery.passed}
              class:failed={!selfTestResult.discovery.passed}
            >
              {selfTestResult.discovery.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if selfTestResult.discovery.reason}
              <div class="check-reason">{selfTestResult.discovery.reason}</div>
            {/if}
          </li>
          <li class="self-test-item">
            <span class="check-name">{t("settings.checkPageFetch")}</span>
            <span
              class="check-status"
              class:passed={selfTestResult.pageFetch.passed}
              class:failed={!selfTestResult.pageFetch.passed}
            >
              {selfTestResult.pageFetch.passed ? t("settings.passed") : t("settings.failed")}
            </span>
            {#if selfTestResult.pageFetch.reason}
              <div class="check-reason">{selfTestResult.pageFetch.reason}</div>
            {/if}
          </li>
        </ul>
      </div>
    {/if}
  </SettingsSection>

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

      <p class="hint" style="margin-top: 1rem;">{t("settings.resetBrowserHint")}</p>
      <div class="reset-row">
        <button type="button" class="reset-btn" onclick={onResetBrowser}
          >{t("settings.resetBrowser")}</button
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
    text-align: left;
  }

  .settings :global(.ren-select [data-select-value]) {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :global(.ren-select-content) {
    z-index: 1100;
    min-width: var(--bits-select-anchor-width);
    max-width: calc(100vw - 1rem);
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
    overflow: hidden;
  }

  :global(.ren-select-content [data-select-viewport]) {
    display: grid;
    gap: 0.15rem;
    padding: 0.35rem;
    max-height: min(18rem, var(--bits-select-content-available-height));
    overflow-y: auto;
  }

  :global(.ren-select-content [data-select-item]) {
    display: flex;
    align-items: center;
    border-radius: 8px;
    padding: 0.45rem 0.65rem;
    font-size: 0.88rem;
    color: var(--ren-fg);
    cursor: pointer;
    user-select: none;
    outline: none;
  }

  :global(.ren-select-content [data-select-item][data-highlighted]) {
    background: var(--ren-tab-hover);
  }

  :global(.ren-select-content [data-select-item][data-selected]) {
    color: var(--ren-accent);
  }

  input:focus {
    border-color: var(--ren-focus);
  }

  input:focus-visible {
    outline: none;
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--ren-focus) 28%, transparent);
  }

  input[type="color"] {
    padding: 0;
    min-height: 2.25rem;
  }

  :global(.settings-slider) {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    height: 0.5rem;
    border: 1px solid var(--ren-border);
    border-radius: 999px;
    background: var(--ren-input-bg);
    cursor: pointer;
    touch-action: none;
    user-select: none;
  }

  :global(.settings-slider-range) {
    height: 100%;
    border-radius: 999px;
    background: var(--ren-accent);
  }

  :global(.settings-slider-thumb) {
    display: block;
    width: 1rem;
    height: 1rem;
    border: 1px solid var(--ren-border-strong, var(--ren-border));
    border-radius: 50%;
    background: var(--ren-fg);
    cursor: grab;
  }

  :global(.settings-slider-thumb:focus-visible) {
    outline: none;
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--ren-focus) 28%, transparent);
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
    overflow-wrap: break-word;
    word-break: normal;
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

  .rns-toggles {
    display: grid;
    gap: 0.45rem;
    margin: 1rem 0;
  }

  .rns-toggles .meta {
    margin: 0;
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

  .sharedInstanceStatus {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-top: 0.35rem;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    padding: 0.15rem 0.5rem;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
    background: var(--ren-surface-muted);
    color: var(--ren-muted);
    border: 1px solid var(--ren-border);
  }

  .status-badge.server {
    background: color-mix(in srgb, var(--ren-accent) 12%, var(--ren-surface-muted));
    color: var(--ren-accent);
    border-color: color-mix(in srgb, var(--ren-accent) 35%, var(--ren-border));
  }

  .status-badge.client {
    background: color-mix(in srgb, var(--ren-success, #3d9a5f) 12%, var(--ren-surface-muted));
    color: var(--ren-success, #3d9a5f);
    border-color: color-mix(in srgb, var(--ren-success, #3d9a5f) 35%, var(--ren-border));
  }

  .sandbox-card {
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-surface-raised);
    padding: 0.85rem 0.9rem;
    display: grid;
    gap: 0.65rem;
  }

  .security-stack {
    display: grid;
    gap: 0.75rem;
  }

  .sandbox-card.active {
    border-color: color-mix(in srgb, var(--ren-success, #3d9a5f) 35%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-success, #3d9a5f) 6%, var(--ren-surface-raised));
  }

  .sandbox-card.warn {
    border-color: color-mix(in srgb, var(--ren-warning, #c9852a) 40%, var(--ren-border));
    background: color-mix(in srgb, var(--ren-warning, #c9852a) 7%, var(--ren-surface-raised));
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

  .sandbox-badge.warn {
    color: var(--ren-warning, #c9852a);
    border-color: color-mix(in srgb, var(--ren-warning, #c9852a) 45%, transparent);
    background: color-mix(in srgb, var(--ren-warning, #c9852a) 12%, transparent);
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

  .self-test-results {
    margin-top: 1rem;
    display: grid;
    gap: 0.75rem;
    background: var(--ren-bg-secondary, #1e1e24);
    padding: 1rem;
    border-radius: var(--ren-radius);
    border: 1px solid var(--ren-border);
  }

  .self-test-summary {
    font-weight: 600;
    font-size: 0.95rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid var(--ren-border);
  }

  .self-test-summary.passed {
    color: var(--ren-success, #3d9a5f);
  }

  .self-test-summary.failed {
    color: var(--ren-danger, #e5484d);
  }

  .self-test-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: grid;
    gap: 0.65rem;
  }

  .self-test-item {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.88rem;
  }

  .check-name {
    color: var(--ren-text);
  }

  .check-status {
    font-weight: 600;
    font-size: 0.82rem;
    padding: 0.12rem 0.4rem;
    border-radius: 4px;
  }

  .check-status.passed {
    background: color-mix(in srgb, var(--ren-success, #3d9a5f) 15%, transparent);
    color: var(--ren-success, #3d9a5f);
  }

  .check-status.failed {
    background: color-mix(in srgb, var(--ren-danger, #e5484d) 15%, transparent);
    color: var(--ren-danger, #e5484d);
  }

  .check-reason {
    grid-column: 1 / -1;
    font-size: 0.8rem;
    color: var(--ren-muted);
    padding-left: 0.5rem;
    border-left: 2px solid var(--ren-danger, #e5484d);
    margin-top: 0.15rem;
  }

  .micron-image-nodes {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .micron-image-nodes-title {
    font-size: 0.85rem;
    color: var(--ren-muted);
  }

  .micron-image-nodes ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .micron-image-nodes li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
  }

  .micron-image-node-hash {
    flex: 1 1 12rem;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.78rem;
    color: var(--ren-fg);
  }

  .micron-image-nodes :global(.micron-image-node-policy) {
    flex-shrink: 0;
    max-width: 8rem;
  }

  .micron-image-node-remove {
    flex-shrink: 0;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: 6px;
    padding: 0.25rem 0.55rem;
    font-size: 0.78rem;
    cursor: pointer;
  }

  .micron-image-node-remove:hover {
    background: var(--ren-tab-hover);
  }
</style>
