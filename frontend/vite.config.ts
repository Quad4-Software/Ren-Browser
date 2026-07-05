import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import wails from "@wailsio/runtime/plugins/vite";
import { MICRON_PARSER_GO_RELEASE_TAG } from "../build/scripts/micron-parser-go-version.mjs";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.join(__dirname, "..");

function isMicronWasmBundled(): boolean {
  const wasmDir = path.join(repoRoot, "frontend", "public", "vendor", "micron-parser-go");
  const wasmFile = path.join(wasmDir, "micron-parser-go.wasm");
  const execFile = path.join(wasmDir, "wasm_exec.js");
  try {
    if (!fs.existsSync(wasmFile) || !fs.existsSync(execFile)) {
      return false;
    }
    return fs.statSync(wasmFile).size >= 8192 && fs.statSync(execFile).size >= 1024;
  } catch {
    return false;
  }
}

function loadMicronWasmIntegrity(): { wasm: string; wasmExec: string } | null {
  const integrityPath = path.join(
    repoRoot,
    "frontend",
    "public",
    "vendor",
    "micron-parser-go",
    "integrity.json",
  );
  try {
    return JSON.parse(fs.readFileSync(integrityPath, "utf8")) as { wasm: string; wasmExec: string };
  } catch {
    return null;
  }
}

const micronWasmBundled = isMicronWasmBundled();
const micronWasmIntegrity = loadMicronWasmIntegrity();

export default defineConfig({
  define: {
    "import.meta.env.VITE_MICRON_WASM_BUNDLED": JSON.stringify(
      micronWasmBundled ? "true" : "false",
    ),
    "import.meta.env.VITE_MICRON_PARSER_GO_RELEASE": JSON.stringify(MICRON_PARSER_GO_RELEASE_TAG),
    __MICRON_WASM_SRI_WASM__: JSON.stringify(micronWasmIntegrity?.wasm ?? ""),
    __MICRON_WASM_SRI_EXEC__: JSON.stringify(micronWasmIntegrity?.wasmExec ?? ""),
  },
  resolve: {
    alias: {
      $lib: path.resolve("src/lib"),
      "micron-parser": path.resolve("node_modules/micron-parser/js/micron-parser.js"),
    },
  },
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [tailwindcss(), svelte(), wails("./bindings")],
});
