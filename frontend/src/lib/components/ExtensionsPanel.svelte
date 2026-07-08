<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { ChevronDown, Cpu, FolderOpen, Package, Plus, X } from "@lucide/svelte";
  import { System } from "@wailsio/runtime";
  import Toggle from "$lib/components/Toggle.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
  import PluginNetworkInstallDialog, {
    type PluginInstallPreview,
    type PluginInstallChoices,
  } from "$lib/components/PluginNetworkInstallDialog.svelte";
  import PluginSignatureBadge from "$lib/components/PluginSignatureBadge.svelte";
  import type { PluginSummary } from "../../../bindings/renbrowser/internal/app/models.js";
  import {
    PickPluginDir,
    PickPluginWasm,
    PickPluginZip,
  } from "../../../bindings/renbrowser/internal/app/browserservice.js";
  import {
    disablePlugin,
    enablePlugin,
    installFromDir,
    installFromWasm,
    installFromZip,
    listPlugins,
    previewInstallFromDir,
    previewInstallFromWasm,
    previewInstallFromZip,
    trustPublisher,
    uninstallPlugin,
  } from "$lib/plugins/api.js";
  import {
    isPluginNetworkInstallWarningSkipped,
    setPluginNetworkInstallWarningSkipped,
  } from "$lib/plugins/plugin-install-warning.js";
  import {
    dismissPluginFinding,
    isPluginFindingDismissed,
    clearDismissedPluginFindings,
  } from "$lib/plugins/plugin-security-dismiss.js";
  import { formatBindingError } from "$lib/browser/binding-errors.js";
  import { formatBytes } from "$lib/browser/download-progress";
  import { t } from "$lib/i18n/i18n.svelte";

  type PluginRow = PluginSummary;

  type Props = {
    pluginsDir: string;
    onChanged?: () => void;
    showTitle?: boolean;
  };

  let { pluginsDir, onChanged, showTitle = true }: Props = $props();

  const desktop = System.IsDesktop();

  let plugins = $state<PluginRow[]>([]);
  let loading = $state(false);
  let picking = $state(false);
  let installMenuOpen = $state(false);
  let error = $state("");
  let uninstallTarget = $state<PluginRow | null>(null);
  let uninstalling = $state(false);
  let networkInstallTarget = $state<{
    path: string;
    method: "zip" | "dir" | "wasm";
    preview: PluginInstallPreview;
  } | null>(null);
  let networkInstalling = $state(false);
  let dismissedFindings = $state<Record<string, true>>({});

  type InstallMethod = "zip" | "dir" | "wasm";

  function findingKey(pluginId: string, findingId: string): string {
    return `${pluginId}:${findingId}`;
  }

  function visibleFindings(plugin: PluginRow) {
    return (plugin.security?.findings ?? []).filter((finding) => {
      if (!finding.id) {
        return true;
      }
      const key = findingKey(plugin.id, finding.id);
      return !dismissedFindings[key] && !isPluginFindingDismissed(plugin.id, finding.id);
    });
  }

  function dismissFinding(pluginId: string, findingId: string) {
    dismissPluginFinding(pluginId, findingId);
    dismissedFindings = { ...dismissedFindings, [findingKey(pluginId, findingId)]: true };
  }

  async function previewInstall(
    method: InstallMethod,
    path: string,
  ): Promise<PluginInstallPreview> {
    switch (method) {
      case "zip":
        return previewInstallFromZip(path);
      case "dir":
        return previewInstallFromDir(path);
      case "wasm":
        return previewInstallFromWasm(path);
    }
  }

  async function runInstall(method: InstallMethod, path: string, granted: string[] = []) {
    switch (method) {
      case "zip":
        await installFromZip(path, granted);
        return;
      case "dir":
        await installFromDir(path, granted);
        return;
      case "wasm":
        await installFromWasm(path, granted);
        return;
    }
  }

  function needsInstallPreview(preview: PluginInstallPreview): boolean {
    if (isPluginNetworkInstallWarningSkipped()) {
      return false;
    }
    if (preview.requiresNetworkFetch) {
      return true;
    }
    if ((preview.permissions?.length ?? 0) > 0) {
      return true;
    }
    if (preview.signature?.present) {
      return true;
    }
    return false;
  }

  async function beginInstall(method: InstallMethod, path: string) {
    const preview = await previewInstall(method, path);
    if (needsInstallPreview(preview)) {
      networkInstallTarget = { path, method, preview };
      return;
    }
    await runInstall(method, path);
    await refresh();
    onChanged?.();
  }

  function cancelNetworkInstall() {
    if (networkInstalling) {
      return;
    }
    networkInstallTarget = null;
  }

  async function confirmNetworkInstall(choices: PluginInstallChoices) {
    const target = networkInstallTarget;
    if (!target || networkInstalling) {
      return;
    }
    networkInstalling = true;
    error = "";
    try {
      if (choices.dontShowAgain) {
        setPluginNetworkInstallWarningSkipped(true);
      }
      if (
        choices.trustPublisher &&
        target.preview.signature.valid &&
        target.preview.signature.signer
      ) {
        await trustPublisher(
          target.preview.signature.signer,
          target.preview.name || target.preview.signature.signerName || "",
        );
      }
      await runInstall(target.method, target.path, choices.grantedPermissions);
      networkInstallTarget = null;
      await refresh();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("extensions.installFailed"));
    } finally {
      networkInstalling = false;
    }
  }

  async function refresh() {
    loading = true;
    error = "";
    try {
      const rows = (await listPlugins()) as PluginRow[];
      plugins = rows ?? [];
    } catch (err) {
      error = formatBindingError(err, t("extensions.loadFailed"));
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

  function requestRemove(plugin: PluginRow) {
    uninstallTarget = plugin;
  }

  function cancelUninstall() {
    if (uninstalling) {
      return;
    }
    uninstallTarget = null;
  }

  async function confirmUninstall() {
    const plugin = uninstallTarget;
    if (!plugin || uninstalling) {
      return;
    }
    uninstalling = true;
    error = "";
    try {
      await uninstallPlugin(plugin.id);
      clearDismissedPluginFindings(plugin.id);
      uninstallTarget = null;
      await refresh();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("extensions.uninstallFailed"));
    } finally {
      uninstalling = false;
    }
  }

  async function pickAndInstallZip() {
    if (!desktop) {
      error = t("extensions.pickerUnavailable");
      return;
    }
    picking = true;
    installMenuOpen = false;
    error = "";
    try {
      const path = await PickPluginZip();
      if (!path?.trim()) {
        return;
      }
      await beginInstall("zip", path.trim());
    } catch (err) {
      error = formatBindingError(err, t("extensions.installFailed"));
    } finally {
      picking = false;
    }
  }

  async function pickAndInstallDir() {
    if (!desktop) {
      error = t("extensions.pickerUnavailable");
      return;
    }
    picking = true;
    installMenuOpen = false;
    error = "";
    try {
      const path = await PickPluginDir();
      if (!path?.trim()) {
        return;
      }
      await beginInstall("dir", path.trim());
    } catch (err) {
      error = formatBindingError(err, t("extensions.installFailed"));
    } finally {
      picking = false;
    }
  }

  async function pickAndInstallWasm() {
    if (!desktop) {
      error = t("extensions.pickerUnavailable");
      return;
    }
    picking = true;
    installMenuOpen = false;
    error = "";
    try {
      const path = await PickPluginWasm();
      if (!path?.trim()) {
        return;
      }
      await beginInstall("wasm", path.trim());
    } catch (err) {
      error = formatBindingError(err, t("extensions.installFailed"));
    } finally {
      picking = false;
    }
  }

  $effect(() => {
    void refresh();
  });
</script>

<section class="extensions">
  <header class:compact={!showTitle}>
    {#if showTitle}
      <h3>{t("extensions.title")}</h3>
    {/if}
    <button type="button" onclick={() => void refresh()} disabled={loading}
      >{t("common.refresh")}</button
    >
  </header>

  {#if error}
    <p class="error panel-error" role="alert">{error}</p>
  {/if}

  <p class="hint">{t("extensions.pluginsDir", { path: pluginsDir || "—" })}</p>

  <div class="install">
    <details class="install-menu" bind:open={installMenuOpen}>
      <summary
        class="pick-btn"
        aria-label={t("extensions.installButton")}
        class:disabled={loading || picking || !desktop}
      >
        <Plus size={16} />
        <span>{t("extensions.installButton")}</span>
        <span class="chevron"><ChevronDown size={14} /></span>
      </summary>
      <div class="install-options" role="menu">
        <button
          type="button"
          role="menuitem"
          disabled={loading || picking || !desktop}
          onclick={() => void pickAndInstallZip()}
        >
          <Package size={16} />
          <span>{t("extensions.installZipButton")}</span>
        </button>
        <button
          type="button"
          role="menuitem"
          disabled={loading || picking || !desktop}
          onclick={() => void pickAndInstallDir()}
        >
          <FolderOpen size={16} />
          <span>{t("extensions.installFolderButton")}</span>
        </button>
        <button
          type="button"
          role="menuitem"
          disabled={loading || picking || !desktop}
          onclick={() => void pickAndInstallWasm()}
        >
          <Cpu size={16} />
          <span>{t("extensions.installWasmButton")}</span>
        </button>
      </div>
    </details>
    {#if !desktop}
      <p class="muted">{t("extensions.pickerUnavailable")}</p>
    {/if}
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
            <div class="title-row">
              <strong>{plugin.name}</strong>
              <PluginSignatureBadge
                signature={plugin.signature}
                tampered={plugin.tampered}
                compact
              />
              <span class="version">v{plugin.version}</span>
              <span class="size">{formatBytes(plugin.sizeBytes ?? 0)}</span>
            </div>
            <span class="id">{plugin.id}</span>
            {#if plugin.description}
              <p>{plugin.description}</p>
            {/if}
            {#if plugin.error}
              <p class="error" role="alert">{formatBindingError(plugin.error)}</p>
            {/if}
            {#if visibleFindings(plugin).length}
              <ul class="security-findings">
                {#each visibleFindings(plugin) as finding (finding.id)}
                  <li data-severity={finding.severity}>
                    <span class="finding-text">{finding.message}</span>
                    <button
                      type="button"
                      class="finding-dismiss"
                      aria-label={t("extensions.dismissWarning")}
                      onclick={() => dismissFinding(plugin.id, finding.id)}
                    >
                      <X size={14} />
                    </button>
                  </li>
                {/each}
              </ul>
            {/if}
          </div>
          <div class="actions">
            <Toggle
              label={t("common.enabled")}
              checked={plugin.enabled}
              onchange={(value) => void toggleEnabled(plugin.id, value)}
            />
            <button
              type="button"
              class="danger"
              disabled={uninstalling}
              onclick={() => requestRemove(plugin)}>{t("extensions.uninstall")}</button
            >
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<PluginNetworkInstallDialog
  open={networkInstallTarget !== null}
  preview={networkInstallTarget?.preview ?? null}
  confirming={networkInstalling}
  onConfirm={(choices) => void confirmNetworkInstall(choices)}
  onCancel={cancelNetworkInstall}
/>

<ConfirmDialog
  open={uninstallTarget !== null}
  title={t("extensions.uninstallTitle")}
  message={uninstallTarget
    ? t("extensions.uninstallMessage", {
        name: uninstallTarget.name,
        id: uninstallTarget.id,
      })
    : ""}
  confirmLabel={t("extensions.uninstall")}
  cancelLabel={t("common.cancel")}
  onConfirm={() => void confirmUninstall()}
  onCancel={cancelUninstall}
/>

<style>
  .extensions {
    display: grid;
    gap: 0.75rem;
    min-width: 0;
    max-width: 100%;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  header.compact {
    justify-content: flex-end;
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
    word-break: break-all;
    overflow-wrap: anywhere;
  }

  .install {
    display: grid;
    gap: 0.5rem;
  }

  .install-menu {
    position: relative;
  }

  .install-menu > summary {
    list-style: none;
  }

  .install-menu > summary::-webkit-details-marker {
    display: none;
  }

  .install-menu > summary.disabled {
    opacity: 0.55;
    cursor: not-allowed;
    pointer-events: none;
  }

  .chevron {
    margin-left: auto;
    opacity: 0.75;
  }

  .install-options {
    margin-top: 0.35rem;
    display: grid;
    gap: 0.35rem;
    border: 1px solid var(--ren-border);
    border-radius: calc(var(--ren-radius) + 2px);
    background: var(--ren-input-bg);
    padding: 0.35rem;
  }

  .install-options button {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    border: 0;
    background: transparent;
    color: var(--ren-fg);
    border-radius: var(--ren-radius);
    padding: 0.5rem 0.6rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    text-align: left;
  }

  .install-options button:hover:not(:disabled) {
    background: var(--ren-tab-hover);
  }

  .install-options button:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .pick-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
    font-size: 0.88rem;
    cursor: pointer;
    text-align: left;
  }

  .pick-btn:hover:not(:disabled) {
    background: var(--ren-tab-hover);
    border-color: var(--ren-border-strong);
  }

  .pick-btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
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

  .title-row {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.35rem 0.5rem;
    min-width: 0;
  }

  .title-row strong {
    min-width: 0;
  }

  .version,
  .size,
  .id {
    color: var(--ren-muted);
    font-size: 0.8rem;
    overflow-wrap: anywhere;
    word-break: break-all;
  }

  .id {
    display: block;
    margin-top: 0.15rem;
  }

  .version,
  .size {
    word-break: normal;
  }

  .security-findings {
    margin: 0.25rem 0 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 0.25rem;
    font-size: 0.8rem;
    color: var(--ren-muted);
  }

  .security-findings li {
    display: flex;
    align-items: flex-start;
    gap: 0.35rem;
    justify-content: space-between;
  }

  .finding-text {
    flex: 1;
    min-width: 0;
  }

  .finding-dismiss {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border: 0;
    background: transparent;
    color: inherit;
    opacity: 0.75;
    padding: 0.1rem;
    border-radius: var(--ren-radius);
    cursor: pointer;
  }

  .finding-dismiss:hover {
    opacity: 1;
    background: color-mix(in srgb, currentColor 12%, transparent);
  }

  .security-findings li[data-severity="high"] {
    color: #ff9b9b;
  }

  .security-findings li[data-severity="warn"] {
    color: #f0c674;
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
    min-width: 0;
  }

  @media (max-width: 640px) {
    .actions {
      flex-direction: column;
      align-items: stretch;
    }

    .actions button {
      width: 100%;
    }
  }

  .error {
    color: var(--ren-danger, #e5484d);
    margin: 0;
    font-size: 0.85rem;
    overflow-wrap: anywhere;
    word-break: break-word;
  }

  .meta .error {
    margin-top: 0.35rem;
  }

  .panel-error {
    padding: 0.55rem 0.65rem;
    border: 1px solid color-mix(in srgb, var(--ren-danger, #e5484d) 45%, var(--ren-border));
    border-radius: 8px;
    background: color-mix(in srgb, var(--ren-danger, #e5484d) 8%, transparent);
  }

  .danger {
    color: #f87171;
  }
</style>
