// SPDX-License-Identifier: MIT
import path from "node:path";
import { fileURLToPath } from "node:url";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vitest/config";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const domTests = [
  "src/lib/auth/api.test.ts",
  "src/lib/app/create-app.test.ts",
  "src/lib/browser/find-in-page.test.ts",
  "src/lib/browser/mobile-gestures.test.ts",
  "src/lib/browser/docs-render.test.ts",
  "src/lib/browser/docs-render.snapshot.test.ts",
  "src/lib/documents/epub.test.ts",
  "src/lib/browser/page-links.test.ts",
  "src/lib/micron/multiline.test.ts",
  "src/lib/components/**/*layout*.test.ts",
  "src/lib/components/mobile-tabs-page.test.ts",
  "src/lib/components/confirm-dialog.test.ts",
  "src/lib/components/empty-state.test.ts",
  "src/lib/components/plugin-toast.test.ts",
  "src/lib/components/page-error-state.test.ts",
];

export default defineConfig({
  plugins: [svelte({ emitCss: false })],
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, "src/lib"),
      "micron-parser": path.resolve(__dirname, "node_modules/micron-parser/js/micron-parser.js"),
    },
    conditions: ["browser"],
  },
  test: {
    coverage: {
      provider: "v8",
      reporter: ["text", "html", "lcov"],
      reportsDirectory: "../coverage/frontend",
      include: ["src/lib/**/*.{ts,svelte}"],
      exclude: ["src/lib/**/*.test.ts", "src/lib/test/**"],
    },
    projects: [
      {
        extends: true,
        test: {
          name: "unit",
          environment: "node",
          include: ["src/**/*.test.ts"],
          exclude: domTests,
        },
      },
      {
        extends: true,
        test: {
          name: "dom",
          environment: "happy-dom",
          include: domTests,
        },
      },
    ],
  },
});
