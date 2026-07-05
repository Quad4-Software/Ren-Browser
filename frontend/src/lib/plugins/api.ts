// SPDX-License-Identifier: MIT
import { Events } from "@wailsio/runtime";
import {
  DisablePlugin,
  EnablePlugin,
  GetContributions,
  GetPluginStorage,
  InstallPluginFromDir,
  InstallPluginFromZip,
  InvokeCommand,
  ListPlugins,
  SetPluginStorage,
  UninstallPlugin,
} from "../../../bindings/renbrowser/internal/app/pluginhost.js";
import type { ContributionsSnapshot, PluginContext, PluginModule } from "./api-types.js";

export async function listPlugins() {
  return ListPlugins();
}

export async function enablePlugin(id: string) {
  return EnablePlugin(id);
}

export async function disablePlugin(id: string) {
  return DisablePlugin(id);
}

export async function installFromZip(path: string) {
  return InstallPluginFromZip(path);
}

export async function installFromDir(path: string) {
  return InstallPluginFromDir(path);
}

export async function uninstallPlugin(id: string) {
  return UninstallPlugin(id);
}

export async function getContributions(): Promise<ContributionsSnapshot> {
  const raw = await GetContributions();
  return {
    panels: (raw.panels ?? []).map((p) => ({
      pluginId: p.pluginId,
      id: p.id,
      title: p.title,
      icon: p.icon,
      entry: p.entry,
      location: p.location,
    })),
    commands: (raw.commands ?? []).map((c) => ({
      pluginId: c.pluginId,
      commandId: c.id,
      title: c.title,
      keybind: c.keybind,
    })),
    devtools: (raw.devtools ?? []).map((d) => ({
      pluginId: d.pluginId,
      id: d.id,
      title: d.title,
      entry: d.entry,
    })),
    urlSchemes: (raw.urlSchemes ?? []).map((s) => ({
      pluginId: s.pluginId,
      scheme: s.scheme,
      handler: s.handler,
    })),
  };
}

export async function invokeCommand(
  pluginId: string,
  commandId: string,
  args: Record<string, string> = {},
) {
  return InvokeCommand(pluginId, commandId, args);
}

export function createPluginContext(
  pluginId: string,
  opts: {
    getCurrentURL: () => string;
    navigate: (url: string) => void;
    showToast: (message: string) => void;
  },
): PluginContext {
  const disposables: Array<{ dispose(): void }> = [];
  return {
    pluginId,
    subscriptions: {
      add(disposable) {
        disposables.push(disposable);
      },
    },
    storage: {
      async get(key) {
        const value = await GetPluginStorage(pluginId, key);
        return value || null;
      },
      async set(key, value) {
        await SetPluginStorage(pluginId, key, value);
      },
    },
    navigation: {
      getCurrentURL: opts.getCurrentURL,
      navigate: opts.navigate,
    },
    events: {
      on(event, fn) {
        const full = event.includes(":") ? event : `plugin:${pluginId}:${event}`;
        const cancel = Events.On(full, (payload) => {
          fn((payload as { data?: unknown }).data ?? payload);
        });
        return { dispose: () => cancel() };
      },
      emit(event, data) {
        const full = event.includes(":") ? event : `plugin:${pluginId}:${event}`;
        Events.Emit(full, data);
      },
    },
    ui: {
      showToast: opts.showToast,
    },
  };
}

export type { PluginModule };
