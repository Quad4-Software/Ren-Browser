// SPDX-License-Identifier: MIT
/**
 * IndexedDB store for multiple micron-parser-go WASM modules (uploads and verified GitHub releases).
 */

import { FetchMicronParserGoRelease } from "../../../bindings/renbrowser/internal/app/browserservice.js";
import { randomId } from "$lib/browser/id";
import { indexedDbName } from "$lib/brand";

export const BUNDLED_MICRON_WASM_PARSER_ID = "bundled";
export const WASM_FILENAME = "micron-parser-go.wasm";
export const MAX_WASM_PARSER_BYTES = 14 * 1024 * 1024;

const DB_NAME = indexedDbName;
const DB_VERSION = 1;
const META_STORE = "meta";
const WASM_STORE = "wasm";

export type MicronWasmParserSource = "bundled" | "github" | "upload";

export type MicronWasmParserMeta = {
  id: string;
  label: string;
  source: Exclude<MicronWasmParserSource, "bundled">;
  releaseTag: string;
  wasmSri: string;
  byteLength: number;
  expectedSha256Hex: string | null;
  addedAt: number;
};

export type MicronWasmParserListEntry = {
  id: string;
  label: string;
  source: MicronWasmParserSource;
  releaseTag: string;
  byteLength: number;
  removable: boolean;
};

type StoredParserRecord = MicronWasmParserMeta & {
  wasmBytes: ArrayBuffer;
};

function bundledReleaseTag(): string {
  const tag = import.meta.env.VITE_MICRON_PARSER_GO_RELEASE;
  return typeof tag === "string" && tag.trim() ? tag.trim() : "v1.0.6";
}

export function bundledMicronWasmReleaseTag(): string {
  return bundledReleaseTag();
}

export function bundledMicronWasmParserEntry(byteLength = 0): MicronWasmParserListEntry {
  const tag = bundledReleaseTag();
  return {
    id: BUNDLED_MICRON_WASM_PARSER_ID,
    label: `Bundled (${tag})`,
    source: "bundled",
    releaseTag: tag,
    byteLength,
    removable: false,
  };
}

function openDb(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onerror = () => reject(req.error ?? new Error("IndexedDB open failed"));
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains(META_STORE)) {
        db.createObjectStore(META_STORE);
      }
      if (!db.objectStoreNames.contains(WASM_STORE)) {
        db.createObjectStore(WASM_STORE);
      }
    };
    req.onsuccess = () => resolve(req.result);
  });
}

export function parseShasums256ForFilename(text: string, filename: string): string | null {
  if (!text || !filename) {
    return null;
  }
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim();
    if (!line || line.startsWith("#")) {
      continue;
    }
    const m = line.match(/^([a-fA-F0-9]{64})\s+\*?(\S+)\s*$/);
    if (!m) {
      continue;
    }
    const name = m[2].trim();
    if (name === filename || name.endsWith(`/${filename}`)) {
      return m[1].toLowerCase();
    }
  }
  return null;
}

export async function sha256HexOfBuffer(buf: ArrayBuffer): Promise<string> {
  const d = await crypto.subtle.digest("SHA-256", buf);
  const bytes = new Uint8Array(d);
  let hex = "";
  for (let i = 0; i < bytes.length; i++) {
    hex += bytes[i].toString(16).padStart(2, "0");
  }
  return hex;
}

export async function computeWasmSriSha384(buf: ArrayBuffer): Promise<string> {
  const hash = await crypto.subtle.digest("SHA-384", buf);
  const base64 = btoa(String.fromCharCode(...new Uint8Array(hash)));
  return `sha384-${base64}`;
}

function assertSri(wasmSri: string): void {
  if (!/^sha384-[A-Za-z0-9+/=]+$/.test(wasmSri)) {
    throw new Error("Micron WASM: invalid SRI format");
  }
}

function decodeBase64Wasm(base64: string): ArrayBuffer {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
}

async function readMeta(id: string): Promise<MicronWasmParserMeta | null> {
  const db = await openDb();
  try {
    const tx = db.transaction(META_STORE, "readonly");
    const raw = await new Promise<StoredParserRecord | MicronWasmParserMeta | undefined>(
      (resolve, reject) => {
        const r = tx.objectStore(META_STORE).get(id);
        r.onsuccess = () =>
          resolve(r.result as StoredParserRecord | MicronWasmParserMeta | undefined);
        r.onerror = () => reject(r.error ?? new Error("IndexedDB read failed"));
      },
    );
    if (!raw || !raw.id) {
      return null;
    }
    return {
      id: raw.id,
      label: raw.label,
      source: raw.source,
      releaseTag: raw.releaseTag,
      wasmSri: raw.wasmSri,
      byteLength: raw.byteLength,
      expectedSha256Hex: raw.expectedSha256Hex,
      addedAt: raw.addedAt,
    };
  } finally {
    db.close();
  }
}

async function readWasmBytes(id: string): Promise<ArrayBuffer | null> {
  const db = await openDb();
  try {
    const tx = db.transaction(WASM_STORE, "readonly");
    const buf = await new Promise<ArrayBuffer | undefined>((resolve, reject) => {
      const r = tx.objectStore(WASM_STORE).get(id);
      r.onsuccess = () => resolve(r.result as ArrayBuffer | undefined);
      r.onerror = () => reject(r.error ?? new Error("IndexedDB read failed"));
    });
    return buf ?? null;
  } finally {
    db.close();
  }
}

async function writeParser(record: StoredParserRecord): Promise<void> {
  assertSri(record.wasmSri);
  if (record.wasmBytes.byteLength > MAX_WASM_PARSER_BYTES) {
    throw new Error(`Micron WASM exceeds maximum size (${MAX_WASM_PARSER_BYTES} bytes)`);
  }
  const db = await openDb();
  try {
    const tx = db.transaction([META_STORE, WASM_STORE], "readwrite");
    const meta: MicronWasmParserMeta = {
      id: record.id,
      label: record.label,
      source: record.source,
      releaseTag: record.releaseTag,
      wasmSri: record.wasmSri,
      byteLength: record.byteLength,
      expectedSha256Hex: record.expectedSha256Hex,
      addedAt: record.addedAt,
    };
    tx.objectStore(META_STORE).put(meta, record.id);
    tx.objectStore(WASM_STORE).put(record.wasmBytes, record.id);
    await new Promise<void>((resolve, reject) => {
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error ?? new Error("IndexedDB write failed"));
      tx.onabort = () => reject(tx.error ?? new Error("IndexedDB write aborted"));
    });
  } finally {
    db.close();
  }
}

export async function listMicronWasmParsers(
  options: { includeBundled?: boolean; bundledByteLength?: number } = {},
): Promise<MicronWasmParserListEntry[]> {
  const entries: MicronWasmParserListEntry[] = [];
  if (options.includeBundled) {
    entries.push(bundledMicronWasmParserEntry(options.bundledByteLength ?? 0));
  }
  const db = await openDb();
  try {
    const tx = db.transaction(META_STORE, "readonly");
    const all = await new Promise<MicronWasmParserMeta[]>((resolve, reject) => {
      const r = tx.objectStore(META_STORE).getAll();
      r.onsuccess = () => resolve((r.result as MicronWasmParserMeta[]) ?? []);
      r.onerror = () => reject(r.error ?? new Error("IndexedDB read failed"));
    });
    all.sort((a, b) => a.addedAt - b.addedAt);
    for (const meta of all) {
      entries.push({
        id: meta.id,
        label: meta.label,
        source: meta.source,
        releaseTag: meta.releaseTag,
        byteLength: meta.byteLength,
        removable: true,
      });
    }
    return entries;
  } finally {
    db.close();
  }
}

export async function hasStoredMicronWasmParsers(): Promise<boolean> {
  const db = await openDb();
  try {
    const tx = db.transaction(META_STORE, "readonly");
    const count = await new Promise<number>((resolve, reject) => {
      const r = tx.objectStore(META_STORE).count();
      r.onsuccess = () => resolve(r.result);
      r.onerror = () => reject(r.error ?? new Error("IndexedDB count failed"));
    });
    return count > 0;
  } finally {
    db.close();
  }
}

export async function getMicronWasmParserWasm(
  parserId: string,
): Promise<{ meta: MicronWasmParserMeta; wasmBytes: ArrayBuffer } | null> {
  if (parserId === BUNDLED_MICRON_WASM_PARSER_ID) {
    return null;
  }
  const meta = await readMeta(parserId);
  const wasmBytes = await readWasmBytes(parserId);
  if (!meta || !wasmBytes) {
    return null;
  }
  return { meta, wasmBytes };
}

export async function removeMicronWasmParser(parserId: string): Promise<void> {
  if (parserId === BUNDLED_MICRON_WASM_PARSER_ID) {
    throw new Error("Cannot remove bundled Micron WASM parser");
  }
  const db = await openDb();
  try {
    const tx = db.transaction([META_STORE, WASM_STORE], "readwrite");
    tx.objectStore(META_STORE).delete(parserId);
    tx.objectStore(WASM_STORE).delete(parserId);
    await new Promise<void>((resolve, reject) => {
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error ?? new Error("IndexedDB delete failed"));
      tx.onabort = () => reject(tx.error ?? new Error("IndexedDB delete aborted"));
    });
  } finally {
    db.close();
  }
}

async function storeWasmParser(
  source: "github" | "upload",
  label: string,
  releaseTag: string,
  wasmBytes: ArrayBuffer,
  expectedSha256Hex: string | null,
): Promise<string> {
  if (wasmBytes.byteLength > MAX_WASM_PARSER_BYTES) {
    throw new Error(`Micron WASM exceeds maximum size (${MAX_WASM_PARSER_BYTES} bytes)`);
  }
  if (wasmBytes.byteLength < 4096) {
    throw new Error("Micron WASM file is too small");
  }
  const wasmSri = await computeWasmSriSha384(wasmBytes);
  const id = randomId();
  await writeParser({
    id,
    label,
    source,
    releaseTag,
    wasmSri,
    byteLength: wasmBytes.byteLength,
    expectedSha256Hex,
    addedAt: Date.now(),
    wasmBytes,
  });
  return id;
}

export async function addMicronWasmParserFromUpload(file: File): Promise<string> {
  if (!file.name.toLowerCase().endsWith(".wasm")) {
    throw new Error("File must be a .wasm module");
  }
  const wasmBytes = await file.arrayBuffer();
  const label = file.name;
  return storeWasmParser("upload", label, file.name, wasmBytes, null);
}

export async function addMicronWasmParserFromGitHub(tag: string): Promise<string> {
  const trimmed = tag.trim();
  if (!trimmed) {
    throw new Error("Release tag is required");
  }
  const result = await FetchMicronParserGoRelease(trimmed);
  const wasmBytes = decodeBase64Wasm(result.wasmBase64);
  const label = `micron-parser-go ${result.releaseTag}`;
  return storeWasmParser("github", label, result.releaseTag, wasmBytes, result.sha256Hex);
}

export function normalizeMicronWasmParserId(
  parserId: string | undefined,
  availableIds: Set<string>,
): string {
  const id = (parserId ?? BUNDLED_MICRON_WASM_PARSER_ID).trim();
  if (id && availableIds.has(id)) {
    return id;
  }
  if (availableIds.has(BUNDLED_MICRON_WASM_PARSER_ID)) {
    return BUNDLED_MICRON_WASM_PARSER_ID;
  }
  const first = [...availableIds][0];
  return first ?? BUNDLED_MICRON_WASM_PARSER_ID;
}
