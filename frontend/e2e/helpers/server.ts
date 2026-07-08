// SPDX-License-Identifier: MIT
import { spawn, type ChildProcess } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
const binName = process.platform === "win32" ? "renbrowser-server.exe" : "renbrowser-server";
const serverBin = path.join(root, "bin", binName);

export const e2ePort = Number(process.env.REN_BROWSER_E2E_PORT || 19284);
export const e2eHost = process.env.REN_BROWSER_E2E_HOST || "127.0.0.1";
export const e2eBaseURL = `http://${e2eHost}:${e2ePort}/`;

export async function waitForURL(url: string, timeoutMs = 90_000): Promise<void> {
  const start = Date.now();
  let lastErr = "";
  while (Date.now() - start < timeoutMs) {
    try {
      const res = await fetch(url, { redirect: "follow" });
      if (res.ok) {
        return;
      }
      lastErr = `HTTP ${res.status}`;
    } catch (err) {
      lastErr = err instanceof Error ? err.message : String(err);
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`server did not become ready at ${url}: ${lastErr}`);
}

export type ServerHandle = {
  process: ChildProcess;
  profileHome: string;
  log: () => string;
  stop: () => Promise<void>;
};

export async function startServer(opts?: {
  port?: number;
  host?: string;
  extraEnv?: Record<string, string>;
}): Promise<ServerHandle> {
  const port = opts?.port ?? e2ePort;
  const host = opts?.host ?? e2eHost;
  const profileHome = await mkdtemp(path.join(os.tmpdir(), "renbrowser-e2e-"));
  let serverLog = "";

  const child = spawn(serverBin, ["--host", host, "--port", String(port)], {
    cwd: root,
    env: {
      ...process.env,
      HOME: profileHome,
      REN_BROWSER_PUBLIC_MODE: "1",
      ...(opts?.extraEnv ?? {}),
    },
    stdio: ["ignore", "pipe", "pipe"],
  });

  child.stdout?.on("data", (chunk) => {
    serverLog += chunk.toString();
  });
  child.stderr?.on("data", (chunk) => {
    serverLog += chunk.toString();
  });

  await waitForURL(`http://${host}:${port}/`);

  return {
    process: child,
    profileHome,
    log: () => serverLog,
    stop: async () => {
      child.kill("SIGTERM");
      await new Promise<void>((resolve) => {
        if (child.exitCode !== null) {
          resolve();
          return;
        }
        child.once("exit", () => resolve());
        setTimeout(() => {
          try {
            child.kill("SIGKILL");
          } catch {
            // already exited
          }
          resolve();
        }, 5_000);
      });
      await rm(profileHome, { recursive: true, force: true });
    },
  };
}

export { root, serverBin };
