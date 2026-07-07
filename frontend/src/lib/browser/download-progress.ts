// SPDX-License-Identifier: MIT
export type ActiveDownloadStatus =
  "pending" | "downloading" | "retrying" | "completed" | "failed" | "canceled" | "interrupted";

export type ActiveDownloadRow = {
  id: string;
  url: string;
  name: string;
  path?: string;
  received: number;
  total: number;
  status: ActiveDownloadStatus;
  error?: string;
  attempt?: number;
  maxAttempts?: number;
  startedAt: number;
  updatedAt: number;
};

export type DownloadProgressView = ActiveDownloadRow & {
  speedBps: number;
  etaSeconds: number | null;
};

type Sample = { at: number; received: number; speedBps: number };

const samples = new Map<string, Sample>();

/**
 * Derives client-side speed (bytes/sec, smoothed) and ETA for each active
 * download by comparing against the last-seen sample for that id. The
 * backend only reports raw byte counts and timestamps, so this state lives
 * on the client and is keyed by download id across successive snapshots.
 */
export function withProgress(rows: ActiveDownloadRow[]): DownloadProgressView[] {
  const seen = new Set<string>();
  const out = rows.map((row) => {
    seen.add(row.id);
    const prev = samples.get(row.id);
    let speedBps = prev?.speedBps ?? 0;
    if (row.status === "downloading" && prev && row.updatedAt > prev.at) {
      const dtSeconds = (row.updatedAt - prev.at) / 1000;
      const deltaBytes = row.received - prev.received;
      if (dtSeconds > 0 && deltaBytes >= 0) {
        const instant = deltaBytes / dtSeconds;
        speedBps = prev.speedBps > 0 ? prev.speedBps * 0.6 + instant * 0.4 : instant;
      }
    }
    if (row.status === "downloading" || row.status === "pending" || row.status === "retrying") {
      samples.set(row.id, { at: row.updatedAt, received: row.received, speedBps });
    } else {
      samples.delete(row.id);
    }
    const etaSeconds =
      row.status === "downloading" && speedBps > 0 && row.total > row.received
        ? (row.total - row.received) / speedBps
        : null;
    return { ...row, speedBps, etaSeconds };
  });
  for (const id of Array.from(samples.keys())) {
    if (!seen.has(id)) {
      samples.delete(id);
    }
  }
  return out;
}

export function formatBytes(bytes: number): string {
  if (!bytes) {
    return "0 B";
  }
  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  const rounded = value >= 10 || unit === 0 ? Math.round(value) : Math.round(value * 10) / 10;
  return `${rounded} ${units[unit]}`;
}

export function formatSpeed(bytesPerSecond: number): string {
  if (!bytesPerSecond || bytesPerSecond <= 0) {
    return "";
  }
  return `${formatBytes(bytesPerSecond)}/s`;
}

export function formatEta(seconds: number | null): string {
  if (seconds === null || !Number.isFinite(seconds) || seconds <= 0) {
    return "";
  }
  const total = Math.round(seconds);
  if (total < 60) {
    return `${total}s`;
  }
  const minutes = Math.floor(total / 60);
  const secs = total % 60;
  if (minutes < 60) {
    return `${minutes}m ${secs}s`;
  }
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  return `${hours}h ${mins}m`;
}
