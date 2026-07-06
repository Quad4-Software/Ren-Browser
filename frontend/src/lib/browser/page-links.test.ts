// SPDX-License-Identifier: MIT
import { describe, expect, it, vi } from "vitest";
import { handlePageLinkClick } from "./page-links";

vi.mock("./micron-links", () => ({
  resolveMicronNavigation: vi.fn(async () => ""),
  resolveNomadDataURL: vi.fn((_currentURL: string, dataUrl: string) => dataUrl),
  resolveLinkURL: vi.fn(() => ""),
}));

function clickEvent(target: HTMLElement): MouseEvent {
  const event = new MouseEvent("click", { bubbles: true, cancelable: true });
  Object.defineProperty(event, "target", { value: target });
  return event;
}

describe("handlePageLinkClick", () => {
  it("throws when an openNode link is missing a destination", async () => {
    const root = document.createElement("div");
    root.innerHTML = `<a data-action="openNode">broken</a>`;
    document.body.appendChild(root);
    const link = root.querySelector("a")!;

    await expect(
      handlePageLinkClick(clickEvent(link), root, "abcd:/page/index.mu", vi.fn()),
    ).rejects.toThrow(/missing a destination/);

    root.remove();
  });

  it("throws when the resolved navigation destination is empty", async () => {
    const root = document.createElement("div");
    root.innerHTML = `<a data-action="openNode" data-destination=":/file/broken.zip">broken</a>`;
    document.body.appendChild(root);
    const link = root.querySelector("a")!;

    await expect(
      handlePageLinkClick(clickEvent(link), root, "abcd:/page/index.mu", vi.fn()),
    ).rejects.toThrow(/Could not resolve link destination/);

    root.remove();
  });

  it("throws when a plain anchor href cannot be resolved", async () => {
    const root = document.createElement("div");
    root.innerHTML = `<a href="not-resolvable">broken</a>`;
    document.body.appendChild(root);
    const link = root.querySelector("a")!;

    await expect(
      handlePageLinkClick(clickEvent(link), root, "abcd:/page/index.mu", vi.fn()),
    ).rejects.toThrow(/Could not resolve link/);

    root.remove();
  });

  it("does nothing when the click target is not a link", async () => {
    const root = document.createElement("div");
    root.innerHTML = `<span>plain text</span>`;
    document.body.appendChild(root);
    const span = root.querySelector("span")!;
    const onNavigate = vi.fn();

    await expect(
      handlePageLinkClick(clickEvent(span), root, "abcd:/page/index.mu", onNavigate),
    ).resolves.toBeUndefined();
    expect(onNavigate).not.toHaveBeenCalled();

    root.remove();
  });
});
