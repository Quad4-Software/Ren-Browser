<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Copy, KeyRound, Plus, Trash2, Upload, Download, Pencil } from "@lucide/svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
  import { formatBindingError } from "$lib/browser/binding-errors.js";
  import { t } from "$lib/i18n/i18n.svelte";
  import {
    CreateIdentity,
    DeleteIdentity,
    ExportIdentity,
    ImportIdentity,
    ListIdentities,
    RenameIdentity,
    SetActiveIdentity,
  } from "../../../bindings/renbrowser/internal/app/browserservice.js";

  export type IdentityRow = {
    id: string;
    name: string;
    hash: string;
    createdAt: number;
    active: boolean;
  };

  type Props = {
    showTitle?: boolean;
    onChanged?: () => void;
  };

  let { showTitle = true, onChanged }: Props = $props();

  let items = $state<IdentityRow[]>([]);
  let loading = $state(false);
  let busy = $state(false);
  let error = $state("");
  let newName = $state("");
  let importName = $state("");
  let renameTarget = $state<IdentityRow | null>(null);
  let renameValue = $state("");
  let deleteTarget = $state<IdentityRow | null>(null);
  let switchTarget = $state<IdentityRow | null>(null);

  function shortHash(hash: string): string {
    if (hash.length <= 16) {
      return hash;
    }
    return `${hash.slice(0, 8)}...${hash.slice(-8)}`;
  }

  async function load() {
    loading = true;
    error = "";
    try {
      const rows = (await ListIdentities()) as IdentityRow[] | null;
      items = Array.isArray(rows) ? rows : [];
    } catch (err) {
      error = formatBindingError(err, t("identity.loadFailed"));
      items = [];
    } finally {
      loading = false;
    }
  }

  async function createIdentity() {
    const name = newName.trim() || t("identity.defaultName", { n: items.length + 1 });
    busy = true;
    error = "";
    try {
      await CreateIdentity(name);
      newName = "";
      await load();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("identity.createFailed"));
    } finally {
      busy = false;
    }
  }

  async function importIdentity() {
    const name = importName.trim() || t("identity.defaultName", { n: items.length + 1 });
    busy = true;
    error = "";
    try {
      await ImportIdentity(name);
      importName = "";
      await load();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("identity.importFailed"));
    } finally {
      busy = false;
    }
  }

  async function exportIdentity(row: IdentityRow) {
    busy = true;
    error = "";
    try {
      await ExportIdentity(row.id);
    } catch (err) {
      error = formatBindingError(err, t("identity.exportFailed"));
    } finally {
      busy = false;
    }
  }

  async function activateIdentity(row: IdentityRow) {
    if (row.active) {
      return;
    }
    busy = true;
    error = "";
    try {
      await SetActiveIdentity(row.id);
      await load();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("identity.switchFailed"));
    } finally {
      busy = false;
      switchTarget = null;
    }
  }

  async function confirmRename() {
    if (!renameTarget) {
      return;
    }
    const name = renameValue.trim();
    if (!name) {
      return;
    }
    busy = true;
    error = "";
    try {
      await RenameIdentity(renameTarget.id, name);
      renameTarget = null;
      renameValue = "";
      await load();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("identity.renameFailed"));
    } finally {
      busy = false;
    }
  }

  async function confirmDelete() {
    if (!deleteTarget) {
      return;
    }
    busy = true;
    error = "";
    try {
      await DeleteIdentity(deleteTarget.id);
      deleteTarget = null;
      await load();
      onChanged?.();
    } catch (err) {
      error = formatBindingError(err, t("identity.deleteFailed"));
    } finally {
      busy = false;
    }
  }

  async function copyHash(hash: string) {
    try {
      await navigator.clipboard.writeText(hash);
    } catch {
      error = t("identity.copyFailed");
    }
  }

  $effect(() => {
    void load();
  });
</script>

<section class="identity-panel">
  {#if showTitle}
    <div class="header">
      <h3>{t("identity.title")}</h3>
      <p class="hint">{t("identity.hint")}</p>
    </div>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <ul class="list">
    {#if loading && items.length === 0}
      <li class="empty">{t("identity.loading")}</li>
    {:else if items.length === 0}
      <li class="empty">
        <EmptyState
          title={t("identity.noIdentities")}
          description={t("identity.noIdentitiesDescription")}
        >
          <KeyRound size={22} />
        </EmptyState>
      </li>
    {:else}
      {#each items as row (row.id)}
        <li class:active={row.active}>
          <div class="card">
            <div class="body">
              {#if renameTarget?.id === row.id}
                <div class="rename-row">
                  <input class="name-input" type="text" bind:value={renameValue} disabled={busy} />
                  <button
                    type="button"
                    class="ren-btn"
                    disabled={busy || !renameValue.trim()}
                    onclick={() => void confirmRename()}
                  >
                    {t("common.save")}
                  </button>
                  <button
                    type="button"
                    class="ren-btn secondary"
                    disabled={busy}
                    onclick={() => {
                      renameTarget = null;
                      renameValue = "";
                    }}
                  >
                    {t("common.cancel")}
                  </button>
                </div>
              {:else}
                <div class="title-row">
                  <span class="name">{row.name}</span>
                  {#if row.active}
                    <span class="badge">{t("identity.active")}</span>
                  {/if}
                </div>
                <button
                  type="button"
                  class="hash"
                  title={t("identity.copyHash")}
                  onclick={() => void copyHash(row.hash)}
                >
                  <span class="hash-text">{shortHash(row.hash)}</span>
                  <Copy size={14} />
                </button>
              {/if}
            </div>
            <div class="actions">
              {#if !row.active}
                <button
                  type="button"
                  class="ren-btn secondary"
                  disabled={busy}
                  onclick={() => {
                    switchTarget = row;
                  }}
                >
                  {t("identity.use")}
                </button>
              {/if}
              <button
                type="button"
                class="ren-icon-btn"
                aria-label={t("identity.export")}
                title={t("identity.export")}
                disabled={busy}
                onclick={() => void exportIdentity(row)}
              >
                <Download size={16} />
              </button>
              <button
                type="button"
                class="ren-icon-btn"
                aria-label={t("identity.rename")}
                title={t("identity.rename")}
                disabled={busy}
                onclick={() => {
                  renameTarget = row;
                  renameValue = row.name;
                }}
              >
                <Pencil size={16} />
              </button>
              <button
                type="button"
                class="ren-icon-btn danger"
                aria-label={t("identity.delete")}
                title={t("identity.delete")}
                disabled={busy || row.active || items.length <= 1}
                onclick={() => {
                  deleteTarget = row;
                }}
              >
                <Trash2 size={16} />
              </button>
            </div>
          </div>
        </li>
      {/each}
    {/if}
  </ul>

  <div class="create-row">
    <input
      class="name-input"
      type="text"
      placeholder={t("identity.newNamePlaceholder")}
      bind:value={newName}
      disabled={busy}
    />
    <button type="button" class="ren-btn" disabled={busy} onclick={() => void createIdentity()}>
      <Plus size={16} />
      {t("identity.create")}
    </button>
  </div>

  <div class="import-row">
    <input
      class="name-input"
      type="text"
      placeholder={t("identity.importNamePlaceholder")}
      bind:value={importName}
      disabled={busy}
    />
    <button
      type="button"
      class="ren-btn secondary"
      disabled={busy}
      onclick={() => void importIdentity()}
    >
      <Upload size={16} />
      {t("identity.import")}
    </button>
  </div>
</section>

<ConfirmDialog
  open={!!switchTarget}
  title={t("identity.switchConfirmTitle")}
  message={t("identity.switchConfirm", { name: switchTarget?.name ?? "" })}
  confirmLabel={t("identity.use")}
  onConfirm={() => switchTarget && void activateIdentity(switchTarget)}
  onCancel={() => {
    switchTarget = null;
  }}
/>

<ConfirmDialog
  open={!!deleteTarget}
  title={t("identity.deleteTitle")}
  message={t("identity.deleteConfirm", { name: deleteTarget?.name ?? "" })}
  confirmLabel={t("identity.delete")}
  onConfirm={() => void confirmDelete()}
  onCancel={() => {
    deleteTarget = null;
  }}
/>

<style>
  .identity-panel {
    display: grid;
    gap: 0.65rem;
  }

  .header h3 {
    margin: 0;
    color: var(--ren-muted);
    font-size: 0.9rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .hint {
    margin: 0.2rem 0 0;
    color: var(--ren-muted);
    font-size: 0.82rem;
  }

  .error {
    margin: 0;
    color: var(--ren-danger);
    font-size: 0.85rem;
  }

  .list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 0.45rem;
    max-height: 40vh;
    overflow: auto;
  }

  .list li {
    border: 1px solid var(--ren-border);
    border-radius: var(--ren-radius);
    background: var(--ren-surface-raised);
  }

  .list li.active {
    border-color: color-mix(in srgb, var(--ren-accent) 45%, var(--ren-border));
  }

  .card {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.65rem;
    padding: 0.6rem 0.7rem;
  }

  .body {
    min-width: 0;
    display: grid;
    gap: 0.25rem;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    flex-wrap: wrap;
  }

  .name {
    font-weight: 600;
    color: var(--ren-fg);
  }

  .badge {
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--ren-accent);
    border: 1px solid color-mix(in srgb, var(--ren-accent) 40%, var(--ren-border));
    border-radius: 999px;
    padding: 0.1rem 0.45rem;
  }

  .hash {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    border: none;
    background: transparent;
    color: var(--ren-muted);
    font: inherit;
    font-size: 0.8rem;
    padding: 0;
    cursor: pointer;
  }

  .hash:hover {
    color: var(--ren-fg);
  }

  .hash-text {
    font-family: var(--ren-mono, ui-monospace, monospace);
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    flex-shrink: 0;
  }

  .actions .ren-btn {
    padding: 0.35rem 0.55rem;
    font-size: 0.82rem;
  }

  .danger {
    color: var(--ren-danger);
  }

  .empty {
    padding: 0.75rem;
    color: var(--ren-muted);
    font-size: 0.85rem;
  }

  .create-row,
  .import-row {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 0.45rem;
    align-items: center;
  }

  .name-input {
    border: 1px solid var(--ren-border);
    background: var(--ren-input-bg);
    color: var(--ren-fg);
    border-radius: calc(var(--ren-radius) + 2px);
    padding: 0.55rem 0.75rem;
    font: inherit;
    min-width: 0;
  }

  .rename-row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    gap: 0.35rem;
    align-items: center;
  }

  .ren-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    white-space: nowrap;
  }
</style>
