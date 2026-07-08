// SPDX-License-Identifier: MIT
import { clearPluginModuleCache } from "./loader.js";
import { clearPluginI18n } from "./plugin-i18n.js";
import type { PluginContext, PluginModule } from "./api-types.js";

const active = new Map<string, { ctx: PluginContext; mod: PluginModule }>();

export function getActivePlugin(pluginId: string) {
  return active.get(pluginId);
}

export function hasActivePlugin(pluginId: string): boolean {
  return active.has(pluginId);
}

export function setActivePlugin(
  pluginId: string,
  entry: { ctx: PluginContext; mod: PluginModule },
): void {
  active.set(pluginId, entry);
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
