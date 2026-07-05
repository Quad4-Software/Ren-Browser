// SPDX-License-Identifier: MIT
let fallbackCounter = 0;

function fallbackId(): string {
  fallbackCounter = (fallbackCounter + 1) % 0xffff;
  const time = Date.now().toString(36);
  const rand = Math.random().toString(36).slice(2, 10);
  const seq = fallbackCounter.toString(36).padStart(2, "0");
  return `ren-${time}-${rand}-${seq}`;
}

export function randomId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    try {
      return crypto.randomUUID();
    } catch {
      return fallbackId();
    }
  }
  return fallbackId();
}
