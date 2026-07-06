// SPDX-License-Identifier: MIT
import path from "node:path";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, "src/lib"),
    },
  },
  test: {
    environment: "node",
    environmentMatchGlobs: [
      ["src/lib/auth/api.test.ts", "happy-dom"],
      ["src/lib/browser/find-in-page.test.ts", "happy-dom"],
      ["src/lib/browser/docs-render.test.ts", "happy-dom"],
    ],
    include: ["src/**/*.test.ts"],
  },
});
