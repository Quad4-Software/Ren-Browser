<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import ContentViewer from "$lib/components/ContentViewer.svelte";
  import MicronEditor from "$lib/components/MicronEditor.svelte";
  import ReticulumConfigEditor from "$lib/components/ReticulumConfigEditor.svelte";
  import SplitPane from "$lib/components/SplitPane.svelte";
  import SplitTabPicker from "$lib/components/SplitTabPicker.svelte";
  import AppSettingsPane from "$lib/components/AppSettingsPane.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import type { AppController } from "$lib/app/create-app.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();
</script>

{#snippet settingsPane()}
  <AppSettingsPane {app} />
{/snippet}

{#snippet primaryPane()}
  {#if app.contentType === "editor"}
    {#if app.loading}
      <div class="editor-loading">{t("editor.loadingMicron")}</div>
    {:else}
      <MicronEditor
        source={app.lastRaw}
        currentURL={app.url}
        micronWasmEnabled={app.micronWasmEnabled}
        micronWasmParserId={app.micronWasmParserId}
        micronWasmReady={app.micronWasmReady}
        onSourceChange={app.updateEditorSource}
        onNavigate={app.openPage}
      />
    {/if}
  {:else if app.contentType === "config"}
    {#if app.loading}
      <div class="editor-loading">{t("editor.loadingConfig")}</div>
    {:else}
      <section class="config-page">
        <ReticulumConfigEditor
          bind:configText={app.configText}
          configPath={app.configPath}
          saving={app.configSaving}
          error={app.configError}
          onChange={(text) => {
            app.configText = text;
          }}
          onSave={() => void app.saveConfigText()}
          onReload={() => void app.reloadConfigFromDisk()}
          onOpenConfigDir={() => void app.openConfigFolder()}
        />
      </section>
    {/if}
  {:else if app.contentType === "settings"}
    {#if app.loading}
      <div class="editor-loading">{t("common.loading")}</div>
    {:else}
      <div class="settings-page">
        {@render settingsPane()}
      </div>
    {/if}
  {:else}
    <ContentViewer
      html={app.html}
      contentType={app.contentType}
      loading={app.loading}
      error={app.error}
      errorKind={app.errorKind}
      pageFg={app.pageFg}
      pageBg={app.pageBg}
      raw={app.lastRaw}
      binaryB64={app.binaryB64}
      fromCache={app.fromCache}
      cachedAt={app.cachedAt}
      showSource={app.showSource}
      currentURL={app.url}
      findOpen={app.findOpen}
      pageHighlight={app.pageHighlight}
      onPageHighlightDone={app.clearPageHighlight}
      micronEngine={app.effectiveMicronEngine}
      micronPreserveLayout={app.micronPreserveLayout}
      mobileGestures={app.mobileUI && app.activePanel === "browser" && !app.mobileTabsOpen}
      canGoBack={app.canGoBack}
      onBack={app.goBack}
      onFindClose={() => (app.findOpen = false)}
      onNavigate={app.openPage}
      onRetry={() => app.openPage(app.url)}
      onReloadFresh={() => app.openPage(app.url, true, { skipCache: true })}
      onShowSourceChange={app.setShowSource}
      onDownloadResult={app.handleDownloadResult}
    />
  {/if}
{/snippet}

{#if app.splitViewOpen}
  {#snippet secondaryPane()}
    {#if app.splitTab}
      {@const splitPage = app.splitTab.page ?? app.emptyPage()}
      <ContentViewer
        html={splitPage.html}
        contentType={splitPage.contentType}
        loading={false}
        error={splitPage.error}
        errorKind={splitPage.errorKind}
        pageFg={splitPage.pageFg}
        pageBg={splitPage.pageBg}
        raw={splitPage.lastRaw}
        binaryB64={splitPage.binaryB64 ?? ""}
        fromCache={splitPage.fromCache}
        cachedAt={splitPage.cachedAt ?? 0}
        showSource={splitPage.showSource ?? false}
        currentURL={app.splitTab.url}
        findOpen={false}
        micronEngine={app.effectiveMicronEngine}
        micronPreserveLayout={app.micronPreserveLayout}
        onFindClose={() => {}}
        onNavigate={(target) => void app.openPage(target, true, { tabId: app.splitTab!.id })}
        onRetry={() => void app.openPage(app.splitTab!.url, false, { tabId: app.splitTab!.id })}
        onReloadFresh={() =>
          void app.openPage(app.splitTab!.url, true, {
            tabId: app.splitTab!.id,
            skipCache: true,
          })}
        onShowSourceChange={(value) => app.setSplitTabShowSource(app.splitTab!.id, value)}
        onDownloadResult={app.handleDownloadResult}
      />
    {:else}
      <SplitTabPicker
        tabs={app.tabs}
        activeTabId={app.activeTabId}
        onSelect={app.selectSplitTab}
        onClose={app.closeSplitView}
      />
    {/if}
  {/snippet}
  <SplitPane
    ratio={app.splitRatio}
    onRatioChange={(value) => (app.splitRatio = value)}
    primary={primaryPane}
    secondary={secondaryPane}
  />
{:else}
  {@render primaryPane()}
{/if}

<style>
  .editor-loading {
    height: 100%;
    display: grid;
    place-items: center;
    color: var(--ren-muted);
  }

  .config-page {
    height: 100%;
    overflow: auto;
    padding: 1rem;
    background: var(--ren-content-bg);
  }

  .settings-page {
    height: 100%;
    width: 100%;
    max-width: 100%;
    min-width: 0;
    overflow: hidden;
  }
</style>
