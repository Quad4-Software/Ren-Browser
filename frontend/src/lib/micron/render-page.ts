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
  },
): MicronEffectiveEngine {
  if (preference === "wasm") {
    if (ctx.wasmEnabled && ctx.wasmAvailable && ctx.wasmReady && isWebAssemblySupported()) {
      return "wasm";
    }
    return "js";
  }
  if (preference === "go") {
    return "go";
  }
  if (preference === "js") {
    return "js";
  }
  if (ctx.wasmEnabled && ctx.wasmAvailable && ctx.wasmReady && isWebAssemblySupported()) {
    return "wasm";
  }
  if (ctx.hasServerHtml) {
    return "go";
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
