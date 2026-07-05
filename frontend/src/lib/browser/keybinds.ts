export type KeybindAction =
  | "focusUrl"
  | "reload"
  | "devtools"
  | "findInPage"
  | "discovery"
  | "settings"
  | "newTab"
  | "newWindow"
  | "closeTab"
  | "fullscreen";

export type KeybindMap = Record<KeybindAction, string>;

export type KeybindSettings = {
  bindings: KeybindMap;
};

let recordingKeybind = false;

export function setKeybindRecording(active: boolean): void {
  recordingKeybind = active;
}

export function isKeybindRecording(): boolean {
  return recordingKeybind;
}

export const KEYBIND_LABELS: Record<KeybindAction, string> = {
  focusUrl: "Focus address bar",
  reload: "Reload page",
  devtools: "Developer tools",
  findInPage: "Find in page",
  discovery: "Discovery panel",
  settings: "Settings panel",
  newTab: "New tab",
  newWindow: "New window",
  closeTab: "Close tab",
  fullscreen: "Toggle fullscreen",
};

export function defaultKeybinds(): KeybindSettings {
  return {
    bindings: {
      focusUrl: "mod+l",
      reload: "mod+r",
      devtools: "mod+shift+i",
      findInPage: "mod+f",
      discovery: "mod+shift+d",
      settings: "mod+,",
      newTab: "mod+t",
      newWindow: "mod+shift+n",
      closeTab: "mod+w",
      fullscreen: "f11",
    },
  };
}

type ParsedChord = {
  mod: boolean;
  shift: boolean;
  alt: boolean;
  key: string;
};

export function parseChord(chord: string): ParsedChord {
  const parts = chord
    .toLowerCase()
    .split("+")
    .map((part) => part.trim())
    .filter(Boolean);
  const key = parts.at(-1) ?? "";
  return {
    mod: parts.includes("mod"),
    shift: parts.includes("shift"),
    alt: parts.includes("alt"),
    key,
  };
}

export function matchKeybind(event: KeyboardEvent, chord: string): boolean {
  if (!chord) {
    return false;
  }
  const parsed = parseChord(chord);
  const mod = event.ctrlKey || event.metaKey;
  if (parsed.mod !== mod) {
    return false;
  }
  if (parsed.shift !== event.shiftKey) {
    return false;
  }
  if (parsed.alt !== event.altKey) {
    return false;
  }
  return event.key.toLowerCase() === parsed.key;
}

export function formatChord(chord: string): string {
  return chord
    .split("+")
    .map((part) => {
      const token = part.trim().toLowerCase();
      if (token === "mod") {
        return navigator.platform.toLowerCase().includes("mac") ? "Cmd" : "Ctrl";
      }
      if (token === "shift") {
        return "Shift";
      }
      if (token === "alt") {
        return "Alt";
      }
      if (token === ",") {
        return ",";
      }
      return token.length === 1 ? token.toUpperCase() : token;
    })
    .join("+");
}

export function chordFromEvent(event: KeyboardEvent): string {
  const parts: string[] = [];
  if (event.ctrlKey || event.metaKey) {
    parts.push("mod");
  }
  if (event.shiftKey) {
    parts.push("shift");
  }
  if (event.altKey) {
    parts.push("alt");
  }
  const key = event.key.toLowerCase();
  if (!["control", "meta", "shift", "alt"].includes(key)) {
    parts.push(key);
  }
  return parts.join("+");
}

export function mergeKeybinds(
  saved: Partial<KeybindSettings> | { bindings?: Partial<KeybindMap> | null } | null | undefined,
): KeybindSettings {
  const defaults = defaultKeybinds();
  return {
    bindings: { ...defaults.bindings, ...(saved?.bindings ?? {}) },
  };
}
