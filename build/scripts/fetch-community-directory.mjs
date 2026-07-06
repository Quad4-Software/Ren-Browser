// SPDX-License-Identifier: MIT
/**
 * Downloads the community interface directory snapshot for offline fallback.
 * Safe to run offline: exits 0 when the network request fails and a snapshot already exists.
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "../..");
const target = path.join(repoRoot, "internal/rns/data/community_directory.json");
const url =
  process.env.REN_BROWSER_COMMUNITY_DIRECTORY_URL ??
  "https://directory.rns.recipes/api/directory/submitted?search=&type=&status=online";

async function main() {
  let response;
  try {
    response = await fetch(url, { signal: AbortSignal.timeout(60_000) });
  } catch (err) {
    if (fs.existsSync(target)) {
      console.log(`fetch-community-directory: network failed, keeping existing ${target}`);
      return;
    }
    throw err;
  }
  if (!response.ok) {
    const body = await response.text();
    if (fs.existsSync(target)) {
      console.log(`fetch-community-directory: HTTP ${response.status}, keeping existing snapshot`);
      return;
    }
    throw new Error(`HTTP ${response.status}: ${body.slice(0, 256)}`);
  }
  const payload = await response.json();
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(target, `${JSON.stringify(payload, null, 2)}\n`, "utf8");
  const count = Array.isArray(payload?.data) ? payload.data.length : 0;
  console.log(`fetch-community-directory: wrote ${count} entries to ${target}`);
}

main().catch((err) => {
  console.error(`fetch-community-directory: ${err instanceof Error ? err.message : err}`);
  process.exit(1);
});
