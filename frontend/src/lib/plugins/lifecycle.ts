// SPDX-License-Identifier: MIT
import { createPluginContext } from "./api.js";
import { loadPluginModule, clearPluginModuleCache } from "./loader.js";
import type { PluginContext, PluginModule } from "./api-types.js";
import { getContributionsSnapshot } from "./registry.js";

type LifecycleOpts = {
  getCurrentURL: () => string;
  navigate: (url: string) => void;
  showToast: (message: string) => void;
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
  for (const pluginId of pluginIds) {
    await activatePlugin(pluginId, "main.js", opts);
  }
}

export async function activatePlugin(pluginId: string, entry: string, opts: LifecycleOpts) {
  if (active.has(pluginId)) {
    return;
  }
  let mod: PluginModule;
  try {
    mod = await loadPluginModule(pluginId, entry);
  } catch {
    return;
  }
  const ctx = createPluginContext(pluginId, opts);
  if (mod.activate) {
    await mod.activate(ctx);
  }
  active.set(pluginId, { ctx, mod });
}

export async function deactivatePlugin(pluginId: string) {
  const item = active.get(pluginId);
  if (!item) {
    return;
  }
  if (item.mod.deactivate) {
    await item.mod.deactivate();
  }
  active.delete(pluginId);
  clearPluginModuleCache(pluginId);
}

export async function deactivateAllPlugins() {
  for (const pluginId of [...active.keys()]) {
    await deactivatePlugin(pluginId);
  }
}

export async function handlePluginScheme(pluginId: string, handler: string, url: string) {
  const item = active.get(pluginId);
  if (!item?.mod.handleScheme) {
    const mod = await loadPluginModule(pluginId, "main.js");
    if (!mod.handleScheme) {
      return null;
    }
    return mod.handleScheme(url);
  }
  if (handler && item.mod.handleScheme) {
    return item.mod.handleScheme(url);
  }
  return item.mod.handleScheme?.(url) ?? null;
}
