// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import {
  downloadFailureMessage,
  isDownloadCanceledError,
  isFileURL,
  pageDownloadName,
  truncateDownloadLabel,
  canceledDownloadToast,
} from "./download";

describe("pageDownloadName", () => {
  it("uses the path leaf for mesh pages", () => {
    expect(
      pageDownloadName("abb3ebcd03cb2388a838e70c001291f9:/page/guide/index.mu", "micron"),
    ).toBe("index.mu");
  });

  it("strips query and backtick suffixes", () => {
    expect(
      pageDownloadName("abb3ebcd03cb2388a838e70c001291f9:/file/guide.zip?token=abc", "plaintext"),
    ).toBe("guide.zip");
  });

  it("labels special pages", () => {
    expect(pageDownloadName("about:", "html")).toBe("about.html");
    expect(pageDownloadName("license:", "html")).toBe("LICENSE");
    expect(pageDownloadName("editor:", "editor")).toBe("editor.mu");
    expect(pageDownloadName("config:", "config")).toBe("reticulum.conf");
  });

  it("falls back to content-type defaults when the path has no leaf", () => {
    const base = "abb3ebcd03cb2388a838e70c001291f9:/";
    expect(pageDownloadName(base, "markdown")).toBe("page.md");
    expect(pageDownloadName(base, "html")).toBe("page.html");
    expect(pageDownloadName(base, "plaintext")).toBe("page.txt");
  });
});

describe("isDownloadCanceledError", () => {
  it("detects plain context canceled errors", () => {
    expect(isDownloadCanceledError(new Error("context canceled"))).toBe(true);
    expect(isDownloadCanceledError("download canceled")).toBe(true);
  });

  it("detects wails runtime error payloads", () => {
    expect(
      isDownloadCanceledError(
        JSON.stringify({ message: "context canceled", cause: {}, kind: "RuntimeError" }),
      ),
    ).toBe(true);
  });

  it("ignores real download failures", () => {
    expect(isDownloadCanceledError(new Error("node response timed out"))).toBe(false);
  });
});

describe("downloadFailureMessage", () => {
  it("returns empty text for canceled downloads", () => {
    expect(downloadFailureMessage(new Error("context canceled"), "fallback")).toBe("");
  });

  it("unwraps wails runtime error messages", () => {
    expect(
      downloadFailureMessage(
        JSON.stringify({ message: "permission denied", kind: "RuntimeError" }),
        "fallback",
      ),
    ).toBe("permission denied");
  });
});

describe("truncateDownloadLabel", () => {
  it("truncates long filenames for toasts", () => {
    const longName = "a".repeat(60) + ".zip";
    const truncated = truncateDownloadLabel(longName);
    expect(truncated.length).toBeLessThanOrEqual(48);
    expect(truncated.endsWith("\u2026")).toBe(true);
  });

  it("keeps short filenames unchanged", () => {
    expect(truncateDownloadLabel("guide.zip")).toBe("guide.zip");
  });
});

describe("canceledDownloadToast", () => {
  const translate = (key: string, params?: Record<string, string>) => {
    if (key === "downloads.canceledNamed" && params?.name) {
      return `Canceled: ${params.name}`;
    }
    return "Canceled";
  };

  it("includes the truncated download name", () => {
    expect(canceledDownloadToast("guide.zip", translate)).toBe("Canceled: guide.zip");
  });

  it("falls back when the name is missing", () => {
    expect(canceledDownloadToast("", translate)).toBe("Canceled");
  });
});

describe("isFileURL", () => {
  it("detects mesh file paths", () => {
    expect(isFileURL("abb3ebcd03cb2388a838e70c001291f9:/file/guide.zip")).toBe(true);
  });

  it("rejects page and special urls", () => {
    expect(isFileURL("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu")).toBe(false);
    expect(isFileURL("about:")).toBe(false);
    expect(isFileURL("license:")).toBe(false);
    expect(isFileURL("editor:")).toBe(false);
    expect(isFileURL("config:")).toBe(false);
    expect(isFileURL("")).toBe(false);
  });
});
