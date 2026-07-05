// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it } from "vitest";
import { clearFindHighlights, highlightFindMatches, scrollToFindMatch } from "./find-in-page";

describe("find-in-page", () => {
  let root: HTMLElement;

  afterEach(() => {
    root?.remove();
  });

  it("highlights matching text nodes", () => {
    root = document.createElement("div");
    root.innerHTML = "<p>Hello Ren Browser</p>";
    document.body.appendChild(root);

    const count = highlightFindMatches(root, "ren");
    expect(count).toBe(1);
    expect(root.querySelectorAll("mark.ren-find-hit")).toHaveLength(1);
  });

  it("clears previous highlights", () => {
    root = document.createElement("div");
    root.innerHTML = "<p>Ren Ren</p>";
    document.body.appendChild(root);

    highlightFindMatches(root, "ren");
    clearFindHighlights(root);
    expect(root.querySelectorAll("mark.ren-find-hit")).toHaveLength(0);
  });

  it("scrolls to the active match", () => {
    root = document.createElement("div");
    root.innerHTML = "<p>one two three</p>";
    document.body.appendChild(root);
    highlightFindMatches(root, "o");

    scrollToFindMatch(root, 1);
    const active = root.querySelectorAll("mark.ren-find-active");
    expect(active).toHaveLength(1);
  });
});
