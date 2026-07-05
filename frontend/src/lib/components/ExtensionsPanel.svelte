<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import Toggle from "$lib/components/Toggle.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import {
    disablePlugin,
    enablePlugin,
    installFromDir,
    installFromZip,
    listPlugins,
    uninstallPlugin,
  } from "$lib/plugins/api.js";
  import { t } from "$lib/i18n/i18n.svelte";

  type PluginRow = {
    id: string;
    name: string;
    version: string;
    description?: string;
    enabled: boolean;
    error?: string;
    permissions?: string[];
  };

  type Props = {
    pluginsDir: string;
    onChanged?: () => void;
  };

  let { pluginsDir, onChanged }: Props = $props();

  let plugins = $state<PluginRow[]>([]);
  let loading = $state(false);
  let error = $state("");
  let zipPath = $state("");
  let dirPath = $state("");

  async function refresh() {
    loading = true;
    error = "";
    try {
      const rows = (await listPlugins()) as PluginRow[];
      plugins = rows ?? [];
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      loading = false;
    }
  }

  async function toggleEnabled(id: string, enabled: boolean) {
    if (enabled) {
      await enablePlugin(id);
    } else {
      await disablePlugin(id);
    }
    await refresh();
    onChanged?.();
  }

  async function remove(id: string) {
    if (!confirm(t("extensions.uninstallConfirm", { id }))) {
      return;
    }
    await uninstallPlugin(id);
    await refresh();
    onChanged?.();
  }

  async function installZip() {
    if (!zipPath.trim()) {
      return;
    }
    await installFromZip(zipPath.trim());
    zipPath = "";
    await refresh();
    onChanged?.();
  }

  async function installDir() {
    if (!dirPath.trim()) {
      return;
    }
    await installFromDir(dirPath.trim());
    dirPath = "";
    await refresh();
    onChanged?.();
  }

  $effect(() => {
    void refresh();
  });
</script>

<section class="extensions">
  <header>
    <h3>{t("extensions.title")}</h3>
    <button type="button" onclick={() => void refresh()} disabled={loading}>{t("common.refresh")}</button>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <p class="hint">{t("extensions.pluginsDir", { path: pluginsDir || "—" })}</p>

  <div class="install">
    <label>
      {t("extensions.installZip")}
      <input
        class="ren-input"
        bind:value={zipPath}
        placeholder={t("extensions.installZipPlaceholder")}
      />
    </label>
    <button type="button" onclick={() => void installZip()}>{t("extensions.installZipButton")}</button>
    <label>
      {t("extensions.installFolder")}
      <input
        class="ren-input"
        bind:value={dirPath}
        placeholder={t("extensions.installFolderPlaceholder")}
      />
    </label>
    <button type="button" onclick={() => void installDir()}>{t("extensions.installFolderButton")}</button>
  </div>

  {#if loading}
    <p class="muted">{t("extensions.loading")}</p>
  {:else if plugins.length === 0}
    <EmptyState
      title={t("extensions.noExtensions")}
      description={t("extensions.noExtensionsDescription")}
    />
  {:else}
    <ul class="list">
      {#each plugins as plugin (plugin.id)}
        <li>
          <div class="meta">
            <strong>{plugin.name}</strong>
            <span class="version">v{plugin.version}</span>
            <span class="id">{plugin.id}</span>
            {#if plugin.description}
              <p>{plugin.description}</p>
            {/if}
            {#if plugin.permissions?.length}
              <p class="perms">{t("common.permissions", { list: plugin.permissions.join(", ") })}</p>
            {/if}
            {#if plugin.error}
              <p class="error">{plugin.error}</p>
            {/if}
          </div>
          <div class="actions">
            <Toggle
              label={t("common.enabled")}
              checked={plugin.enabled}
              onchange={(value) => void toggleEnabled(plugin.id, value)}
            />
            <button type="button" class="danger" onclick={() => void remove(plugin.id)}
              >{t("extensions.uninstall")}</button
            >
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .extensions {
    display: grid;
    gap: 0.75rem;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  h3 {
    margin: 0;
    font-size: 1rem;
  }

  .hint,
  .muted {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.85rem;
  }

  code {
    font-family: var(--ren-mono, monospace);
    font-size: 0.8rem;
  }

  .install {
    display: grid;
    gap: 0.5rem;
  }

  label {
    display: grid;
    gap: 0.25rem;
    font-size: 0.85rem;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.75rem;
  }

  li {
    border: 1px solid var(--ren-border);
    border-radius: 10px;
    padding: 0.75rem;
    display: grid;
    gap: 0.5rem;
  }

  .meta p {
    margin: 0.25rem 0 0;
    color: var(--ren-muted);
    font-size: 0.85rem;
  }

  .version,
  .id {
    margin-left: 0.5rem;
    color: var(--ren-muted);
    font-size: 0.8rem;
  }

  .perms {
    font-family: var(--ren-mono, monospace);
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .error {
    color: #f87171;
    margin: 0;
  }

  .danger {
    color: #f87171;
  }
</style>
