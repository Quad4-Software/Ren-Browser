import { describe, expect, it } from "vitest";
import { resolveLinkURL } from "./micron-links";

describe("resolveLinkURL", () => {
  const current = "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu";

  it("appends query fields to the current path", () => {
    expect(resolveLinkURL(current, "?user=alice&action=go")).toBe(
      "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice&action=go",
    );
  });

  it("appends backtick fields to the current path", () => {
    expect(resolveLinkURL(current, "`user=alice|action=go")).toBe(
      "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu`user=alice|action=go",
    );
  });

  it("resolves relative file paths", () => {
    expect(resolveLinkURL(current, "/file/guide.zip")).toBe(
      "abb3ebcd03cb2388a838e70c001291f9:/file/guide.zip",
    );
  });

  it("preserves field suffixes on the current path", () => {
    const withFields = "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?user=alice";
    expect(resolveLinkURL(withFields, "?action=go")).toBe(
      "abb3ebcd03cb2388a838e70c001291f9:/page/form.mu?action=go",
    );
  });
});
