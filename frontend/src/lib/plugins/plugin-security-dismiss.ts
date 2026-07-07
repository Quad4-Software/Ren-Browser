// SPDX-License-Identifier: MIT
const STORAGE_PREFIX = "renbrowser:dismiss-plugin-finding:v1";

function storageKey(pluginId: string, findingId: string): string {
  return `${STORAGE_PREFIX}:${pluginId}:${findingId}`;
}

export function isPluginFindingDismissed(pluginId: string, findingId: string): boolean {
  try {
    return localStorage.getItem(storageKey(pluginId, findingId)) === "1";
  } catch {
    return false;
  }
}

export function dismissPluginFinding(pluginId: string, findingId: string): void {
  try {
    localStorage.setItem(storageKey(pluginId, findingId), "1");
  } catch {
    // ignore storage failures
  }
}

export function clearDismissedPluginFindings(pluginId?: string): void {
  try {
    if (!pluginId) {
      for (let i = localStorage.length - 1; i >= 0; i -= 1) {
        const key = localStorage.key(i);
        if (key?.startsWith(`${STORAGE_PREFIX}:`)) {
          localStorage.removeItem(key);
        }
      }
      return;
    }
    const prefix = `${STORAGE_PREFIX}:${pluginId}:`;
    for (let i = localStorage.length - 1; i >= 0; i -= 1) {
      const key = localStorage.key(i);
      if (key?.startsWith(prefix)) {
        localStorage.removeItem(key);
      }
    }
  } catch {
    // ignore storage failures
  }
}
