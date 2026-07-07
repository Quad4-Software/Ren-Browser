// SPDX-License-Identifier: MIT
import JSZip from "jszip";
import { describe, expect, it } from "vitest";
import { parseEpub } from "./epub";

async function buildNamespacedEpub(): Promise<Uint8Array> {
  const zip = new JSZip();
  zip.file(
    "META-INF/container.xml",
    `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`,
  );
  zip.file(
    "OEBPS/content.opf",
    `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata>
    <dc:title xmlns:dc="http://purl.org/dc/elements/1.1/">Test Book</dc:title>
  </metadata>
  <manifest>
    <item id="c1" href="chapter.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="c1"/>
  </spine>
</package>`,
  );
  zip.file(
    "OEBPS/chapter.xhtml",
    `<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head><title>Chapter One</title></head>
  <body><h1>Chapter One</h1><p>Hello epub.</p></body>
</html>`,
  );
  return await zip.generateAsync({ type: "uint8array" });
}

describe("parseEpub", () => {
  it("parses namespaced OPF packages", async () => {
    const data = await buildNamespacedEpub();
    const book = await parseEpub(data);
    expect(book.title).toBe("Test Book");
    expect(book.chapters).toHaveLength(1);
    expect(book.chapters[0]?.title).toBe("Chapter One");
    expect(book.chapters[0]?.html).toContain("Hello epub.");
  });
});
