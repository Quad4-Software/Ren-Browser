<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import DiscoveryPanel from "$lib/components/DiscoveryPanel.svelte";
  import HistoryPanel from "$lib/components/HistoryPanel.svelte";
  import SearchPanel from "$lib/components/SearchPanel.svelte";
  import DevToolsPanel from "$lib/components/DevToolsPanel.svelte";
  import PluginPanelHost from "$lib/components/PluginPanelHost.svelte";
  import AppSettingsPane from "$lib/components/AppSettingsPane.svelte";
  import { pluginLabel } from "$lib/plugins/plugin-label.js";
  import type { AppController } from "$lib/app/create-app.svelte";

  type Props = {
    app: AppController;
  };

  let { app }: Props = $props();
</script>

{#if app.activePanel === "search"}
  <aside class="side-pane">
    <SearchPanel
      history={app.history}
      nodes={app.nodes}
      favorites={app.favorites}
      onOpen={app.browseURL}
    />
  </aside>
{:else if app.activePanel === "discovery"}
  <aside class="side-pane">
    <DiscoveryPanel
      nodes={app.nodes}
      favorites={app.favorites}
      slowMode={app.discoverySlowMode}
      onSlowModeChange={app.saveDiscoverySlowMode}
      onOpen={app.browseURL}
      onFavorite={(favUrl) => void app.addFavoriteUrl(favUrl)}
    />
  </aside>
{:else if app.activePanel === "history"}
  <aside class="side-pane">
    <HistoryPanel history={app.history} onOpen={app.browseURL} onClear={app.requestClearHistory} />
  </aside>
{:else if app.activePanel === "devtools"}
  <aside class="side-pane">
    <DevToolsPanel
      logs={app.logs}
      network={app.network}
      raw={app.lastRaw}
      logLevel={app.logLevel}
      contentType={app.contentType}
      durationMs={app.durationMs}
      hops={app.hops}
      fromCache={app.fromCache}
      cachedAt={app.cachedAt}
      micronRendererBadge={app.micronRendererBadge}
      pluginTabs={app.pluginContributions.devtools}
      onClear={() => void app.clearDevLogsPanel()}
      onExport={() => void app.exportDevLogsFile()}
      onLogLevel={(level) => void app.setDevLogLevel(level)}
    />
  </aside>
{:else if app.activePanel === "settings"}
  <aside class="side-pane">
    <AppSettingsPane {app} />
  </aside>
{:else if app.activePluginPanel}
  <aside class="side-pane">
    <PluginPanelHost
      pluginId={app.activePluginPanel.pluginId}
      panelId={app.activePluginPanel.id}
      title={pluginLabel(app.activePluginPanel.pluginId, app.activePluginPanel.title)}
      entry={app.activePluginPanel.entry}
      {...app.pluginHostOpts(app.activePluginPanel.pluginId)}
    />
  </aside>
{/if}

<style>
  .side-pane {
    min-height: 0;
    min-width: 0;
    border-top: 1px solid var(--ren-border);
    border-left: 1px solid var(--ren-border);
    background: var(--ren-chrome-bg);
    box-shadow: var(--ren-shadow);
  }
</style>
