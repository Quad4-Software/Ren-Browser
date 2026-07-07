// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { withProgress, type ActiveDownloadRow } from "./download-progress";

function row(overrides: Partial<ActiveDownloadRow>): ActiveDownloadRow {
  return {
    id: "dl-1",
    url: "deadbeef:/file/a.bin",
    name: "a.bin",
    received: 0,
    total: 0,
    status: "pending",
    startedAt: 1,
    updatedAt: 1,
    ...overrides,
  };
}

describe("download progress", () => {
  it("tracks retrying downloads like active transfers", () => {
    const views = withProgress([
      row({ status: "retrying", attempt: 2, maxAttempts: 4, updatedAt: 100 }),
    ]);
    expect(views[0]?.status).toBe("retrying");
    expect(views[0]?.attempt).toBe(2);
  });
});
