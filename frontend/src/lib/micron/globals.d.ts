// SPDX-License-Identifier: MIT
import type {} from "vite/client";

declare global {
  // Go wasm_exec.js
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  var Go: any;

  var micronConvert:
    ((markup: string, darkTheme: boolean, forceMonospace: boolean) => string) | undefined;
}

export {};
