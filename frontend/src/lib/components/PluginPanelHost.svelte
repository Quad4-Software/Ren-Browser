<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { onMount } from "svelte";
  import { createPluginContext } from "$lib/plugins/api.js";
  import { loadPluginModule } from "$lib/plugins/loader.js";

  type Props = {
    pluginId: string;
    panelId: string;
    title: string;
    entry: string;
    getCurrentURL?: () => string;
    navigate?: (url: string) => void;
    showToast?: (message: string) => void;
  };

  let {
    pluginId,
    panelId,
    title,
    entry,
    getCurrentURL = () => "",
    navigate = () => {},
    showToast = () => {},
  }: Props = $props();

  let error = $state("");

  onMount(() => {
    let cancelled = false;
    void (async () => {
      try {
        const mod = await loadPluginModule(pluginId, entry);
        if (cancelled || !mod.activate) {
          return;
        }
        const ctx = createPluginContext(pluginId, { getCurrentURL, navigate, showToast });
        await mod.activate(ctx);
      } catch (err) {
        error = err instanceof Error ? err.message : String(err);
      }
    })();
    return () => {
      cancelled = true;
    };
  });
</script>

<div class="plugin-panel">
  <header>
    <h3>{title}</h3>
    <span class="id">{pluginId}:{panelId}</span>
  </header>
  {#if error}
    <p class="error">{error}</p>
  {:else}
    <p class="muted">Extension panel loaded from <code>{entry}</code></p>
  {/if}
</div>

<style>
  .plugin-panel {
    padding: 0.75rem;
    display: grid;
    gap: 0.5rem;
  }

  header {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
  }

  h3 {
    margin: 0;
    font-size: 1rem;
  }

  .id {
    color: var(--ren-muted);
    font-size: 0.75rem;
  }

  .muted {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.85rem;
  }

  .error {
    margin: 0;
    color: #f87171;
  }

  code {
    font-family: var(--ren-mono, monospace);
  }
</style>
