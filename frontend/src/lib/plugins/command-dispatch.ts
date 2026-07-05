// SPDX-License-Identifier: MIT
import { invokeCommand } from "./api.js";
import { getContributionsSnapshot } from "./registry.js";
import { matchKeybind, parseChord, type KeybindAction } from "$lib/browser/keybinds";

export type PluginKeybindHandler = (pluginId: string, commandId: string) => void;

export function pluginKeybindChords(): Array<{
  chord: string;
  pluginId: string;
  commandId: string;
}> {
  const out: Array<{ chord: string; pluginId: string; commandId: string }> = [];
  for (const cmd of getContributionsSnapshot().commands) {
    if (cmd.keybind) {
      out.push({ chord: cmd.keybind, pluginId: cmd.pluginId, commandId: cmd.commandId });
    }
  }
  return out;
}

export function matchPluginKeybind(
  event: KeyboardEvent,
): { pluginId: string; commandId: string } | null {
  for (const entry of pluginKeybindChords()) {
    if (matchKeybind(event, entry.chord)) {
      return { pluginId: entry.pluginId, commandId: entry.commandId };
    }
  }
  return null;
}

export async function dispatchPluginCommand(pluginId: string, commandId: string) {
  await invokeCommand(pluginId, commandId, {});
}

export function formatPluginChord(chord: string): string {
  return parseChord(chord).key;
}

export type { KeybindAction };
