// SPDX-License-Identifier: MIT
import MicronParser, { parseMicronHeaderTagColors } from "./parser";
import { renderNomadPageByPath } from "./nomad-renderer";
import {
  BUNDLED_MICRON_WASM_PARSER_ID,
  bundledMicronWasmParserEntry,
  listMicronWasmParsers,
  normalizeMicronWasmParserId,
  type MicronWasmParserListEntry,
} from "./wasm-store";
import {
  getBundledMicronWasmByteLength,
  getLoadedMicronWasmParserId,
  isMicronWasmAvailable,
  isMicronWasmBundled,
  isWebAssemblySupported,
  preloadNomadMicronWasm,
  probeBundledMicronWasmByteLength,
} from "./wasm-loader";

export type MicronRendererPreference = "auto" | "wasm" | "go" | "js";
export type MicronEffectiveEngine = "wasm" | "go" | "js";

/** Prefer server-rendered HTML above this raw size to avoid main-thread re-parse. */
export const LARGE_MICRON_RAW_BYTES = 48 * 1024;

export function usesClientMicronRenderer(engine: MicronEffectiveEngine): boolean {
  return engine === "js" || engine === "wasm";
}

export function normalizeMicronRendererPreference(
  value: string | undefined,
): MicronRendererPreference {
  const v = (value ?? "auto").toLowerCase();
  if (v === "wasm" || v === "go" || v === "js" || v === "auto") {
    return v;
  }
  return "auto";
}

export function shouldPreloadMicronWasm(
  preference: MicronRendererPreference,
  wasmEnabled: boolean,
): boolean {
  return wasmEnabled && (preference === "auto" || preference === "wasm");
}

export function resolveEffectiveMicronEngine(
  preference: MicronRendererPreference,
  ctx: {
    wasmEnabled: boolean;
    wasmAvailable: boolean;
    wasmReady: boolean;
    hasServerHtml: boolean;
    rawBytes?: number;
  },
): MicronEffectiveEngine {
  const large = (ctx.rawBytes ?? 0) >= LARGE_MICRON_RAW_BYTES;

  // Large pages already have Go HTML from Navigate; re-parsing with WASM/JS on
  // the UI thread (DOMPurify + {@html}) is the main source of multi-second lag.
  if (large && ctx.hasServerHtml && preference !== "js") {
    return "go";
  }

  if (preference === "wasm") {
    if (ctx.wasmEnabled && ctx.wasmAvailable && ctx.wasmReady && isWebAssemblySupported()) {
      return "wasm";
    }
    return ctx.hasServerHtml ? "go" : "js";
  }
  if (preference === "go") {
    return "go";
  }
  if (preference === "js") {
    return "js";
  }
  // Auto: prefer server Go HTML when present (same micron-parser-go as WASM,
  // without a second main-thread parse). Then WASM, then JS.
  if (ctx.hasServerHtml) {
    return "go";
  }
  if (ctx.wasmEnabled && ctx.wasmAvailable && ctx.wasmReady && isWebAssemblySupported()) {
    return "wasm";
  }
  return "js";
}

export function micronEffectiveLabel(effective: MicronEffectiveEngine, parserLabel = ""): string {
  if (effective === "go") {
    return "Micron Go";
  }
  if (effective === "wasm") {
    return parserLabel ? `Micron WASM (${parserLabel})` : "Micron WASM";
  }
  return "Micron JS";
}

export function micronRendererBadgeLabel(
  preference: MicronRendererPreference,
  effective: MicronEffectiveEngine,
  parserLabel = "",
): string {
  const effectiveLabel = micronEffectiveLabel(effective, parserLabel);
  if (preference === "auto") {
    return `Auto: ${effectiveLabel}`;
  }
  return effectiveLabel;
}

export function nodeHashFromURL(url: string): string {
  const hash = url.split(":/")[0]?.trim().toLowerCase() ?? "";
  return /^[a-f0-9]{32}$/.test(hash) ? hash : "";
}

export function pagePathFromURL(url: string): string {
  const rest = url.includes(":/") ? (url.split(":/")[1] ?? "") : url;
  const bare = rest.split(/[?`]/)[0] ?? rest;
  return bare.startsWith("/") ? bare : `/${bare}`;
}

export function renderClientMicronPage(
  url: string,
  source: string,
  engine: MicronEffectiveEngine,
): string {
  const path = pagePathFromURL(url);
  const nodeHash = nodeHashFromURL(url);
  const useWasm = engine === "wasm" && typeof globalThis.micronConvert === "function";
  return renderNomadPageByPath(path, source, {}, MicronParser, {
    nomadDestinationHash: nodeHash,
    nomad_micron_wasm_use: useWasm,
    darkTheme: darkThemeFromDocument(),
    // Always emit monospace cells so ASCII/box art stays aligned.
    // ContentViewer preserve-layout is CSS-only and does not affect this.
    forceMonospace: true,
    renderMarkdown: true,
    renderHtml: true,
    renderPlaintext: true,
  });
}

export function parseMicronHeaderColors(source: string): { fg: string; bg: string } {
  return parseMicronHeaderTagColors(source, darkThemeFromDocument());
}

export function darkThemeFromDocument(): boolean {
  if (typeof document === "undefined") {
    return true;
  }
  return document.documentElement.dataset.theme !== "light";
}

export async function listAvailableMicronWasmParsers(): Promise<MicronWasmParserListEntry[]> {
  const includeBundled = isMicronWasmBundled();
  const bundledBytes = includeBundled
    ? getBundledMicronWasmByteLength() || (await probeBundledMicronWasmByteLength())
    : 0;
  return listMicronWasmParsers({ includeBundled, bundledByteLength: bundledBytes });
}

export async function resolveMicronWasmParserLabel(parserId: string): Promise<string> {
  const parsers = await listAvailableMicronWasmParsers();
  const match = parsers.find((entry) => entry.id === parserId);
  return match?.label ?? bundledMicronWasmParserEntry().label;
}

export async function ensureMicronWasmReady(
  wasmEnabled: boolean,
  parserId = BUNDLED_MICRON_WASM_PARSER_ID,
): Promise<boolean> {
  if (!wasmEnabled || !(await isMicronWasmAvailable())) {
    return false;
  }
  const parsers = await listAvailableMicronWasmParsers();
  const availableIds = new Set(parsers.map((entry) => entry.id));
  const selected = normalizeMicronWasmParserId(parserId, availableIds);
  return preloadNomadMicronWasm(selected);
}

export function activeMicronWasmParserMatches(
  selectedParserId: string,
  loadedParserId: string | null,
): boolean {
  const selected = selectedParserId || BUNDLED_MICRON_WASM_PARSER_ID;
  return loadedParserId === selected;
}

export { BUNDLED_MICRON_WASM_PARSER_ID, getLoadedMicronWasmParserId, isWebAssemblySupported };
