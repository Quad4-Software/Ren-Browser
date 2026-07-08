// SPDX-License-Identifier: MIT
import js from "@eslint/js";
import nounsanitized from "eslint-plugin-no-unsanitized";
import pluginSecurity from "eslint-plugin-security";
import svelte from "eslint-plugin-svelte";
import globals from "globals";
import ts from "typescript-eslint";

export default ts.config(
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs.recommended,
  pluginSecurity.configs.recommended,
  nounsanitized.configs.recommended,
  {
    rules: {
      "security/detect-object-injection": "off",
    },
  },
  {
    files: ["**/*.test.ts", "**/*.test.js", "**/*.spec.ts", "**/*.spec.js"],
    rules: {
      "security/detect-non-literal-fs-filename": "off",
    },
  },
  {
    files: ["**/keybinds.ts"],
    rules: {
      "security/detect-possible-timing-attacks": "off",
    },
  },
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ["**/*.svelte", "**/*.svelte.ts"],
    languageOptions: {
      parserOptions: {
        parser: ts.parser,
      },
    },
  },
  {
    ignores: [
      "dist/**",
      "bindings/**",
      "node_modules/**",
      "src/lib/micron/parser.ts",
      "src/lib/micron/nomad-renderer.ts",
    ],
  },
);
