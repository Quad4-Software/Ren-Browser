<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Activity, FileCode, Terminal } from "@lucide/svelte";
  import { Slider, Tabs } from "bits-ui";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import PluginPanelHost from "$lib/components/PluginPanelHost.svelte";
  import { t } from "$lib/i18n/i18n.svelte";
  import type { PluginDevToolsContribution } from "$lib/plugins/api-types.js";
  type LogEntry = {
    time: number;
    level: string;
    message: string;
    detail?: string;
  };

  type NetworkEntry = {
    time: number;
    url: string;
    nodeHash: string;
    path: string;
    durationMs: number;
    bytes: number;
    fromCache: boolean;
    hops: number;
    interface?: string;
    error?: string;
    source?: string;
    pluginId?: string;
    method?: string;
    statusCode?: number;
  };

  function networkSourceLabel(row: NetworkEntry): string {
    if (row.source === "plugin") {
      const plugin = row.pluginId ? ` ${row.pluginId}` : "";
      const method = row.method ? ` ${row.method}` : "";
      return `${t("devtools.pluginFetch")}${plugin}${method}`.trim();
    }
    return row.interface || "—";
  }

  function networkPathLabel(row: NetworkEntry): string {
    if (row.source === "plugin") {
      return row.url || row.path;
    }
    return row.path;
  }

  type Props = {
    logs: LogEntry[];
    network: NetworkEntry[];
    raw: string;
    logLevel: number;
    contentType: string;
    durationMs: number;
    hops: number;
    fromCache: boolean;
    cachedAt: number;
    micronRendererBadge?: string;
    pluginTabs?: PluginDevToolsContribution[];
    onClear: () => void;
    onExport: () => void;
    onLogLevel: (level: number) => void;
  };

  let {
    logs,
    network,
    raw,
    logLevel,
    contentType,
    durationMs,
    hops,
    fromCache,
    cachedAt,
    micronRendererBadge = "",
    pluginTabs = [],
    onClear,
    onExport,
    onLogLevel,
  }: Props = $props();

  let tab = $state<string>("console");
  let rawMode = $state<"text" | "hex">("text");
  let logQuery = $state("");

  const filteredLogs = $derived.by(() => {
    const q = logQuery.trim().toLowerCase();
    if (!q) {
      return logs;
    }
    return logs.filter((entry) => {
      const hay = `${entry.level} ${entry.message} ${entry.detail ?? ""}`.toLowerCase();
      return hay.includes(q);
    });
  });

  function formatTime(ms: number): string {
    return new Date(ms).toLocaleTimeString();
  }

  function toHex(text: string): string {
    const bytes = new TextEncoder().encode(text);
    const parts: string[] = [];
    for (const b of bytes) {
      parts.push(b.toString(16).padStart(2, "0"));
    }
    return parts.join(" ");
  }

  function formatHops(value: number): string {
    if (value < 0) {
      return "—";
    }
    return value === 1
      ? t("devtools.hopsCount", { count: value })
      : t("devtools.hopsCountPlural", { count: value });
  }
</script>

<section class="devtools">
  <Tabs.Root bind:value={tab} class="devtools-tabs">
    <header>
      <Tabs.List class="tabs">
        <Tabs.Trigger value="console">{t("devtools.console")}</Tabs.Trigger>
        <Tabs.Trigger value="network">{t("devtools.network")}</Tabs.Trigger>
        <Tabs.Trigger value="raw">{t("devtools.raw")}</Tabs.Trigger>
        {#each pluginTabs as pluginTab (pluginTab.pluginId + ":" + pluginTab.id)}
          <Tabs.Trigger value={`plugin:${pluginTab.pluginId}:${pluginTab.id}`}>
            {pluginTab.title}
          </Tabs.Trigger>
        {/each}
      </Tabs.List>
      <div class="actions">
        <label>
          {t("devtools.log")}
          <Slider.Root
            type="single"
            class="devtools-slider"
            value={logLevel}
            min={1}
            max={7}
            step={1}
            onValueChange={onLogLevel}
          >
            <Slider.Range class="devtools-slider-range" />
            <Slider.Thumb index={0} class="devtools-slider-thumb" aria-label={t("devtools.log")} />
          </Slider.Root>
        </label>
        <button type="button" onclick={onExport}>{t("devtools.export")}</button>
        <button type="button" onclick={onClear}>{t("devtools.clear")}</button>
      </div>
    </header>

    <div class="page-info">
      <span>{contentType || t("common.unknown")}</span>
      {#if micronRendererBadge}
        <span class="renderer-badge">{micronRendererBadge}</span>
      {/if}
      {#if hops >= 0}
        <span>{formatHops(hops)}</span>
      {/if}
      {#if durationMs > 0}
        <span>{durationMs} ms</span>
      {/if}
      {#if fromCache}
        <span
          >{t("devtools.cached", {
            when: cachedAt > 0 ? new Date(cachedAt).toLocaleString() : "",
          })}</span
        >
      {/if}
    </div>

    <Tabs.Content value="console" class="panel logs">
      <input
        class="search ren-input"
        type="search"
        bind:value={logQuery}
        aria-label={t("devtools.searchLogs")}
        placeholder={t("devtools.searchLogs")}
        spellcheck="false"
        autocomplete="off"
      />
      {#if logs.length === 0}
        <EmptyState
          title={t("devtools.consoleEmpty")}
          description={t("devtools.consoleEmptyDescription")}
        >
          <Terminal size={22} />
        </EmptyState>
      {:else if filteredLogs.length === 0}
        <EmptyState
          title={t("devtools.noMatchingLogs")}
          description={t("common.nothingMatches", { query: logQuery.trim() })}
        >
          <Terminal size={22} />
        </EmptyState>
      {:else}
        {#each filteredLogs as entry (entry.time + entry.message)}
          <div class="entry" data-level={entry.level}>
            <span class="time">{formatTime(entry.time)}</span>
            <span class="level">{entry.level}</span>
            <span class="message">{entry.message}</span>
            {#if entry.detail}
              <pre>{entry.detail}</pre>
            {/if}
          </div>
        {/each}
      {/if}
    </Tabs.Content>
    <Tabs.Content value="network" class="panel network">
      {#if network.length === 0}
        <EmptyState
          title={t("devtools.noNetwork")}
          description={t("devtools.noNetworkDescription")}
        >
          <Activity size={22} />
        </EmptyState>
      {:else}
        <table>
          <thead>
            <tr>
              <th>{t("devtools.time")}</th>
              <th>{t("devtools.path")}</th>
              <th>{t("devtools.hops")}</th>
              <th>{t("devtools.interface")}</th>
              <th>{t("devtools.status")}</th>
              <th>ms</th>
              <th>{t("devtools.bytes")}</th>
              <th>{t("devtools.cache")}</th>
              <th>{t("devtools.error")}</th>
            </tr>
          </thead>
          <tbody>
            {#each network as row (row.time + row.url)}
              <tr data-source={row.source ?? "page"}>
                <td>{formatTime(row.time)}</td>
                <td>{networkPathLabel(row)}</td>
                <td>{row.source === "plugin" ? "—" : formatHops(row.hops ?? -1)}</td>
                <td>{networkSourceLabel(row)}</td>
                <td>{row.source === "plugin" ? row.statusCode || "—" : "—"}</td>
                <td>{row.durationMs}</td>
                <td>{row.bytes}</td>
                <td>{row.fromCache ? t("common.yes") : t("common.no")}</td>
                <td>{row.error ?? ""}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </Tabs.Content>
    <Tabs.Content value="raw" class="panel raw">
      <div class="raw-actions">
        <button class:active={rawMode === "text"} onclick={() => (rawMode = "text")}
          >{t("devtools.text")}</button
        >
        <button class:active={rawMode === "hex"} onclick={() => (rawMode = "hex")}
          >{t("devtools.hex")}</button
        >
      </div>
      {#if raw.trim().length === 0}
        <EmptyState title={t("devtools.noRaw")} description={t("devtools.noRawDescription")}>
          <FileCode size={22} />
        </EmptyState>
      {:else}
        <pre>{rawMode === "text" ? raw : toHex(raw)}</pre>
      {/if}
    </Tabs.Content>
    {#each pluginTabs as pluginTab (pluginTab.pluginId + ":" + pluginTab.id)}
      {#if tab === `plugin:${pluginTab.pluginId}:${pluginTab.id}`}
        <PluginPanelHost
          pluginId={pluginTab.pluginId}
          panelId={pluginTab.id}
          title={pluginTab.title}
          entry={pluginTab.entry}
        />
      {/if}
    {/each}
  </Tabs.Root>
</section>

<style>
  .devtools {
    height: 100%;
    display: flex;
    flex-direction: column;
    background: var(--ren-content-bg);
  }

  header {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    align-items: center;
    padding: 0.55rem 0.75rem;
    border-bottom: 1px solid var(--ren-border);
    flex-wrap: wrap;
    background: var(--ren-chrome-bg);
  }

  .page-info {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    padding: 0.45rem 0.85rem;
    border-bottom: 1px solid var(--ren-border);
    color: var(--ren-muted);
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    background: var(--ren-chrome-bg);
  }

  .renderer-badge {
    color: var(--ren-accent);
  }

  .devtools :global(.devtools-tabs) {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    min-width: 0;
  }

  .devtools :global(.tabs),
  .actions {
    display: flex;
    gap: 0.35rem;
    align-items: center;
  }

  .devtools :global(.tabs button),
  .actions button,
  .raw-actions button {
    border: 1px solid var(--ren-border);
    background: var(--ren-surface-muted);
    color: var(--ren-fg-secondary);
    border-radius: 10px;
    padding: 0.35rem 0.65rem;
    cursor: pointer;
    font: inherit;
    font-size: 0.82rem;
    transition:
      background 0.15s ease,
      color 0.15s ease,
      border-color 0.15s ease;
  }

  .devtools :global(.tabs button:hover),
  .actions button:hover,
  .raw-actions button:hover {
    background: var(--ren-tab-hover);
    color: var(--ren-fg);
  }

  .devtools :global(.tabs button.active),
  .devtools :global(.tabs button[data-state="active"]),
  .raw-actions button.active {
    border-color: var(--ren-accent);
    background: var(--ren-accent);
    color: #fff;
  }

  .devtools :global(.panel) {
    flex: 1;
    overflow: auto;
    padding: 0.75rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.82rem;
  }

  .devtools :global(.panel[hidden]) {
    display: none;
  }

  .search {
    margin-bottom: 0.65rem;
    width: 100%;
    font-family: inherit;
  }

  .entry {
    display: grid;
    grid-template-columns: auto auto 1fr;
    gap: 0.5rem;
    padding: 0.35rem 0;
    border-bottom: 1px solid color-mix(in srgb, var(--ren-border) 55%, transparent);
  }

  .entry[data-level="error"] .level {
    color: var(--ren-danger);
  }

  .time,
  .level {
    color: var(--ren-muted);
  }

  pre {
    grid-column: 1 / -1;
    margin: 0.25rem 0 0;
    white-space: pre-wrap;
    color: var(--ren-muted);
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  th,
  td {
    border-bottom: 1px solid var(--ren-border);
    text-align: left;
    padding: 0.35rem 0.25rem;
    vertical-align: top;
  }

  label {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    color: var(--ren-muted);
    font-size: 0.78rem;
  }

  :global(.devtools-slider) {
    position: relative;
    display: flex;
    align-items: center;
    flex-shrink: 0;
    width: 5.5rem;
    height: 0.45rem;
    border: 1px solid var(--ren-border);
    border-radius: 999px;
    background: var(--ren-input-bg);
    cursor: pointer;
    touch-action: none;
    user-select: none;
  }

  :global(.devtools-slider-range) {
    height: 100%;
    border-radius: 999px;
    background: var(--ren-accent);
  }

  :global(.devtools-slider-thumb) {
    display: block;
    width: 0.85rem;
    height: 0.85rem;
    border: 1px solid var(--ren-border-strong, var(--ren-border));
    border-radius: 50%;
    background: var(--ren-fg);
    cursor: grab;
  }

  :global(.devtools-slider-thumb:focus-visible) {
    outline: none;
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--ren-focus) 28%, transparent);
  }
</style>
