<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Activity, FileCode, Terminal } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import PluginPanelHost from "$lib/components/PluginPanelHost.svelte";
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
    error?: string;
  };

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
    return String(value);
  }
</script>

<section class="devtools">
  <header>
    <div class="tabs">
      <button class:active={tab === "console"} onclick={() => (tab = "console")}>Console</button>
      <button class:active={tab === "network"} onclick={() => (tab = "network")}>Network</button>
      <button class:active={tab === "raw"} onclick={() => (tab = "raw")}>Raw</button>
      {#each pluginTabs as pluginTab (pluginTab.pluginId + ":" + pluginTab.id)}
        <button
          class:active={tab === `plugin:${pluginTab.pluginId}:${pluginTab.id}`}
          onclick={() => (tab = `plugin:${pluginTab.pluginId}:${pluginTab.id}`)}
        >
          {pluginTab.title}
        </button>
      {/each}
    </div>
    <div class="actions">
      <label>
        Log
        <input
          type="range"
          min="1"
          max="7"
          value={logLevel}
          oninput={(event) => onLogLevel(Number((event.currentTarget as HTMLInputElement).value))}
        />
      </label>
      <button onclick={onExport}>Export</button>
      <button onclick={onClear}>Clear</button>
    </div>
  </header>

  <div class="page-info">
    <span>{contentType || "unknown"}</span>
    {#if micronRendererBadge}
      <span class="renderer-badge">{micronRendererBadge}</span>
    {/if}
    {#if hops >= 0}
      <span>{hops} hop{hops === 1 ? "" : "s"}</span>
    {/if}
    {#if durationMs > 0}
      <span>{durationMs} ms</span>
    {/if}
    {#if fromCache}
      <span>cached {cachedAt > 0 ? new Date(cachedAt).toLocaleString() : ""}</span>
    {/if}
  </div>

  {#if tab === "console"}
    <div class="panel logs">
      <input
        class="search ren-input"
        type="search"
        bind:value={logQuery}
        placeholder="Search logs..."
        spellcheck="false"
        autocomplete="off"
      />
      {#if logs.length === 0}
        <EmptyState
          title="Console is empty"
          description="Application logs from Reticulum and the browser shell will show up here."
        >
          <Terminal size={22} />
        </EmptyState>
      {:else if filteredLogs.length === 0}
        <EmptyState
          title="No matching logs"
          description={'Nothing matches "' + logQuery.trim() + '".'}
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
    </div>
  {:else if tab === "network"}
    <div class="panel network">
      {#if network.length === 0}
        <EmptyState
          title="No network activity"
          description="Mesh page requests and their timing will be recorded here."
        >
          <Activity size={22} />
        </EmptyState>
      {:else}
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Path</th>
              <th>Hops</th>
              <th>ms</th>
              <th>Bytes</th>
              <th>Cache</th>
              <th>Error</th>
            </tr>
          </thead>
          <tbody>
            {#each network as row (row.time + row.url)}
              <tr>
                <td>{formatTime(row.time)}</td>
                <td>{row.path}</td>
                <td>{formatHops(row.hops ?? -1)}</td>
                <td>{row.durationMs}</td>
                <td>{row.bytes}</td>
                <td>{row.fromCache ? "yes" : "no"}</td>
                <td>{row.error ?? ""}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
    </div>
  {:else if tab === "raw"}
    <div class="panel raw">
      <div class="raw-actions">
        <button class:active={rawMode === "text"} onclick={() => (rawMode = "text")}>Text</button>
        <button class:active={rawMode === "hex"} onclick={() => (rawMode = "hex")}>Hex</button>
      </div>
      {#if raw.trim().length === 0}
        <EmptyState
          title="No raw page data"
          description="Open a mesh page to inspect its source bytes here."
        >
          <FileCode size={22} />
        </EmptyState>
      {:else}
        <pre>{rawMode === "text" ? raw : toHex(raw)}</pre>
      {/if}
    </div>
  {:else}
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
  {/if}
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

  .tabs,
  .actions {
    display: flex;
    gap: 0.35rem;
    align-items: center;
  }

  .tabs button,
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

  .tabs button:hover,
  .actions button:hover,
  .raw-actions button:hover {
    background: var(--ren-tab-hover);
    color: var(--ren-fg);
  }

  .tabs button.active,
  .raw-actions button.active {
    border-color: var(--ren-accent);
    background: var(--ren-accent);
    color: #fff;
  }

  .panel {
    flex: 1;
    overflow: auto;
    padding: 0.75rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.82rem;
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
</style>
