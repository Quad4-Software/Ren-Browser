// SPDX-License-Identifier: MIT
import { PersistedState } from "runed";
import { localStorageKey } from "$lib/brand";
import { clampSidebarWidth, SIDEBAR_DEFAULT_PX } from "./pane-resize";

const SIDEBAR_WIDTH_KEY = localStorageKey("layout-sidebar-width", 1);
const SIDEBAR_STORE_SPAN_PX = 1200;

let sidebarWidth: PersistedState<number> | undefined;

// Lazily created so storage is resolved on first use, matching the previous
// direct localStorage reads. The window fallback keeps tests that only stub
// the localStorage global working under a non-DOM environment.
function sidebarWidthState(): PersistedState<number> {
  sidebarWidth ??= new PersistedState<number>(SIDEBAR_WIDTH_KEY, SIDEBAR_DEFAULT_PX, {
    storage: "local",
    syncTabs: false,
    window:
      typeof window !== "undefined"
        ? window
        : ({ localStorage: globalThis.localStorage } as typeof globalThis & Window),
    serializer: {
      serialize: (value) => String(value),
      deserialize: (raw) => {
        const parsed = Number.parseFloat(raw);
        return Number.isFinite(parsed) ? parsed : SIDEBAR_DEFAULT_PX;
      },
    },
  });
  return sidebarWidth;
}

export function readSidebarWidth(): number {
  try {
    return clampSidebarWidth(sidebarWidthState().current, SIDEBAR_STORE_SPAN_PX);
  } catch {
    return SIDEBAR_DEFAULT_PX;
  }
}

export function writeSidebarWidth(width: number): void {
  try {
    sidebarWidthState().current = clampSidebarWidth(width, SIDEBAR_STORE_SPAN_PX);
  } catch {
    // Storage can be unavailable in private windows or tests.
  }
}
