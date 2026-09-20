// SPDX-License-Identifier: MIT
import { PersistedState } from "runed";

const STORAGE_KEY = "renbrowser:skip-plugin-network-install-warning:v1";

let skipped: PersistedState<boolean> | undefined;

// Lazily created so storage is resolved on first use. The window fallback
// keeps tests that only stub the localStorage global working.
function skippedState(): PersistedState<boolean> {
  skipped ??= new PersistedState<boolean>(STORAGE_KEY, false, {
    storage: "local",
    syncTabs: false,
    window:
      typeof window !== "undefined"
        ? window
        : ({ localStorage: globalThis.localStorage } as typeof globalThis & Window),
    serializer: {
      serialize: (value) => (value ? "1" : "0"),
      deserialize: (raw) => raw === "1",
    },
  });
  return skipped;
}

export function isPluginNetworkInstallWarningSkipped(): boolean {
  try {
    return skippedState().current === true;
  } catch {
    return false;
  }
}

export function setPluginNetworkInstallWarningSkipped(skip: boolean): void {
  try {
    skippedState().current = skip;
  } catch {
    // ignore storage failures
  }
}
