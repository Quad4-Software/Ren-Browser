// SPDX-License-Identifier: MIT
import type { ContributionsSnapshot } from "./api-types.js";

let snapshot: ContributionsSnapshot = {
  panels: [],
  commands: [],
  devtools: [],
  urlSchemes: [],
};

export function setContributions(next: ContributionsSnapshot) {
  snapshot = {
    panels: [...next.panels],
    commands: [...next.commands],
    devtools: [...next.devtools],
    urlSchemes: [...next.urlSchemes],
  };
}

export function getContributionsSnapshot(): ContributionsSnapshot {
  return snapshot;
}

export function panelKey(pluginId: string, panelId: string): `plugin:${string}` {
  return `plugin:${pluginId}:${panelId}`;
}

export function parsePanelKey(key: string): { pluginId: string; panelId: string } | null {
  if (!key.startsWith("plugin:")) {
    return null;
  }
  const rest = key.slice("plugin:".length);
  const idx = rest.indexOf(":");
  if (idx <= 0) {
    return null;
  }
  return { pluginId: rest.slice(0, idx), panelId: rest.slice(idx + 1) };
}

export function findPanel(key: string) {
  const parsed = parsePanelKey(key);
  if (!parsed) {
    return null;
  }
  return (
    snapshot.panels.find((p) => p.pluginId === parsed.pluginId && p.id === parsed.panelId) ?? null
  );
}
