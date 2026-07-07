// SPDX-License-Identifier: MIT
const STORAGE_KEY = "renbrowser:skip-plugin-network-install-warning:v1";

export function isPluginNetworkInstallWarningSkipped(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

export function setPluginNetworkInstallWarningSkipped(skip: boolean): void {
  try {
    if (skip) {
      localStorage.setItem(STORAGE_KEY, "1");
    } else {
      localStorage.removeItem(STORAGE_KEY);
    }
  } catch {
    // ignore storage failures
  }
}
