// SPDX-License-Identifier: MIT
import { localStorageKey } from "$lib/brand";
import { clampSidebarWidth, SIDEBAR_DEFAULT_PX } from "./pane-resize";

const SIDEBAR_WIDTH_KEY = localStorageKey("layout-sidebar-width", 1);
const SIDEBAR_STORE_SPAN_PX = 1200;

export function readSidebarWidth(): number {
  if (typeof localStorage === "undefined") {
    return SIDEBAR_DEFAULT_PX;
  }
  try {
    const raw = localStorage.getItem(SIDEBAR_WIDTH_KEY);
    if (!raw) {
      return SIDEBAR_DEFAULT_PX;
    }
    return clampSidebarWidth(Number.parseFloat(raw), SIDEBAR_STORE_SPAN_PX);
  } catch {
    return SIDEBAR_DEFAULT_PX;
  }
}

export function writeSidebarWidth(width: number): void {
  if (typeof localStorage === "undefined") {
    return;
  }
  try {
    localStorage.setItem(
      SIDEBAR_WIDTH_KEY,
      String(clampSidebarWidth(width, SIDEBAR_STORE_SPAN_PX)),
    );
  } catch {
    // Storage can be unavailable in private windows or tests.
  }
}
