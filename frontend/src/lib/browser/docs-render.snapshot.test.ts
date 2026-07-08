// SPDX-License-Identifier: MIT
import { describe, expect, it } from "vitest";
import { renderDocsMarkdown } from "./docs-render";

describe("docs render snapshots", () => {
  it("matches markdown chrome snapshot", () => {
    const html = renderDocsMarkdown(
      "# Getting started\n\n- Install\n- Configure\n\n[Next](installation.md)",
      "en",
      "getting-started",
    );
    expect(html).toMatchSnapshot();
  });

  it("matches external link rewrite snapshot", () => {
    const html = renderDocsMarkdown(
      "See [Reticulum](https://reticulum.network/) and [local](faq.md).",
      "en",
      "",
    );
    expect(html).toMatchSnapshot();
  });
});
