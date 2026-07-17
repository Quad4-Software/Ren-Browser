<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import SettingsPanel from "$lib/components/SettingsPanel.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();
</script>

<SettingsPanel
  bind:theme={app.theme}
  systemFonts={app.systemFonts}
  keybinds={app.keybinds}
  interfaces={app.interfaces}
  reticulumStatus={app.reticulumStatus}
  configPath={app.configPath}
  pluginsDir={app.pluginsDir}
  sandboxStatus={app.sandboxStatus}
  bind:downloadDir={app.downloadDir}
  uiLanguage={app.uiLanguage}
  onChangeUILanguage={app.saveUILanguage}
  openLinksInNewTab={app.openLinksInNewTab}
  tabHoverPreviews={app.tabHoverPreviews}
  nativeTitlebar={app.nativeTitlebar}
  micronRenderer={app.micronRenderer}
  micronWasmEnabled={app.micronWasmEnabled}
  micronWasmParserId={app.micronWasmParserId}
  micronPreserveLayout={app.micronPreserveLayout}
  desktopChrome={app.desktopChrome}
  mobileUI={app.mobileUI}
  mobileDevTools={app.mobileDevTools}
  publicMode={app.publicMode}
  bind:configText={app.configText}
  configSaving={app.configSaving}
  configError={app.configError}
  communityItems={app.communityItems}
  communityLoading={app.communityLoading}
  communityImporting={app.communityImporting}
  communityError={app.communityError}
  communityFromBundle={app.communityFromBundle}
  bind:communityFilter={app.communityFilter}
  communitySelected={app.communitySelected}
  sectionsCollapsed={app.settingsSectionsCollapsed}
  onChangeSectionsCollapsed={app.saveSettingsSectionsCollapsed}
  onChange={app.saveTheme}
  onChangeKeybinds={app.saveKeybinds}
  onChangeDownloadDir={app.saveDownloadDir}
  onPickDownloadDir={app.pickDownloadDir}
  onChangeOpenLinksInNewTab={app.saveOpenLinksInNewTab}
  onChangeTabHoverPreviews={app.saveTabHoverPreviews}
  onChangeMobileDevTools={app.saveMobileDevTools}
  onOpenSearch={() => app.setPanel("search")}
  onChangeNativeTitlebar={app.saveNativeTitlebar}
  onChangeMicronRenderer={app.saveMicronRenderer}
  onChangeMicronWasmEnabled={app.saveMicronWasmEnabled}
  onChangeMicronPreserveLayout={app.saveMicronPreserveLayout}
  onChangeMicronWasmParser={app.saveMicronWasmParser}
  onMicronWasmReadyChange={app.setMicronWasmReady}
  onResetDefaults={app.resetDefaults}
  onResetBrowser={app.requestResetBrowser}
  onRestartReticulum={app.requestRestartReticulum}
  onShutdown={app.requestShutdown}
  selfTestResult={app.selfTestResult}
  selfTestRunning={app.selfTestRunning}
  onRunSelfTest={app.runSelfTest}
  onToggleInterface={app.toggleInterface}
  onToggleTransport={app.toggleTransport}
  onToggleShareInstance={app.toggleShareInstance}
  onExportTheme={() => void app.exportThemeFile()}
  onImportTheme={(json) => void app.importThemeFile(json)}
  onConfigChange={(text) => {
    app.configText = text;
  }}
  onConfigSave={() => void app.saveConfigText()}
  onConfigReload={() => void app.reloadConfigFromDisk()}
  onOpenConfigDir={() => void app.openConfigFolder()}
  onClearPageCache={() => void app.clearPageCache()}
  onChangePageCacheEnabled={(value) => void app.savePageCacheEnabled(value)}
  pageCacheRAMEntries={app.pageCacheRAMEntries}
  pageCacheDiskEntries={app.pageCacheDiskEntries}
  pageCacheClearing={app.pageCacheClearing}
  pageCacheEnabled={app.pageCacheEnabled}
  onCommunityRefresh={() => void app.loadCommunityInterfaces()}
  onCommunityFilter={(value) => {
    app.communityFilter = value;
  }}
  onCommunityToggle={app.toggleCommunitySelection}
  onCommunityImport={() => void app.importCommunitySelection()}
  onPluginsChanged={() => void app.reloadPlugins()}
/>
