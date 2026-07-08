// SPDX-License-Identifier: MIT
import { formatBindingError } from "$lib/browser/binding-errors.js";
import { getUILocale } from "$lib/i18n/i18n.svelte";
import { createPluginContext, listPlugins } from "./api.js";
import { loadPluginModule } from "./loader.js";
import { reportPluginFailure } from "./plugin-errors.js";
import { getActivePlugin, hasActivePlugin, setActivePlugin } from "./plugin-runtime.js";
import { ensurePluginI18n, preloadPluginI18n } from "./plugin-i18n.js";
import type { PluginModule } from "./api-types.js";
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
  if (hasActivePlugin(pluginId)) {
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
  setActivePlugin(pluginId, { ctx, mod });
}

export { deactivateAllPlugins, deactivatePlugin } from "./plugin-runtime.js";

export async function handlePluginScheme(pluginId: string, handler: string, url: string) {
  const item = getActivePlugin(pluginId);
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
