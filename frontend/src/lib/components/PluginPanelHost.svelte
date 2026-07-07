<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { createPluginContext } from "$lib/plugins/api.js";
  import { ensurePluginI18n } from "$lib/plugins/plugin-i18n.js";
  import { getUILocale } from "$lib/i18n/i18n.svelte";
  import { loadPluginModule } from "$lib/plugins/loader.js";
  import { formatBindingError } from "$lib/browser/binding-errors.js";
  import { reportPluginFailure } from "$lib/plugins/plugin-errors.js";
  import type { ActivePageSnapshot } from "$lib/plugins/api-types.js";

  type Props = {
    pluginId: string;
    panelId: string;
    title: string;
    entry: string;
    getCurrentURL?: () => string;
    navigate?: (url: string) => void;
    showToast?: (message: string) => void;
    getActivePage?: () => ActivePageSnapshot;
    updateActivePage?: (patch: Partial<ActivePageSnapshot>) => void;
    networkFetch?: boolean;
    wasmBackend?: boolean;
  };

  let {
    pluginId,
    panelId,
    title,
    entry,
    getCurrentURL = () => "",
    navigate = () => {},
    showToast = () => {},
    getActivePage = () => ({ url: "", path: "", contentType: "", html: "", raw: "" }),
    updateActivePage = () => {},
    networkFetch = false,
    wasmBackend = false,
  }: Props = $props();

  let error = $state("");
  let panelEl: HTMLElement | undefined = $state();
  const locale = $derived(getUILocale());

  $effect(() => {
    const activeLocale = locale;
    const mountEl = panelEl;
    if (!mountEl) {
      return;
    }
    let cancelled = false;
    void (async () => {
      try {
        error = "";
        const mod = await loadPluginModule(pluginId, entry);
        if (cancelled) {
          return;
        }
        const ctx = createPluginContext(pluginId, {
          getCurrentURL,
          navigate,
          showToast,
          getActivePage,
          updateActivePage,
          networkFetch,
          wasmBackend,
          i18n: await ensurePluginI18n(pluginId, activeLocale),
        });
        if (mod.activate) {
          await mod.activate(ctx);
        }
        if (!cancelled && mod.mount && panelEl) {
          mountEl.replaceChildren();
          await mod.mount(mountEl);
        }
      } catch (err) {
        if (!cancelled) {
          error = formatBindingError(err, "Extension panel failed to load");
          await reportPluginFailure(pluginId, "panel", err);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  });
</script>

<div class="plugin-panel" role="region" aria-label={title} data-panel-id={panelId}>
  {#if error}
    <p class="error">{error}</p>
  {/if}
  <div class="body" bind:this={panelEl}></div>
</div>

<style>
  .plugin-panel {
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-height: 0;
    height: 100%;
  }

  .body {
    min-height: 0;
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: auto;
  }

  .body :global(> :not(style)) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .error {
    margin: 0;
    color: #f87171;
  }
</style>
