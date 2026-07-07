// SPDX-License-Identifier: MIT
import { formatBindingError } from "$lib/browser/binding-errors.js";
import { getUILocale } from "$lib/i18n/i18n.svelte";
import { createPluginContext, listPlugins } from "./api.js";
import { loadPluginModule, clearPluginModuleCache } from "./loader.js";
import { reportPluginFailure } from "./plugin-errors.js";
import { clearPluginI18n, ensurePluginI18n, preloadPluginI18n } from "./plugin-i18n.js";
import type { PluginContext, PluginModule } from "./api-types.js";
import { getContributionsSnapshot } from "./registry.js";

type LifecycleOpts = {
  getCurrentURL: () => string;
  navigate: (url: string) => void;
  showToast: (message: string) => void;
  getActivePage: () => import("./api-types.js").ActivePageSnapshot;
  updateActivePage: (patch: Partial<import("./api-types.js").ActivePageSnapshot>) => void;
  networkFetch?: boolean;
  wasmBackend?: boolean;
  onPluginError?: (pluginId: string, message: string) => void;
};

const active = new Map<string, { ctx: PluginContext; mod: PluginModule }>();

export async function activateAllPlugins(opts: LifecycleOpts) {
  const snapshot = getContributionsSnapshot();
  const pluginIds = new Set(snapshot.panels.map((p) => p.pluginId));
  for (const cmd of snapshot.commands) {
    pluginIds.add(cmd.pluginId);
  }
  for (const scheme of snapshot.urlSchemes) {
    pluginIds.add(scheme.pluginId);
  }
  const rows = await listPlugins();
  const grantedById = new Map((rows ?? []).map((row) => [row.id, row.grantedPermissions ?? []]));
  await preloadPluginI18n(pluginIds, getUILocale());
  for (const pluginId of pluginIds) {
    const granted = grantedById.get(pluginId) ?? [];
    await activatePlugin(pluginId, "main.js", {
      ...opts,
      networkFetch: granted.includes("network.fetch"),
    });
  }
}

export async function activatePlugin(pluginId: string, entry: string, opts: LifecycleOpts) {
  if (active.has(pluginId)) {
    return;
  }
  let mod: PluginModule;
  try {
    mod = await loadPluginModule(pluginId, entry);
  } catch (err) {
    opts.onPluginError?.(pluginId, formatBindingError(err, "Extension failed"));
    await reportPluginFailure(pluginId, "load", err);
    return;
  }
  const ctx = createPluginContext(pluginId, {
    ...opts,
    i18n: await ensurePluginI18n(pluginId, getUILocale()),
  });
  try {
    if (mod.activate) {
      await mod.activate(ctx);
    }
  } catch (err) {
    await reportPluginFailure(pluginId, "activate", err);
    return;
  }
  active.set(pluginId, { ctx, mod });
}

export async function deactivatePlugin(pluginId: string) {
  const item = active.get(pluginId);
  if (!item) {
    return;
  }
  try {
    if (item.mod.deactivate) {
      await item.mod.deactivate();
    }
  } catch {
    // Ignore teardown errors while unloading a broken plugin.
  }
  active.delete(pluginId);
  clearPluginModuleCache(pluginId);
  clearPluginI18n(pluginId);
}

export async function deactivateAllPlugins() {
  for (const pluginId of [...active.keys()]) {
    await deactivatePlugin(pluginId);
  }
}

export async function handlePluginScheme(pluginId: string, handler: string, url: string) {
  const item = active.get(pluginId);
  if (!item?.mod.handleScheme) {
    let mod: PluginModule;
    try {
      mod = await loadPluginModule(pluginId, "main.js");
    } catch (err) {
      await reportPluginFailure(pluginId, "scheme", err);
      return null;
    }
    if (!mod.handleScheme) {
      return null;
    }
    try {
      return await mod.handleScheme(url);
    } catch (err) {
      await reportPluginFailure(pluginId, "scheme", err);
      return null;
    }
  }
  try {
    if (handler && item.mod.handleScheme) {
      return await item.mod.handleScheme(url);
    }
    return (await item.mod.handleScheme?.(url)) ?? null;
  } catch (err) {
    await reportPluginFailure(pluginId, "scheme", err);
    return null;
  }
}
