// SPDX-License-Identifier: MIT
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export function micronWasmRepoRoot() {
  return path.join(__dirname, "..", "..");
}

export function micronWasmVendorPaths(repoRoot = micronWasmRepoRoot()) {
  const dir = path.join(repoRoot, "frontend", "public", "vendor", "micron-parser-go");
  return {
    dir,
    wasm: path.join(dir, "micron-parser-go.wasm"),
    wasmExec: path.join(dir, "wasm_exec.js"),
  };
}

export function isMicronWasmBundled(repoRoot = micronWasmRepoRoot()) {
  const { wasm, wasmExec } = micronWasmVendorPaths(repoRoot);
  try {
    if (!fs.existsSync(wasm) || !fs.existsSync(wasmExec)) {
      return false;
    }
    return fs.statSync(wasm).size >= 8192 && fs.statSync(wasmExec).size >= 1024;
  } catch {
    return false;
  }
}
