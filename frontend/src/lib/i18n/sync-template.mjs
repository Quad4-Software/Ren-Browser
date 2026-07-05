// SPDX-License-Identifier: MIT
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const localesDir = join(dirname(fileURLToPath(import.meta.url)), "locales");
const en = JSON.parse(readFileSync(join(localesDir, "en.json"), "utf8")) as Record<string, unknown>;

function templateize(value: unknown): unknown {
  if (typeof value === "string") {
    return "";
  }
  const out: Record<string, unknown> = {};
  for (const [key, nested] of Object.entries(value as Record<string, unknown>)) {
    if (key.startsWith("_")) {
      continue;
    }
    out[key] = templateize(nested);
  }
  return out;
}

writeFileSync(
  join(localesDir, "_template.json"),
  `${JSON.stringify(templateize(en), null, 2)}\n`,
  "utf8",
);

console.log("Updated locales/_template.json from en.json");
