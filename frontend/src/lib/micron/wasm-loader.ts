// SPDX-License-Identifier: MIT
/**
 * Lazy-load micron-parser-go WASM for in-browser Micron rendering.
 * Bundled artifacts live under /vendor/micron-parser-go/; additional modules in IndexedDB.
 */

import {
  BUNDLED_MICRON_WASM_PARSER_ID,
  getMicronWasmParserWasm,
  hasStoredMicronWasmParsers,
} from "./wasm-store";
import { wasmExecDomId, wasmGlobalFlag } from "$lib/brand";

declare const __MICRON_WASM_SRI_WASM__: string | undefined;
declare const __MICRON_WASM_SRI_EXEC__: string | undefined;

type GoRuntime = {
  importObject: WebAssembly.Imports;
  run: (instance: WebAssembly.Instance) => void;
};

type IntegrityHashes = {
  wasm: string;
  wasmExec: string;
};

let resolvedPromise: Promise<boolean> | null = null;
let loadedParserId: string | null = null;
let integrityHashes: IntegrityHashes | null = null;
let bundledWasmByteLength = 0;

export function isWebAssemblySupported(): boolean {
  try {
    return typeof WebAssembly !== "undefined" && typeof WebAssembly.instantiate === "function";
  } catch {
    return false;
  }
}

export function isMicronWasmBundled(): boolean {
  return import.meta.env.VITE_MICRON_WASM_BUNDLED === "true";
}

export function isMicronWasmCapable(): boolean {
  return isWebAssemblySupported();
}

export async function isMicronWasmAvailable(): Promise<boolean> {
  if (!isMicronWasmCapable()) {
    return false;
  }
  if (isMicronWasmBundled() && bundledWasmByteLength > 0) {
    return true;
  }
  if (isMicronWasmBundled()) {
    return hasStoredMicronWasmParsers();
  }
  return hasStoredMicronWasmParsers();
}

export async function hasMicronWasmExec(): Promise<boolean> {
  if (!isMicronWasmCapable()) {
    return false;
  }
  if (isMicronWasmBundled()) {
    return true;
  }
  try {
    const res = await fetch(`${baseUrl()}/wasm_exec.js`, { method: "HEAD" });
    return res.ok;
  } catch {
    return false;
  }
}

export function getBundledMicronWasmByteLength(): number {
  return bundledWasmByteLength;
}

export async function probeBundledMicronWasmByteLength(): Promise<number> {
  if (bundledWasmByteLength > 0) {
    return bundledWasmByteLength;
  }
  if (!isMicronWasmBundled()) {
    return 0;
  }
  const wasmUrl = `${baseUrl()}/micron-parser-go.wasm`;
  try {
    const head = await fetch(wasmUrl, { method: "HEAD" });
    if (head.ok) {
      const contentLength = Number(head.headers.get("content-length"));
      if (Number.isFinite(contentLength) && contentLength > 0) {
        bundledWasmByteLength = contentLength;
        return contentLength;
      }
    }
    const res = await fetch(wasmUrl);
    if (!res.ok) {
      return 0;
    }
    const buf = await res.arrayBuffer();
    bundledWasmByteLength = buf.byteLength;
    return buf.byteLength;
  } catch {
    return 0;
  }
}

function baseUrl(): string {
  const root = import.meta.env.BASE_URL || "/";
  return `${root.replace(/\/?$/, "/")}vendor/micron-parser-go`;
}

async function computeSriHash(buf: ArrayBuffer): Promise<string> {
  const hash = await crypto.subtle.digest("SHA-384", buf);
  const base64 = btoa(String.fromCharCode(...new Uint8Array(hash)));
  return `sha384-${base64}`;
}

function injectMicronWasmStyles(): void {
  if (document.getElementById("micron-wasm-monospace-styles")) {
    return;
  }
  const styleEl = document.createElement("style");
  styleEl.id = "micron-wasm-monospace-styles";
  styleEl.textContent = `
    .Mu-nl { cursor: pointer; }
    .Mu-mnt {
      display: inline-block;
      min-width: 1ch;
      text-align: center;
      white-space: pre;
      text-decoration: inherit;
      font-variant-numeric: tabular-nums;
    }
    .Mu-mws { text-decoration: inherit; display: inline; }
  `;
  document.head.appendChild(styleEl);
}

async function getIntegrityHashes(): Promise<IntegrityHashes | null> {
  if (integrityHashes !== null) {
    return integrityHashes;
  }
  const embeddedWasm =
    typeof __MICRON_WASM_SRI_WASM__ !== "undefined" ? __MICRON_WASM_SRI_WASM__ : "";
  const embeddedExec =
    typeof __MICRON_WASM_SRI_EXEC__ !== "undefined" ? __MICRON_WASM_SRI_EXEC__ : "";
  if (embeddedWasm && embeddedExec) {
    integrityHashes = { wasm: embeddedWasm, wasmExec: embeddedExec };
    return integrityHashes;
  }
  try {
    const res = await fetch(`${baseUrl()}/integrity.json`);
    if (!res.ok) {
      return null;
    }
    integrityHashes = (await res.json()) as IntegrityHashes;
    return integrityHashes;
  } catch {
    return null;
  }
}

async function verifySri(buf: ArrayBuffer, expectedHash: string, name: string): Promise<void> {
  if (!expectedHash) {
    return;
  }
  const actualHash = await computeSriHash(buf);
  if (actualHash !== expectedHash) {
    throw new Error(`Micron WASM: SRI hash mismatch for ${name}`);
  }
}

async function injectScript(src: string, expectedHash: string): Promise<void> {
  const id = wasmExecDomId;
  if (document.getElementById(id)) {
    return;
  }
  const res = await fetch(src);
  if (!res.ok) {
    throw new Error(`Micron WASM: failed to fetch script ${src} (${res.status})`);
  }
  const buf = await res.arrayBuffer();
  await verifySri(buf, expectedHash, "wasm_exec.js");
  const blob = new Blob([buf], { type: "application/javascript" });
  const blobUrl = URL.createObjectURL(blob);
  await new Promise<void>((resolve, reject) => {
    const script = document.createElement("script");
    script.id = id;
    script.async = true;
    script.src = blobUrl;
    script.onload = () => {
      URL.revokeObjectURL(blobUrl);
      resolve();
    };
    script.onerror = () => {
      URL.revokeObjectURL(blobUrl);
      reject(new Error(`Micron WASM: failed to load script ${src}`));
    };
    document.head.appendChild(script);
  });
}

function isWasmWrapped(fn: unknown): boolean {
  return (
    typeof fn === "function" &&
    Boolean((fn as Record<string, boolean | undefined>)[wasmGlobalFlag])
  );
}

function wrapMicronConvertForNarrowJsSurface(): void {
  const inner = globalThis.micronConvert;
  if (typeof inner !== "function" || isWasmWrapped(inner)) {
    return;
  }
  const wrapped = (markup: unknown, darkTheme: boolean, forceMonospace: boolean) =>
    inner(String(markup ?? ""), darkTheme === true, forceMonospace === true);
  const tagged = wrapped as typeof wrapped & Record<string, boolean | undefined>;
  tagged[wasmGlobalFlag] = true;
  globalThis.micronConvert = tagged as typeof globalThis.micronConvert;
}

async function instantiateWasmBuffer(buf: ArrayBuffer, go: GoRuntime): Promise<void> {
  try {
    const result = await WebAssembly.instantiateStreaming(
      new Response(buf, { headers: { "content-type": "application/wasm" } }),
      go.importObject,
    );
    go.run(result.instance);
    return;
  } catch {
    const result = await WebAssembly.instantiate(buf, go.importObject);
    go.run(result.instance);
  }
}

async function loadWasmBytes(parserId: string): Promise<ArrayBuffer> {
  if (parserId === BUNDLED_MICRON_WASM_PARSER_ID) {
    const integrity = await getIntegrityHashes();
    const res = await fetch(`${baseUrl()}/micron-parser-go.wasm`);
    if (!res.ok) {
      throw new Error(`Micron WASM: fetch failed (${res.status})`);
    }
    const buf = await res.arrayBuffer();
    bundledWasmByteLength = buf.byteLength;
    await verifySri(buf, integrity?.wasm ?? "", "micron-parser-go.wasm");
    return buf;
  }
  const stored = await getMicronWasmParserWasm(parserId);
  if (!stored) {
    throw new Error(`Micron WASM: parser ${parserId} not found`);
  }
  await verifySri(stored.wasmBytes, stored.meta.wasmSri, `${stored.meta.label} (stored)`);
  return stored.wasmBytes;
}

async function instantiateOnce(parserId: string): Promise<void> {
  if (!isWebAssemblySupported()) {
    throw new Error("Micron WASM: WebAssembly is not available in this webview");
  }
  const root = baseUrl();
  const integrity = await getIntegrityHashes();
  const execUrl = `${root}/wasm_exec.js`;
  if (integrity?.wasmExec) {
    await injectScript(execUrl, integrity.wasmExec);
  } else if (await hasMicronWasmExec()) {
    await injectScript(execUrl, "");
  } else {
    throw new Error("Micron WASM: wasm_exec.js is not available");
  }
  if (typeof globalThis.Go === "undefined") {
    throw new Error("Micron WASM: Go runtime missing after wasm_exec.js load");
  }
  const go = new globalThis.Go() as GoRuntime;
  const wasmBuf = await loadWasmBytes(parserId);
  await instantiateWasmBuffer(wasmBuf, go);
  if (typeof globalThis.micronConvert !== "function") {
    throw new Error("Micron WASM: micronConvert was not registered");
  }
  wrapMicronConvertForNarrowJsSurface();
  loadedParserId = parserId;
}

export function teardownNomadMicronWasmRuntime(): void {
  document.getElementById(wasmExecDomId)?.remove();
  globalThis.micronConvert = undefined;
  globalThis.Go = undefined;
  loadedParserId = null;
}

export function invalidateNomadMicronWasmPreload(): void {
  resolvedPromise = null;
  integrityHashes = null;
  teardownNomadMicronWasmRuntime();
}

export function preloadNomadMicronWasm(parserId = BUNDLED_MICRON_WASM_PARSER_ID): Promise<boolean> {
  if (!isMicronWasmCapable()) {
    return Promise.resolve(false);
  }
  const targetId = parserId || BUNDLED_MICRON_WASM_PARSER_ID;
  if (typeof globalThis.micronConvert === "function" && loadedParserId === targetId) {
    injectMicronWasmStyles();
    return Promise.resolve(true);
  }
  if (loadedParserId !== targetId) {
    invalidateNomadMicronWasmPreload();
  }
  if (resolvedPromise === null) {
    resolvedPromise = (async () => {
      try {
        await instantiateOnce(targetId);
        const ok = typeof globalThis.micronConvert === "function";
        if (ok) {
          injectMicronWasmStyles();
        }
        return ok;
      } catch (err) {
        console.warn(err);
        resolvedPromise = null;
        loadedParserId = null;
        teardownNomadMicronWasmRuntime();
        return false;
      }
    })();
  }
  return resolvedPromise;
}

export function getLoadedMicronWasmParserId(): string | null {
  return loadedParserId;
}
