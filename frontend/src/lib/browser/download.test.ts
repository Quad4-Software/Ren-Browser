import { describe, expect, it } from "vitest";
import { isFileURL, pageDownloadName } from "./download";

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
    expect(pageDownloadName("editor:", "editor")).toBe("editor.mu");
  });

  it("falls back to content-type defaults when the path has no leaf", () => {
    const base = "abb3ebcd03cb2388a838e70c001291f9:/";
    expect(pageDownloadName(base, "markdown")).toBe("page.md");
    expect(pageDownloadName(base, "html")).toBe("page.html");
    expect(pageDownloadName(base, "plaintext")).toBe("page.txt");
  });
});

describe("isFileURL", () => {
  it("detects mesh file paths", () => {
    expect(isFileURL("abb3ebcd03cb2388a838e70c001291f9:/file/guide.zip")).toBe(true);
  });

  it("rejects page and special urls", () => {
    expect(isFileURL("abb3ebcd03cb2388a838e70c001291f9:/page/index.mu")).toBe(false);
    expect(isFileURL("about:")).toBe(false);
    expect(isFileURL("editor:")).toBe(false);
    expect(isFileURL("")).toBe(false);
  });
});
