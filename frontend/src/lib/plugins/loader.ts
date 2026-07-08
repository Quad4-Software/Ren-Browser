// SPDX-License-Identifier: MIT
import type { PluginModule } from "./api-types.js";

const cache = new Map<string, PluginModule>();

export async function loadPluginModule(pluginId: string, entry: string): Promise<PluginModule> {
  const key = `${pluginId}:${entry}`;
  const cached = cache.get(key);
  if (cached) {
    return cached;
  }
  const url = `/_plugins/${encodeURIComponent(pluginId)}/${entry.split("/").map(encodeURIComponent).join("/")}`;
  // eslint-disable-next-line no-unsanitized/method
  const mod = (await import(/* @vite-ignore */ url)) as PluginModule;
  cache.set(key, mod);
  return mod;
}

export function clearPluginModuleCache(pluginId?: string) {
  if (!pluginId) {
    cache.clear();
    return;
  }
  for (const key of cache.keys()) {
    if (key.startsWith(`${pluginId}:`)) {
      cache.delete(key);
    }
  }
}
